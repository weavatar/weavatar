package geetest

type Ticket struct {
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
}

type Response struct {
	// set when the request fails
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	// set when the request succeeds
	Result      string         `json:"result"`
	Reason      string         `json:"reason"`
	CaptchaArgs map[string]any `json:"captcha_args"`
}
