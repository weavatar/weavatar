package mail

import (
	"bytes"
	"html/template"
	"time"
)

func CodeTmpl(company, code string) string {
	tmpl, _ := template.New("code").Parse(codeTmpl)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]any{
		"Company": company,
		"Code":    code,
		"Year":    time.Now().Year(),
	})
	if err != nil {
		return ""
	}

	return buf.String()
}

var codeTmpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>验证码</title>
    <style>
        body {
            color: #333;
            line-height: 1.6;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            padding: 20px 0;
        }
        .text-logo {
            font-size: 28px;
            font-weight: bold;
            color: #2563eb;
            letter-spacing: 1px;
            margin: 0;
            text-decoration: none;
        }
        .content {
            background-color: #ffffff;
            padding: 30px;
            border-radius: 6px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
        }
        .code-container {
            background-color: #f5f7fa;
            border-radius: 4px;
            padding: 15px;
            margin: 20px 0;
            text-align: center;
        }
        .verification-code {
            font-size: 32px;
            font-weight: bold;
            letter-spacing: 5px;
            color: #2563eb;
            margin: 0;
        }
        .footer {
            text-align: center;
            font-size: 12px;
            color: #888;
            margin-top: 30px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="text-logo">{{.Company}}</div>
        </div>
        <div class="content">
            <h2>您好！</h2>
            <p>您正在进行身份验证，请使用以下验证码完成操作：</p>
            <div class="code-container">
                <p class="verification-code">{{.Code}}</p>
            </div>
            <p>验证码有效期为 <strong>5 分钟</strong>。如果这不是您本人的操作，请忽略此邮件。</p>
            <p>此敬<br><b>{{.Company}}</b></p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>© {{.Year}} {{.Company}}. 保留所有权利。</p>
        </div>
    </div>
</body>
</html>
`
