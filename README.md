# cmb-sdk

招商银行新直联（免前置）Go SDK，支持账户交易查询、回单查询与财务变动通知。

## 功能特性

- 完整的国密算法支持（SM2签名/验签、SM4加密/解密）
- 账户交易查询（`trsQryByBreakPoint`），支持自动续传
- 单笔回单查询（`DCSIGREC`），支持 PDF/OFD
- 企银支付单笔经办（`BB1PAYOP`）
- 企银支付业务查询（`BB1PAYQR`）
- 财务变动通知（`YQN01010`）与支付结果通知（`YQN02030`）
- 并发控制（默认最多 5 个并发请求）
- **可插拔日志系统**（支持自定义日志实例，兼容 zap/logrus/slog 等）
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
    // 方式1：使用内置 Debug 开关（向后兼容）
    Debug:            true,
    // 方式2：传入自定义 Logger（优先级更高）
    // Logger:        cmb.NewDefaultLogger(cmb.LogLevelInfo, os.Stdout),
}

client, err := cmb.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

### 3. 日志系统

SDK 提供可插拔的日志接口，支持两种风格：
- **结构化风格**（zap-style）：`Debug/Info/Warn/Error`，使用 key-value 对
- **格式化风格**（printf-style）：`Debugf/Infof/Warnf/Errorf`，使用 format + args

支持 4 个级别：`Debug` > `Info` > `Warn` > `Error` > `Silent`

#### 使用内置日志

```go
// Debug 级别，输出到 stderr（默认）
config.Logger = cmb.NewDefaultLogger(cmb.LogLevelDebug, nil)

// Info 级别，输出到文件
f, _ := os.OpenFile("cmb-sdk.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
config.Logger = cmb.NewDefaultLogger(cmb.LogLevelInfo, f)

// 同时输出到标准输出和日志文件（推荐）
config.Logger, _ = cmb.NewDefaultLoggerWithFile(cmb.LogLevelDebug, "/var/log/cmb-sdk.log")

// 完全静默
config.Logger = cmb.NewNopLogger()
```

#### 对接自定义日志（如 zap）

只需实现 `cmb.Logger` 接口（8 个方法）：

```go
type Logger interface {
    // 结构化风格（key-value 对）
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})

    // 格式化风格（printf-style）
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

示例（对接 zap）：

```go
type ZapLogger struct {
    logger *zap.SugaredLogger
}

func (z *ZapLogger) Debug(msg string, kv ...interface{})  { z.logger.Debugw(msg, kv...) }
func (z *ZapLogger) Info(msg string, kv ...interface{})   { z.logger.Infow(msg, kv...) }
func (z *ZapLogger) Warn(msg string, kv ...interface{})   { z.logger.Warnw(msg, kv...) }
func (z *ZapLogger) Error(msg string, kv ...interface{})  { z.logger.Errorw(msg, kv...) }
func (z *ZapLogger) Debugf(f string, a ...interface{})    { z.logger.Debugf(f, a...) }
func (z *ZapLogger) Infof(f string, a ...interface{})     { z.logger.Infof(f, a...) }
func (z *ZapLogger) Warnf(f string, a ...interface{})     { z.logger.Warnf(f, a...) }
func (z *ZapLogger) Errorf(f string, a ...interface{})    { z.logger.Errorf(f, a...) }

config.Logger = &ZapLogger{logger: zap.S()}
```

#### 运行时切换日志

```go
// 运行时动态切换（线程安全）
client.SetLogger(cmb.NewDefaultLogger(cmb.LogLevelDebug, os.Stdout))
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

### 企银支付业务查询

```go
reqBody := &cmb.PaymentQueryRequestBody{
    X1: []cmb.PaymentQueryX1{
        {
            BusCod: "N02030",
            YurRef: "YOUR_YUR_REF_123",
        },
    },
}

respBody, head, err := client.PaymentQuery(reqBody, "")
if err != nil {
    log.Fatal(err)
}

if len(respBody.Z1) > 0 {
    res := respBody.Z1[0]
    fmt.Printf("状态: %s, 结果: %s, 原因: %s\n", res.ReqSts, res.RtnFlg, res.RtnNar)
}
```

### 支付结果通知处理 (YQN02030)

支持解析业务完成通知 (`FINS`) 和支付退票通知 (`FINB`)。

```go
err := client.HandleNotification(data, func(msg *cmb.NotificationMessage) error {
    switch msg.MsgTyp {
    case "FINS": // 业务完成
        info := msg.MsgDat.TrsInfo
        fmt.Printf("业务%s完成，金额: %s\n", info.YurRef, info.TrsAmt)
    case "FINB": // 支付退票
        back := msg.MsgDat.BackInfo
        fmt.Printf("业务%s退票，原因: %s\n", back.YurRef, back.RtnNar)
    }
    return nil
})
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
