package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ErrMatchFunc 判断 err 是否匹配某条规则（例如 errors.Is）。
type ErrMatchFunc func(err, target error) bool

// Mapper 将业务 error 映射为 *ErrorCode，供 HTTP 响应使用。规则按注册顺序匹配，命中即停止。
type Mapper struct {
	rules []errMapRule
	def   *ErrorCode
}

type errMapRule struct {
	match  ErrMatchFunc
	target error
	code   *ErrorCode
}

// NewMapper 创建空错误映射表，便于链式注册规则。
func NewMapper() *Mapper {
	return &Mapper{}
}

// On 注册一条规则：当 match(err, target) 为 true 时使用 code。
func (m *Mapper) On(match ErrMatchFunc, target error, code *ErrorCode) *Mapper {
	m.rules = append(m.rules, errMapRule{match: match, target: target, code: code})
	return m
}

// OnIs 等价于 On(errors.Is, target, code)。
func (m *Mapper) OnIs(target error, code *ErrorCode) *Mapper {
	return m.On(errors.Is, target, code)
}

// Default 未命中任何规则时使用的错误码；若不设置，Lookup/Write 回落到 ErrInternalServer。
func (m *Mapper) Default(code *ErrorCode) *Mapper {
	m.def = code
	return m
}

// Lookup 返回 err 对应的业务错误码；err 为 nil 时返回 nil。
func (m *Mapper) Lookup(err error) *ErrorCode {
	if err == nil {
		return nil
	}
	if m != nil {
		for i := range m.rules {
			r := &m.rules[i]
			if r.match(err, r.target) {
				return r.code
			}
		}
		if m.def != nil {
			return m.def
		}
	}
	return ErrInternalServer
}

// Write 将 err 映射为 JSON 错误响应；err 为 nil 时不写入响应。
func (m *Mapper) Write(c *gin.Context, err error) {
	if err == nil || c == nil {
		return
	}
	Error(c, m.Lookup(err))
}
