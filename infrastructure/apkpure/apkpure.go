package apkpure

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"go.uber.org/zap"

	"infrastructure/android"
	"infrastructure/logger"
	"internal/core"
)

type ApkPure struct {
	bundleID string
	uri      string
}

func (ap *ApkPure) Name() string {
	return `apkpure`
}

func (ap *ApkPure) URI() string {
	return ap.uri
}

func (ap *ApkPure) GetLastRelease(ctx context.Context, _ core.Release) (release core.Release, err error) {
	agentHeader := `user-agent: ` +
		`Mozilla/5.0 (Macintosh; ` +
		`Intel Mac OS X 10_15_7) ` +
		`AppleWebKit/537.36 (KHTML, like Gecko) ` +
		`Chrome/133.0.0.0 Safari/537.36`

	// TODO @funco.lin http2 支持似乎有问题，先用curl处理
	start := time.Now()
	cmd := exec.Command(`curl`, `-L`, ap.URI(), `-H`, agentHeader)
	apkBytes, err := cmd.Output()
	logger.Write(ctx, zap.InfoLevel, `android.GetVersion.curl`,
		zap.Duration(`runtime/ms`, time.Now().Sub(start)/time.Millisecond),
		zap.Int(`size`, len(apkBytes)),
		zap.String(`cmd`, cmd.String()),
	)
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `android.GetVersion`, zap.Error(err))
		return nil, err
	}

	// contribute core.Release
	apk, err := android.NewRelease(apkBytes)
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `android.GetVersion`, zap.Error(err))
		return nil, err
	}
	return apk, nil
}

func NewReleaseChannel(bundleID string) core.ReleaseChannel {
	releaseChannel := &ApkPure{
		bundleID: bundleID,
		uri:      fmt.Sprintf(`https://d.apkpure.com/b/APK/%s?version=latest`, bundleID),
	}
	return releaseChannel
}
