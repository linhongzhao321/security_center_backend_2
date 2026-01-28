package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"infrastructure/logger"
)

func Notify(ctx context.Context, webhook, content string, recipientIDs ...string) (postErr error) {
	if len(recipientIDs) > 0 {
		content = fmt.Sprintf("%s\n\n%s", userBlocks(recipientIDs...), content)
	}
	postErr = slack.PostWebhookContext(
		ctx,
		webhook,
		&slack.WebhookMessage{
			Text: content,
		},
	)
	logger.Write(ctx, zapcore.ErrorLevel, content)
	if postErr != nil {
		logger.Write(ctx, zap.FatalLevel, `slack.Notifier.Notify()`, zap.Error(postErr))
		return postErr
	}
	return postErr
}

func userBlocks(slackIDs ...string) string {
	var blocks []string
	for _, slackID := range slackIDs {
		blocks = append(blocks, userBlock(slackID))
	}
	return strings.Join(blocks, ``)
}

func userBlock(slackID string) string {
	return fmt.Sprintf(`<@%s>`, slackID)
}
