# cmb-sdk

招商银行新直联（免前置）Go SDK，支持账户交易查询、回单查询与财务变动通知。

## 功能特性

- 完整的国密算法支持（SM2签名/验签、SM4加密/解密）
- 账户交易查询（`trsQryByBreakPoint`），支持自动续传
- 单笔回单查询（`DCSIGREC`），支持 PDF/OFD
- 财务变动通知处理（`YQN01010`）
- 并发控制（默认最多 5 个并发请求）
- 调试日志支持
- 完整的错误处理与类型安全 API

## 安装

```bash
go get github.com/ceiyo/cmb-sdk
```

## 快速开始

### 1. 准备配置

请参考 `config/cmb.yaml.example`。注意：YAML 中写 `30` 可能被解析为 `30ns`，SDK 会自动按秒纠正，推荐明确写成 `30s`。

```yaml
domain: "http://cdctest.cmburl.cn/cdcserver/api/v2"
userid: "N000000000"
sm4_key: "0123456789abcdef"
sm2_private_key: "BASE64_OR_PEM"
sm2_bank_public_key: "BASE64_OR_PEM"
account_nbr: "755000000000000"

timeout: 30
max_concurrent: 5
debug: false
```

### 2. 初始化客户端

```go
import "github.com/ceiyo/cmb-sdk/cmb"

config := &cmb.Config{
    Domain:           "http://cdctest.cmburl.cn/cdcserver/api/v2",
    UserID:           "N000000000",
    SM4Key:           "0123456789abcdef",
    SM2PrivateKey:    "...",
    SM2BankPublicKey: "...",
    Debug:            true,
}

client, err := cmb.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

## 使用示例

### 账户交易查询（单次，最多200条）

```go
reqBody := &cmb.TransQueryRequestBody{
    X1: []cmb.TransQueryX1{
        {
            CardNbr:      "755947919810515",
            BeginDate:    "20240101",
            EndDate:      "20240131",
            CurrencyCode: "10",
        },
    },
}

respBody, head, err := client.QueryAccountTransaction(reqBody, "")
if err != nil {
    log.Fatal(err)
}

for _, tx := range respBody.Z2 {
    fmt.Printf("日期: %s, 金额: %s, 对方: %s\n",
        tx.TransDate, tx.TransAmount, tx.CtpAcctName)
}
```

### 账户交易查询（自动续传）

```go
allTxns, err := client.QueryAccountTransactionAll(
    "755947919810515",  // 户口号
    "20240101",         // 开始日期
    "20240131",         // 结束日期
    "10",               // 币种
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("共查询到 %d 条交易\n", len(allTxns))
```

### 单笔回单查询与下载

```go
reqBody := &cmb.ReceiptQueryRequestBody{
    EacNbr: "755947919810515",
    QueDat: "2024-01-15",
    TrsSeq: "C0446BM00026EYZ",
    PriMod: "PDF",
}

respBody, _, err := client.QuerySingleReceipt(reqBody, "")
if err != nil {
    log.Fatal(err)
}

// respBody.FilDat 包含 BASE64 编码的回单文件

checkCode, err := client.DownloadReceipt(
    "755947919810515",
    "2024-01-15",
    "C0446BM00026EYZ",
    "/tmp/receipt.pdf",
    "PDF",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("验证码: %s\n", checkCode)
```

### 财务变动通知处理

```go
// 在 HTTP handler 中使用
func handleCMBNotification(c *gin.Context) {
    data, _ := io.ReadAll(c.Request.Body)

    err := client.HandleNotification(data, func(msg *cmb.NotificationMessage) error {
        fmt.Printf("收到%s通知，金额：%s\n",
            msg.MsgTyp, msg.MsgDat.CTrsAmt)
        return nil
    })

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "ok"})
}
```

## 文档

- SDK 代码：`cmb/`
- 示例代码：`examples/cmb_demo.go`
- 官方文档整理：`docs/official/`

## 测试

```bash
go test -v ./...
```

## 注意事项

1. 生产环境必须使用 HTTPS。
2. 单用户同时请求数不超过 5 个，单接口 TPS 不超过 10 次/秒。
3. 签名时间校验：服务端会校验签名时间，前后相差超过 1 小时会报错。
4. 生产环境请勿在代码中硬编码密钥。

## 免责声明

`docs/official/` 中的内容为对招商银行官方资料的整理，仅用于接口对接参考，版权归原作者/权利人所有。

## License

MIT
