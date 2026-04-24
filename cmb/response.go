package cmb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResponseParser 响应解析器
type ResponseParser struct {
	crypto *CryptoManager
	logger Logger
}

// NewResponseParser 创建响应解析器
func NewResponseParser(crypto *CryptoManager, logger Logger) *ResponseParser {
	return &ResponseParser{
		crypto: crypto,
		logger: logger,
	}
}

// ParseResponse 解析响应报文
// encryptedData: BASE64编码的加密响应数据
// result: 用于接收解析后的响应体（传入指针）
// 返回: 响应头和错误
func (rp *ResponseParser) ParseResponse(encryptedData string, result interface{}) (*ResponseHead, error) {
	rp.logger.Debug("response decrypting", "encrypted_length", len(encryptedData))

	// 1. SM4解密
	decryptedData, err := rp.crypto.Decrypt(encryptedData)
	if err != nil {
		rp.logger.Error("decrypt response failed", "error", err)
		return nil, fmt.Errorf("decrypt response failed: %w", err)
	}

	rp.logger.Debugf("response decrypted, raw JSON:\n%s", string(decryptedData))

	// 2. 解析JSON
	var resp Response
	err = json.Unmarshal(decryptedData, &resp)
	if err != nil {
		rp.logger.Errorf("unmarshal response JSON failed: %v", err)
		return nil, fmt.Errorf("%w: unmarshal response: %v", ErrInvalidResponse, err)
	}

	// 3. 验签
	err = rp.verifySignature(string(decryptedData), resp.Signature.SigDat)
	if err != nil {
		rp.logger.Error("verify response signature failed", "error", err)
		return nil, fmt.Errorf("verify signature failed: %w", err)
	}

	rp.logger.Debug("response signature verified")

	// 4. 检查业务结果码
	head := resp.Response.Head
	rp.logger.Info("response head parsed",
		"funcode", head.FunCode,
		"result_code", head.ResultCode,
		"result_msg", head.ResultMsg,
		"reqid", head.ReqID,
		"rspid", head.RspID,
	)

	if !ResultCode(head.ResultCode).IsSuccess() {
		rp.logger.Errorf("business error: code=%s, msg=%s", head.ResultCode, head.ResultMsg)
		return &head, fmt.Errorf("business error: code=%s, msg=%s", head.ResultCode, head.ResultMsg)
	}

	// 5. 解析业务响应体
	if result != nil {
		bodyBytes, err := json.Marshal(resp.Response.Body)
		if err != nil {
			return &head, fmt.Errorf("%w: marshal body: %v", ErrInvalidResponse, err)
		}

		// 响应体详细日志
		prettyBody, _ := json.MarshalIndent(resp.Response.Body, "", "  ")
		rp.logger.Debugf("response body:\n%s", string(prettyBody))

		err = json.Unmarshal(bodyBytes, result)
		if err != nil {
			rp.logger.Errorf("unmarshal response body failed: %v", err)
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
