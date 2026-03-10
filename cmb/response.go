package cmb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResponseParser 响应解析器
type ResponseParser struct {
	crypto *CryptoManager
}

// NewResponseParser 创建响应解析器
func NewResponseParser(crypto *CryptoManager) *ResponseParser {
	return &ResponseParser{
		crypto: crypto,
	}
}

// ParseResponse 解析响应报文
// encryptedData: BASE64编码的加密响应数据
// result: 用于接收解析后的响应体（传入指针）
// 返回: 响应头和错误
func (rp *ResponseParser) ParseResponse(encryptedData string, result interface{}) (*ResponseHead, error) {
	// 1. SM4解密
	decryptedData, err := rp.crypto.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt response failed: %w", err)
	}

	// 2. 解析JSON
	var resp Response
	err = json.Unmarshal(decryptedData, &resp)
	if err != nil {
		return nil, fmt.Errorf("%w: unmarshal response: %v", ErrInvalidResponse, err)
	}

	// 3. 验签
	err = rp.verifySignature(string(decryptedData), resp.Signature.SigDat)
	if err != nil {
		return nil, fmt.Errorf("verify signature failed: %w", err)
	}

	// 4. 检查业务结果码
	head := resp.Response.Head
	if !ResultCode(head.ResultCode).IsSuccess() {
		return &head, fmt.Errorf("business error: code=%s, msg=%s", head.ResultCode, head.ResultMsg)
	}

	// 5. 解析业务响应体
	if result != nil {
		bodyBytes, err := json.Marshal(resp.Response.Body)
		if err != nil {
			return &head, fmt.Errorf("%w: marshal body: %v", ErrInvalidResponse, err)
		}

		err = json.Unmarshal(bodyBytes, result)
		if err != nil {
			return &head, fmt.Errorf("%w: unmarshal body: %v", ErrInvalidResponse, err)
		}
	}

	return &head, nil
}

// verifySignature 验证响应签名
func (rp *ResponseParser) verifySignature(rawResponse, sigdat string) error {
	if sigdat == "" {
		return fmt.Errorf("signature not found in response")
	}

	// 按官方示例：直接在原始响应串中替换 sigdat，再验签
	verifyStr := strings.Replace(rawResponse, sigdat, SignaturePlaceholder, 1)

	// SM2验签
	err := rp.crypto.Verify([]byte(verifyStr), sigdat)
	if err != nil {
		return err
	}

	return nil
}
