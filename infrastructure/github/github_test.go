package github

import (
	"context"
	"fmt"
	_ "github.com/go-redis/redis/v7"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"internal/core"
)

func TestClient_GetRelease(t *testing.T) {
	pemFile, err := os.Open(`coinexcom.pem`)
	if err != nil {
		t.Fatal(err)
	}
	pem, err := io.ReadAll(pemFile)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		t.Fatal(err)
	}
	startAt := time.Now()
	client, err := NewClientByApp(`Iv23li8ZIT0l9Kb4QOI5`, privateKey, 59511683)
	if err != nil {
		t.Fatal(err)
	}
	releases, err := client.GetAppReleaseFromAssets(context.Background(), `coinexcom`, `coinex_exchange_android`, ``, 0, core.ReleaseTypeAPK)
	t.Logf(`get release from assets took %v`, time.Since(startAt))
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) == 0 {
		t.Fatal(fmt.Errorf(`releases is empty`))
	}
	if len(releases[0].CheckSums()) == 0 {
		t.Fatal(fmt.Errorf(`check-sums is empty`))
	}
}

func Test_getInstallationToken(t *testing.T) {
	// 创建一个 mock server
	mockServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// 可以在这里检查请求路径、header、method 等，然后构造对应的返回
			if r.URL.Path != "/token" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("123456"))
			if err != nil {
				t.Fatal(err)
			}
		},
	))
	defer mockServer.Close() // 测试结束时关闭

	// 2. 调用被测试的函数，使用 mock server 的 URL
	resp, err := http.Get(mockServer.URL + "/token")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := `123456`
	if string(body) != expected {
		t.Fatalf("unexpected response: got %q want %q", body, expected)
	}
}

func TestClient_GetAppRelease(t *testing.T) {
	price := 99.99 // 程序员想：99.99 × 1000 得到 99990 分？不，先用 99.99 × 100
	quantity := 100.0
	totalCent := price * quantity                   // float64 相乘
	fmt.Printf("float64  计算: %.20f\n", totalCent)   // 9998.99999999999954500000
	fmt.Printf("强制转 int64: %d\n", int64(totalCent)) // 9998！少了1分钱！
}
