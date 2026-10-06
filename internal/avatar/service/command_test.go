package service_test

import (
	"path/filepath"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/avatar/service"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
)

func TestHashCommandDirFlagOverridesConfig(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "configured")
	flagged := filepath.Join(t.TempDir(), "flagged")

	err := service.HashCommand(appinfo.HashDir(configured)).Run(t.Context(), []string{"hash", "stat"})
	must.Error(t, err)
	check.Contains(t, err.Error(), configured)

	err = service.HashCommand(appinfo.HashDir(configured)).Run(t.Context(), []string{"hash", "stat", "--dir", flagged})
	must.Error(t, err)
	check.Contains(t, err.Error(), flagged)
}
