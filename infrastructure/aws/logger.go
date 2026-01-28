package aws

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/logging"
	"go.uber.org/zap"

	selfLogger "infrastructure/logger"
)

type Logger struct {
}

func (l *Logger) Logf(classification logging.Classification, format string, v ...interface{}) {
	if classification == logging.Debug {
		selfLogger.Write(context.Background(), zap.DebugLevel, fmt.Sprintf(format, v...))
	}
}
