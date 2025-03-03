package config

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Options 配置选项
type Options struct {
	ConfigType  string   // 配置文件类型（yaml, json等）
	ConfigName  string   // 配置文件名（不含扩展名）
	ConfigPaths []string // 配置文件搜索路径
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() *Options {
	return &Options{
		ConfigType:  "yaml",
		ConfigName:  "config",
		ConfigPaths: []string{".", "./config", "/etc/app"},
	}
}

// Config 配置管理器
type Config struct {
	v          *viper.Viper
	configFile string
	configType string
	isWatching bool                 // 是否正在监听配置文件变更
	callback   func(fsnotify.Event) // 用户注册的配置变更回调函数
}

// New 创建一个新的配置管理器
func New(opts *Options) (*Config, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	v := viper.New()
	v.SetConfigType(opts.ConfigType)
	v.SetConfigName(opts.ConfigName)

	// 添加配置文件搜索路径
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// 获取当前使用的配置文件
	configFile := v.ConfigFileUsed()

	// 创建配置管理器实例
	config := &Config{
		v:          v,
		configFile: configFile,
		configType: opts.ConfigType,
		isWatching: false,
	}

	return config, nil
}

// StopWatch 停止配置文件监听
func (c *Config) StopWatch() {
	if c.isWatching {
		// 创建新的viper实例以停止当前实例的监听
		newV := viper.New()
		newV.SetConfigType(c.configType)
		// 复制所有配置到新实例
		for k, v := range c.v.AllSettings() {
			newV.Set(k, v)
		}
		c.v = newV
		c.v.OnConfigChange(nil) // 清除回调
		c.isWatching = false
	}
}

// 修改后的 StartWatch 方法
func (c *Config) StartWatch() {
	if !c.isWatching {
		// 确保每次启动都重新绑定回调
		c.v.OnConfigChange(func(in fsnotify.Event) {
			// 添加事件类型过滤
			if in.Op&fsnotify.Write == fsnotify.Write || in.Op&fsnotify.Create == fsnotify.Create {
				// 延迟读取避免编辑器原子保存造成的中间状态
				time.Sleep(100 * time.Millisecond)
				if err := c.v.ReadInConfig(); err == nil {
					c.configFile = c.v.ConfigFileUsed()
					if c.callback != nil {
						c.callback(in)
					}
				}
			}
		})
		c.v.WatchConfig()
		c.isWatching = true
	}
}

// OnConfigChange 注册配置变更回调函数
func (c *Config) OnConfigChange(callback func(fsnotify.Event)) {
	c.callback = callback
}

// GetConfigFile 获取当前使用的配置文件路径
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// GetAll 获取所有配置
func (c *Config) GetAll() map[string]interface{} {
	return c.v.AllSettings()
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// WriteConfig 将当前配置写入文件
func (c *Config) WriteConfig() error {
	return c.v.WriteConfig()
}

// SafeWriteConfig 安全地将当前配置写入文件（不覆盖已存在的文件）
func (c *Config) SafeWriteConfig() error {
	return c.v.SafeWriteConfig()
}

// GetString 获取字符串类型的配置值
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt 获取整数类型的配置值
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool 获取布尔类型的配置值
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// Get 获取原始配置值
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// UnmarshalKey 将指定key的配置解析到结构体
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	return c.v.UnmarshalKey(key, rawVal)
}
