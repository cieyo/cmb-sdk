package cmb

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm4"
	"github.com/tjfoc/gmsm/x509"
)

// CryptoManager 加密管理器
type CryptoManager struct {
	sm4Key             []byte          // 业务接口SM4对称密钥
	sm4NotifyKey       []byte          // 主动通知SM4对称密钥
	sm2PrivateKey      *sm2.PrivateKey // SM2私钥（用于签名）
	sm2BankPublicKey   *sm2.PublicKey  // 招商银行SM2公钥（用于验签）
	sm2NotifyPublicKey *sm2.PublicKey  // 招商银行主动通知验签公钥
	userID             string          // 用户ID（用作IV向量）
}

type sm2DER struct {
	R, S *big.Int
}

// NewCryptoManager 创建加密管理器
func NewCryptoManager(config *Config) (*CryptoManager, error) {
	sm4Key, err := parseSM4Key(config.SM4Key)
	if err != nil {
		return nil, err
	}
	sm4NotifyKey, err := parseSM4Key(config.SM4NotifyKey)
	if err != nil {
		return nil, err
	}
	cm := &CryptoManager{
		sm4Key:       sm4Key,
		sm4NotifyKey: sm4NotifyKey,
		userID:       config.UserID,
	}

	// 解析SM2私钥
	privateKey, err := parseSM2PrivateKey(config.SM2PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse SM2 private key failed: %w", err)
	}
	cm.sm2PrivateKey = privateKey

	// 解析招商银行SM2公钥
	publicKey, err := parseSM2PublicKey(config.SM2BankPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse SM2 bank public key failed: %w", err)
	}
	cm.sm2BankPublicKey = publicKey
	notifyPubKey, err := parseSM2PublicKey(config.SM2NotifyPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse SM2 notify public key failed: %w", err)
	}
	cm.sm2NotifyPublicKey = notifyPubKey

	return cm, nil
}

// Sign SM2签名
// data: 待签名数据
// 返回: BASE64编码的签名结果
func (cm *CryptoManager) Sign(data []byte) (string, error) {
	if cm.sm2PrivateKey == nil {
		return "", ErrInvalidSM2PrivateKey
	}

	uid := cm.getSM2UID()
	r, s, err := sm2.Sm2Sign(cm.sm2PrivateKey, data, uid, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSignatureFailed, err)
	}

	sign := encodeRawRS(r, s)

	// BASE64编码
	return base64.StdEncoding.EncodeToString(sign), nil
}

// Verify SM2验签
// data: 原始数据
// signBase64: BASE64编码的签名
// 返回: 验签是否通过
func (cm *CryptoManager) Verify(data []byte, signBase64 string) error {
	return cm.verifyWithUIDAndKey(cm.sm2BankPublicKey, data, signBase64, cm.getSM2UID())
}

// VerifyWithFixedUID 使用固定UID进行验签（主动通知场景）
func (cm *CryptoManager) VerifyWithFixedUID(data []byte, signBase64, uid string) error {
	pub := cm.sm2NotifyPublicKey
	if pub == nil {
		pub = cm.sm2BankPublicKey
	}
	return cm.verifyWithUIDAndKey(pub, data, signBase64, []byte(uid))
}

func (cm *CryptoManager) verifyWithUIDAndKey(pub *sm2.PublicKey, data []byte, signBase64 string, uid []byte) error {
	if pub == nil {
		return ErrInvalidSM2PublicKey
	}

	// BASE64解码
	sign, err := base64.StdEncoding.DecodeString(signBase64)
	if err != nil {
		return fmt.Errorf("%w: base64 decode failed: %v", ErrVerifySignatureFailed, err)
	}

	// 优先兼容银行返回 R||S（64字节）签名格式
	if len(sign) == 64 {
		r := new(big.Int).SetBytes(sign[:32])
		s := new(big.Int).SetBytes(sign[32:])
		if sm2.Sm2Verify(pub, data, uid, r, s) {
			return nil
		}
	}

	// 兼容 ASN.1 DER
	if r, s, err := decodeDERSign(sign); err == nil {
		if sm2.Sm2Verify(pub, data, uid, r, s) {
			return nil
		}
	}

	// 最后回退默认库逻辑（默认ID）
	if pub.Verify(data, sign) {
		return nil
	}

	return ErrVerifySignatureFailed
}

// Encrypt SM4加密
// plaintext: 明文数据
// 返回: BASE64编码的密文
func (cm *CryptoManager) Encrypt(plaintext []byte) (string, error) {
	// 设置IV向量：用户ID补0到16位
	sm4.IV = cm.getIV()

	// SM4 CBC加密（mode=true表示加密）
	ciphertext, err := sm4.Sm4Cbc(cm.sm4Key, plaintext, true)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	// BASE64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt SM4解密
// ciphertextBase64: BASE64编码的密文
// 返回: 明文数据
func (cm *CryptoManager) Decrypt(ciphertextBase64 string) ([]byte, error) {
	// BASE64解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %v", ErrDecryptFailed, err)
	}

	// 设置IV向量：用户ID补0到16位
	sm4.IV = cm.getIV()

	// SM4 CBC解密（mode=false表示解密）
	plaintext, err := sm4.Sm4Cbc(cm.sm4Key, ciphertext, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return plaintext, nil
}

// DecryptCTRWithSeed 使用SM4-CTR解密，IV由seed右补'0'至16位
func (cm *CryptoManager) DecryptCTRWithSeed(ciphertextBase64, seed string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %v", ErrDecryptFailed, err)
	}

	block, err := sm4.NewCipher(cm.sm4NotifyKey)
	if err != nil {
		return nil, fmt.Errorf("%w: init sm4 ctr failed: %v", ErrDecryptFailed, err)
	}

	iv := cm.buildIV(seed)
	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}

// EncryptCTRWithSeed 使用SM4-CTR加密，IV由seed右补'0'至16位
func (cm *CryptoManager) EncryptCTRWithSeed(plaintext []byte, seed string) (string, error) {
	block, err := sm4.NewCipher(cm.sm4NotifyKey)
	if err != nil {
		return "", fmt.Errorf("%w: init sm4 ctr failed: %v", ErrEncryptFailed, err)
	}

	iv := cm.buildIV(seed)
	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// getIV 获取IV向量
// 规则：用户ID不足16位时，右侧用字符'0'（0x30）补齐到16位
// 参考文档: iv = USER_ID.ljust(16, '0')[:16]
// 注意：是ASCII字符'0'，不是空字节0x00！
func (cm *CryptoManager) getIV() []byte {
	return cm.buildIV(cm.userID)
}

func (cm *CryptoManager) buildIV(seed string) []byte {
	iv := make([]byte, 16)
	// 先用字符'0'填满
	for i := range iv {
		iv[i] = '0'
	}
	// 再把seed覆盖前面的部分
	copy(iv, []byte(seed))
	return iv
}

func (cm *CryptoManager) getSM2UID() []byte {
	return cm.getIV()
}

func parseSM4Key(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 16 {
		return []byte(raw), nil
	}
	if len(raw) == 32 {
		bs, err := hex.DecodeString(raw)
		if err == nil && len(bs) == 16 {
			return bs, nil
		}
	}
	return nil, ErrInvalidSM4Key
}

// parseSM2PrivateKey 解析SM2私钥
// 支持两种格式：
// 1. PEM格式（-----BEGIN PRIVATE KEY-----...-----END PRIVATE KEY-----）
// 2. 裸Base64格式（招行xlsx配置中提供的格式，如 "NBtl7WnuUtA2v5FaebEkU0/Jj1IodLGT6lQqwkzmd2E="）
func parseSM2PrivateKey(keyStr string) (*sm2.PrivateKey, error) {
	// 尝试1：PEM格式
	block, _ := pem.Decode([]byte(keyStr))
	if block != nil {
		privateKey, err := x509.ParsePKCS8UnecryptedPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PEM SM2 private key: %w", err)
		}
		return privateKey, nil
	}

	// 尝试2：裸Base64格式（招行提供的32字节私钥原始值）
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	// 尝试按PKCS8解析
	privateKey, err := x509.ParsePKCS8UnecryptedPrivateKey(keyBytes)
	if err == nil {
		return privateKey, nil
	}

	// 尝试直接作为SM2 D值（32字节的大整数）
	if len(keyBytes) == 32 {
		curve := sm2.P256Sm2()
		priv := new(sm2.PrivateKey)
		priv.D = new(big.Int).SetBytes(keyBytes)
		priv.PublicKey.Curve = curve
		priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(keyBytes)
		return priv, nil
	}

	return nil, fmt.Errorf("unsupported SM2 private key format (len=%d)", len(keyBytes))
}

// parseSM2PublicKey 解析SM2公钥
// 支持两种格式：
// 1. PEM格式（-----BEGIN PUBLIC KEY-----...-----END PUBLIC KEY-----）
// 2. 裸Base64格式（招行xlsx配置中提供的格式，65字节04开头的未压缩公钥点）
func parseSM2PublicKey(keyStr string) (*sm2.PublicKey, error) {
	// 尝试1：PEM格式
	block, _ := pem.Decode([]byte(keyStr))
	if block != nil {
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PEM SM2 public key: %w", err)
		}
		publicKey, ok := pubInterface.(*sm2.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not a SM2 public key")
		}
		return publicKey, nil
	}

	// 尝试2：裸Base64格式（招行提供的65字节未压缩公钥 04||X||Y）
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
	}

	// 尝试按PKIX解析
	pubInterface, err := x509.ParsePKIXPublicKey(keyBytes)
	if err == nil {
		publicKey, ok := pubInterface.(*sm2.PublicKey)
		if ok {
			return publicKey, nil
		}
	}

	// 尝试按未压缩公钥点格式解析（04 || X(32字节) || Y(32字节)）
	if len(keyBytes) == 65 && keyBytes[0] == 0x04 {
		curve := sm2.P256Sm2()
		x := new(big.Int).SetBytes(keyBytes[1:33])
		y := new(big.Int).SetBytes(keyBytes[33:65])
		pub := &sm2.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		}
		return pub, nil
	}

	return nil, fmt.Errorf("unsupported SM2 public key format (len=%d)", len(keyBytes))
}

func encodeRawRS(r, s *big.Int) []byte {
	out := make([]byte, 64)

	rb := r.Bytes()
	sb := s.Bytes()

	copy(out[32-len(rb):32], rb)
	copy(out[64-len(sb):64], sb)
	return out
}

func decodeDERSign(sign []byte) (*big.Int, *big.Int, error) {
	var der sm2DER
	if _, err := asn1.Unmarshal(sign, &der); err != nil {
		return nil, nil, err
	}
	if der.R == nil || der.S == nil {
		return nil, nil, fmt.Errorf("invalid DER signature")
	}
	return der.R, der.S, nil
}
