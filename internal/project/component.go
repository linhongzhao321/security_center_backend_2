package project

import (
	"context"

	"internal/core"
)

type Component struct {
	name              string
	releaseChannels   []core.ReleaseChannel
	vcsReleasesGetter core.VCSReleasesGetter
}

func (c *Component) GetVCSReleases(ctx context.Context, latest uint8) (core.Releases, error) {
	return c.vcsReleasesGetter(ctx, latest)
}

func (c *Component) Name() string {
	return c.name
}

func (c *Component) ReleaseChannels() []core.ReleaseChannel {
	return c.releaseChannels
}

func NewComponent(name string, options ...ComponentOption) (core.Component, error) {
	component := &Component{name: name}
	for _, option := range options {
		err := option(component)
		if err != nil {
			return nil, err
		}
	}
	return component, nil
}

type ComponentOption func(component *Component) error

func WithReleaseChannel(releaseChannel ...core.ReleaseChannel) ComponentOption {
	return func(component *Component) error {
		component.releaseChannels = append(component.releaseChannels, releaseChannel...)
		return nil
	}
}

func WithVersionControlService(vcsReleasesGetter core.VCSReleasesGetter) ComponentOption {
	return func(component *Component) error {
		component.vcsReleasesGetter = vcsReleasesGetter
		return nil
	}
}
