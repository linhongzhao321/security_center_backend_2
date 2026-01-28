package core

import (
	"context"
	"io"
)

type Project interface {
	Name() string
	Components() []Component
	Run(ctx context.Context) []MonitorResult
}

type Component interface {
	Name() string
	ReleaseChannels() []ReleaseChannel
	// GetVCSReleases 获取 git 中最新的 latest 个 release
	GetVCSReleases(ctx context.Context, latest uint8) (Releases, error)
}

// ReleaseChannel
// 发行渠道的抽象
type ReleaseChannel interface {
	// Name 渠道名称
	Name() string
	// URI 渠道链接
	URI() string
	// GetLastRelease 获取当前渠道的发行版
	GetLastRelease(ctx context.Context, shouldBeRelease Release) (release Release, err error)
}

type ReleaseGetter func() (reader io.Reader, err error)

type VCSReleasesGetter func(ctx context.Context, latest uint8) (release Releases, err error)

type Release interface {
	CheckSums() map[string][]byte
}

type Releases []Release

type ReleaseType string

const (
	ReleaseTypeAPK ReleaseType = "apk"
	ReleaseTypeIPA ReleaseType = "ipa"
)

type Monitor interface {
	Name() string
	Description() string
	Run(ctx context.Context, project Project) []MonitorResult
}

type MonitorResult interface {
	MonitorName() string
	ProjectName() string
	ComponentName() string
	ReleaseChannelName() string
	Extensions() map[string]string
	IsFail() bool
	String() string
	Error() error
}

type MonitorNotifier func(ctx context.Context, result MonitorResult) error
type MonitorNotifiers []MonitorNotifier

func (notifiers MonitorNotifiers) Notify(ctx context.Context, result MonitorResult) error {
	for _, notifier := range notifiers {
		err := notifier(ctx, result)
		if err != nil {
			return err
		}
	}
	return nil
}
