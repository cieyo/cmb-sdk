package cmb

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/sm2"
)

const (
	notificationVerifyUID = "1234567812345678"
)

type activeNotificationEnvelope struct {
	CorpNbr  string `json:"corpNbr"`
	EnptData string `json:"enptData"`
}

type activeNotificationPayload struct {
	SigTim string `json:"sigtim"`
	SigDat string `json:"sigdat"`
	NotDat string `json:"notdat"`
	NotKey string `json:"notkey"`
	UsrNbr string `json:"usrnbr"`
	NotNbr string `json:"notnbr"`
	NotTyp string `json:"nottyp"`
}

const (
	notificationTypeFinanceChange = "YQN01010"
	notificationTypePaymentResult = "YQN02030"
)

// ParseNotification 解析财务变动通知（YQN01010）
// 用于解析招商银行推送的财务变动通知
// data: 通知数据（JSON格式或加密后的BASE64字符串）
// 返回: 通知消息和错误
func (c *Client) ParseNotification(data []byte) (*NotificationMessage, error) {
	c.logger.Infof("========== PARSE NOTIFICATION START ==========")
	c.logger.Debug("parse notification start", "data_length", len(data))
	c.logger.Debugf("notification raw data:\n%s", string(data))

	raw := strings.TrimSpace(string(data))
	if strings.HasPrefix(raw, "{") && strings.Contains(raw, "\"enptData\"") {
		// 主动通知外层包装（debug模式可不验签）。检测到外层包装时不再回退，
		// 避免把真实错误覆盖成"整段JSON不是base64"。
		c.logger.Debugf("detected active notification envelope format")
		return c.parseNotificationEnvelope(data, false)
	}

	var msg NotificationMessage

	// 尝试直接解析JSON
	err := json.Unmarshal(data, &msg)
	if err == nil && (msg.MsgTyp != "" || msg.MsgDat.RefNbr != "") {
		c.logger.Infof("notification parsed (plain JSON): type=%s, acc=%s, amount=%s, ref=%s",
			msg.MsgTyp, msg.MsgDat.AccNbr, msg.MsgDat.CTrsAmt, msg.MsgDat.RefNbr)
		c.logger.Infof("========== PARSE NOTIFICATION END ==========")
		return &msg, nil
	}

	// 主动通知外层包装（debug模式可不验签）
	activeMsg, err := c.parseNotificationEnvelope(data, false)
	if err == nil && activeMsg != nil {
		return activeMsg, nil
	}
	var unsupportedTypeErr *ErrUnsupportedNotificationType
	if errors.As(err, &unsupportedTypeErr) {
		c.logger.Warn("unsupported notification type", "error", err)
		return nil, err
	}

	// 如果直接解析失败，尝试解密后再解析（兼容旧纯密文格式）
	c.logger.Debugf("trying legacy encrypted format")
	decryptedData, err := c.crypto.Decrypt(raw)
	if err != nil {
		c.logger.Errorf("decrypt notification failed: %v", err)
		return nil, fmt.Errorf("parse notification failed: %w", err)
	}

	c.logger.Debugf("decrypted notification data:\n%s", string(decryptedData))

	err = json.Unmarshal(decryptedData, &msg)
	if err != nil {
		c.logger.Errorf("unmarshal decrypted notification failed: %v", err)
		return nil, fmt.Errorf("parse notification failed: %w", err)
	}

	c.logger.Infof("notification parsed (legacy encrypted): type=%s, acc=%s, amount=%s, ref=%s",
		msg.MsgTyp, msg.MsgDat.AccNbr, msg.MsgDat.CTrsAmt, msg.MsgDat.RefNbr)
	c.logger.Infof("========== PARSE NOTIFICATION END ==========")
	return &msg, nil
}

// ParseNotificationFromEncrypted 从加密数据中解析通知
// encryptedData: BASE64编码的加密数据
// 返回: 通知消息和错误
func (c *Client) ParseNotificationFromEncrypted(encryptedData string) (*NotificationMessage, error) {
	raw := strings.TrimSpace(encryptedData)
	// 官方主动通知外层：{"corpNbr":"...","enptData":"..."}，生产环境需验签
	activeMsg, err := c.parseNotificationEnvelope([]byte(raw), true)
	if err == nil && activeMsg != nil {
		return activeMsg, nil
	}
	// 若报文是JSON对象，说明调用方传的是官方外层包装；此时应直接返回包装解析错误，
	// 避免继续按“纯base64”回退导致错误信息被掩盖。
	if strings.HasPrefix(raw, "{") {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("invalid active notification payload")
	}

	// 兼容旧格式（直接传enptData）
	// 解密
	decryptedData, err := c.crypto.Decrypt(raw)
	if err != nil {
		return nil, fmt.Errorf("decrypt notification failed: %w", err)
	}

	// 解析JSON
	var msg NotificationMessage
	err = json.Unmarshal(decryptedData, &msg)
	if err != nil {
		return nil, fmt.Errorf("parse notification failed: %w", err)
	}

	return &msg, nil
}

func (c *Client) parseNotificationEnvelope(data []byte, verify bool) (*NotificationMessage, error) {
	var env activeNotificationEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	env.CorpNbr = strings.TrimSpace(env.CorpNbr)
	env.EnptData = strings.TrimSpace(env.EnptData)
	if env.CorpNbr == "" || env.EnptData == "" {
		return nil, fmt.Errorf("invalid active notification envelope")
	}

	c.logger.Debugf("decrypting notification envelope: corp_nbr=%s, verify=%v, enpt_data_length=%d",
		env.CorpNbr, verify, len(env.EnptData))

	decryptedData, err := c.crypto.DecryptCTRWithSeed(env.EnptData, env.CorpNbr)
	if err != nil {
		c.logger.Errorf("decrypt active notification failed: corp_nbr=%s, error=%v", env.CorpNbr, err)
		return nil, fmt.Errorf("decrypt active notification failed: %w", err)
	}

	c.logger.Debugf("decrypted notification payload:\n%s", string(decryptedData))

	var payload activeNotificationPayload
	if err = json.Unmarshal(decryptedData, &payload); err != nil {
		c.logger.Error("parse active notification payload failed", "error", err)
		return nil, fmt.Errorf("parse active notification payload failed: %w", err)
	}
	if payload.NotTyp != "" &&
		!strings.EqualFold(strings.TrimSpace(payload.NotTyp), notificationTypeFinanceChange) &&
		!strings.EqualFold(strings.TrimSpace(payload.NotTyp), notificationTypePaymentResult) {
		return nil, &ErrUnsupportedNotificationType{NotTyp: strings.TrimSpace(payload.NotTyp)}
	}

	if verify {
		c.logger.Debug("verifying notification signature", "not_typ", payload.NotTyp)
		sigdat := strings.TrimSpace(payload.SigDat)
		if sigdat == "" {
			return nil, fmt.Errorf("active notification signature is empty")
		}
		verifyStr := strings.Replace(string(decryptedData), sigdat, SignaturePlaceholder, 1)
		if err = c.crypto.VerifyWithFixedUID([]byte(verifyStr), sigdat, notificationVerifyUID); err != nil {
			c.logger.Error("verify active notification signature failed", "error", err)
			return nil, fmt.Errorf("verify active notification signature failed: %w", err)
		}
		c.logger.Debug("notification signature verified")
	}

	notdat := strings.TrimSpace(payload.NotDat)
	if notdat == "" {
		return nil, fmt.Errorf("active notification notdat is empty")
	}

	var msg NotificationMessage
	if err = json.Unmarshal([]byte(notdat), &msg); err != nil {
		return nil, fmt.Errorf("parse active notification notdat failed: %w", err)
	}

	if strings.EqualFold(payload.NotTyp, notificationTypeFinanceChange) {
		c.logger.Infof("notification parsed (envelope): type=%s, acc=%s, amount=%s, ref=%s, corp=%s",
			msg.MsgTyp, msg.MsgDat.AccNbr, msg.MsgDat.CTrsAmt, msg.MsgDat.RefNbr, env.CorpNbr)
	} else if strings.EqualFold(payload.NotTyp, notificationTypePaymentResult) {
		var ref, amt string
		if msg.MsgDat.TrsInfo != nil {
			ref = msg.MsgDat.TrsInfo.YurRef
			amt = msg.MsgDat.TrsInfo.TrsAmt
		} else if msg.MsgDat.BackInfo != nil {
			ref = msg.MsgDat.BackInfo.YurRef
			if a, ok := msg.MsgDat.BackInfo.TrsAmt.(string); ok {
				amt = a
			} else {
				amt = fmt.Sprintf("%v", msg.MsgDat.BackInfo.TrsAmt)
			}
		}
		c.logger.Infof("payment notification parsed (envelope): type=%s, ref=%s, amount=%s, corp=%s",
			msg.MsgTyp, ref, amt, env.CorpNbr)
	}

	c.logger.Infof("========== PARSE NOTIFICATION END ==========")
	return &msg, nil
}

// NotificationHandler 通知处理器函数类型
type NotificationHandler func(msg *NotificationMessage) error

// HandleNotification 处理通知的辅助函数
// 这个函数可以在HTTP handler中使用
// data: 从HTTP请求body中读取的数据
// handler: 业务处理函数
// 返回: 错误
func (c *Client) HandleNotification(data []byte, handler NotificationHandler) error {
	// 解析通知
	msg, err := c.ParseNotification(data)
	if err != nil {
		return err
	}

	// 执行业务处理
	if handler != nil {
		err = handler(msg)
		if err != nil {
			return fmt.Errorf("handle notification failed: %w", err)
		}
	}

	return nil
}

// BuildNotificationEnvelope 构建一份完整的招行主动通知加密报文。
// 生成的 JSON 可直接 POST 到 /cmb/notify，走完整的解密+验签流程。
// corpNbr: 企业编号（如 P0204921），用作 CTR 加密的 IV seed
// msg:     要发送的通知消息体
func (c *Client) BuildNotificationEnvelope(corpNbr string, msg *NotificationMessage) ([]byte, error) {
	if corpNbr == "" {
		return nil, fmt.Errorf("corpNbr is required")
	}
	if msg == nil {
		return nil, fmt.Errorf("notification message is required")
	}

	// 1. 将 msg 序列化为 notdat
	notdatBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal notification message failed: %w", err)
	}

	// 2. 构造 payload（sigdat 先放占位符，用于签名）
	now := time.Now()
	payload := activeNotificationPayload{
		SigTim: now.Format("20060102150405"),
		SigDat: SignaturePlaceholder,
		NotDat: string(notdatBytes),
		NotTyp: notificationTypeFinanceChange,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %w", err)
	}

	// 3. SM2 签名（使用固定 UID：1234567812345678）
	r, s, err := sm2.Sm2Sign(c.crypto.sm2PrivateKey, payloadBytes, []byte(notificationVerifyUID), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("sign payload failed: %w", err)
	}
	payload.SigDat = base64.StdEncoding.EncodeToString(encodeRawRS(r, s))

	// 4. 重新序列化带真实签名的 payload
	signedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal signed payload failed: %w", err)
	}

	// 5. SM4-CTR 加密
	enptData, err := c.crypto.EncryptCTRWithSeed(signedPayload, corpNbr)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload failed: %w", err)
	}

	// 6. 组装外层信封
	env := activeNotificationEnvelope{
		CorpNbr:  corpNbr,
		EnptData: enptData,
	}
	return json.Marshal(env)
}
