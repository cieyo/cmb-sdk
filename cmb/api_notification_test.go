package cmb

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm4"
)

func TestParseNotificationFromEncrypted_ActiveEnvelope(t *testing.T) {
	client := mustNewTestClient(t)
	corpNbr := "P0204921"

	notdat := `{"msgtyp":"NCCRTTRS","msgdat":{"accnbr":"755915680110101","trsdat":"20260305","trstim":"091122","c_ccynbr":"10","c_trsamt":"100.00","refnbr":"C0147DJ0000FYBZ","amtcdr":"C","rpyacc":"测试企业"}}`
	payload := activeNotificationPayload{
		SigTim: "20260305170100",
		SigDat: SignaturePlaceholder,
		NotDat: notdat,
		NotTyp: "YQN01010",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	// 主动通知按固定UID验签
	r, s, err := sm2.Sm2Sign(client.crypto.sm2PrivateKey, payloadBytes, []byte(notificationVerifyUID), rand.Reader)
	if err != nil {
		t.Fatalf("sign payload failed: %v", err)
	}
	payload.SigDat = base64.StdEncoding.EncodeToString(encodeRawRS(r, s))

	signedPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal signed payload failed: %v", err)
	}

	ciphertext, err := encryptCTRBase64([]byte(client.config.SM4Key), corpNbr, signedPayload)
	if err != nil {
		t.Fatalf("encrypt payload failed: %v", err)
	}

	env := activeNotificationEnvelope{
		CorpNbr:  corpNbr,
		EnptData: ciphertext,
	}
	body, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope failed: %v", err)
	}

	msg, err := client.ParseNotificationFromEncrypted(string(body))
	if err != nil {
		t.Fatalf("ParseNotificationFromEncrypted failed: %v", err)
	}
	if msg == nil {
		t.Fatal("message is nil")
	}
	if msg.MsgDat.RefNbr != "C0147DJ0000FYBZ" {
		t.Fatalf("unexpected refnbr: %s", msg.MsgDat.RefNbr)
	}
	if msg.MsgDat.AmtCdr != "C" {
		t.Fatalf("unexpected amtcdr: %s", msg.MsgDat.AmtCdr)
	}
}

func TestParseNotification_ActiveEnvelopeSkipVerify(t *testing.T) {
	client := mustNewTestClient(t)
	corpNbr := "P0204921"

	payload := activeNotificationPayload{
		SigTim: "20260305170500",
		// debug 解析路径可不验签
		SigDat: "invalid-signature-for-debug",
		NotDat: `{"msgtyp":"NCCRTTRS","msgdat":{"accnbr":"755915680110101","trsdat":"20260305","trstim":"101122","c_ccynbr":"10","c_trsamt":"88.88","refnbr":"C0147DJ0000FYZZ","amtcdr":"C"}}`,
		NotTyp: "YQN01010",
	}
	signedPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	ciphertext, err := encryptCTRBase64([]byte(client.config.SM4Key), corpNbr, signedPayload)
	if err != nil {
		t.Fatalf("encrypt payload failed: %v", err)
	}

	env := activeNotificationEnvelope{
		CorpNbr:  corpNbr,
		EnptData: ciphertext,
	}
	body, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope failed: %v", err)
	}

	msg, err := client.ParseNotification(body)
	if err != nil {
		t.Fatalf("ParseNotification failed: %v", err)
	}
	if msg == nil {
		t.Fatal("message is nil")
	}
	if msg.MsgDat.RefNbr != "C0147DJ0000FYZZ" {
		t.Fatalf("unexpected refnbr: %s", msg.MsgDat.RefNbr)
	}
}

func mustNewTestClient(t *testing.T) *Client {
	t.Helper()

	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate sm2 key failed: %v", err)
	}

	d := privateKey.D.FillBytes(make([]byte, 32))
	pub := make([]byte, 65)
	pub[0] = 0x04
	copy(pub[1:33], privateKey.PublicKey.X.FillBytes(make([]byte, 32)))
	copy(pub[33:65], privateKey.PublicKey.Y.FillBytes(make([]byte, 32)))

	client, err := NewClient(&Config{
		Domain:           "http://127.0.0.1",
		UserID:           "N123456789012345",
		SM4Key:           "1234567890abcdef",
		SM2PrivateKey:    base64.StdEncoding.EncodeToString(d),
		SM2BankPublicKey: base64.StdEncoding.EncodeToString(pub),
		Timeout:          1,
		MaxConcurrent:    1,
	})
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	return client
}

func encryptCTRBase64(key []byte, seed string, plaintext []byte) (string, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, 16)
	for i := range iv {
		iv[i] = '0'
	}
	copy(iv, []byte(seed))

	stream := cipher.NewCTR(block, iv)
	out := make([]byte, len(plaintext))
	stream.XORKeyStream(out, plaintext)
	return base64.StdEncoding.EncodeToString(out), nil
}
