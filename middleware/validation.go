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
	"github.com/shrimps80/go-service-utils/utils"
)

var (
	trans ut.Translator // 全局翻译器实例
)

// InitValidator 初始化验证器和翻译器
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册JSON字段名处理
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// 初始化中文翻译器
		zh := zh.New()
		uni := ut.New(zh, zh)
		trans, _ = uni.GetTranslator("zh")

		// 注册默认翻译
		_ = zh_translations.RegisterDefaultTranslations(v, trans)

		// 注册自定义验证规则
		registerCustomValidations(v)
	}
}

// ValidationMiddleware 验证中间件
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅处理需要验证的请求
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// 在实际业务处理中自动验证，此处只设置错误处理
			c.Next()

			// 检查是否存在验证错误
			if len(c.Errors) > 0 {
				var validationErrors validator.ValidationErrors
				for _, ginErr := range c.Errors {
					if errs, ok := ginErr.Err.(validator.ValidationErrors); ok {
						validationErrors = errs
						break
					}
				}

				// 转换验证错误信息
				if validationErrors != nil {
					errMessages := make(map[string]string)
					for _, err := range validationErrors {
						errMessages[err.Field()] = err.Translate(trans)
					}

					// 使用统一的错误响应格式
					utils.Error(c, utils.ErrInvalidParam, utils.WithMessage("请求参数验证失败"))
					return
				}
			}
		}
		c.Next()
	}
}

// registerCustomValidations 注册自定义验证规则
func registerCustomValidations(v *validator.Validate) {
	// 示例：手机号验证
	v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return len(value) == 11 && strings.HasPrefix(value, "1")
	})

	// 添加对应的翻译
	_ = v.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return ut.Add("mobile", "{0}格式不正确", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})
}

// ShouldBindWithValidation 带自动验证的绑定方法
func ShouldBindWithValidation(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errMessages := make(map[string]string)
			for _, err := range validationErrors {
				errMessages[err.Field()] = err.Translate(trans)
			}
			return &gin.Error{
				Err:  validationErrors,
				Type: gin.ErrorTypeBind,
				Meta: errMessages,
			}
		}
		return err
	}
	return nil
}
