package cmb

// RequestHead 请求头
type RequestHead struct {
	// FunCode 接口功能码
	FunCode string `json:"funcode"`

	// UserID 客户用户ID
	UserID string `json:"userid"`

	// ReqID 请求唯一标识（建议规则：时间戳+随机串）
	ReqID string `json:"reqid"`
}

// ResponseHead 响应头
type ResponseHead struct {
	// FunCode 接口功能码
	FunCode string `json:"funcode"`

	// BizCode 业务代码
	BizCode string `json:"bizcode"`

	// ReqID 请求唯一标识
	ReqID string `json:"reqid"`

	// ResultCode 处理结果码（000000或SUC0000为成功）
	ResultCode string `json:"resultcode"`

	// ResultMsg 处理结果描述
	ResultMsg string `json:"resultmsg"`

	// RspID 响应唯一标识
	RspID string `json:"rspid"`

	// UserID 客户用户ID
	UserID string `json:"userid"`
}

// Signature 签名信息
type Signature struct {
	// SigTim 签名时间（格式：YYYYMMDDHHmmss）
	SigTim string `json:"sigtim"`

	// SigDat 签名数据（BASE64编码）
	SigDat string `json:"sigdat"`
}

// Request 请求报文
type Request struct {
	Request struct {
		Head RequestHead `json:"head"`
		Body interface{} `json:"body"`
	} `json:"request"`
	Signature Signature `json:"signature"`
}

// Response 响应报文
type Response struct {
	Response struct {
		Head ResponseHead `json:"head"`
		Body interface{}  `json:"body"`
	} `json:"response"`
	Signature Signature `json:"signature"`
}

// ======================== 账户交易查询接口（trsQryByBreakPoint） ========================

// TransQueryX1 账户交易查询单记录参数
type TransQueryX1 struct {
	// CardNbr 户口号
	CardNbr string `json:"cardNbr"`

	// BeginDate 开始日期（格式：YYYYMMDD）
	BeginDate string `json:"beginDate"`

	// EndDate 结束日期（格式：YYYYMMDD）
	EndDate string `json:"endDate"`

	// TransactionSequence 起始记账序号（可选，默认从第1笔开始）
	TransactionSequence string `json:"transactionSequence"`

	// CurrencyCode 币种（可选，建议传入）
	CurrencyCode string `json:"currencyCode"`

	// QueryAcctNbr 继续查询账号（首次查询无需传，续传时赋值响应报文Z1中的对应字段）
	QueryAcctNbr string `json:"queryAcctNbr"`

	// Reserve 保留字段
	Reserve string `json:"reserve"`

	// LoanCode 借贷码（C=贷方，D=借方）
	LoanCode string `json:"loanCode,omitempty"`
}

// TransQueryY1 账户交易查询多记录参数（续传用）
type TransQueryY1 struct {
	// AcctNbr 账号
	AcctNbr string `json:"acctNbr"`

	// TransDate 交易日期（当前查询最后一笔交易日期）
	TransDate string `json:"transDate"`

	// ExpectNextSequence 期望下一记账序号
	ExpectNextSequence string `json:"expectNextSequence"`
}

// TransQueryRequestBody 账户交易查询请求体
type TransQueryRequestBody struct {
	X1 []TransQueryX1 `json:"TRANSQUERYBYBREAKPOINT_X1"`
	Y1 []TransQueryY1 `json:"TRANSQUERYBYBREAKPOINT_Y1,omitempty"`
}

// TransQueryZ1 账户交易查询汇总信息
type TransQueryZ1 struct {
	// CtnFlag 未传完标记（Y=还有记录需查询，N=查询完毕）
	CtnFlag string `json:"ctnFlag"`

	// QueryAcctNbr 继续查询账号（ctnFlag为Y时，下一次请求需携带）
	QueryAcctNbr string `json:"queryAcctNbr"`

	// DebitNums 借方笔数
	DebitNums string `json:"debitNums"`

	// DebitAmount 借方金额
	DebitAmount string `json:"debitAmount"`

	// CreditNums 贷方笔数
	CreditNums string `json:"creditNums"`

	// CreditAmount 贷方金额
	CreditAmount string `json:"creditAmount"`

	// Reserve 保留字
	Reserve string `json:"reserve,omitempty"`
}

// TransQueryZ2 账户交易明细信息
type TransQueryZ2 struct {
	// TransDate 交易日
	TransDate string `json:"transDate"`

	// TransSequenceIdn 流水号
	TransSequenceIdn string `json:"transSequenceIdn"`

	// TransTime 交易时间
	TransTime string `json:"transTime,omitempty"`

	// ValueDate 起息日
	ValueDate string `json:"valueDate,omitempty"`

	// LoanCode 借贷码（C=贷方，D=借方）
	LoanCode string `json:"loanCode,omitempty"`

	// TransAmount 交易金额
	TransAmount string `json:"transAmount"`

	// CurrencyNbr 币种
	CurrencyNbr string `json:"currencyNbr"`

	// TextCode 交易类型
	TextCode string `json:"textCode,omitempty"`

	// BillNumber 票据号
	BillNumber string `json:"billNumber,omitempty"`

	// RemarkTextClt 你方摘要
	RemarkTextClt string `json:"remarkTextClt,omitempty"`

	// ReversalFlag 冲帐标志（*=冲帐，X=补帐）
	ReversalFlag string `json:"reversalFlag,omitempty"`

	// AcctOnlineBal 余额
	AcctOnlineBal string `json:"acctOnlineBal"`

	// ExtendedRemark 扩展摘要
	ExtendedRemark string `json:"extendedRemark,omitempty"`

	// CtpAcctNbr 收付方帐号
	CtpAcctNbr string `json:"ctpAcctNbr,omitempty"`

	// CtpAcctName 收付方名称
	CtpAcctName string `json:"ctpAcctName,omitempty"`

	// CtpBankName 收付方开户行行名
	CtpBankName string `json:"ctpBankName,omitempty"`

	// CtpBankAddress 收付方开户行地址
	CtpBankAddress string `json:"ctpBankAddress,omitempty"`

	// FatOrSonAccount 母子公司帐号
	FatOrSonAccount string `json:"fatOrSonAccount,omitempty"`

	// FatOrSonCompanyName 母子公司名称
	FatOrSonCompanyName string `json:"fatOrSonCompanyName,omitempty"`

	// FatOrSonBankName 母子公司开户行行名
	FatOrSonBankName string `json:"fatOrSonBankName,omitempty"`

	// FatOrSonBankAddress 母子公司开户行地址
	FatOrSonBankAddress string `json:"fatOrSonBankAddress,omitempty"`

	// InfoFlag 信息标志（空=付方帐号和子公司；1=收方帐号和子公司；2=收方帐号和母公司；3=原收方帐号和子公司）
	InfoFlag string `json:"infoFlag,omitempty"`

	// BusinessName 业务名称
	BusinessName string `json:"businessName,omitempty"`

	// BusinessText 网银业务摘要
	BusinessText string `json:"businessText,omitempty"`

	// RequestNbr 网银流程实例号
	RequestNbr string `json:"requestNbr,omitempty"`

	// YurRef 网银业务参考号
	YurRef string `json:"yurRef,omitempty"`

	// VirtualNbr 虚拟户编号
	VirtualNbr string `json:"virtualNbr,omitempty"`

	// MchOrderNbr 商务支付订单号
	MchOrderNbr string `json:"mchOrderNbr,omitempty"`

	// TransCardNbr 记账卡号
	TransCardNbr string `json:"transCardNbr,omitempty"`

	// Reserve 保留字
	Reserve string `json:"reserve,omitempty"`
}

// TransQueryResponseBody 账户交易查询响应体
type TransQueryResponseBody struct {
	Y1 []TransQueryY1 `json:"TRANSQUERYBYBREAKPOINT_Y1"`
	Z1 []TransQueryZ1 `json:"TRANSQUERYBYBREAKPOINT_Z1"`
	Z2 []TransQueryZ2 `json:"TRANSQUERYBYBREAKPOINT_Z2"`
}

// ======================== 单笔回单查询接口（DCSIGREC） ========================

// ReceiptQueryRequestBody 单笔回单查询请求体
type ReceiptQueryRequestBody struct {
	// EacNbr 账号
	EacNbr string `json:"eacnbr"`

	// QueDat 查询日期（yyyy-MM-dd格式）
	QueDat string `json:"quedat"`

	// TrsSeq 交易流水号
	TrsSeq string `json:"trsseq"`

	// PriMod 打印模式（空时默认PDF，可选PDF或者OFDEX）
	PriMod string `json:"primod,omitempty"`
}

// ReceiptQueryResponseBody 单笔回单查询响应体
type ReceiptQueryResponseBody struct {
	// CheCod 验证码
	CheCod string `json:"checod,omitempty"`

	// FilDat 单笔回单返回的数据流（Base64编码的PDF或OFD文件）
	FilDat string `json:"fildat,omitempty"`

	// IstNbr 回单实例号
	IstNbr string `json:"istnbr,omitempty"`
}

// ======================== 财务变动通知接口（YQN01010） ========================

// NotificationMessage 通知消息
type NotificationMessage struct {
	// MsgTyp 通知子类型（NCCRTTRS=到账通知，NCDBTTRS=付款通知）
	MsgTyp string `json:"msgtyp"`

	// MsgDat 通知内容
	MsgDat NotificationData `json:"msgdat"`
}

// NotificationData 通知内容数据
type NotificationData struct {
	// AccNbr 账号
	AccNbr string `json:"accnbr,omitempty"`

	// AccNam 户名
	AccNam string `json:"accnam,omitempty"`

	// TrsDat 交易日期
	TrsDat string `json:"trsdat"`

	// TrsTim 交易时间
	TrsTim string `json:"trstim"`

	// CCcyNbr 币种
	CCcyNbr string `json:"c_ccynbr,omitempty"`

	// CTrsAmt 交易金额
	CTrsAmt string `json:"c_trsamt"`

	// RpyAcc 收/付方帐户名称
	RpyAcc string `json:"rpyacc,omitempty"`

	// RpyNam 收/付方的转入或转出帐号
	RpyNam string `json:"rpynam,omitempty"`

	// BlvAmt 帐户的联机余额
	BlvAmt string `json:"blvamt,omitempty"`

	// RefNbr 流水号
	RefNbr string `json:"refnbr"`

	// VltDat 起息日
	VltDat string `json:"vltdat,omitempty"`

	// TrsCod 交易类型
	TrsCod string `json:"trscod,omitempty"`

	// NarYur 摘要
	NarYur string `json:"naryur,omitempty"`

	// AmtCdr 借贷标记（C=贷，D=借）
	AmtCdr string `json:"amtcdr"`

	// ReqNbr 流程实例号
	ReqNbr string `json:"reqnbr,omitempty"`

	// BusNam 业务名称
	BusNam string `json:"busnam,omitempty"`

	// NUsage 用途
	NUsage string `json:"nusage,omitempty"`

	// YurRef 业务参考号
	YurRef string `json:"yurref,omitempty"`

	// BusNar 业务摘要
	BusNar string `json:"busnar,omitempty"`

	// OtrNar 其它摘要
	OtrNar string `json:"otrnar,omitempty"`

	// RpyBbk 收/付方开户地区分行号
	RpyBbk string `json:"rpybbk,omitempty"`

	// RpyBbn 收/付方开户行行号
	RpyBbn string `json:"rpybbn,omitempty"`

	// RpyBnk 收/付方开户行名
	RpyBnk string `json:"rpybnk,omitempty"`

	// RpyAdr 收/付方开户行地址
	RpyAdr string `json:"rpyadr,omitempty"`

	// GsbBbk 母/子公司所在地区分行
	GsbBbk string `json:"gsbbbk,omitempty"`

	// GsbAcc 母/子公司帐号
	GsbAcc string `json:"gsbacc,omitempty"`

	// GsbNam 母/子公司名称
	GsbNam string `json:"gsbnam,omitempty"`

	// InfFlg 信息标志（空=付方帐号和子公司；1=收方帐号和子公司；2=收方帐号和母公司；3=原收方帐号和子公司）
	InfFlg string `json:"infflg,omitempty"`

	// AthFlg 有否附件信息标志（Y=是，N=否）
	AthFlg string `json:"athflg,omitempty"`

	// ChkNbr 票据号
	ChkNbr string `json:"chknbr,omitempty"`

	// RsvFlg 冲帐标志（*=冲帐，X=补帐）
	RsvFlg string `json:"rsvflg,omitempty"`

	// NarExt 扩展摘要
	NarExt string `json:"narext,omitempty"`

	// TrsAnl 交易分析码
	TrsAnl string `json:"trsanl,omitempty"`

	// RefSub 商务支付订单号
	RefSub string `json:"refsub,omitempty"`

	// FrmCod 企业识别码
	FrmCod string `json:"frmcod,omitempty"`
}
