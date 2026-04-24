package cmb

import (
	"encoding/hex"
	"time"
)

// Config 招商银行新直联配置
type Config struct {
	// Domain API域名
	// 测试环境：http://cdctest.cmburl.cn/cdcserver/api/v2
	// 生产环境：https://cdc.cmbchina.com/cdcserver/api/v2
	Domain string `yaml:"domain"`

	// UserID 用户ID（从招商银行获取）
	UserID string `yaml:"userid"`

	// SM4Key SM4对称密钥（16位字符串，用于加密解密）
	SM4Key string `yaml:"sm4_key"`

	// SM4NotifyKey 主动通知SM4对称密钥（为空时回退 SM4Key）
	SM4NotifyKey string `yaml:"sm4_notify_key"`

	// SM2PrivateKey SM2私钥（PEM格式，用于签名）
	SM2PrivateKey string `yaml:"sm2_private_key"`

	// SM2BankPublicKey 招商银行SM2公钥（PEM格式，用于验签）
	SM2BankPublicKey string `yaml:"sm2_bank_public_key"`

	// SM2NotifyPublicKey 主动通知验签公钥（为空时回退 SM2BankPublicKey）
	SM2NotifyPublicKey string `yaml:"sm2_notify_public_key"`

	// Timeout 请求超时时间
	Timeout time.Duration `yaml:"timeout"`

	// MaxConcurrent 最大并发数（招商银行限制单用户同时请求数不超过5个）
	MaxConcurrent int `yaml:"max_concurrent"`

	// Debug 是否开启调试日志（向后兼容，设为 true 时如果 Logger 为 nil 则自动创建 Debug 级别日志）
	Debug bool `yaml:"debug"`

	// Logger 日志实例，外部可传入自定义实现
	// 如果不设置，Debug=true 时使用默认 Debug 级别日志，否则使用静默日志
	Logger Logger `yaml:"-" json:"-"`

	// AccountNbr 招行收款账号（用于交易查询/回单下载）
	AccountNbr string `yaml:"account_nbr"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Domain:        "http://cdctest.cmburl.cn/cdcserver/api/v2",
		Timeout:       30 * time.Second,
		MaxConcurrent: 5,
		Debug:         false,
		Logger:        nil, // normalize() 时自动设置
	}
}

// normalize 标准化配置，兼容常见配置写法
// 1. timeout<=0 时回退默认值 30s
// 2. timeout 很小（如 YAML 中写 30 被反序列化为 30ns）时按秒解释
// 3. max_concurrent<=0 时回退默认值 5
func (c *Config) normalize() {
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	} else if c.Timeout < 10*time.Millisecond {
		c.Timeout = c.Timeout * time.Second
	}

	if c.MaxConcurrent <= 0 {
		c.MaxConcurrent = 5
	}
	if c.SM4NotifyKey == "" {
		c.SM4NotifyKey = c.SM4Key
	}
	if c.SM2NotifyPublicKey == "" {
		c.SM2NotifyPublicKey = c.SM2BankPublicKey
	}

	// 初始化日志
	if c.Logger == nil {
		if c.Debug {
			c.Logger = NewDefaultLogger(LogLevelDebug, nil)
		} else {
			c.Logger = NewNopLogger()
		}
	}
}

// Validate 验证配置是否完整
func (c *Config) Validate() error {
	if c.Domain == "" {
		return ErrInvalidConfig
	}
	if c.UserID == "" {
		return ErrInvalidConfig
	}
	if !isValidSM4Key(c.SM4Key) {
		return ErrInvalidSM4Key
	}
	if c.SM4NotifyKey != "" && !isValidSM4Key(c.SM4NotifyKey) {
		return ErrInvalidSM4Key
	}
	if c.SM2PrivateKey == "" {
		return ErrInvalidSM2PrivateKey
	}
	if c.SM2BankPublicKey == "" {
		return ErrInvalidSM2PublicKey
	}
	return nil
}

func isValidSM4Key(k string) bool {
	if len(k) == 16 {
		return true
	}
	if len(k) != 32 {
		return false
	}
	_, err := hex.DecodeString(k)
	return err == nil
}
