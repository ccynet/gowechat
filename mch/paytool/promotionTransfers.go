package paytool

import (
	"errors"
	"fmt"
	"github.com/yaotian/gowechat/mch/base"
	"github.com/yaotian/gowechat/util"
)

//官方文档： https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2

//TransfersInput 发红包的配置
type TransfersInput struct {
	ToOpenID string //接红包的OpenID
	Amount int    //分为单位
	PartnerTradeNo string //商户订单号
	Desc   string //注 String(256)
	SpbillCreateIp string //该IP同在商户平台设置的IP白名单中的IP没有关联，该IP可传用户端或者服务端的IP

	//非必填，但大于200元，此必填, 有8个选项可供选择
	SceneID string
}

//Check check input
func (m *TransfersInput) checkTransfers() (isGood bool, err error) {
	if m.ToOpenID == "" || m.Amount == 0 || m.Desc == "" || m.SpbillCreateIp == "" || m.PartnerTradeNo == "" {
		err = fmt.Errorf("%s", "Input有必填项没有值")
		return
	}

	if m.Amount >= 200*100 && m.SceneID == "" {
		err = fmt.Errorf("%s", "大于200元的红包，必须设置SceneID")
		return
	}
	return true, nil
}

//SendRedPack 发红包
func (c *PayTool) SendTransfers(input TransfersInput) (isSuccess bool, err error) {
	if isGood, err := input.checkTransfers(); !isGood {
		return false, err
	}

	var signMap = make(map[string]string)
	signMap["mch_appid"] = c.AppID
	signMap["mchid"] = c.MchID
	signMap["nonce_str"] = util.RandomStr(5)
	signMap["device_info"] = "WEB"
	signMap["partner_trade_no"] = input.PartnerTradeNo
	signMap["openid"] = input.ToOpenID
	signMap["check_name"] = "NO_CHECK"
	signMap["amount"] = util.ToStr(input.Amount)
	signMap["desc"] = input.Desc
	signMap["spbill_create_ip"] = input.SpbillCreateIp
	signMap["sign"] = base.Sign(signMap, c.MchAPIKey, nil)
	//fmt.Println(signMap)

	respMap, err := c.SendTransfersRaw(signMap)
	if err != nil {
		return false, err
	}

	resultCode, ok := respMap["result_code"]
	if !ok {
		err = errors.New("no result_code")
		return false, err
	}

	if resultCode != "SUCCESS" {
		returnMsg, _ := respMap["return_msg"]
		errMsg, _ := respMap["err_code_des"]
		errCode, _ := respMap["err_code"]

		if errCode == "NOTENOUGH" {
			return false, errors.New("No enough money")
		}

		err = fmt.Errorf("Err:%s return_msg:%s err_code:%s err_code_des:%s", "result code is not success", returnMsg, errCode, errMsg)
		return false, err
	}

	return true, nil
}
