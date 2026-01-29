package release

import (
	"io"

	"internal/core"
)

type Channel struct {
	name          string
	releaseGetter core.ReleaseGetter
}

func (c *Channel) GetRelease() (reader io.Reader, err error) {
	return c.releaseGetter()
}

func (c *Channel) SetReleaseGetter(getter core.ReleaseGetter) {
	c.releaseGetter = getter
}

func (c *Channel) Name() string {
	return c.name
}
