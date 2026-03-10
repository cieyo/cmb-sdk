package cmb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Client 招商银行新直联客户端
type Client struct {
	config  *Config
	crypto  *CryptoManager
	request *RequestBuilder
	parser  *ResponseParser
	http    *http.Client

	// 并发控制
	semaphore chan struct{}
	mu        sync.Mutex
}

var reqIDCounter uint32

// NewClient 创建招商银行客户端
func NewClient(config *Config) (*Client, error) {
	config.normalize()

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建加密管理器
	crypto, err := NewCryptoManager(config)
	if err != nil {
		return nil, fmt.Errorf("create crypto manager failed: %w", err)
	}

	client := &Client{
		config:    config,
		crypto:    crypto,
		request:   NewRequestBuilder(crypto, config),
		parser:    NewResponseParser(crypto),
		http:      &http.Client{Timeout: config.Timeout},
		semaphore: make(chan struct{}, config.MaxConcurrent),
	}

	return client, nil
}

// doRequest 执行HTTP请求
// funcode: 接口功能码
// reqID: 请求唯一标识
// reqBody: 请求体
// respBody: 响应体（传入指针）
// 返回: 响应头和错误
func (c *Client) doRequest(funcode, reqID string, reqBody, respBody interface{}) (*ResponseHead, error) {
	// 并发控制
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	default:
		return nil, ErrConcurrencyLimit
	}

	// 1. 构建并加密请求
	encryptedData, err := c.request.BuildRequest(funcode, reqID, reqBody)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}

	// 2. 构造表单参数
	formData := url.Values{}
	formData.Set("UID", c.config.UserID)
	formData.Set("FUNCODE", funcode)
	formData.Set("ALG", "SM")
	formData.Set("DATA", encryptedData)

	// 调试日志
	if c.config.Debug {
		fmt.Printf("[CMB DEBUG] Request: funcode=%s, reqid=%s\n", funcode, reqID)
		fmt.Printf("[CMB DEBUG] Form params: UID=%s, FUNCODE=%s, ALG=SM\n", c.config.UserID, funcode)
	}

	// 3. 发送HTTP请求
	resp, err := c.http.PostForm(c.config.Domain, formData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	// 4. 读取响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read response body: %v", ErrRequestFailed, err)
	}

	// 调试日志
	if c.config.Debug {
		fmt.Printf("[CMB DEBUG] Response status: %d\n", resp.StatusCode)
		fmt.Printf("[CMB DEBUG] Response length: %d bytes\n", len(respData))
		if resp.StatusCode != 200 || len(respData) < 500 {
			fmt.Printf("[CMB DEBUG] Response body: %s\n", string(respData))
		}
	}

	// 检查HTTP状态码，非200时银行通常返回纯文本错误
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrRequestFailed, resp.StatusCode, string(respData))
	}

	// 5. 解析响应（响应数据本身就是加密的BASE64字符串）
	head, err := c.parser.ParseResponse(string(respData), respBody)
	if err != nil {
		return head, fmt.Errorf("parse response failed: %w", err)
	}

	return head, nil
}

// GenerateReqID 生成请求唯一标识
// 格式：YYYYMMDDHHmmssSSS + 随机字符串
func GenerateReqID(prefix string) string {
	now := time.Now()
	timestamp := now.Format("20060102150405") + fmt.Sprintf("%03d", now.Nanosecond()/1e6)

	if prefix == "" {
		prefix = "REQ"
	}

	counter := atomic.AddUint32(&reqIDCounter, 1)
	buf := make([]byte, 2)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s%s%04X", timestamp, prefix, counter&0xFFFF)
	}

	return fmt.Sprintf("%s%s%s%04X", timestamp, prefix, hex.EncodeToString(buf), counter&0xFFFF)
}
