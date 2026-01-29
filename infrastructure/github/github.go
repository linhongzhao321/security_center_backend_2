package github

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v76/github"
	errors2 "github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"infrastructure/android"
	ioslib "infrastructure/ios"
	"infrastructure/logger"
	"internal/core"
	baseReleae "internal/release"
)

const iatMaxOffset = 60

// 根据官方文档，jwt 时间不能超过 10 分钟
// 这里仅允许 9 分钟
const expireOffset = 540
const maxRedirect = 3
const downloadRetryLimit = 5
const downloadRetryInterval = 5 * time.Second

type Client struct {
	githubClient *github.Client
	mutex        sync.Mutex
	expiresAt    *github.Timestamp
}

var clients = sync.Map{}

func GetClientByApp(clientID string, privateKey *rsa.PrivateKey, installationID int64, options ...Option) (*Client, error) {
	cliKey := fmt.Sprintf(`%s.%d`, clientID, installationID)
	client, isExist := clients.Load(cliKey)
	var err error
	// 不存在或即将过期则刷新
	if !isExist || time.Now().Before(client.(*Client).expiresAt.Time.Add(-time.Minute)) {
		client, err = NewClientByApp(clientID, privateKey, installationID, options...)
		if err != nil {
			return nil, err
		}
		clients.Store(cliKey, client)
	}
	return client.(*Client), nil
}

func NewClientByApp(clientID string, privateKey *rsa.PrivateKey, installationID int64, options ...Option) (*Client, error) {
	installationToken, err := getInstallationToken(clientID, privateKey, installationID)
	if err != nil {
		return nil, err
	}
	client := &Client{
		githubClient: github.NewClient(nil).WithAuthToken(installationToken.GetToken()),
		mutex:        sync.Mutex{},
		expiresAt:    installationToken.ExpiresAt,
	}

	for _, option := range options {
		err := option(client)
		if err != nil {
			logger.Write(context.Background(), zap.FatalLevel, `github client error`, zap.Error(err))
			return nil, err
		}
	}

	return client, nil
}

func getInstallationToken(clientID string, privateKey *rsa.PrivateKey, installationID int64) (*github.InstallationToken, error) {
	logger.Write(context.Background(), zap.InfoLevel, `getInstallationToken()`)
	now := time.Now().Unix()
	tokenGenerator := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now - iatMaxOffset,
		"exp": now + expireOffset,
		"iss": clientID,
	})

	logger.Write(context.Background(), zap.InfoLevel, `signed app token`)
	appToken, err := tokenGenerator.SignedString(privateKey)
	if err != nil {
		logger.Write(context.Background(), zap.FatalLevel, `sign github token error`, zap.Error(err))
		return nil, errors2.Wrap(err, `github token error`)
	}
	githubClient := github.NewClient(nil).WithAuthToken(appToken)
	logger.Write(context.Background(), zap.InfoLevel, `signed installation token`)
	installationToken, resp, err := githubClient.Apps.CreateInstallationToken(context.Background(), installationID, nil)
	if err != nil {
		logger.Write(context.Background(), zap.FatalLevel, `CreateInstallationToken() error`,
			zap.Error(err), zap.Int(`status code`, resp.StatusCode), zap.Any(`resp`, resp))
		return nil, errors2.Wrap(err, `CreateInstallationToken() error`)
	}
	logger.Write(context.Background(), zap.InfoLevel, `getInstallationToken() end`)
	return installationToken, nil
}

type Option func(client *Client) error

// CommitsBranchResp
// response for https://api.github.com/repos/:owner/:repo/commits/:branch
type CommitsBranchResp struct {
	SHA    string
	Commit struct {
		Committer struct {
			Name  string
			Email string
			Date  string
		}
		Message string
	}
}

// GetWebReleases 获取 web 端项目的 release
// 由于 web 端一般不存在发布延迟的问题，因此不必返回最近的多个版本
func (c *Client) GetWebReleases(ctx context.Context, owner, repo string, corePaths ...string) (core.Releases, error) {
	c.mutex.Lock()
	url, _, err := c.githubClient.Repositories.GetArchiveLink(ctx, owner, repo, github.Zipball, nil, maxRedirect)
	c.mutex.Unlock()
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `get artchive link fail`, zap.Any(`header`, zap.Error(err)))
		return nil, err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `get artchive link fail`, zap.Any(`resp`, resp), zap.Error(err))
		return nil, err
	}
	buff, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(buff)
	zipReader, err := zip.NewReader(reader, int64(len(buff)))
	if err != nil {
		return nil, err
	}
	coreFiles := map[string]*zip.File{}
	root := zipReader.File[0].Name
	for _, file := range zipReader.File[1:] {
		ext := path.Ext(file.Name)
		if ext != `.js` {
			continue
		}
		filename := strings.Replace(file.Name, root, ``, 1)
		inWhitelist := false
		for _, corePath := range corePaths {
			if strings.Index(filename, corePath) == 0 {
				inWhitelist = true
				break
			}
		}
		if !inWhitelist {
			continue
		}
		coreFiles[filename] = file
	}
	coreFilenames := maps.Keys(coreFiles)
	sort.Strings(coreFilenames)

	checkSumItems := map[string][]byte{}
	for _, filename := range coreFilenames {
		bCRC := make([]byte, 4)
		binary.LittleEndian.PutUint32(bCRC, coreFiles[filename].CRC32)
		_, filename = filepath.Split(filename)
		checkSumItems[filename] = bCRC
	}

	return core.Releases{baseReleae.NewBaseRelease(checkSumItems)}, nil
}

func (c *Client) GetAppReleaseFromAssets(
	ctx context.Context, owner, repo, binaryFilename string, latest int, releaseType core.ReleaseType,
) (core.Releases, error) {
	listOptions := &github.ListOptions{PerPage: latest, Page: 1}
	c.mutex.Lock()
	releases, resp, err := c.githubClient.Repositories.ListReleases(ctx, owner, repo, listOptions)
	c.mutex.Unlock()
	if err != nil {
		logger.Write(ctx, zap.ErrorLevel, `get artchive link fail`, zap.Any(`resp`, resp), zap.Error(err))
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			`ListReleases() the status code should be 200, but in reality it is %d`, resp.StatusCode,
		)
		return nil, err
	}

	var retReleases core.Releases
	for _, release := range releases {
		if releaseType == core.ReleaseTypeAPK {
			retRelease, err := c.getAPK(ctx, owner, repo, release, release.Assets[0].GetID())
			if err != nil {
				return nil, err
			}
			retReleases = append(retReleases, retRelease)
		} else if releaseType == core.ReleaseTypeIPA {
			retRelease, err := c.getIPABytes(ctx, owner, repo, binaryFilename, release.Assets[0].GetID())
			if err != nil {
				return nil, err
			}
			retReleases = append(retReleases, retRelease)
		} else {
			return nil, fmt.Errorf("invalid release type: %s", releaseType)
		}
	}
	return retReleases, nil
}

func (c *Client) getAPK(ctx context.Context, owner string, repo string, release *github.RepositoryRelease, assetID int64) (core.Release, error) {
	isAAB := path.Ext(release.Assets[0].GetName()) == `.aab`

	versionCode, err := stringToVersionCode(release.GetName())
	if err != nil {
		return nil, err
	}

	// download bytes
	logger.Write(ctx, zap.InfoLevel, `DownloadReleaseAsset`,
		zap.String(`owner`, owner),
		zap.String(`repo`, repo),
	)
	retry := 0
	var apkReader io.ReadCloser
	for retry < downloadRetryLimit {
		c.mutex.Lock()
		apkReader, _, err = c.githubClient.Repositories.
			DownloadReleaseAsset(ctx, owner, repo, assetID, c.githubClient.Client())
		c.mutex.Unlock()
		logger.Write(ctx, zap.InfoLevel, `DownloadReleaseAsset done`,
			zap.String(`owner`, owner),
			zap.String(`repo`, repo),
			zap.Int(`retry`, retry),
			zap.Error(err),
		)
		if err == nil {
			break
		}
		retry += 1
		time.Sleep(downloadRetryInterval)
	}
	if retry == downloadRetryLimit {
		return nil, errors.New(`retry limit exceeded, failed to download resource`)
	}
	apkBytes, err := io.ReadAll(apkReader)
	_ = apkReader.Close()
	if err != nil {
		return nil, err
	}

	logger.Write(ctx, zap.InfoLevel, `android.NewRelease`,
		zap.String(`owner`, owner),
		zap.String(`repo`, repo),
	)
	apkRelease, err := android.NewRelease(apkBytes, android.WithVersionCode(versionCode), android.WithAAB(isAAB))
	logger.Write(ctx, zap.InfoLevel, `android.NewRelease done`,
		zap.String(`owner`, owner),
		zap.String(`repo`, repo),
	)

	return apkRelease, err
}

func (c *Client) getIPABytes(ctx context.Context, owner string, repo string, binaryFilename string, assetID int64) (core.Release, error) {
	// download bytes
	retry := 0
	var ipaReader io.ReadCloser
	var err error
	for retry < downloadRetryLimit {
		c.mutex.Lock()
		ipaReader, _, err = c.githubClient.Repositories.
			DownloadReleaseAsset(ctx, owner, repo, assetID, c.githubClient.Client())
		c.mutex.Unlock()
		logger.Write(ctx, zap.InfoLevel, `DownloadIOSReleaseAsset done`,
			zap.String(`owner`, owner),
			zap.String(`repo`, repo),
			zap.Int(`retry`, retry),
			zap.Error(err),
		)
		if err == nil {
			break
		}
		retry += 1
		time.Sleep(downloadRetryInterval)
	}
	if retry == downloadRetryLimit {
		return nil, errors.New(`retry limit exceeded, failed to download resource`)
	}
	ipaBytes, err := io.ReadAll(ipaReader)
	_ = ipaReader.Close()
	if err != nil {
		return nil, err
	}
	fileSize := len(ipaBytes)

	ipaBytesReader := bytes.NewReader(ipaBytes)
	return ioslib.NewIPA(ipaBytesReader, int64(fileSize), binaryFilename)
}

const versionSplitNum = 3

// version is string for "v1.02.3" format
func stringToVersionCode(version string) (int64, error) {
	split := strings.Split(version[1:], `.`)
	if len(split) != versionSplitNum {
		return 0, errors.New(`incorrect reading of version number, len(splitted) != 3`)
	}

	major, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return 0, err
	}
	minor, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return 0, err
	}
	revision, err := strconv.ParseInt(split[2], 10, 64)
	if err != nil {
		return 0, err
	}
	return major*1000 + minor*10 + revision, nil
}
