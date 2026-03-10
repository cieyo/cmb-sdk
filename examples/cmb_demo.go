package main

import (
	"fmt"
	"log"
	"github.com/ceiyo/cmb-sdk/cmb"
)

func main() {
	// ==========================================
	// 1. 初始化客户端
	// ==========================================
	config := &cmb.Config{
		Domain: "http://cdctest.cmburl.cn/cdcserver/api/v2", // 测试环境
		UserID: "N002432758",                                // 替换为实际的用户ID

		// SM4对称密钥（16位）
		SM4Key: "1234567890123456", // 替换为实际的SM4密钥

		// SM2私钥（PEM格式）
		SM2PrivateKey: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg...
-----END PRIVATE KEY-----`, // 替换为实际的SM2私钥

		// 招商银行SM2公钥（PEM格式）
		SM2BankPublicKey: `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE...
-----END PUBLIC KEY-----`, // 替换为实际的银行公钥

		Debug: true, // 开启调试日志
	}

	// 创建客户端
	client, err := cmb.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	fmt.Println("招商银行新直联客户端初始化成功")

	// ==========================================
	// 2. 示例：账户交易查询（trsQryByBreakPoint）
	// ==========================================
	fmt.Println("\n--- 查询账户交易 ---")

	// 方式1：单次查询（最多200条）
	queryReq := &cmb.TransQueryRequestBody{
		X1: []cmb.TransQueryX1{
			{
				CardNbr:      "755947919810515", // 户口号
				BeginDate:    "20240101",        // 开始日期
				EndDate:      "20240131",        // 结束日期
				CurrencyCode: "10",              // 币种：10=人民币
			},
		},
	}

	respBody, head, err := client.QueryAccountTransaction(queryReq, "")
	if err != nil {
		log.Printf("查询交易失败: %v", err)
	} else {
		fmt.Printf("查询成功，返回 %d 条交易记录\n", len(respBody.Z2))
		fmt.Printf("请求ID: %s\n", head.ReqID)

		// 打印前3条交易
		for i, tx := range respBody.Z2 {
			if i >= 3 {
				break
			}
			fmt.Printf("  交易日期: %s, 金额: %s, 借贷: %s, 对方: %s\n",
				tx.TransDate, tx.TransAmount, tx.LoanCode, tx.CtpAcctName)
		}

		// 检查是否还有更多数据
		if len(respBody.Z1) > 0 && respBody.Z1[0].CtnFlag == "Y" {
			fmt.Println("  还有更多数据需要续传查询...")
		}
	}

	// 方式2：自动查询所有数据（自动处理续传）
	fmt.Println("\n--- 查询所有交易（自动续传） ---")
	allTxns, err := client.QueryAccountTransactionAll(
		"755947919810515", // 户口号
		"20240101",        // 开始日期
		"20240131",        // 结束日期
		"10",              // 币种
	)
	if err != nil {
		log.Printf("查询所有交易失败: %v", err)
	} else {
		fmt.Printf("共查询到 %d 条交易记录\n", len(allTxns))
	}

	// ==========================================
	// 3. 示例：单笔回单查询（DCSIGREC）
	// ==========================================
	fmt.Println("\n--- 查询单笔回单 ---")

	// 从上面的交易中获取流水号
	if len(respBody.Z2) > 0 {
		tx := respBody.Z2[0]

		// 方式1：仅查询回单数据
		receiptReq := &cmb.ReceiptQueryRequestBody{
			EacNbr: "755947919810515",   // 账号
			QueDat: "2024-01-15",        // 查询日期（yyyy-MM-dd格式）
			TrsSeq: tx.TransSequenceIdn, // 交易流水号
			PriMod: "PDF",               // 打印模式：PDF或OFDEX
		}

		receiptResp, _, err := client.QuerySingleReceipt(receiptReq, "")
		if err != nil {
			log.Printf("查询回单失败: %v", err)
		} else {
			fmt.Printf("回单查询成功，验证码: %s\n", receiptResp.CheCod)
			fmt.Printf("回单数据长度: %d 字节（BASE64）\n", len(receiptResp.FilDat))
		}

		// 方式2：直接下载回单到文件
		checkCode, err := client.DownloadReceipt(
			"755947919810515",   // 账号
			"2024-01-15",        // 查询日期
			tx.TransSequenceIdn, // 交易流水号
			"/tmp/receipt.pdf",  // 输出文件路径
			"PDF",               // 打印模式
		)
		if err != nil {
			log.Printf("下载回单失败: %v", err)
		} else {
			fmt.Printf("回单已下载到 /tmp/receipt.pdf，验证码: %s\n", checkCode)
		}
	}

	// ==========================================
	// 4. 示例：财务变动通知处理（YQN01010）
	// ==========================================
	fmt.Println("\n--- 处理财务变动通知 ---")

	// 模拟接收到的通知数据（实际场景中这是从HTTP请求body中读取的）
	mockNotificationJSON := `{
		"msgtyp": "NCCRTTRS",
		"msgdat": {
			"trsdat": "20220608",
			"trstim": "122258",
			"c_trsamt": "1000.50",
			"amtcdr": "C",
			"refnbr": "C0446BM00026EYZ",
			"accnbr": "755936495310539",
			"accnam": "测试账户",
			"rpyacc": "755941221510216",
			"rpynam": "付款方公司",
			"blvamt": "50000.00"
		}
	}`

	// 定义通知处理函数
	notificationHandler := func(msg *cmb.NotificationMessage) error {
		fmt.Printf("收到通知类型: %s\n", msg.MsgTyp)
		fmt.Printf("  交易日期: %s %s\n", msg.MsgDat.TrsDat, msg.MsgDat.TrsTim)
		fmt.Printf("  交易金额: %s\n", msg.MsgDat.CTrsAmt)
		fmt.Printf("  借贷标记: %s\n", msg.MsgDat.AmtCdr)
		fmt.Printf("  流水号: %s\n", msg.MsgDat.RefNbr)
		fmt.Printf("  账户余额: %s\n", msg.MsgDat.BlvAmt)

		// 在这里执行业务逻辑（如更新数据库、发送通知等）

		return nil
	}

	// 处理通知
	err = client.HandleNotification([]byte(mockNotificationJSON), notificationHandler)
	if err != nil {
		log.Printf("处理通知失败: %v", err)
	} else {
		fmt.Println("通知处理成功")
	}

	fmt.Println("\n所有示例执行完毕")
}
