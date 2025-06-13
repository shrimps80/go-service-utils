package middleware

import (
	"reflect"
	"strings"

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

		// 注册自定义验证规则和翻译
		registerCustomValidations(v)
		registerCustomTranslations(v)
	}
}

// registerCustomValidations 注册自定义验证规则
func registerCustomValidations(v *validator.Validate) {
	// 示例：手机号验证
	v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return len(value) == 11 && strings.HasPrefix(value, "1")
	})
}

// registerCustomTranslations 注册自定义翻译
func registerCustomTranslations(v *validator.Validate) {
	// 只注册自定义验证规则的翻译，不重复注册默认规则

	// 注册手机号验证的翻译
	_ = v.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return ut.Add("mobile", "{0}格式不正确", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})

	// 可以在这里添加其他自定义验证规则的翻译
}

// ShouldBindWithValidation 带自动验证的绑定方法
func ShouldBindWithValidation(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
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

func (e *ValidationError) Error() string {
	return e.Message
}
