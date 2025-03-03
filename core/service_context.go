package core

import (
	"context"
	"sync"
)

// ServiceContext 提供服务级别的上下文管理和依赖注入
type ServiceContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
	deps   map[string]interface{}
}

// NewServiceContext 创建一个新的服务上下文
func NewServiceContext() *ServiceContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceContext{
		ctx:    ctx,
		cancel: cancel,
		deps:   make(map[string]interface{}),
	}
}

// Context 返回底层的context.Context
func (sc *ServiceContext) Context() context.Context {
	return sc.ctx
}

// Cancel 取消服务上下文
func (sc *ServiceContext) Cancel() {
	sc.cancel()
}

// Register 注册一个依赖
func (sc *ServiceContext) Register(name string, dependency interface{}) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.deps[name] = dependency
}

// Get 获取一个依赖
func (sc *ServiceContext) Get(name string) interface{} {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.deps[name]
}

// MustGet 获取一个依赖，如果不存在则panic
func (sc *ServiceContext) MustGet(name string) interface{} {
	if dep := sc.Get(name); dep != nil {
		return dep
	}
	panic("dependency not found: " + name)
}