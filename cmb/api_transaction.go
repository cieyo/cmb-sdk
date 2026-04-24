package cmb

import "fmt"

// QueryAccountTransaction 账户交易查询（trsQryByBreakPoint）
// 用于查询账户的交易明细信息，支持断点续传
// reqID: 请求唯一标识（如果为空则自动生成）
// reqBody: 请求参数
// 返回: 响应体和错误
func (c *Client) QueryAccountTransaction(reqBody *TransQueryRequestBody, reqID string) (*TransQueryResponseBody, *ResponseHead, error) {
	if reqID == "" {
		reqID = GenerateReqID("TXN")
	}
	normalizeTransQueryRequest(reqBody)

	var respBody TransQueryResponseBody
	head, err := c.doRequest("trsQryByBreakPoint", reqID, reqBody, &respBody)
	if err != nil {
		return nil, head, err
	}

	return &respBody, head, nil
}

// QueryAccountTransactionAll 查询账户所有交易（自动处理续传）
// 会自动循环查询直到所有数据获取完毕
// cardNbr: 户口号
// beginDate: 开始日期（格式：YYYYMMDD）
// endDate: 结束日期（格式：YYYYMMDD）
// 返回: 所有交易明细和错误
func (c *Client) QueryAccountTransactionAll(cardNbr, beginDate, endDate string, _ ...string) ([]TransQueryZ2, error) {
	c.logger.Info("query all transactions start",
		"card_nbr", cardNbr,
		"begin_date", beginDate,
		"end_date", endDate,
	)

	var allTransactions []TransQueryZ2
	var y1Data []TransQueryY1
	var queryAcctNbr string
	seenStates := make(map[string]struct{})

	const maxPages = 200

	for page := 1; page <= maxPages; page++ {
		// 构建请求
		reqBody := &TransQueryRequestBody{
			X1: []TransQueryX1{
				{
					CardNbr:      cardNbr,
					BeginDate:    beginDate,
					EndDate:      endDate,
					CurrencyCode: "",
					QueryAcctNbr: queryAcctNbr,
				},
			},
			Y1: y1Data,
		}
		normalizeTransQueryRequest(reqBody)

		// 发送请求
		respBody, _, err := c.QueryAccountTransaction(reqBody, "")
		if err != nil {
			c.logger.Error("query transaction page failed", "page", page, "error", err)
			return nil, err
		}

		// 追加交易明细
		allTransactions = append(allTransactions, respBody.Z2...)

		c.logger.Debug("transaction page fetched",
			"page", page,
			"page_count", len(respBody.Z2),
			"total_count", len(allTransactions),
		)

		// 检查是否还有更多数据
		if len(respBody.Z1) == 0 {
			break
		}

		z1 := respBody.Z1[0]
		if z1.CtnFlag != "Y" {
			// 查询完毕
			break
		}

		if z1.QueryAcctNbr == "" {
			return nil, fmt.Errorf("%w: ctnFlag=Y but queryAcctNbr is empty", ErrPaginationLoop)
		}

		nextState := fmt.Sprintf("acct=%s;y1=%v", z1.QueryAcctNbr, respBody.Y1)
		if _, exists := seenStates[nextState]; exists {
			c.logger.Error("pagination loop detected", "page", page, "state", nextState)
			return nil, fmt.Errorf("%w: repeated continuation state on page %d", ErrPaginationLoop, page)
		}
		seenStates[nextState] = struct{}{}

		// 准备下一次续传查询
		queryAcctNbr = z1.QueryAcctNbr
		y1Data = respBody.Y1
	}

	if len(seenStates) >= maxPages {
		c.logger.Error("exceeded max pages", "max_pages", maxPages)
		return nil, fmt.Errorf("%w: exceeded max pages (%d)", ErrPaginationLoop, maxPages)
	}

	c.logger.Info("query all transactions completed",
		"card_nbr", cardNbr,
		"total_transactions", len(allTransactions),
		"total_pages", len(seenStates)+1,
	)

	return allTransactions, nil
}

func normalizeTransQueryRequest(reqBody *TransQueryRequestBody) {
	if reqBody == nil {
		return
	}

	for i := range reqBody.X1 {
		if reqBody.X1[i].TransactionSequence == "" {
			reqBody.X1[i].TransactionSequence = "1"
		}
	}
}
