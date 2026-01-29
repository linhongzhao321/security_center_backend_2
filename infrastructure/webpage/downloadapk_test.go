package webpage

import (
	"context"
	"testing"
)

func TestFindURLByRegExp(t *testing.T) {
	releaseChannel, err := NewReleaseChannelForAPK(
		FindURLByRegExp(`https://api.wallet.coinex.com/res/walletapp/url`, RegexpApkURI),
	)
	if err != nil {
		t.Fatal(err)
	}
	apk, err := releaseChannel.GetLastRelease(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if apk == nil {
		t.Error(`apk is nil`)
	}
}

func TestDownloadURL(t *testing.T) {
	releaseChannel, err := NewReleaseChannelForAPK(
		DownloadURL(`https://www.viabtc.com/res/common/app/download/android`),
	)
	if err != nil {
		t.Fatal(err)
	}
	apk, err := releaseChannel.GetLastRelease(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if apk == nil {
		t.Error(`apk is nil`)
	}
}
