package googleplay

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"infrastructure/android"
	"infrastructure/logger"
	"internal/core"
)

const retryLimit = 5
const retryInterval = 5 * time.Second

type ReleaseChannel struct {
	bundleID string
	client   *DeviceClient
	uri      string
}

func (releaseChannel *ReleaseChannel) Name() string {
	return `google play device simulation`
}

func (releaseChannel *ReleaseChannel) URI() string {
	return releaseChannel.uri
}

func (releaseChannel *ReleaseChannel) GetLastRelease(ctx context.Context, _ core.Release) (core.Release, error) {
	retry := 0
	for retry < 5 {
		apk, err := releaseChannel.client.DownloadAPK(ctx, releaseChannel.bundleID)
		if err == nil {
			return android.NewRelease(apk)
		}
		retry += 1
		logger.Write(ctx, zap.ErrorLevel, ` download apk from google play with oauth`,
			zap.String(`bundleID`, releaseChannel.bundleID),
		)
		time.Sleep(retryInterval)
	}
	return nil, errors.New(`retry limit exceeded, failed to download apk from google play with simulate device`)
}

func NewReleaseChannel(tokenFilename, deviceFilename, bundleID string) (core.ReleaseChannel, error) {
	googlePlay := NewDeviceClient(tokenFilename, deviceFilename)
	return &ReleaseChannel{
		bundleID: bundleID,
		uri:      fmt.Sprintf(`https://play.google.com/store/apps/details?id=%s`, bundleID),
		client:   googlePlay,
	}, nil
}

func MustNewReleaseChannel(tokenFilename, deviceFilename, bundleID string) core.ReleaseChannel {
	releaseChannel, err := NewReleaseChannel(tokenFilename, deviceFilename, bundleID)
	if err != nil {
		panic(err)
	}
	return releaseChannel
}

type OAuthReleaseChannel struct {
	bundleID string
	client   *OAuthClient
	uri      string
}

func (releaseChannel *OAuthReleaseChannel) Name() string {
	return `google play oauth`
}

func (releaseChannel *OAuthReleaseChannel) URI() string {
	return releaseChannel.uri
}

func (releaseChannel *OAuthReleaseChannel) GetLastRelease(
	ctx context.Context, shouldBeRelease core.Release,
) (core.Release, error) {
	apkRelease, ok := shouldBeRelease.(*android.APK)
	if !ok {
		return nil, errors.New(`unable to retrieve version code from github-apk-release`)
	}
	retry := 0
	for retry < retryLimit {
		apk, err := releaseChannel.client.DownloadApk(ctx, releaseChannel.bundleID, apkRelease.VersionCode())
		if err == nil {
			return android.NewRelease(apk)
		}
		retry += 1
		logger.Write(ctx, zap.ErrorLevel, ` download apk from google play with oauth`,
			zap.String(`bundleID`, releaseChannel.bundleID),
			zap.Int64(`versionCode`, apkRelease.VersionCode()),
		)
		time.Sleep(retryInterval)
	}
	return nil, errors.New(`retry limit exceeded, failed to download apk from google play with oauth`)
}

func NewOAuthReleaseChannel(bundleID, code, credentialFile string) (core.ReleaseChannel, error) {
	googlePlay, err := NewOAuthClient(code, credentialFile)
	if err != nil {
		return nil, err
	}
	return &OAuthReleaseChannel{
		bundleID: bundleID,
		uri:      fmt.Sprintf(`https://play.google.com/store/apps/details?id=%s`, bundleID),
		client:   googlePlay,
	}, nil
}

func MustNewOAuthReleaseChannel(bundleID, code, credentialFile string) core.ReleaseChannel {
	releaseChannel, err := NewOAuthReleaseChannel(bundleID, code, credentialFile)
	if err != nil {
		panic(err)
	}
	return releaseChannel
}
