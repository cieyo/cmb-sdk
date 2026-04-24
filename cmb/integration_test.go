package cmb

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// 招行测试环境配置（来自 xlsx: 招行标准直联对接免前置.xlsx）
// 注意：这些是测试环境密钥，严禁用于生产环境
func getTestConfig() *Config {
	userID := getEnvOrDefault("CMB_USER_ID", "N002463244")
	if userID == "" {
		userID = "N002463244"
	}

	return &Config{
		Domain:           getEnvOrDefault("CMB_DOMAIN", "http://cdctest.cmburl.cn/cdcserver/api/v2"),
		UserID:           userID,
		SM4Key:           getEnvOrDefault("CMB_SM4_KEY", "VuAzSWQhsoNqzn0K"),
		SM2PrivateKey:    getEnvOrDefault("CMB_SM2_PRIVATE_KEY", "NBtl7WnuUtA2v5FaebEkU0/Jj1IodLGT6lQqwkzmd2E="),
		SM2BankPublicKey: getEnvOrDefault("CMB_SM2_BANK_PUBLIC_KEY", "BNsIe9U0x8IeSe4h/dxUzVEz9pie0hDSfMRINRXc7s1UIXfkExnYECF4QqJ2SnHxLv3z/99gsfDQrQ6dzN5lZj0="),
		Timeout:          30 * time.Second,
		MaxConcurrent:    5,
		Debug:            true,
	}
}

func getTestAccountNbr() string {
	return getEnvOrDefault("CMB_TEST_ACCOUNT_NBR", "755915680110101")
}

func getEnvOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return fallback
}

// TestIntegration_CreateClient 测试客户端初始化（验证密钥解析是否正确）
func TestIntegration_CreateClient(t *testing.T) {
	config := getTestConfig()

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	if client == nil {
		t.Fatal("客户端不应为nil")
	}

	t.Log("✅ 客户端创建成功，密钥解析通过")
}

// TestIntegration_QueryTransaction 测试账户交易查询
// 运行前确保网络可达: telnet cdctest.cmburl.cn 80
func TestIntegration_QueryTransaction(t *testing.T) {
	if os.Getenv("CMB_INTEGRATION") != "1" {
		t.Skip("跳过集成测试。设置 CMB_INTEGRATION=1 环境变量以运行")
	}

	config := getTestConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 查询最近1个月的交易
	now := time.Now()
	beginDate := now.AddDate(0, -1, 0).Format("20060102")
	endDate := now.Format("20060102")

	reqBody := &TransQueryRequestBody{
		X1: []TransQueryX1{
			{
				CardNbr:   getTestAccountNbr(),
				BeginDate: beginDate,
				EndDate:   endDate,
			},
		},
	}

	t.Logf("查询交易: 账号=%s, 日期范围=%s ~ %s", getTestAccountNbr(), beginDate, endDate)

	respBody, head, err := client.QueryAccountTransaction(reqBody, "")
	if err != nil {
		t.Fatalf("交易查询失败: %v", err)
	}

	t.Logf("✅ 请求成功: resultcode=%s, reqid=%s", head.ResultCode, head.ReqID)

	if len(respBody.Z1) > 0 {
		z1 := respBody.Z1[0]
		t.Logf("  汇总: 借方笔数=%s, 借方金额=%s, 贷方笔数=%s, 贷方金额=%s, 续传=%s",
			z1.DebitNums, z1.DebitAmount, z1.CreditNums, z1.CreditAmount, z1.CtnFlag)
	}

	t.Logf("  交易明细数量: %d", len(respBody.Z2))
	for i, tx := range respBody.Z2 {
		if i >= 5 {
			t.Logf("  ... 共 %d 条，仅显示前5条", len(respBody.Z2))
			break
		}
		t.Logf("  [%d] 日期=%s, 流水号=%s, 借贷=%s, 金额=%s, 对方=%s",
			i+1, tx.TransDate, tx.TransSequenceIdn, tx.LoanCode, tx.TransAmount, tx.CtpAcctName)
	}
}

// TestIntegration_QueryReceipt 测试单笔回单查询
func TestIntegration_QueryReceipt(t *testing.T) {
	if os.Getenv("CMB_INTEGRATION") != "1" {
		t.Skip("跳过集成测试。设置 CMB_INTEGRATION=1 环境变量以运行")
	}

	config := getTestConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 注意：需要使用上面交易查询返回的真实流水号和日期
	// 这里使用占位值，你需要替换为实际值
	trsSeq := os.Getenv("CMB_TEST_TRS_SEQ")
	trsDate := os.Getenv("CMB_TEST_TRS_DATE") // 格式: yyyy-MM-dd

	if trsSeq == "" || trsDate == "" {
		t.Skip("跳过回单查询测试。设置 CMB_TEST_TRS_SEQ 和 CMB_TEST_TRS_DATE 环境变量")
	}

	reqBody := &ReceiptQueryRequestBody{
		EacNbr: getTestAccountNbr(),
		QueDat: trsDate,
		TrsSeq: trsSeq,
		PriMod: "PDF",
	}

	t.Logf("查询回单: 账号=%s, 日期=%s, 流水号=%s", getTestAccountNbr(), trsDate, trsSeq)

	respBody, head, err := client.QuerySingleReceipt(reqBody, "")
	if err != nil {
		t.Fatalf("回单查询失败: %v", err)
	}

	t.Logf("✅ 回单查询成功: resultcode=%s", head.ResultCode)
	t.Logf("  验证码=%s, 实例号=%s, 数据长度=%d", respBody.CheCod, respBody.IstNbr, len(respBody.FilDat))

	// 可选：保存到文件
	if respBody.FilDat != "" {
		outputPath := fmt.Sprintf("/tmp/cmb_receipt_%s.pdf", trsSeq)
		_, err := client.DownloadReceipt(getTestAccountNbr(), trsDate, trsSeq, outputPath, "PDF")
		if err == nil {
			t.Logf("  📄 回单已保存到: %s", outputPath)
		}
	}
}

// TestIntegration_QueryAllTransactions 测试自动续传查询所有交易
func TestIntegration_QueryAllTransactions(t *testing.T) {
	if os.Getenv("CMB_INTEGRATION") != "1" {
		t.Skip("跳过集成测试。设置 CMB_INTEGRATION=1 环境变量以运行")
	}

	config := getTestConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	now := time.Now()
	beginDate := now.AddDate(0, -1, 0).Format("20060102")
	endDate := now.Format("20060102")

	t.Logf("全量查询交易: 账号=%s, 日期范围=%s ~ %s", getTestAccountNbr(), beginDate, endDate)

	allTx, err := client.QueryAccountTransactionAll(getTestAccountNbr(), beginDate, endDate, "")
	if err != nil {
		t.Fatalf("全量交易查询失败: %v", err)
	}

	t.Logf("✅ 查询完毕，共 %d 笔交易", len(allTx))
}

// TestIntegration_ParseNotification 测试通知解析（使用文档中的示例数据）
func TestIntegration_ParseNotification(t *testing.T) {
	config := getTestConfig()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 使用文档中的示例通知内容
	sampleNotification := []byte(`{
		"msgdat": {
			"chknbr": " ",
			"infflg": "2",
			"refsub": " ",
			"refnbr": "C0446BM00026EYZ",
			"rpyacc": "755941221510216",
			"trscod": "CPUA",
			"gsbacc": " ",
			"otrnar": " ",
			"rpynam": "当当科技公司",
			"amtcdr": "C",
			"naryur": "摘要",
			"vltdat": "20220608",
			"yurref": "DBJBZDH20220608122130262",
			"accnam": "招行测试账户",
			"gsbnam": " ",
			"narext": " ",
			"trsanl": " ",
			"nusage": " ",
			"trsdat": "20220608",
			"reqnbr": "7000021370",
			"trstim": "122258",
			"rpybnk": "招商银行 深圳分行深圳分行蛇口支行",
			"gsbbbk": " ",
			"frmcod": " ",
			"athflg": "N",
			"rpybbn": " ",
			"rsvflg": "N",
			"accnbr": "755936495310539",
			"busnam": "支付",
			"rpybbk": " ",
			"c_trsamt": "1",
			"c_ccynbr": "人民币",
			"busnar": " ",
			"blvamt": "76.08",
			"rpyadr": "广东省深圳市"
		},
		"msgtyp": "NCCRTTRS"
	}`)

	msg, err := client.ParseNotification(sampleNotification)
	if err != nil {
		t.Fatalf("解析通知失败: %v", err)
	}

	t.Logf("✅ 通知类型: %s", msg.MsgTyp)
	t.Logf("  账号: %s, 户名: %s", msg.MsgDat.AccNbr, msg.MsgDat.AccNam)
	t.Logf("  借贷: %s, 金额: %s, 余额: %s", msg.MsgDat.AmtCdr, msg.MsgDat.CTrsAmt, msg.MsgDat.BlvAmt)
	t.Logf("  流水号: %s, 交易日期: %s %s", msg.MsgDat.RefNbr, msg.MsgDat.TrsDat, msg.MsgDat.TrsTim)
	t.Logf("  对方: %s (%s)", msg.MsgDat.RpyNam, msg.MsgDat.RpyAcc)

	// 验证关键字段
	if msg.MsgTyp != "NCCRTTRS" {
		t.Errorf("预期通知类型 NCCRTTRS，实际 %s", msg.MsgTyp)
	}
	if msg.MsgDat.AmtCdr != "C" {
		t.Errorf("预期借贷标记 C，实际 %s", msg.MsgDat.AmtCdr)
	}
	if msg.MsgDat.RefNbr != "C0446BM00026EYZ" {
		t.Errorf("预期流水号 C0446BM00026EYZ，实际 %s", msg.MsgDat.RefNbr)
	}
}
