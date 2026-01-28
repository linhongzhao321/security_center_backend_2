package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/aws"
	"infrastructure/logger"
	"infrastructure/sshkeygen"
)

func GithubCA() *cobra.Command {
	var ttl int
	var senderEmail, ca, userKeysDir, ipAllowList, validPeriod string
	var senderPassword, awsSesKey, awsSesSecret, awsSesRegion string
	var sender Sender

	command := &cobra.Command{
		Use:   `github-ca`,
		Short: `generate github ssh ca user certificate`,
		Long: `
执行步骤：
  1. 在 resources/members 下以成员的github username 创建目录
     如 ./resources/members/funcolin ./resources/members/weil1024
  2. 将成员的公钥放至上一步创建的成员目录下，公钥文件名可以包含多个，但公钥文件名必须符合正则 ^[a-zA-Z0-9_]+\.pub$，程序将自动为每一份公钥生成一份证书并合并邮箱发送
     如 ./resources/members/funcolin/user_key.pub ./resources/members/weil1024/user_key.pub
  3. 在 resources/githubca 下创建 ca 秘钥对: 
     ssh-keygen -t rsa -b 4096 -f resources/githubca/user_ca -C user_ca
  4. 执行命令

说明：
  1. 如果提供 aws-kms 相关参数，则程序将使用 kms 对传入的 --email-password 参数进行解密
  2. 如果使用的是企业 gmail 邮箱，不能使用真实邮箱密码，请前往 “Google个人账号页面 - 安全性 - 两步验证 - 应用专用密码” 功能页面，申请应用专用密码
`,
		Example: `ssh github-ca --ca <ca filename>` +
			`--user-keys-dir resources/members --sender-email 发件人邮箱` +
			`--email-password 发件人密码 --valid-period +14d` +
			`--ip-allow-list <ipv4 白名单，用","分隔> funcolin:funco.lin@vinotech.com  weil1024:weil.liu@vinotech.com`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if aws.IsInitialized() {
				err = aws.DecryptStrings(&senderPassword, &awsSesKey, &awsSesSecret)
				if err != nil {
					return err
				}
			}
			switch {
			case senderPassword != ``:
				sender = getGomailSender(senderEmail, senderPassword)
			case awsSesKey != ``:
				sender, err = getSesSender(cmd.Context(), awsSesRegion, awsSesKey, awsSesSecret, senderEmail)
				if err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, users []string) error {
			ips := strings.Split(ipAllowList, `,`)
			runner := getRunner(cmd.Context(), users, ips, ca, userKeysDir, validPeriod, sender)
			if ttl <= 0 {
				return runner()
			}

			for {
				if err := runner(); err != nil {
					return err
				}
				time.Sleep(time.Duration(ttl) * time.Hour * 24)
			}
		},
	}

	command.Flags().IntVarP(&ttl, `ttl`, `t`, 0, `time units day`)
	command.Flags().StringVar(&ipAllowList, `ip-allow-list`, ``, `which IPs can certificate be used by`)
	command.Flags().StringVar(&validPeriod, `valid-period`, ``, `such as: "+14d"`)
	command.Flags().StringVar(&userKeysDir, `user-keys-dir`, `./`, "such as: \ngen-ca --config-dir ./members funcolin:funco.lin@vinotech.com\n\tThe program will get public-key from ./members/funcolin/user_key.pub\n\n")
	command.Flags().StringVar(&senderEmail, `sender-email`, ``, `if empty, certificate content will printed to stdout`)
	_ = command.MarkFlagRequired(`sender-email`)
	command.Flags().StringVar(&senderPassword, `sender-password`, ``, `email(sender, password) will be used to send certificates, if not empty`)
	command.Flags().StringVar(&awsSesKey, `aws-ses-key`, ``, `aws-ses(key, secret, region, sender-email) will be used to send certificates, if not empty`)
	command.Flags().StringVar(&awsSesSecret, `aws-ses-secret`, ``, ``)
	command.Flags().StringVar(&awsSesRegion, `aws-ses-region`, ``, ``)
	command.MarkFlagsRequiredTogether(`aws-ses-key`, `aws-ses-secret`, `aws-ses-region`)
	command.MarkFlagsOneRequired(`sender-password`, `aws-ses-key`)
	command.Flags().StringVar(&ca, `ca`, ``, `ca key filename`)
	_ = command.MarkFlagRequired(`ca`)

	return command
}

func getRunner(ctx context.Context, users, allowIPs []string, ca, userKeysDir, validPeriod string, sender Sender) func() error {
	return func() error {
		for _, githubUserName := range users {
			publicKeyFilenames, err := getAllPublicKeys(userKeysDir + string(os.PathSeparator) + githubUserName)
			if err != nil {
				return err
			}
			var certFilenames []string
			recipientEmail := ``
			for _, pubKeyFilename := range publicKeyFilenames {
				publicKeyFile, err := os.Open(pubKeyFilename)
				if err != nil {
					return err
				}
				publicKey, err := io.ReadAll(publicKeyFile)
				if err != nil {
					return err
				}
				_ = publicKeyFile.Close()
				publicKeyEmail, specifiedIPs, err := parseComment(string(publicKey))
				if err != nil {
					return err
				}
				if recipientEmail == `` {
					recipientEmail = publicKeyEmail
				} else if recipientEmail != publicKeyEmail {
					return errors.New(`multiple certificates of the same user using different receiving email addresses`)
				}

				certFilename, err := genCertFromPubKey(ctx, githubUserName, validPeriod, append(specifiedIPs, allowIPs...), ca, pubKeyFilename)
				if err != nil {
					return err
				}
				certFilenames = append(certFilenames, certFilename)
			}
			err = sender(recipientEmail, certFilenames...)
			if err != nil {
				logger.Write(ctx, zap.FatalLevel, `send cert file error`, zap.Error(err))
				return err
			}
		}
		return nil
	}
}

func genCertFromPubKey(ctx context.Context, userName string, validPeriod string, allowIPs []string, ca string, publicKeyFilename string) (string, error) {
	options := []sshkeygen.Option{
		sshkeygen.CertificateIdentity(userName),
		sshkeygen.GithubUsername(userName),
	}
	if validPeriod != `` {
		options = append(options, sshkeygen.ValidPeriod(validPeriod))
	}
	if len(allowIPs) > 0 {
		options = append(options, sshkeygen.SourceAddress(allowIPs...))
	}
	certFilename, err := sshkeygen.Execute(ctx, ca, publicKeyFilename, options...)
	if err != nil {
		logger.Write(ctx, zap.FatalLevel, `generate cert file error`, zap.Error(err), zap.String(`user`, userName))
		return "", err
	}
	return certFilename, nil
}

var pubKeyFilenameRegex, _ = regexp.Compile(`^[a-zA-Z0-9_]+\.pub`)

func getAllPublicKeys(dir string) ([]string, error) {
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return nil, err
	}

	var publicFilenames []string
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		isMatched := pubKeyFilenameRegex.MatchString(fileInfo.Name())
		if !isMatched {
			continue
		}
		fullName := dir + string(os.PathSeparator) + fileInfo.Name()
		publicFilenames = append(publicFilenames, fullName)
	}
	return publicFilenames, nil
}

var ErrMissingParam = errors.New(`missing param`)

func parseComment(publicKey string) (email string, ips []string, err error) {
	parts := strings.Split(publicKey, ` `)
	if len(parts) < 3 {
		return ``, nil, ErrMissingParam
	}
	ips = []string{}
	if len(parts) >= 4 {
		ips = strings.Split(strings.TrimSpace(parts[3]), `,`)
	}
	return strings.TrimSpace(parts[2]), ips, nil
}
