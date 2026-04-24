package cmb

import (
	"testing"
	"time"
)

// TestConfig 测试配置验证
func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &Config{
				Domain:           "http://cdctest.cmburl.cn/cdcserver/api/v2",
				UserID:           "N002432758",
				SM4Key:           "1234567890123456",
				SM2PrivateKey:    "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
				SM2BankPublicKey: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
				Timeout:          30 * time.Second,
				MaxConcurrent:    5,
			},
			wantErr: false,
		},
		{
			name: "SM4密钥长度错误",
			config: &Config{
				Domain:           "http://cdctest.cmburl.cn/cdcserver/api/v2",
				UserID:           "N002432758",
				SM4Key:           "123456",
				SM2PrivateKey:    "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
				SM2BankPublicKey: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			},
			wantErr: true,
		},
		{
			name: "缺少UserID",
			config: &Config{
				Domain:           "http://cdctest.cmburl.cn/cdcserver/api/v2",
				UserID:           "",
				SM4Key:           "1234567890123456",
				SM2PrivateKey:    "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
				SM2BankPublicKey: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestResultCode 测试结果码
func TestResultCode(t *testing.T) {
	tests := []struct {
		code    ResultCode
		success bool
	}{
		{ResultCodeSuccess, true},
		{"000000", true},
		{ResultCodeAccountNotExist, false},
		{ResultCodeSignatureFailed, false},
		{ResultCodeDecryptFailed, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.IsSuccess(); got != tt.success {
				t.Errorf("ResultCode.IsSuccess() = %v, want %v", got, tt.success)
			}
		})
	}
}

// TestGenerateReqID 测试请求ID生成
func TestGenerateReqID(t *testing.T) {
	id1 := GenerateReqID("TEST")
	id2 := GenerateReqID("TEST")

	if id1 == id2 {
		t.Error("GenerateReqID() should generate unique IDs")
	}

	if len(id1) < 14 {
		t.Errorf("GenerateReqID() length = %d, want >= 14", len(id1))
	}
}

func TestConfigNormalize(t *testing.T) {
	cfg := &Config{
		Domain:           "http://cdctest.cmburl.cn/cdcserver/api/v2",
		UserID:           "N002432758",
		SM4Key:           "1234567890123456",
		SM2PrivateKey:    "k",
		SM2BankPublicKey: "k",
		Timeout:          30, // YAML 中常见写法，反序列化到 time.Duration 会变成 30ns
		MaxConcurrent:    0,
	}

	cfg.normalize()

	if cfg.Timeout != 30*time.Second {
		t.Fatalf("normalize timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}

	if cfg.MaxConcurrent != 5 {
		t.Fatalf("normalize max_concurrent = %d, want 5", cfg.MaxConcurrent)
	}
}

// TestSortedJSONString 测试JSON排序
func TestSortedJSONString(t *testing.T) {
	input := map[string]interface{}{
		"z": "last",
		"a": "first",
		"m": map[string]interface{}{
			"y": "y",
			"x": "x",
		},
	}

	result, err := sortedJSONString(input)
	if err != nil {
		t.Fatalf("sortedJSONString() error = %v", err)
	}

	// 验证key的顺序
	expected := `{"a":"first","m":{"x":"x","y":"y"},"z":"last"}`
	if result != expected {
		t.Errorf("sortedJSONString() = %v, want %v", result, expected)
	}
}

// TestCryptoManagerIV 测试IV生成
func TestCryptoManagerIV(t *testing.T) {
	tests := []struct {
		userID   string
		wantLen  int
		wantData []byte
	}{
		{
			userID:   "N002432758",
			wantLen:  16,
			wantData: []byte("N002432758000000"),
		},
		{
			userID:   "1234567890123456",
			wantLen:  16,
			wantData: []byte("1234567890123456"),
		},
		{
			userID:   "ABC",
			wantLen:  16,
			wantData: []byte("ABC0000000000000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.userID, func(t *testing.T) {
			cm := &CryptoManager{userID: tt.userID}
			iv := cm.getIV()

			if len(iv) != tt.wantLen {
				t.Errorf("getIV() length = %d, want %d", len(iv), tt.wantLen)
			}

			for i := 0; i < tt.wantLen; i++ {
				if iv[i] != tt.wantData[i] {
					t.Errorf("getIV()[%d] = %v, want %v", i, iv[i], tt.wantData[i])
				}
			}
		})
	}
}

// TestTransQueryRequestBody 测试交易查询请求结构
func TestTransQueryRequestBody(t *testing.T) {
	req := &TransQueryRequestBody{
		X1: []TransQueryX1{
			{
				CardNbr:      "755947919810515",
				BeginDate:    "20240101",
				EndDate:      "20240131",
				CurrencyCode: "10",
			},
		},
	}

	if len(req.X1) != 1 {
		t.Errorf("TransQueryRequestBody.X1 length = %d, want 1", len(req.X1))
	}

	if req.X1[0].CardNbr != "755947919810515" {
		t.Errorf("TransQueryRequestBody.X1[0].CardNbr = %s, want 755947919810515", req.X1[0].CardNbr)
	}
}

// 注意：以下测试需要实际的密钥才能运行，这里仅作为框架示例

// BenchmarkSign 签名性能基准测试
func BenchmarkSign(b *testing.B) {
	// 需要实际的密钥才能运行
	b.Skip("需要实际的SM2密钥")

	config := &Config{
		Domain:           "http://test.example.com",
		UserID:           "TEST",
		SM4Key:           "1234567890123456",
		SM2PrivateKey:    "...",
		SM2BankPublicKey: "...",
	}

	crypto, _ := NewCryptoManager(config)
	data := []byte("test data for signing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = crypto.Sign(data)
	}
}

// BenchmarkEncrypt 加密性能基准测试
func BenchmarkEncrypt(b *testing.B) {
	// 需要实际的密钥才能运行
	b.Skip("需要实际的SM2密钥")

	config := &Config{
		Domain:           "http://test.example.com",
		UserID:           "TEST",
		SM4Key:           "1234567890123456",
		SM2PrivateKey:    "...",
		SM2BankPublicKey: "...",
	}

	crypto, _ := NewCryptoManager(config)
	data := []byte("test data for encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = crypto.Encrypt(data)
	}
}
