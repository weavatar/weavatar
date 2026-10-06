package conf_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/platform/conf"
)

func writeConfig(t *testing.T, yaml string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	must.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	t.Setenv("APP_CONFIG", path)
}

const minimal = `
app:
  name: "test-app"
  key: "a-long-string-with-32-characters"
http:
  address: ":3000"
  domain: "weavatar.test"
database:
  dsn: "postgres://postgres:postgres@127.0.0.1:5432/weavatar?sslmode=disable"
`

func TestLoadAppliesDefaults(t *testing.T) {
	writeConfig(t, minimal)

	c, err := conf.Load()
	must.NoError(t, err)

	must.Equal(t, c.App.Name, "test-app")
	must.Equal(t, c.HTTP.Domain, "weavatar.test")
	must.Equal(t, c.HTTP.BodyLimit, 4096)
	must.Equal(t, c.HTTP.ProxyHeader, "", must.Msgf("no proxy is trusted by default"))
	must.Equal(t, c.Log.Level, "info")
	must.Equal(t, c.Log.Output, "file")
	must.Equal(t, c.Log.Path, "storage/logs/app.log")
	must.Equal(t, c.Hash.Dir, "storage/hash")
	must.Equal(t, c.Code.Expire, 5*time.Minute)
	must.Equal(t, c.Audit.Driver, "aliyun")
}

func TestLoadDecodesNestedValues(t *testing.T) {
	writeConfig(t, minimal+`
log:
  output: "stderr"
code:
  expire: "10m"
sms:
  aliyun:
    access_key_id: "ak"
  tencent:
    sdk_app_id: 1400000000
oauth:
  base_url: "https://account.haozi.net/"
  client_id: "client"
audit:
  driver: "cos"
cdn:
  drivers: ["baishan", "cloudflare"]
  cloudflare:
    zone_id: "zone"
`)

	c, err := conf.Load()
	must.NoError(t, err)

	must.Equal(t, c.Log.Output, "stderr")
	must.Equal(t, c.Code.Expire, 10*time.Minute)
	must.Equal(t, c.SMS.Aliyun.AccessKeyID, "ak")
	must.Equal(t, c.SMS.Tencent.SDKAppID, "1400000000") // numbers weakly decode into strings
	must.Equal(t, c.OAuth.BaseURL, "https://account.haozi.net", must.Msgf("trailing slash trimmed"))
	must.Equal(t, c.OAuth.ClientID, "client")
	must.Equal(t, c.Audit.Driver, "cos")
	must.DeepEqual(t, c.CDN.Drivers, []string{"baishan", "cloudflare"})
	must.Equal(t, c.CDN.Cloudflare.ZoneID, "zone")
}

func TestLoadProxySettings(t *testing.T) {
	writeConfig(t, strings.Replace(minimal, `  domain: "weavatar.test"`, `  domain: "weavatar.test"
  proxy_header: "X-Real-IP"`, 1))

	c, err := conf.Load()
	must.NoError(t, err)

	must.Equal(t, c.HTTP.ProxyHeader, "X-Real-IP")
}

func TestLoadIgnoresEnvironment(t *testing.T) {
	writeConfig(t, minimal)
	t.Setenv("APP_HTTP__ADDRESS", ":8080")
	t.Setenv("APP_DATABASE__DSN", "postgres://other")

	c, err := conf.Load()
	must.NoError(t, err)

	must.Equal(t, c.HTTP.Address, ":3000")
	must.Equal(t, c.Database.DSN, "postgres://postgres:postgres@127.0.0.1:5432/weavatar?sslmode=disable")
}

func TestLoadExampleConfig(t *testing.T) {
	t.Setenv("APP_CONFIG", filepath.Join("..", "..", "..", "config", "config.example.yml"))

	c, err := conf.Load()
	must.NoError(t, err)

	must.Equal(t, c.App.Name, "weavatar")
	must.Equal(t, c.HTTP.Domain, "weavatar.com")
	must.DeepEqual(t, c.HTTP.CorsOrigins, []string{"*"})
	must.Equal(t, c.HTTP.ProxyHeader, "", must.Msgf("no proxy is trusted by default"))
	must.Equal(t, c.HTTP.ReadTimeout, 10*time.Second)
	must.Equal(t, c.Log.Output, "both")
	must.Equal(t, c.Code.Expire, 5*time.Minute)
	must.DeepEqual(t, c.CDN.Drivers, []string{"baishan"})
	must.Equal(t, c.OAuth.BaseURL, "https://account.haozi.net")
}

func TestLoadRejectsBadValues(t *testing.T) {
	for name, tc := range map[string]struct {
		from, to string // replaced in minimal; empty from appends to
		wantErr  string
	}{
		"short key":          {from: "a-long-string-with-32-characters", to: "short", wantErr: "app.key"},
		"missing address":    {from: `":3000"`, to: `""`, wantErr: "http.address"},
		"missing domain":     {from: `"weavatar.test"`, to: `""`, wantErr: "http.domain"},
		"domain with scheme": {from: `"weavatar.test"`, to: `"https://weavatar.test"`, wantErr: "http.domain"},
		"missing dsn":        {from: "postgres://postgres:postgres@127.0.0.1:5432/weavatar?sslmode=disable", to: "", wantErr: "database.dsn"},
		"bad log level":      {to: "log:\n  level: verbose\n", wantErr: "log.level"},
		"bad log output":     {to: "log:\n  output: syslog\n", wantErr: "log.output"},
		"stdout log output":  {to: "log:\n  output: stdout\n", wantErr: "log.output"},
		"short code expire":  {to: "code:\n  expire: 30s\n", wantErr: "code.expire"},
		"bad audit driver":   {to: "audit:\n  driver: baidu\n", wantErr: "audit.driver"},
	} {
		t.Run(name, func(t *testing.T) {
			yaml := minimal + tc.to
			if tc.from != "" {
				yaml = strings.Replace(minimal, tc.from, tc.to, 1)
			}
			writeConfig(t, yaml)
			_, err := conf.Load()
			must.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestLoadMissingFileFails(t *testing.T) {
	t.Setenv("APP_CONFIG", filepath.Join(t.TempDir(), "absent.yml"))
	_, err := conf.Load()
	must.Error(t, err)
}
