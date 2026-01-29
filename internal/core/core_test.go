package core

import (
	"context"
	"crypto/cipher"
	"fmt"
	"math"
)

type TestReleaseChannel struct {
}

func (t TestReleaseChannel) Name() string {
	return `test-channel`
}

func (t TestReleaseChannel) URI() string {
	return `http://localhost:8080`
}

type TestRelease struct {
}

func (t *TestRelease) CheckSums() map[string][]byte {
	return map[string][]byte{
		`1.txt`: []byte("rdKZBQ8QbexN55up"),
		`2.txt`: []byte("dBxyekz34aHnF71w"),
		`3.txt`: []byte("EpJilNm02RUTNrns"),
	}
}

func (t TestReleaseChannel) GetLastRelease(ctx context.Context, shouldBeRelease Release) (release Release, err error) {
	cipher.NewCBCEncrypter()
	return &TestRelease{}, nil
}
