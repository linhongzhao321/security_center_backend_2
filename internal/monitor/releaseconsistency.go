package monitor

import (
	"bytes"
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"infrastructure/logger"
	"internal/core"
)

type ReleaseConsistencyMonitor struct {
	notifiers core.MonitorNotifiers
}

func (monitor *ReleaseConsistencyMonitor) Description() string {
	return `this monitor is used to monitor the version consistency ` +
		`between the product from release channel and the  product contributed from version control service`
}

func (monitor *ReleaseConsistencyMonitor) Name() string {
	return `version-consistency`
}

func (monitor *ReleaseConsistencyMonitor) Run(ctx context.Context, project core.Project) []core.MonitorResult {
	var results []core.MonitorResult
	for _, component := range project.Components() {
		for _, releaseChannel := range component.ReleaseChannels() {
			logger.Write(ctx, zap.InfoLevel, `compareCheckSum() start`,
				zap.String(`project`, project.Name()),
				zap.String(`name`, releaseChannel.Name()),
			)
			result := monitor.compareCheckSum(ctx, project, releaseChannel, component)
			logger.Write(ctx, zap.InfoLevel, `compareCheckSum() end`,
				zap.String(`project`, project.Name()),
				zap.String(`name`, releaseChannel.Name()),
			)
			monitor.notify(ctx, result)
			results = append(results, result)
		}
	}
	return results
}

func (monitor *ReleaseConsistencyMonitor) notify(ctx context.Context, result core.MonitorResult) {
	if len(monitor.notifiers) == 0 {
		logger.Write(ctx, zap.WarnLevel, `notifiers is empty`,
			zap.String(`monitor`, monitor.Name()),
			zap.Any(`result`, result),
		)
		return
	}

	err := monitor.notifiers.Notify(ctx, result)
	if err != nil {
		logger.Write(ctx, zap.WarnLevel, `notify fail`,
			zap.Error(err),
			zap.String(`monitor`, monitor.Name()),
			zap.Any(`result`, result),
		)
	}
}

const compareRetryLimit = 5

func (monitor *ReleaseConsistencyMonitor) compareCheckSum(
	ctx context.Context, project core.Project, releaseChannel core.ReleaseChannel, component core.Component,
) core.MonitorResult {
	logger.Write(ctx, zap.InfoLevel, `GetLastRelease`,
		zap.String(`project`, project.Name()),
		zap.String(`component`, component.Name()),
		zap.String(`channel`, releaseChannel.Name()),
		zap.String(`name`, monitor.Name()),
	)
	shouldBeReleases, err := component.GetVCSReleases(ctx, compareRetryLimit+1)
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `get vcs-release error`,
			zap.String(`project`, project.Name()),
			zap.String(`component`, component.Name()),
			zap.String(`channel`, releaseChannel.Name()),
			zap.Error(err),
		)
		err = fmt.Errorf(`get %s.%s.%s vcs-release error. %s`,
			project.Name(), component.Name(), releaseChannel.Name(), err.Error())
		return NewResult(monitor.Name(), project.Name(), component.Name(), releaseChannel.Name(), WithError(err))
	}

	// 从最新的 release 开始比对，最近 defaultMaxLastOffset 个release都不符合要求则报错
	var retryCount uint8 = 0
	for _, shouldBeRelease := range shouldBeReleases {
		retryCount += 1
		actualRelease, err := releaseChannel.GetLastRelease(ctx, shouldBeRelease)
		if err != nil {
			logger.Write(ctx, zap.WarnLevel, `get release error`,
				zap.String(`project`, project.Name()),
				zap.String(`component`, component.Name()),
				zap.String(`channel`, releaseChannel.Name()),
				zap.Error(err),
			)
			continue
		}

		logger.Write(ctx, zap.InfoLevel, `compare check sum`,
			zap.String(`project`, project.Name()),
			zap.String(`component`, component.Name()),
			zap.String(`channel`, releaseChannel.Name()),
			zap.String(`name`, monitor.Name()),
			zap.Uint8(`retryCount`, retryCount),
		)
		shouldBeCheckSums := shouldBeRelease.CheckSums()
		isPass := true
		for key, checkSum := range actualRelease.CheckSums() {
			shouldBeCheckSum, ok := shouldBeCheckSums[key]
			if !ok {
				logger.Write(ctx, zapcore.WarnLevel, `compare fail`,
					zap.String(`project`, project.Name()),
					zap.String(`component`, component.Name()),
					zap.String(`channel`, releaseChannel.Name()),
				)
				continue
			}
			if !bytes.Equal(checkSum, shouldBeCheckSum) {
				logger.Write(ctx, zapcore.WarnLevel, `checksum inconsistent`,
					zap.Uint8(`retryCount`, retryCount),
					zap.String(`key`, key),
					zap.ByteString(`shouldBe`, shouldBeCheckSum),
					zap.ByteString(`actual`, checkSum),
				)
				isPass = false
				break
			}
		}
		if isPass {
			break
		}
	}
	logger.Write(ctx, zap.InfoLevel, `compare done`,
		zap.String(`project`, project.Name()),
		zap.String(`component`, component.Name()),
		zap.String(`channel`, releaseChannel.Name()),
		zap.String(`name`, monitor.Name()),
		zap.Uint8(`retryCount`, retryCount),
	)
	if retryCount > compareRetryLimit {
		return NewResult(monitor.Name(), project.Name(), component.Name(), releaseChannel.Name(),
			Extension(`uri`, releaseChannel.URI()),
			Error(fmt.Sprintf(`there are no matching version in the last %d releases`, retryCount)),
		)
	}
	return NewResult(monitor.Name(), project.Name(), component.Name(), releaseChannel.Name())
}

func NewReleaseConsistencyMonitor(notifiers ...core.MonitorNotifier) core.Monitor {
	return &ReleaseConsistencyMonitor{
		notifiers: notifiers,
	}
}
