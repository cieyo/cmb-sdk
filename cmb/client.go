package cmb

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	logger  Logger

	// 并发控制
	semaphore chan struct{}
	mu        sync.Mutex
}

var reqIDCounter uint32

// NewClient 创建招商银行客户端
func NewClient(config *Config) (*Client, error) {
	config.normalize()

	logger := config.Logger

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
		parser:    NewResponseParser(crypto, logger),
		http:      &http.Client{Timeout: config.Timeout},
		logger:    logger,
		semaphore: make(chan struct{}, config.MaxConcurrent),
	}

	logger.Info("cmb client initialized",
		"domain", config.Domain,
		"userid", config.UserID,
		"timeout", config.Timeout.String(),
		"max_concurrent", config.MaxConcurrent,
	)

	return client, nil
}

// doRequest 执行HTTP请求
// funcode: 接口功能码
// reqID: 请求唯一标识
// reqBody: 请求体
// respBody: 响应体（传入指针）
// 返回: 响应头和错误
func (c *Client) doRequest(funcode, reqID string, reqBody, respBody interface{}) (*ResponseHead, error) {
	start := time.Now()

	// 打印详细的请求体内容
	reqBodyJSON, _ := json.MarshalIndent(reqBody, "", "  ")
	c.logger.Infof("========== REQUEST START [%s] ==========", funcode)
	c.logger.Info("request info",
		"funcode", funcode,
		"reqid", reqID,
		"userid", c.config.UserID,
		"domain", c.config.Domain,
	)
	c.logger.Debugf("request body:\n%s", string(reqBodyJSON))

	// 并发控制
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	default:
		c.logger.Warn("concurrency limit exceeded", "funcode", funcode, "reqid", reqID)
		return nil, ErrConcurrencyLimit
	}

	// 1. 构建并加密请求
	encryptedData, err := c.request.BuildRequest(funcode, reqID, reqBody)
	if err != nil {
		c.logger.Errorf("[%s] build request failed: %v", funcode, err)
		return nil, fmt.Errorf("build request failed: %w", err)
	}

	// 2. 构造表单参数
	formData := url.Values{}
	formData.Set("UID", c.config.UserID)
	formData.Set("FUNCODE", funcode)
	formData.Set("ALG", "SM")
	formData.Set("DATA", encryptedData)

	c.logger.Debugf("http form params: UID=%s, FUNCODE=%s, ALG=SM, DATA_LENGTH=%d",
		c.config.UserID, funcode, len(encryptedData))

	// 3. 发送HTTP请求
	c.logger.Debugf("sending POST to %s", c.config.Domain)
	resp, err := c.http.PostForm(c.config.Domain, formData)
	if err != nil {
		c.logger.Errorf("[%s] http request failed: %v", funcode, err)
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	// 4. 读取响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("[%s] read response body failed: %v", funcode, err)
		return nil, fmt.Errorf("%w: read response body: %v", ErrRequestFailed, err)
	}

	c.logger.Infof("http response: status=%d, content_length=%d, content_type=%s",
		resp.StatusCode, len(respData), resp.Header.Get("Content-Type"))

	// 检查HTTP状态码，非200时银行通常返回纯文本错误
	if resp.StatusCode != 200 {
		c.logger.Errorf("[%s] http status error: %d, body: %s", funcode, resp.StatusCode, string(respData))
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrRequestFailed, resp.StatusCode, string(respData))
	}

	c.logger.Debugf("encrypted response data (first 200 chars): %.200s...", string(respData))

	// 5. 解析响应（响应数据本身就是加密的BASE64字符串）
	head, err := c.parser.ParseResponse(string(respData), respBody)
	if err != nil {
		c.logger.Errorf("[%s] parse response failed: %v", funcode, err)
		if head != nil {
			c.logger.Errorf("[%s] response head: result_code=%s, result_msg=%s",
				funcode, head.ResultCode, head.ResultMsg)
		}
		return head, fmt.Errorf("parse response failed: %w", err)
	}

	// 打印详细的响应体内容
	if respBody != nil {
		respBodyJSON, _ := json.MarshalIndent(respBody, "", "  ")
		c.logger.Debugf("response body (parsed):\n%s", string(respBodyJSON))
	}

	duration := time.Since(start)
	c.logger.Infof("========== REQUEST END [%s] (%s) ==========", funcode, duration)
	c.logger.Info("request completed",
		"funcode", funcode,
		"reqid", reqID,
		"result_code", head.ResultCode,
		"result_msg", head.ResultMsg,
		"rspid", head.RspID,
		"duration", duration.String(),
	)

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

// SetLogger 设置日志实例（运行时切换）
// 传入 nil 则使用静默日志
func (c *Client) SetLogger(logger Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if logger == nil {
		logger = NewNopLogger()
	}
	c.logger = logger
	c.request.logger = logger
	c.parser.logger = logger
	c.config.Logger = logger
}
