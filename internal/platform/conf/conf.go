// Package conf loads the typed application configuration.
package conf

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/pkg/audit"
	"github.com/weavatar/weavatar/pkg/cdn"
	"github.com/weavatar/weavatar/pkg/sms"
)

// exampleKey ships in config.example.yml and must never be used in production.
const exampleKey = "a-long-string-with-32-characters"

type Config struct {
	App      App          `koanf:"app"`
	HTTP     HTTP         `koanf:"http"`
	Log      Log          `koanf:"log"`
	Database Database     `koanf:"database"`
	Hash     Hash         `koanf:"hash"`
	Gravatar Gravatar     `koanf:"gravatar"`
	Mail     Mail         `koanf:"mail"`
	Geetest  Geetest      `koanf:"geetest"`
	Code     Code         `koanf:"code"`
	SMS      sms.Config   `koanf:"sms"`
	OAuth    OAuth        `koanf:"oauth"`
	Audit    audit.Config `koanf:"audit"`
	CDN      cdn.Config   `koanf:"cdn"`
}

type App struct {
	Name string `koanf:"name"`
	// Key is the 32-byte secret that signs login tokens.
	Key string `koanf:"key"`
	// Locale selects validator messages (zh_Hans, zh_Hant, ja, ko, es, ru); default English.
	Locale string `koanf:"locale"`
	// Debug skips the geetest captcha check; never enable it in production.
	Debug bool `koanf:"debug"`
}

type HTTP struct {
	Debug   bool   `koanf:"debug"`
	Address string `koanf:"address"`
	// Domain is the public host used for avatar URLs, OAuth callbacks and CDN purges.
	Domain string `koanf:"domain"`
	// DebugAddress serves pprof/expvar on a private port when set; never expose it.
	DebugAddress string `koanf:"debug_address"`
	// CorsOrigins allows cross-origin requests; empty = same-origin only.
	CorsOrigins []string `koanf:"cors_origins"`
	// ProxyHeader names the header carrying the client IP behind a reverse
	// proxy, e.g. X-Real-IP behind nginx; empty uses the TCP peer address.
	// It is believed from any peer, so never expose the port directly.
	ProxyHeader string `koanf:"proxy_header"`
	// Docs serves the OpenAPI document and UI at /openapi.json and /docs.
	Docs bool `koanf:"docs"`
	// BodyLimit is the maximum request body size in KB; zero means 4096.
	BodyLimit int `koanf:"body_limit"`
	// HeaderLimit is the maximum request header size in bytes.
	HeaderLimit       int           `koanf:"header_limit"`
	ReadTimeout       time.Duration `koanf:"read_timeout"`
	WriteTimeout      time.Duration `koanf:"write_timeout"`
	IdleTimeout       time.Duration `koanf:"idle_timeout"`
	ReduceMemoryUsage bool          `koanf:"reduce_memory_usage"`
}

type Log struct {
	// Level is one of debug | info | warn | error.
	Level string `koanf:"level"`
	// Output is one of file | stderr | both; stdout is left to command output.
	Output string `koanf:"output"`
	Path   string `koanf:"path"`
}

type Database struct {
	// Debug logs every statement.
	Debug bool `koanf:"debug"`
	// DSN is a PostgreSQL connection string (URL or keyword/value form).
	DSN string `koanf:"dsn"`
	// Pool knobs; zero keeps the driver default.
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
}

type Hash struct {
	// Dir holds the QQ hash tables; missing tables only disable the QQ fallback.
	Dir string `koanf:"dir"`
}

type Gravatar struct {
	// URL is the Gravatar origin or a mirror of it, reachable from the server;
	// avatars are fetched from <url>/avatar/<hash>.
	URL string `koanf:"url"`
}

type Mail struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

type Geetest struct {
	ID  string `koanf:"id"`
	Key string `koanf:"key"`
}

type Code struct {
	// Expire is how long a verification code stays valid.
	Expire time.Duration `koanf:"expire"`
}

type OAuth struct {
	BaseURL      string `koanf:"base_url"`
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}

// Load reads $APP_CONFIG (default config/config.yml). Values come from the
// file only; the environment never overrides them.
func Load() (*Config, error) {
	path := os.Getenv("APP_CONFIG")
	if path == "" {
		path = "config/config.yml"
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}

	conf := &Config{}
	if err := k.UnmarshalWithConf("", conf, koanf.UnmarshalConf{
		Tag: "koanf",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
			Result:           conf,
			WeaklyTypedInput: true,
		},
	}); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	conf.fillDefaults()
	if err := conf.check(); err != nil {
		return nil, err
	}

	return conf, nil
}

// SlogLevel returns the parsed slog level; check guarantees it is valid.
func (l Log) SlogLevel() slog.Level {
	var level slog.Level
	_ = level.UnmarshalText([]byte(l.Level))
	return level
}

func (c *Config) fillDefaults() {
	if c.HTTP.BodyLimit <= 0 {
		c.HTTP.BodyLimit = 4096
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Output == "" {
		c.Log.Output = "file"
	}
	if c.Log.Path == "" {
		c.Log.Path = "storage/logs/app.log"
	}
	if c.Hash.Dir == "" {
		c.Hash.Dir = "storage/hash"
	}
	if c.Gravatar.URL == "" {
		c.Gravatar.URL = "https://gravatar.com"
	}
	c.Gravatar.URL = strings.TrimRight(c.Gravatar.URL, "/")
	if c.Code.Expire <= 0 {
		c.Code.Expire = 5 * time.Minute
	}
	if c.Audit.Driver == "" {
		c.Audit.Driver = "aliyun"
	}
	c.OAuth.BaseURL = strings.TrimRight(c.OAuth.BaseURL, "/")
}

func (c *Config) check() error {
	if len(c.App.Key) != 32 {
		return fmt.Errorf("app.key must be exactly 32 characters, got %d", len(c.App.Key))
	}
	if c.App.Key == exampleKey {
		fmt.Fprintln(os.Stderr, "[WARN] app.key is still the example value, generate your own before deploying")
	}
	if c.HTTP.Address == "" {
		return errors.New("http.address must not be empty")
	}
	if c.HTTP.Domain == "" {
		return errors.New("http.domain must not be empty")
	}
	if strings.Contains(c.HTTP.Domain, "/") {
		return fmt.Errorf("http.domain must be a bare host such as weavatar.com, got %q", c.HTTP.Domain)
	}
	if !strings.HasPrefix(c.Gravatar.URL, "http://") && !strings.HasPrefix(c.Gravatar.URL, "https://") {
		return fmt.Errorf("gravatar.url must start with http:// or https://, got %q", c.Gravatar.URL)
	}
	if c.Database.DSN == "" {
		return errors.New("database.dsn must not be empty")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log.level must be debug, info, warn or error, got %q", c.Log.Level)
	}
	switch c.Log.Output {
	case "file", "stderr", "both":
	default:
		return fmt.Errorf("log.output must be file, stderr or both, got %q", c.Log.Output)
	}
	// the SMS templates render the expiry in whole minutes
	if c.Code.Expire < time.Minute {
		return fmt.Errorf("code.expire must be at least 1m, got %s", c.Code.Expire)
	}
	switch c.Audit.Driver {
	case "aliyun", "cos":
	default:
		return fmt.Errorf("audit.driver must be aliyun or cos, got %q", c.Audit.Driver)
	}

	return nil
}
