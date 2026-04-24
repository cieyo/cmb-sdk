package cmb

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	// SignaturePlaceholder 签名占位符
	SignaturePlaceholder = "__signature_sigdat__"
)

// RequestBuilder 请求构建器
type RequestBuilder struct {
	crypto *CryptoManager
	config *Config
	logger Logger
}

// NewRequestBuilder 创建请求构建器
func NewRequestBuilder(crypto *CryptoManager, config *Config) *RequestBuilder {
	return &RequestBuilder{
		crypto: crypto,
		config: config,
		logger: config.Logger,
	}
}

// BuildRequest 构建请求报文
// funcode: 接口功能码
// reqID: 请求唯一标识
// body: 业务参数
// 返回: 加密后的BASE64字符串和错误
func (rb *RequestBuilder) BuildRequest(funcode, reqID string, body interface{}) (string, error) {
	// 1. 构建原始请求报文
	req := Request{}
	req.Request.Head = RequestHead{
		FunCode: funcode,
		UserID:  rb.config.UserID,
		ReqID:   reqID,
	}
	req.Request.Body = body

	// 2. 添加signature占位符
	sigtim := time.Now().Format("20060102150405")
	req.Signature = Signature{
		SigTim: sigtim,
		SigDat: SignaturePlaceholder,
	}

	// 3. 生成待签名字符串（排序+去空格换行）
	signStr, err := rb.generateSignString(req)
	if err != nil {
		return "", fmt.Errorf("generate sign string failed: %w", err)
	}

	rb.logger.Debugf("sign string generated (length=%d):\n%s", len(signStr), signStr)

	// 4. SM2签名
	sigdat, err := rb.crypto.Sign([]byte(signStr))
	if err != nil {
		return "", fmt.Errorf("sign failed: %w", err)
	}

	// 5. 替换signature.sigdat
	req.Signature.SigDat = sigdat

	// 6. SM4加密
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("%w: marshal request: %v", ErrInvalidJSON, err)
	}

	rb.logger.Debugf("request JSON (before encrypt, length=%d):\n%s", len(reqJSON), string(reqJSON))

	encryptedData, err := rb.crypto.Encrypt(reqJSON)
	if err != nil {
		return "", fmt.Errorf("encrypt failed: %w", err)
	}

	return encryptedData, nil
}

// generateSignString 生成待签名字符串
// 1. 所有KEY按ASCII码升序排序
// 2. 移除报文中所有空格和换行符
func (rb *RequestBuilder) generateSignString(req Request) (string, error) {
	// 排序并序列化
	sortedJSON, err := sortedJSONString(req)
	if err != nil {
		return "", err
	}

	// 移除空格和换行符
	signStr := strings.ReplaceAll(sortedJSON, " ", "")
	signStr = strings.ReplaceAll(signStr, "\n", "")
	signStr = strings.ReplaceAll(signStr, "\r", "")
	signStr = strings.ReplaceAll(signStr, "\t", "")

	return signStr, nil
}

// sortedJSONString 将对象转换为KEY排序后的JSON字符串
func sortedJSONString(v interface{}) (string, error) {
	// 先转成map
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var data interface{}
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return "", err
	}

	// 递归排序
	sorted := sortMapRecursive(data)

	// 转回JSON（不转义HTML字符，不缩进）
	sortedBytes, err := json.Marshal(sorted)
	if err != nil {
		return "", err
	}

	return string(sortedBytes), nil
}

// sortMapRecursive 递归排序map的key
func sortMapRecursive(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// 创建有序的map
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sorted := make(map[string]interface{}, len(val))
		for _, k := range keys {
			sorted[k] = sortMapRecursive(val[k])
		}
		return sorted

	case []interface{}:
		// 数组元素递归排序
		for i, item := range val {
			val[i] = sortMapRecursive(item)
		}
		return val

	default:
		return v
	}
}
