package ios

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"infrastructure/logger"
	"internal/core"
)

const retryIPADownloadLimit = 5
const retryIPADownloadInterval = 5 * time.Second

type AppStore struct {
	passphrase     string
	bundleID       string
	binaryFilename string
	uri            string
}

func (appStore *AppStore) Name() string {
	return `App Store`
}

func (appStore *AppStore) URI() string {
	return appStore.uri
}

func (appStore *AppStore) GetLastRelease(ctx context.Context, _ core.Release) (release core.Release, err error) {
	// generate temp-file
	ipaFile, err := os.CreateTemp(``, ``)
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `ios.GetVersion:os.CreateTemp`, zap.Error(err))
		return nil, err
	}
	err = ipaFile.Close()
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `ios.GetVersion:ipaFile.Close`, zap.Error(err))
		return nil, err
	}
	defer os.Remove(ipaFile.Name())

	// download ipa
	err = appStore.ipatoolDownload(ctx, ipaFile)
	if err != nil {
		return nil, err
	}

	// reopen & get release
	ipaFile, err = os.Open(ipaFile.Name())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = ipaFile.Close()
	}()
	ipaFileStat, err := ipaFile.Stat()
	if err != nil {
		return nil, err
	}
	return NewIPA(ipaFile, ipaFileStat.Size(), appStore.binaryFilename)
}

func (appStore *AppStore) ipatoolDownload(ctx context.Context, ipaFile *os.File) error {
	cmd := exec.Command(`ipatool`, `download`,
		`--non-interactive`,
		`-b`, appStore.bundleID,
		`-o`, ipaFile.Name(),
	)
	if appStore.passphrase != `` {
		cmd.Args = append(cmd.Args, `--keychain-passphrase`, appStore.passphrase)
	}
	logger.Write(ctx, zap.DebugLevel, `run ipatool`, zap.String(`bundle-id`, appStore.bundleID))
	retry := 0
	for retry < retryIPADownloadLimit {
		err := cmd.Run()
		if err == nil {
			return nil
		}
		retry += 1
		output := ``
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			output = exitErr.String() + ":" + string(exitErr.Stderr)
		}
		logger.Write(ctx, zap.ErrorLevel, `ios.GetVersion`,
			zap.Error(err),
			zap.String(`cmd`, cmd.String()),
			zap.String(`output`, output),
		)
		time.Sleep(retryIPADownloadInterval)
	}
	return errors.New(`retry limit exceeded, failed to download ipa by ipatool`)
}

func NewAppStore(ctx context.Context, passphrase, bundleID, binaryFilename string) *AppStore {
	uri, err := getURI(bundleID)
	if err != nil {
		logger.Write(ctx, zap.WarnLevel, `get app store uri error, possible lack of auxiliary information during alarm`,
			zap.Error(err))
	}
	return &AppStore{
		passphrase:     passphrase,
		bundleID:       bundleID,
		uri:            uri,
		binaryFilename: binaryFilename,
	}
}

func getURI(bundleID string) (string, error) {
	uri := `https://itunes.apple.com/lookup?bundleId=` + bundleID
	resp, err := http.Get(uri)
	if err != nil {
		return ``, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ``, err
	}
	result := gjson.GetBytes(respBytes, `results.0.trackViewUrl`)
	if result.String() == `` {
		return ``, fmt.Errorf(`get trackViewUrl fail. resp:%+v`, resp)
	}
	return result.String(), nil
}
