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

// testConfig points APP_CONFIG at a copy of the example config whose side
// effects land in a temp dir; the environment cannot override values.
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

// TestGeneratedGraphs never contacts the database: the PostgreSQL handle
// connects lazily and migrations are not run.
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

// TestUserCleanupsAreMerged reads wire_gen.go because a Multibind declared
// inside the user module would silently inject an empty collection.
func TestUserCleanupsAreMerged(t *testing.T) {
	src, err := os.ReadFile("wire_gen.go")
	must.NoError(t, err)
	must.Contains(t, string(src), "make(registry.UserCleanups, 0, 1)")
}
