package middleware

import (
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	trans ut.Translator // 全局翻译器实例
)

// ValidationError 自定义验证错误类型
type ValidationError struct {
	Message string      `json:"message"`
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
}

// InitValidator 初始化验证器和翻译器
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册字段名处理函数，优先使用label标签，其次使用json标签
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// 优先使用label标签作为字段名
			if label := fld.Tag.Get("label"); label != "" {
				return label
			}
			// 其次使用json标签
			if jsonTag := fld.Tag.Get("json"); jsonTag != "" {
				name := strings.SplitN(jsonTag, ",", 2)[0]
				if name != "-" {
					return name
				}
			}
			// 最后使用字段名
			return fld.Name
		})

		// 初始化中文翻译器
		zh := zh.New()
		uni := ut.New(zh, zh)
		trans, _ = uni.GetTranslator("zh")

		// 注册默认翻译
		_ = zh_translations.RegisterDefaultTranslations(v, trans)

		// 注册手机号验证
		_ = v.RegisterValidation("phone", validatePhone)
		// 可以注册其他自定义验证器
		_ = v.RegisterValidation("password", validatePassword)

	}
}

// validatePhone 手机号验证
func validatePhone(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// validatePassword 密码强度验证
func validatePassword(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 密码长度检查
	if len(password) < 6 || len(password) > 100 {
		return false
	}

	// 检查是否包含数字和字母
	var hasLetter, hasNumber bool
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		}
	}

	return hasLetter && hasNumber
}

// ShouldBindWithValidation 带自动验证的绑定方法 - 绑定 JSON
func ShouldBindWithValidation(c *gin.Context, obj interface{}) error {
	// 明确使用 JSON 绑定
	if err := c.ShouldBindJSON(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// 获取第一个验证错误的中文信息
			if len(validationErrors) > 0 {
				firstError := validationErrors[0]
				chineseMsg := firstError.Translate(trans)

				// 创建一个包含中文错误信息的自定义错误
				return &ValidationError{
					Message: chineseMsg,
					Field:   firstError.Field(),
					Tag:     firstError.Tag(),
					Value:   firstError.Value(),
				}
			}
		}
		return err
	}
	return nil
}

// ShouldBindWithValidationQuery 绑定查询参数
func ShouldBindWithValidationQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			if len(validationErrors) > 0 {
				firstError := validationErrors[0]
				chineseMsg := firstError.Translate(trans)
				return &ValidationError{
					Message: chineseMsg,
					Field:   firstError.Field(),
					Tag:     firstError.Tag(),
					Value:   firstError.Value(),
				}
			}
		}
		return err
	}
	return nil
}

// ShouldBindWithValidationUri 绑定 URI 参数
func ShouldBindWithValidationUri(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			if len(validationErrors) > 0 {
				firstError := validationErrors[0]
				chineseMsg := firstError.Translate(trans)
				return &ValidationError{
					Message: chineseMsg,
					Field:   firstError.Field(),
					Tag:     firstError.Tag(),
					Value:   firstError.Value(),
				}
			}
		}
		return err
	}
	return nil
}

// ShouldBindWithValidationForm 绑定表单数据
func ShouldBindWithValidationForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			if len(validationErrors) > 0 {
				firstError := validationErrors[0]
				chineseMsg := firstError.Translate(trans)
				return &ValidationError{
					Message: chineseMsg,
					Field:   firstError.Field(),
					Tag:     firstError.Tag(),
					Value:   firstError.Value(),
				}
			}
		}
		return err
	}
	return nil
}

func (e *ValidationError) Error() string {
	return e.Message
}
