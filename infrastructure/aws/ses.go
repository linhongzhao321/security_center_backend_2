package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

const emailRawBoundary = `NextPart`

var sesClient *sesv2.Client
var senderEmail string

func InitSES(ctx context.Context, region string, key string, secret string, sender string) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, ``)),
	)
	if err != nil {
		return err
	}

	sesClient = sesv2.NewFromConfig(cfg)
	senderEmail = sender
	return nil
}

const templateHeaderContent = `From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-type: Multipart/Mixed; boundary="%s"

--%s
Content-Type: text/plain; charset="UTF-8"

%s
`

const templateAttachment = `
--%[1]s
Content-Type: text/plain; name="%[2]s"
Content-Disposition: attachment; filename="%[2]s"
Content-Transfer-Encoding: base64
Content-ID:<%[2]s>

%[3]s
`

func SendMessage(sub string, destinations []string, content string, b64Attachments, attachmentFilenames []string) error {
	sender := senderEmail
	headerContent := fmt.Sprintf(templateHeaderContent, sender,
		strings.Join(destinations, `,`), sub, emailRawBoundary, emailRawBoundary, content)
	attachments := make([]string, len(b64Attachments))
	for i, b64Attachment := range b64Attachments {
		filename := `unnamed`
		if i < len(attachmentFilenames) {
			filename = attachmentFilenames[i]
		}
		attachments[i] = fmt.Sprintf(templateAttachment, emailRawBoundary, filename, b64Attachment)
	}
	raw := fmt.Sprintf("%s\n\n%s\n\n--%s--\n", headerContent, strings.Join(attachments, "\n\n"), emailRawBoundary)
	params := &sesv2.SendEmailInput{Content: &types.EmailContent{Raw: &types.RawMessage{Data: []byte(raw)}}}
	_, err := sesClient.SendEmail(context.Background(), params)
	return err
}
