package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/assert/must"
)

// TestGeneratedGraphs needs no database: the PostgreSQL handle connects
// lazily and migrations are not run.
func TestGeneratedGraphs(t *testing.T) {
	testConfig(t)

	application, cleanup, err := InitializeApp("test")
	must.NoError(t, err)
	must.NotNil(t, application)

	resp, err := application.router.Test(httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	must.NoError(t, err)
	must.Equal(t, resp.StatusCode, 200)
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	must.NoError(t, resp.Body.Close())
	must.Contains(t, string(body), `"version": "test"`)
	must.Contains(t, string(body), `"title": "weavatar"`)

	resp, err = application.router.Test(httptest.NewRequest(http.MethodGet, "/healthz", nil))
	must.NoError(t, err)
	must.Equal(t, resp.StatusCode, 200)
	must.NoError(t, resp.Body.Close())

	must.NoError(t, cleanup())
	must.NoError(t, cleanup(), "generated cleanup must be idempotent")

	management, cleanupCLI, err := InitializeCLI()
	must.NoError(t, err)
	must.NotNil(t, management)
	must.NoError(t, cleanupCLI())
	must.NoError(t, cleanupCLI(), "generated cleanup must be idempotent")
}

func TestUserCleanupsAreMerged(t *testing.T) {
	testConfig(t)

	cleanups, cleanup, err := InitializeUserCleanups()
	must.NoError(t, err)
	names := make([]string, len(cleanups))
	for i, c := range cleanups {
		names[i] = c.Name
		must.NotNil(t, c.Run, c.Name)
	}
	must.DeepEqual(t, names, []string{"avatar"})
	must.NoError(t, cleanup())
}

// testConfig rewrites the example config so side effects land in a temp dir;
// config values come from the file only, so the test edits the file.
func testConfig(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()

	k := koanf.New(".")
	must.NoError(t, k.Load(file.Provider("../../config/config.example.yml"), yaml.Parser()))
	for key, value := range map[string]any{
		"log.output": "file",
		"log.path":   filepath.Join(tmp, "test.log"),
		"hash.dir":   filepath.Join(tmp, "hash"),
		"http.docs":  true,
	} {
		must.NoError(t, k.Set(key, value))
	}
	data, err := k.Marshal(yaml.Parser())
	must.NoError(t, err)

	path := filepath.Join(tmp, "config.yml")
	must.NoError(t, os.WriteFile(path, data, 0o600))
	t.Setenv("APP_CONFIG", path)
}
