package ssh

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/gomail.v2"

	"infrastructure/aws"
)

type Sender func(recipientEmail string, caFilenames ...string) error

func getGomailSender(sender, password string) Sender {
	return func(recipientEmail string, caFilenames ...string) error {
		content := `
1. 请下载证书文件到你本地
2. 请重命名文件，使证书文件名与你公钥前缀一致，如：你的公钥文件名为 id_rsa.pub，则你的证书文件名应为 id_rsa-cert.pub
3. 如果你使用开发机，开发机 IP 默认不在 SSH CA 的 IP 白名单中，请提供您的开发机 IP 给管理员以获取新的 SSH CA 证书
4. 该证书有效期为 14 天. 请在该证书过期前及时更新下一次证书
5. 如果你需要其他帮助，请直接 slack 联系管理员
`

		m := gomail.NewMessage()
		m.SetHeader(`From`, sender)
		m.SetHeader(`To`, recipientEmail)
		m.SetHeader(`Subject`, `Please rewrite your SSH CA certificate`)
		m.SetBody(`text/plain`, content)
		for _, caFilename := range caFilenames {
			m.Attach(caFilename)
		}

		dialer := gomail.NewDialer(`smtp.gmail.com`, 587, sender, password)
		return dialer.DialAndSend(m)
	}
}

func getSesSender(ctx context.Context, region, key, secret, sender string) (Sender, error) {
	err := aws.InitSES(ctx, region, key, secret, sender)
	return sesSender, err
}

const caEmailContent = `
1. 附件为您的证书文件，请下载证书文件到您的计算机，证书文件名与您所使用的公钥文件前缀一致
      例如：
          若你的公钥文件为 ~/.ssh/id_rsa.pub
          则你的证书文件名应为 ~/.ssh/id_rsa-cert.pub
2. 如果您的证书还未过期，可以忽略
3. 如果您短时间内收到多个证书，一般情况下，任意证书都可以使用
4. 如果您使用开发机，请提供您的开发机 IP 给管理员以获取将开发机 IP 加入白名单的 SSH 证书
5. 该证书有效期为 14 天. 请在该证书过期前及时更新下一次证书 —— 新证书将定期通过邮箱发放
6. 如果你需要其他帮助，请直接 slack 联系 @funco.lin
`

func sesSender(recipientEmail string, caFilenames ...string) error {
	var filenames []string
	var b64Attachments []string
	for _, filename := range caFilenames {
		file, err := os.Open(filename)
		if err != nil {
			return nil
		}
		filenames = append(filenames, filepath.Base(file.Name()))
		attachment, err := io.ReadAll(file)
		if err != nil {
			return nil
		}
		b64Attachments = append(b64Attachments, base64.StdEncoding.EncodeToString(attachment))

	}
	return aws.SendMessage(`GitHub SSH Certificate`, []string{recipientEmail}, caEmailContent, b64Attachments, filenames)
}
