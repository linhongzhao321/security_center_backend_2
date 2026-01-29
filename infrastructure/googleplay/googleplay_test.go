package googleplay

import (
	"context"
	"fmt"
	"testing"
)

func TestGeneratedApks(t *testing.T) {
	client, err := NewOAuthClient(`4/0Ab32j93gA-hPZx_piRK6eMCjBuDq_iTybNVvrU1y1hZ_t1A17rwXp7r-_6bfzuzVv7jrjg`, `credential.json`)
	if err != nil {
		panic(err)
	}
	raw, err := client.DownloadApk(context.Background(), `com.coinex.trade.play`, 4014)
	if err != nil {
		panic(err)
	}
	if len(raw) == 0 {
		panic(fmt.Errorf(`apk raw is empty`))
	}
}
