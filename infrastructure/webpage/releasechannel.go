package webpage

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"path/filepath"

	"go.uber.org/zap"

	"infrastructure/logger"
	"internal/core"
)

type ReleaseChannel struct {
	domain         string
	uri            string
	checkSumGetter CheckSumGetter
}

func (channel *ReleaseChannel) Name() string {
	return `webpage`
}

func (channel *ReleaseChannel) URI() string {
	return channel.uri
}

func (channel *ReleaseChannel) GetLastRelease(ctx context.Context, shouldBeRelease core.Release) (core.Release, error) {
	fileCheckSums := map[string][]byte{}
	for filename := range shouldBeRelease.CheckSums() {
		url := fmt.Sprintf(`%s/%s`, channel.uri, filename)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			// 某些项目发布后可能会将目录结构拍平
			_, filename = filepath.Split(filename)
			url = fmt.Sprintf(`%s/_nuxt/%s`, channel.uri, filename)
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			// 找不到的文件不参与比较即可，有一些文件通过线上无法获取
			if resp.StatusCode != http.StatusOK {
				logger.Write(ctx, zap.DebugLevel, `javascript not found`, zap.String(`filename`, url))
				continue
			}
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		crc := crc32.NewIEEE()
		_, err = crc.Write(body)
		if err != nil {
			return nil, err
		}
		bCRC := make([]byte, 4)
		binary.LittleEndian.PutUint32(bCRC, crc.Sum32())
		fileCheckSums[filename] = bCRC
	}

	return NewRelease(fileCheckSums), nil
}

func NewReleaseChannel(domain string, options ...ReleaseChannelOption) (core.ReleaseChannel, error) {
	rc := &ReleaseChannel{
		domain: domain,
		uri:    domain,
	}
	for _, option := range options {
		err := option(rc)
		if err != nil {
			return nil, err
		}
	}

	return rc, nil
}

func MustNewReleaseChannel(domain string, options ...ReleaseChannelOption) core.ReleaseChannel {
	rc, err := NewReleaseChannel(domain, options...)
	if err != nil {
		panic(err)
	}
	return rc
}

type CheckSumGetter func(ctx context.Context, dirname string) (string, error)
type ReleaseChannelOption func(rc *ReleaseChannel) error

func WithCheckSumGetter(getter CheckSumGetter) ReleaseChannelOption {
	return func(rc *ReleaseChannel) error {
		rc.checkSumGetter = getter
		return nil
	}
}
