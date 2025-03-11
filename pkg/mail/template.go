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
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>验证码</title>
    <style>
      body {
        color: #333;
        line-height: 1.6;
        margin: 0;
        padding: 0;
        background-color: #f0f0f0;
      }
      .container {
        max-width: 600px;
        margin: 0 auto;
        padding: 0;
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
        border-radius: 0;
        box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
        margin: 0 60px;
      }
      .title {
        font-size: 24px;
        font-weight: bold;
        margin-bottom: 20px;
      }
      .code-container {
        background-color: #f5f7fa;
        border-radius: 4px;
        padding: 10px;
        margin: 20px 0;
        text-align: center;
      }
      .verification-code {
        font-size: 36px;
        font-weight: bold;
        letter-spacing: 8px;
        margin: 0;
      }
      .footer {
        text-align: center;
        font-size: 12px;
        color: #888;
        margin-top: 20px;
        padding: 0 60px;
      }
      strong {
        font-weight: bold;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="header">
        <div class="text-logo">{{.Company}}</div>
      </div>
      <div class="content">
        <div class="title">验证您的邮箱以继续</div>
        <p>请使用以下验证码完成操作：</p>
        <div class="code-container">
          <pre class="verification-code">{{.Code}}</pre>
        </div>
        <p>
          如果这不是您本人操作，请直接忽略此邮件。出于安全原因，验证码将在
          <strong>5 分钟</strong>后失效。
        </p>
        <p>此敬<br /><b>{{.Company}}</b></p>
      </div>
      <div class="footer">
        <p>此邮件由系统自动发送，请勿回复。</p>
        <p>© {{.Year}} {{.Company}} 保留所有权利。</p>
      </div>
    </div>
  </body>
</html>

`
