package main

import (
	"fmt"
	"log"
	"os"

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

		// ===== 日志配置（五种方式任选其一） =====

		// 方式1：简单开关（向后兼容，输出 Debug 级别日志到 stderr）
		Debug: true,

		// 方式2：使用内置日志，控制级别和输出目标
		// Logger: cmb.NewDefaultLogger(cmb.LogLevelInfo, os.Stdout),

		// 方式3：同时输出到标准输出和日志文件
		// Logger: func() cmb.Logger {
		//     l, _ := cmb.NewDefaultLoggerWithFile(cmb.LogLevelDebug, "/var/log/cmb-sdk.log")
		//     return l
		// }(),

		// 方式4：只输出到日志文件
		// Logger: func() cmb.Logger {
		//     f, _ := os.OpenFile("cmb.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		//     return cmb.NewDefaultLogger(cmb.LogLevelInfo, f)
		// }(),

		// 方式5：对接自定义日志框架（传入实现 cmb.Logger 接口的实例）
		// Logger: &YourZapLogger{},
	}

	// 创建客户端
	client, err := cmb.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 运行时可以切换日志级别（线程安全）
	// client.SetLogger(cmb.NewDefaultLogger(cmb.LogLevelDebug, os.Stdout))

	// 也可以完全静默日志
	// client.SetLogger(cmb.NewNopLogger())

	_ = os.Stdout // 避免 import 未使用

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
	// 4. 示例：企银支付单笔经办（BB1PAYOP）
	// ==========================================
	fmt.Println("\n--- 企银支付单笔经办 ---")

	payReq := &cmb.PaymentOperateRequestBody{
		BusinessMode: []cmb.PayBusinessMode{
			{
				BusCod: "N02030", // 支付
				BusMod: "00034",  // 业务模式编号
			},
		},
		PaymentDetails: []cmb.PaymentDetail{
			{
				DbtAcc: "769900135910504",         // 转出帐号
				CrtAcc: "755903332110404",         // 收方帐号
				CrtNam: "测试收款方",                   // 收方户名
				CcyNbr: "10",                      // 人民币
				TrsAmt: "1.05",                    // 交易金额
				NusAge: "测试支付",                    // 用途
				BnkFlg: "Y",                       // 行内
				YurRef: cmb.GenerateReqID("YREF"), // 业务参考号（必须唯一）
			},
		},
	}

	payResp, payHead, err := client.PaymentOperate(payReq, "")
	if err != nil {
		log.Printf("支付经办失败: %v", err)
	} else {
		fmt.Printf("请求ID: %s\n", payHead.ReqID)
		if len(payResp.Results) > 0 {
			result := payResp.Results[0]
			fmt.Printf("流程实例号: %s\n", result.ReqNbr)
			fmt.Printf("请求状态: %s\n", result.ReqSts)
			fmt.Printf("错误码: %s\n", result.ErrCod)

			if result.NeedApproval() {
				fmt.Println("状态：等待审批")
				// 打印审批人信息
				for _, approver := range payResp.Approvers {
					fmt.Printf("  审批人: %s (%s)\n", approver.UsrNam, approver.UsrLgn)
				}
			} else if result.IsBankProcessing() {
				fmt.Println("状态：银行处理中")
			} else if result.IsFinished() {
				if result.IsSuccess() {
					fmt.Println("状态：支付成功")
				} else if result.IsFailed() {
					fmt.Printf("状态：支付失败，原因：%s\n", result.ErrTxt)
				}
			}
		}
	}

	// ==========================================
	// 4.5 示例：企银支付业务查询（BB1PAYQR）
	// ==========================================
	fmt.Println("\n--- 企银支付业务查询 ---")

	queryReqBody := &cmb.PaymentQueryRequestBody{
		X1: []cmb.PaymentQueryX1{
			{
				BusCod: "N02030",
				YurRef: "YREF20240424190000", // 替换为之前经办使用的参考号
			},
		},
	}

	qResp, qHead, err := client.PaymentQuery(queryReqBody, "")
	if err != nil {
		log.Printf("查询支付业务失败: %v", err)
	} else {
		fmt.Printf("请求ID: %s\n", qHead.ReqID)
		if len(qResp.Z1) > 0 {
			res := qResp.Z1[0]
			fmt.Printf("参考号: %s, 流程实例号: %s\n", res.YurRef, res.ReqNbr)
			fmt.Printf("处理状态: %s, 处理结果: %s, 失败原因: %s\n", res.ReqSts, res.RtnFlg, res.RtnNar)

			if res.IsSuccess() {
				fmt.Println("状态：支付已成功")
			} else if res.IsFailed() {
				fmt.Println("状态：支付已失败")
			}
		}
	}

	// ==========================================
	// 5. 示例：财务变动通知处理（YQN01010）
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
	}

	// ==========================================
	// 6. 示例：支付结果通知处理（YQN02030）
	// ==========================================
	fmt.Println("\n--- 处理支付结果通知 ---")

	// 模拟接收到的支付结果通知数据 (FINS - 业务完成)
	mockResultJSON := `{
		"msgtyp": "FINS",
		"msgdat": {
			"trsInfo": {
				"reqNbr": "5480037057",
				"busCod": "N02030",
				"yurRef": "YREF20240424190000",
				"reqSts": "FIN",
				"rtnFlg": "S",
				"trsAmt": "1.05"
			}
		}
	}`

	// 支付结果处理逻辑
	resultHandler := func(msg *cmb.NotificationMessage) error {
		switch msg.MsgTyp {
		case "FINS":
			info := msg.MsgDat.TrsInfo
			fmt.Printf("收到支付完成通知: 参考号=%s, 金额=%s, 结果=%s\n", info.YurRef, info.TrsAmt, info.RtnFlg)
		case "FINB":
			back := msg.MsgDat.BackInfo
			fmt.Printf("收到支付退票通知: 参考号=%s, 原因=%s\n", back.YurRef, back.RtnNar)
		}
		return nil
	}

	err = client.HandleNotification([]byte(mockResultJSON), resultHandler)
	if err != nil {
		log.Printf("处理结果通知失败: %v", err)
	} else {
		fmt.Println("结果通知处理成功")
	}

	fmt.Println("\n所有示例执行完毕")
}
