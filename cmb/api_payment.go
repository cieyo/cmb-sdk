package cmb

// PaymentOperate 企银支付单笔经办（BB1PAYOP）
// 用于发起单笔支付经办
// reqBody: 请求参数
// reqID: 请求唯一标识（如果为空则自动生成）
// 返回: 响应体、响应头和错误
func (c *Client) PaymentOperate(reqBody *PaymentOperateRequestBody, reqID string) (*PaymentOperateResponseBody, *ResponseHead, error) {
	if reqID == "" {
		reqID = GenerateReqID("PAY")
	}
	
	var respBody PaymentOperateResponseBody
	head, err := c.doRequest("BB1PAYOP", reqID, reqBody, &respBody)
	if err != nil {
		return nil, head, err
	}
	
	return &respBody, head, nil
}

// PaymentQuery 企银支付业务查询（BB1PAYQR）
// 用于支付业务提交之后，按照“业务参考号”对支付业务处理结果的查询
// reqBody: 请求参数
// reqID: 请求唯一标识
// 返回: 响应体、响应头和错误
func (c *Client) PaymentQuery(reqBody *PaymentQueryRequestBody, reqID string) (*PaymentQueryResponseBody, *ResponseHead, error) {
	if reqID == "" {
		reqID = GenerateReqID("PAYQRY")
	}
	
	var respBody PaymentQueryResponseBody
	head, err := c.doRequest("BB1PAYQR", reqID, reqBody, &respBody)
	if err != nil {
		return nil, head, err
	}
	
	return &respBody, head, nil
}
