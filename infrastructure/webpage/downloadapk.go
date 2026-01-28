package webpage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"

	"go.uber.org/zap"

	"infrastructure/android"
	"infrastructure/logger"
	"internal/core"
)

// use for ReleaseChannelForAPK.downloadFlag
// const redirect = 0b001
const isRegExp = 0b010
const isURL = 0b100
const DownloadType = 0b110

type ReleaseChannelForAPK struct {
	name         string
	regexp       *regexp.Regexp
	pageURL      string
	downloadURL  string
	downloadFlag uint
}

func (releaseChannel *ReleaseChannelForAPK) Name() string {
	return releaseChannel.name
}

func (releaseChannel *ReleaseChannelForAPK) URI() string {
	return releaseChannel.pageURL
}

func (releaseChannel *ReleaseChannelForAPK) GetLastRelease(ctx context.Context, _ core.Release) (release core.Release, err error) {
	var apk core.Release

	switch releaseChannel.downloadFlag & DownloadType {
	case isRegExp:
		resp, err := http.Get(releaseChannel.pageURL)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		apkURL := releaseChannel.regexp.FindString(string(body))
		resp, err = http.Get(apkURL)
		if err != nil {
			return nil, err
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// contribute core.Release
		apk, err = android.NewRelease(body)
		if err != nil {
			logger.Write(ctx, zap.ErrorLevel, `android.GetVersion`, zap.Error(err))
			return nil, err
		}
	case isURL:
		resp, err := http.Get(releaseChannel.downloadURL)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// contribute core.Release
		apk, err = android.NewRelease(body)
		if err != nil {
			logger.Write(ctx, zap.ErrorLevel, `android.GetVersion`, zap.Error(err))
			return nil, err
		}
	}

	return apk, nil
}

const RegexpApkURI = `https\:[a-zA-Z0-9\_\-\.\/]+\.apk`

type Option func(apk *ReleaseChannelForAPK) error

func FindURLByRegExp(downloadPageURL string, exp string) Option {
	return func(downloadAPK *ReleaseChannelForAPK) error {
		if downloadAPK.downloadFlag&DownloadType != 0 {
			return errors.New(`Conflicting download URL settings`)
		}
		objRegExp, err := regexp.Compile(exp)
		if err != nil {
			return err
		}
		downloadAPK.regexp = objRegExp
		downloadAPK.downloadFlag = isRegExp
		downloadAPK.pageURL = downloadPageURL
		return nil
	}
}

func DownloadURL(u string) Option {
	return func(downloadAPK *ReleaseChannelForAPK) error {
		if downloadAPK.downloadFlag&DownloadType != 0 {
			return errors.New(`Conflicting download URL settings`)
		}
		downloadAPK.downloadURL = u
		downloadAPK.downloadFlag = isURL
		return nil
	}
}

func NewReleaseChannelForAPK(options ...Option) (core.ReleaseChannel, error) {
	releaseChannel := &ReleaseChannelForAPK{
		name: `download apk from web`,
	}
	for _, option := range options {
		err := option(releaseChannel)
		if err != nil {
			return nil, err
		}
	}

	return releaseChannel, nil
}

func MustReleaseChannelForAPK(options ...Option) core.ReleaseChannel {
	releaseChannel, err := NewReleaseChannelForAPK(options...)
	if err != nil {
		panic(err)
	}
	return releaseChannel
}
