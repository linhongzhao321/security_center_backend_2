package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
)

func ExampleSendMessage() {
	InitSES(context.Background(), os.Getenv(`SES_REGION`), os.Getenv(`AWS_ACCESS_KEY`), os.Getenv(`AWS_ACCESS_SECRET`), os.Getenv(`SENDER_EMAIL`))
	htmlContent := `
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.2/dist/css/bootstrap.min.css" integrity="sha384-xOolHFLEh07PJGoPkLv1IbcEPTNtaed2xpHsD9ESMhqIYd0nLMwNLD69Npy4HI+N" crossorigin="anonymous">
    <title>GitHub SSH Certificate</title>
  </head>
  <body>
	<div class="container-fluid">
		<div class="row-fluid">
			<div class="span12">
				<ol>
					<li>
						请下载证书文件到您的计算机，证书文件名与您所使用的公钥文件前缀一致
						<ul>
							<li>
								例如：
							</li>
							<li>
								若你的公钥文件为 ~/.ssh/id_rsa.pub
							</li>
							<li>
								则你的证书文件名应为 ~/.ssh/id_rsa-cert.pub
							</li>
						</ul>
					</li>
					<li>
						如果您的证书还未过期，可以忽略
					</li>
					<li>
						如果您短时间内收到多个证书，一般情况下，任意证书都可以使用
					</li>
					<li>
						如果您使用开发机，请提供您的开发机 IP 给管理员以获取将开发机 IP 加入白名单的 SSH 证书
					</li>
					<li>
						该证书有效期为 14 天. 请在该证书过期前及时更新下一次证书 —— 新证书将定期通过邮箱发放
					</li>
					<li>
						如果你同时在参与多个存储库的开发，且存在存储库未迁移，仍在 viabtc 下<br>
						请参阅 <a href="https://docs.google.com/document/d/1CPvAtpE6BHmqzHlzOCaFsq0-8eNcpjLg9sBUuDVp9f0/edit?usp=sharing">迁移指引(精简版)</a> 配置你的 ssh/config，使不同 repo 用不同的认证方式
					</li>
					<li>
						如果你需要其他帮助，请直接 slack 联系  <a href="https://vino-tech.slack.com/team/U05V8BXHDK9">@funco.lin</a>
					</li>
				</ol>
			</div>
		</div>
	</div>
  </body>
</html>
`
	attachment := []byte(`test file`)
	err := SendMessage(`GitHub SSH Certificate`, []string{`funco.lin@vinotech.com`}, htmlContent, []string{base64.StdEncoding.EncodeToString(attachment), base64.StdEncoding.EncodeToString(attachment)}, []string{`test-cert.pub`})
	fmt.Println(err == nil)
	// Output:
	// True
}
