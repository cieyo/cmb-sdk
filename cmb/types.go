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

// ======================== 企银支付单笔经办接口（BB1PAYOP） ========================

// PayBusinessMode 调用企业信息（bb1paybmx1）
type PayBusinessMode struct {
	// BusMod 业务模式（模式编号），可通过"可经办业务模式查询(DCLISMOD)"接口获得
	BusMod string `json:"busMod"`

	// BusCod 业务类型（业务代码），N02030: 支付
	BusCod string `json:"busCod"`
}

// PaymentDetail 支付明细信息（bb1payopx1）
type PaymentDetail struct {
	// DbtAcc 转出帐号（必输）
	DbtAcc string `json:"dbtAcc"`

	// DmaNbr 记账子单元编号（目前支持1-10位）
	DmaNbr string `json:"dmaNbr,omitempty"`

	// CrtAcc 收方帐号（必输）
	CrtAcc string `json:"crtAcc"`

	// CrtNam 收方户名（必输，最长100汉字）
	CrtNam string `json:"crtNam"`

	// CrtBnk 收方开户行名称（最长30汉字）
	CrtBnk string `json:"crtBnk,omitempty"`

	// CrtAdr 收方开户行地址（最长30汉字）
	CrtAdr string `json:"crtAdr,omitempty"`

	// BrdNbr 收方行联行号
	BrdNbr string `json:"brdNbr,omitempty"`

	// CcyNbr 币种（只支持10人民币）
	CcyNbr string `json:"ccyNbr"`

	// TrsAmt 交易金额
	TrsAmt string `json:"trsAmt"`

	// BnkFlg 系统内标志（收方为招行户：传空或Y；收方为他行户：传N）
	BnkFlg string `json:"bnkFlg,omitempty"`

	// EptDat 期望日
	EptDat string `json:"eptDat,omitempty"`

	// EptTim 期望时间
	EptTim string `json:"eptTim,omitempty"`

	// StlChn 结算通道（G: 普通；Q: 快速；R: 实时-超网；I: 智能路由，默认Q）
	StlChn string `json:"stlChn,omitempty"`

	// NusAge 用途（展示在回单，最长100汉字，必输）
	NusAge string `json:"nusAge"`

	// CrtSqn 收方编号
	CrtSqn string `json:"crtSqn,omitempty"`

	// YurRef 业务参考号（必须唯一，必输）
	YurRef string `json:"yurRef"`

	// BusNar 业务摘要（最长100汉字）
	BusNar string `json:"busNar,omitempty"`

	// NtfCh1 通知方式一（邮箱）
	NtfCh1 string `json:"ntfCh1,omitempty"`

	// NtfCh2 通知方式二（手机号）
	NtfCh2 string `json:"ntfCh2,omitempty"`

	// TrsTyp 业务种类（100001: 普通汇兑（默认值）；101001: 慈善捐款；101002: 其他）
	TrsTyp string `json:"trsTyp,omitempty"`

	// RcvChk 行内收方账号户名校验（1：校验；空或其他值：不校验）
	RcvChk string `json:"rcvChk,omitempty"`

	// DrpFlg 直汇普通标志（A-普通；B-直汇（失败后不落人工处理））
	DrpFlg string `json:"drpFlg,omitempty"`
}

// PayCoupon 优惠券信息（bb1payopx5）
type PayCoupon struct {
	// CopNbr 优惠券编号
	CopNbr string `json:"copNbr,omitempty"`
}

// PaymentOperateRequestBody 企银支付单笔经办请求体
type PaymentOperateRequestBody struct {
	// BusinessMode 模式信息（长度为1）
	BusinessMode []PayBusinessMode `json:"bb1paybmx1"`

	// PaymentDetails 支付信息（长度为1）
	PaymentDetails []PaymentDetail `json:"bb1payopx1"`

	// Coupons 优惠券信息（可选）
	Coupons []PayCoupon `json:"bb1payopx5,omitempty"`
}

// PaymentOperateResult 企银支付单笔经办响应详情（bb1payopz1）
type PaymentOperateResult struct {
	// ReqNbr 流程实例号
	ReqNbr string `json:"reqNbr,omitempty"`

	// EvtIst 事件实例号
	EvtIst string `json:"evtIst,omitempty"`

	// ReqSts 请求状态（AUT: 等待审批；NTE: 终审完毕；BNK/WRF: 银行处理中；FIN: 完成；OPR: 数据接收中）
	ReqSts string `json:"reqSts"`

	// RtnFlg 业务处理结果（reqSts='FIN'时有意义：S-成功；F-失败；B-退票；R-否决；D-过期；C-撤消；U-银行挂账）
	RtnFlg string `json:"rtnFlg,omitempty"`

	// OprSqn 待处理操作序列
	OprSqn string `json:"oprSqn,omitempty"`

	// OprAls 操作别名
	OprAls string `json:"oprAls,omitempty"`

	// ErrCod 错误码（SUC0000：成功）
	ErrCod string `json:"errCod,omitempty"`

	// ErrTxt 错误文本
	ErrTxt string `json:"errTxt,omitempty"`

	// MsgTxt 提示文本
	MsgTxt string `json:"msgTxt,omitempty"`

	// PrdIst 产品实例号
	PrdIst string `json:"prdIst,omitempty"`

	// BakAppNbr 流程中台流水号
	BakAppNbr string `json:"bakAppNbr,omitempty"`
}

// PaymentApprover 下一级审批人信息（bb1payopz3）
type PaymentApprover struct {
	// UsrNbr 统一用户号
	UsrNbr string `json:"usrNbr,omitempty"`

	// UsrNam 用户姓名（下一级审批人姓名）
	UsrNam string `json:"usrNam,omitempty"`

	// UsrLgn 登录名（下一级审批人登录名）
	UsrLgn string `json:"usrLgn,omitempty"`

	// UsrId 用户编号（下一级审批人用户编号）
	UsrId string `json:"usrId,omitempty"`

	// Rsv50z 保留字50
	Rsv50z string `json:"rsv50z,omitempty"`
}

// PaymentOperateResponseBody 企银支付单笔经办响应体
type PaymentOperateResponseBody struct {
	// Results 经办结果详情（bb1payopz1）
	Results []PaymentOperateResult `json:"bb1payopz1"`

	// Approvers 下一级审批人信息（bb1payopz3）
	Approvers []PaymentApprover `json:"bb1payopz3,omitempty"`
}

// IsFinished 判断支付是否已到终态
func (r *PaymentOperateResult) IsFinished() bool {
	return r.ReqSts == "FIN"
}

// IsSuccess 判断支付是否成功（仅在 IsFinished() 为 true 时有意义）
func (r *PaymentOperateResult) IsSuccess() bool {
	return r.ReqSts == "FIN" && r.RtnFlg == "S"
}

// IsFailed 判断支付是否失败（包括失败、退票、否决、过期、撤消）
func (r *PaymentOperateResult) IsFailed() bool {
	if r.ReqSts != "FIN" {
		return false
	}
	switch r.RtnFlg {
	case "F", "B", "R", "D", "C":
		return true
	default:
		return false
	}
}

// NeedApproval 判断是否需要审批
func (r *PaymentOperateResult) NeedApproval() bool {
	return r.ReqSts == "AUT"
}

// IsBankProcessing 判断是否银行处理中
func (r *PaymentOperateResult) IsBankProcessing() bool {
	return r.ReqSts == "BNK" || r.ReqSts == "WRF"
}
// ======================== 企银支付业务查询接口（BB1PAYQR） ========================

// PaymentQueryX1 支付业务查询条件（bb1payqrx1）
type PaymentQueryX1 struct {
	// BusCod 业务类型（固定值：N02030）
	BusCod string `json:"busCod"`

	// YurRef 业务参考号
	YurRef string `json:"yurRef"`
}

// PaymentQueryRequestBody 企银支付业务查询请求体
type PaymentQueryRequestBody struct {
	X1 []PaymentQueryX1 `json:"bb1payqrx1"`
}

// PaymentQueryResult 支付业务查询结果详情（bb1payqrz1）
type PaymentQueryResult struct {
	// ReqNbr 流程实例号
	ReqNbr string `json:"reqNbr"`

	// BusCod 业务编码
	BusCod string `json:"busCod"`

	// BusMod 业务模式
	BusMod string `json:"busMod"`

	// DbtBbk 转出分行号
	DbtBbk string `json:"dbtBbk"`

	// DbtAcc 付方帐号
	DbtAcc string `json:"dbtAcc"`

	// DmaNbr 付方记账子单元编号
	DmaNbr string `json:"dmaNbr"`

	// DbtNam 付方帐户名
	DbtNam string `json:"dbtNam"`

	// CrtBbk 收方分行号
	CrtBbk string `json:"crtBbk"`

	// CrtAcc 收方帐号
	CrtAcc string `json:"crtAcc"`

	// CrtNam 收方名称
	CrtNam string `json:"crtNam"`

	// CrtBnk 收方行名称
	CrtBnk string `json:"crtBnk"`

	// CrtAdr 收方行地址
	CrtAdr string `json:"crtAdr"`

	// CcyNbr 币种
	CcyNbr string `json:"ccyNbr"`

	// TrsAmt 交易金额
	TrsAmt string `json:"trsAmt"`

	// EptDat 期望日
	EptDat string `json:"eptDat"`

	// EptTim 期望时间
	EptTim string `json:"eptTim"`

	// BnkFlg 系统内外标志
	BnkFlg string `json:"bnkFlg"`

	// StlChn 结算通路
	StlChn string `json:"stlChn"`

	// NusAge 用途
	NusAge string `json:"nusAge"`

	// NtfCh1 通知方式一
	NtfCh1 string `json:"ntfCh1"`

	// NtfCh2 通知方式二
	NtfCh2 string `json:"ntfCh2"`

	// OprDat 经办日期
	OprDat string `json:"oprDat"`

	// YurRef 参考号
	YurRef string `json:"yurRef"`

	// BusNar 业务摘要
	BusNar string `json:"busNar"`

	// ReqSts 请求状态
	ReqSts string `json:"reqSts"`

	// RtnFlg 业务处理结果
	RtnFlg string `json:"rtnFlg"`

	// OprSqn 待处理操作序列
	OprSqn string `json:"oprSqn"`

	// OprAls 操作别名
	OprAls string `json:"oprAls"`

	// LgnNam 用户名
	LgnNam string `json:"lgnNam"`

	// UsrNam 用户姓名
	UsrNam string `json:"usrNam"`

	// RtnNar 失败原因
	RtnNar string `json:"rtnNar"`

	// AthFlg 是否有附件信息
	AthFlg string `json:"athFlg"`

	// RcvBrd 收方大额行号
	RcvBrd string `json:"rcvBrd"`

	// TrsTyp 业务种类
	TrsTyp string `json:"trsTyp"`

	// TrxSet 账务套号
	TrxSet string `json:"trxSet"`

	// TrxSeq 账务流水
	TrxSeq string `json:"trxSeq"`

	// TrxSeqBackward 退票的记账流水号
	TrxSeqBackward string `json:"trxSeqBackward"`

	// BackwardDay 退票日期
	BackwardDay string `json:"backwardDay"`
}

// PaymentQueryResponseBody 企银支付业务查询响应体
type PaymentQueryResponseBody struct {
	Z1 []PaymentQueryResult `json:"bb1payqrz1"`
}

// IsFinished 判断支付是否已到终态
func (r *PaymentQueryResult) IsFinished() bool {
	return r.ReqSts == "FIN"
}

// IsSuccess 判断支付是否成功（仅在 IsFinished() 为 true 时有意义）
func (r *PaymentQueryResult) IsSuccess() bool {
	return r.ReqSts == "FIN" && r.RtnFlg == "S"
}

// IsFailed 判断支付是否失败（包括失败、退票、否决、过期、撤消）
func (r *PaymentQueryResult) IsFailed() bool {
	if r.ReqSts != "FIN" {
		return false
	}
	switch r.RtnFlg {
	case "F", "B", "R", "D", "C":
		return true
	default:
		return false
	}
}

// ======================== 企银支付单笔经办接口（BB1PAYOP） ========================

// ======================== 财务变动通知接口（YQN01010） ========================

// NotificationMessage 通知消息（通用结构）
type NotificationMessage struct {
	// MsgTyp 通知子类型
	// 财务变动通知：NCCRTTRS=到账通知，NCDBTTRS=付款通知
	// 支付结果通知：FINS=业务完成通知，FINB=支付退票通知
	MsgTyp string `json:"msgtyp"`

	// MsgDat 通知内容
	MsgDat NotificationData `json:"msgdat"`
}

// NotificationData 通知内容数据（通用结构，包含财务变动和支付结果字段）
type NotificationData struct {
	// --- 支付结果通知专用字段 (YQN02030) ---

	// TrsInfo 业务完成通知（FINS）
	TrsInfo *PaymentQueryResult `json:"trsInfo,omitempty"`

	// BackInfo 支付退票通知（FINB）
	BackInfo *PaymentBackDetail `json:"backInfo,omitempty"`

	// --- 财务变动通知专用字段 (YQN01010) ---

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

// PaymentBackDetail 支付退票明细（backInfo）
type PaymentBackDetail struct {
	// ReqNbr 流程实例号
	ReqNbr string `json:"reqNbr"`

	// YurRef 业务参考号
	YurRef string `json:"yurRef"`

	// BusNbr 汇款编号
	BusNbr string `json:"busNbr"`

	// OutTyp 汇款方式
	OutTyp string `json:"outTyp"`

	// BusTyp 转账汇款种类
	BusTyp string `json:"busTyp"`

	// BusLvl 汇款优先级
	BusLvl string `json:"busLvl"`

	// BusSts 汇款业务状态
	BusSts string `json:"busSts"`

	// SndClt 付方客户号
	SndClt string `json:"sndClt"`

	// ClrSts 清算状态
	ClrSts string `json:"clrSts"`

	// IsuCnl 汇款发起通道
	IsuCnl string `json:"isuCnl"`

	// IsuDat 发起日期
	IsuDat string `json:"isuDat"`

	// TrsBbk 处理分行
	TrsBbk string `json:"trsBbk"`

	// TrsBrn 处理机构
	TrsBrn string `json:"trsBrn"`

	// CcyNbr 交易货币
	CcyNbr string `json:"ccyNbr"`

	// TrsAmt 金额
	TrsAmt interface{} `json:"trsAmt"` // 可能是 string 或 float64，视银行返回而定，示例中有 888.88 数字

	// DbtAcc 付方户口号
	DbtAcc string `json:"dbtAcc"`

	// DbtNam 付方户名
	DbtNam string `json:"dbtNam"`

	// SndBrn 付方开户机构
	SndBrn string `json:"sndBrn"`

	// CrtAcc 收方户口号
	CrtAcc string `json:"crtAcc"`

	// CrtNam 收方户名
	CrtNam string `json:"crtNam"`

	// CrtBnk 收方开户行
	CrtBnk string `json:"crtBnk"`

	// RcvEaa 收方开户地
	RcvEaa string `json:"rcvEaa"`

	// NarTxt 摘要
	NarTxt string `json:"narTxt"`

	// FeeAmt 费用总额
	FeeAmt interface{} `json:"feeAmt"`

	// FeeCcy 币种
	FeeCcy string `json:"feeCcy"`

	// PsbTyp 提出凭证种类
	PsbTyp string `json:"psbTyp"`

	// PsbNbr 凭证号码
	PsbNbr string `json:"psbNbr"`

	// CtyFlg 同城异地标志
	CtyFlg string `json:"ctyFlg"`

	// SysFlg 系统内外标志
	SysFlg string `json:"sysFlg"`

	// RcvTyp 收方公私标志
	RcvTyp string `json:"rcvTyp"`

	// WatRcn 资金停留原因
	WatRcn string `json:"watRcn"`

	// WatTrs 资金停留流水
	WatTrs string `json:"watTrs"`

	// UpdDat 更新日期
	UpdDat string `json:"updDat"`

	// RtnCod 退票理由代码
	RtnCod string `json:"rtnCod"`

	// RtnNar 退票原因
	RtnNar string `json:"rtnNar"`

	// RcdSts 记录状态
	RcdSts string `json:"rcdSts"`
}
