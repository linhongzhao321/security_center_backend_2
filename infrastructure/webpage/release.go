package webpage

import (
	"internal/core"
)

type Release struct {
	checkSumItem map[string][]byte
}

func (r *Release) CheckSums() map[string][]byte {
	return r.checkSumItem
}

func NewRelease(checkSumItems map[string][]byte) core.Release {
	release := &Release{
		checkSumItem: checkSumItems,
	}
	return release
}
