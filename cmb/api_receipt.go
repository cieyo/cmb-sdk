package cmb

import (
	"encoding/base64"
	"fmt"
	"os"
)

// QuerySingleReceipt 单笔回单查询（DCSIGREC）
// 用于根据流水号查询单笔交易回单
// reqID: 请求唯一标识（如果为空则自动生成）
// reqBody: 请求参数
// 返回: 响应体和错误
func (c *Client) QuerySingleReceipt(reqBody *ReceiptQueryRequestBody, reqID string) (*ReceiptQueryResponseBody, *ResponseHead, error) {
	if reqID == "" {
		reqID = GenerateReqID("RCP")
	}

	var respBody ReceiptQueryResponseBody
	head, err := c.doRequest("DCSIGREC", reqID, reqBody, &respBody)
	if err != nil {
		return nil, head, err
	}

	return &respBody, head, nil
}

// DownloadReceipt 下载回单并保存到文件
// eacNbr: 账号
// queDat: 查询日期（格式：yyyy-MM-dd）
// trsSeq: 交易流水号
// outputPath: 输出文件路径（根据priMod决定扩展名.pdf或.ofd）
// priMod: 打印模式（PDF或OFDEX，默认PDF）
// 返回: 验证码和错误
func (c *Client) DownloadReceipt(eacNbr, queDat, trsSeq, outputPath, priMod string) (string, error) {
	if priMod == "" {
		priMod = "PDF"
	}

	// 查询回单
	reqBody := &ReceiptQueryRequestBody{
		EacNbr: eacNbr,
		QueDat: queDat,
		TrsSeq: trsSeq,
		PriMod: priMod,
	}

	respBody, _, err := c.QuerySingleReceipt(reqBody, "")
	if err != nil {
		return "", err
	}

	// 解码BASE64数据
	fileData, err := base64.StdEncoding.DecodeString(respBody.FilDat)
	if err != nil {
		return "", fmt.Errorf("decode receipt data failed: %w", err)
	}

	// 保存文件
	err = os.WriteFile(outputPath, fileData, 0644)
	if err != nil {
		return "", fmt.Errorf("save receipt file failed: %w", err)
	}

	return respBody.CheCod, nil
}
