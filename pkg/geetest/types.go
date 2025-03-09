package geetest

type Ticket struct {
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
}

type Response struct {
	// 失败时返回
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	// 成功时返回
	Result      string         `json:"result"`
	Reason      string         `json:"reason"`
	CaptchaArgs map[string]any `json:"captcha_args"`
}
