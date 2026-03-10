package cmb

import "errors"

var (
	// ErrInvalidConfig 配置错误
	ErrInvalidConfig = errors.New("invalid config")

	// ErrInvalidSM4Key SM4密钥无效（必须是16位）
	ErrInvalidSM4Key = errors.New("invalid SM4 key: must be 16 characters")

	// ErrInvalidSM2PrivateKey SM2私钥无效
	ErrInvalidSM2PrivateKey = errors.New("invalid SM2 private key")

	// ErrInvalidSM2PublicKey SM2公钥无效
	ErrInvalidSM2PublicKey = errors.New("invalid SM2 public key")

	// ErrSignatureFailed 签名失败
	ErrSignatureFailed = errors.New("signature failed")

	// ErrEncryptFailed 加密失败
	ErrEncryptFailed = errors.New("encrypt failed")

	// ErrDecryptFailed 解密失败
	ErrDecryptFailed = errors.New("decrypt failed")

	// ErrVerifySignatureFailed 验签失败
	ErrVerifySignatureFailed = errors.New("verify signature failed")

	// ErrInvalidResponse 响应报文格式错误
	ErrInvalidResponse = errors.New("invalid response")

	// ErrRequestFailed 请求失败
	ErrRequestFailed = errors.New("request failed")

	// ErrInvalidJSON JSON格式错误
	ErrInvalidJSON = errors.New("invalid JSON")

	// ErrConcurrencyLimit 并发限制
	ErrConcurrencyLimit = errors.New("concurrency limit exceeded")

	// ErrPaginationLoop 续传查询疑似死循环
	ErrPaginationLoop = errors.New("pagination loop detected")
)

type ErrUnsupportedNotificationType struct {
	NotTyp string
}

func (e *ErrUnsupportedNotificationType) Error() string {
	if e == nil || e.NotTyp == "" {
		return "unsupported notification type"
	}
	return "unsupported notification type: " + e.NotTyp
}

// ResultCode 招商银行返回码
type ResultCode string

const (
	// ResultCodeSuccess 成功
	ResultCodeSuccess ResultCode = "SUC0000"

	// ResultCodeAccountNotExist 账户不存在
	ResultCodeAccountNotExist ResultCode = "FAAQ086"

	// ResultCodeSignatureFailed 签名验证失败
	ResultCodeSignatureFailed ResultCode = "FAAQ999"

	// ResultCodeDecryptFailed 报文解密失败
	ResultCodeDecryptFailed ResultCode = "FAAQ888"

	// ResultCodeConcurrencyLimit 请求超时/并发超限
	ResultCodeConcurrencyLimit ResultCode = "FAAQ777"
)

// IsSuccess 判断结果码是否成功
func (r ResultCode) IsSuccess() bool {
	return r == ResultCodeSuccess || r == "000000"
}

// String 返回结果码字符串
func (r ResultCode) String() string {
	return string(r)
}
