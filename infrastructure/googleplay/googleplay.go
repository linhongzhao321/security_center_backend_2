package googleplay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"154.pages.dev/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type DeviceClient struct {
	acquire        bool
	code           string
	device         bool
	tokenFilename  string
	deviceFilename string
	platform       ABI
	v              log.Level
}

func NewDeviceClient(tokenFilename, deviceFilename string) *DeviceClient {
	client := &DeviceClient{
		tokenFilename:  tokenFilename,
		deviceFilename: deviceFilename,
	}
	return client
}

var ErrEmptyVersionCode = errors.New(`query version code fail, result is empty`)

func (f *DeviceClient) DownloadAPK(ctx context.Context, bundleID string) ([]byte, error) {
	var (
		auth    GoogleAuth
		checkin Checkin
	)
	err := f.client(ctx, &auth, &checkin)
	if err != nil {
		return nil, err
	}

	// get latest version
	details, err := checkin.Details(ctx, auth, bundleID, true)
	if err != nil {
		return nil, err
	}
	versionCode, hasVersionCode := details.version_code()
	if !hasVersionCode {
		return nil, ErrEmptyVersionCode
	}

	// query delivery info, and parse download-url
	deliver, err := checkin.Delivery(ctx, auth, bundleID, versionCode, true)
	if err != nil {
		return nil, err
	}
	downloadURL, ok := deliver.URL()
	if !ok {
		return nil, fmt.Errorf(`download url is unparsed`)
	}

	return f.download(ctx, downloadURL)
}

func (f *DeviceClient) client(ctx context.Context, auth *GoogleAuth, checkin *Checkin) error {
	var (
		token Token
		err   error
	)
	token.Data, err = os.ReadFile(f.tokenFilename)
	if err != nil {
		return err
	}
	if err := token.Unmarshal(); err != nil {
		return err
	}
	if err := auth.Auth(ctx, token); err != nil {
		return err
	}
	checkin.Data, err = os.ReadFile(fmt.Sprint(f.deviceFilename))
	if err != nil {
		return err
	}
	return checkin.Unmarshal()
}

func (f *DeviceClient) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`download fail, http code: %d`, res.StatusCode)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	return io.ReadAll(res.Body)
}

type OAuthConfig struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectUris            []string `json:"redirect_uris"`
}

type OAuthClient struct {
	config     *OAuthConfig
	httpClient *http.Client
}

func NewOAuthClient(code, credentialFile string) (*OAuthClient, error) {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, err
	}
	httpClient, err := getClient(config, code)
	if err != nil {
		return nil, err
	}
	client := &OAuthClient{
		httpClient: httpClient,
	}

	return client, nil
}

func (client *OAuthClient) DownloadApk(ctx context.Context, packageName string, versionCode int64) ([]byte, error) {
	srv, err := androidpublisher.NewService(ctx, option.WithHTTPClient(client.httpClient))
	if err != nil {
		return nil, err
	}
	resp, err := srv.Generatedapks.List(packageName, versionCode).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.GeneratedApks) == 0 {
		return nil, fmt.Errorf(`generatedApks is empty`)
	}
	resp, err = srv.Generatedapks.List(packageName, versionCode).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.GeneratedApks) == 0 {
		return nil, fmt.Errorf(`generatedApks is empty`)
	}
	httpResp, err := srv.Generatedapks.
		Download(packageName, versionCode, resp.GeneratedApks[0].GeneratedUniversalApk.DownloadId).
		Download() //nolint:bodyclose
	defer googleapi.CloseBody(httpResp)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(httpResp.Body)
}

var httpClients = make(map[string]*http.Client)

func getClient(config *oauth2.Config, authCode string) (*http.Client, error) {
	if httpClients[authCode] != nil {
		return httpClients[authCode], nil
	}
	tok, err := config.Exchange(context.TODO(), authCode, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		err = fmt.Errorf(`%+v, get code from %s`, err, config.AuthCodeURL(``, oauth2.AccessTypeOffline, oauth2.ApprovalForce))
		return nil, err
	}
	httpClients[authCode] = config.Client(context.Background(), tok)
	return httpClients[authCode], nil
}
