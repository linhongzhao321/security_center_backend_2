package release

import "internal/core"

type BaseRelease struct {
	checkSumItems map[string][]byte
}

func (release *BaseRelease) CheckSums() map[string][]byte {
	return release.checkSumItems
}

func NewBaseRelease(items map[string][]byte) core.Release {
	return &BaseRelease{checkSumItems: items}
}
