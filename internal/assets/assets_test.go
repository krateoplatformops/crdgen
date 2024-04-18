//go:build integration
// +build integration

package assets_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/krateoplatformops/crdgen/internal/assets"
)

func TestRender(t *testing.T) {
	ds := map[string]string{
		"module": "github.com/krateoplatformops/form1",
	}

	err := assets.Render(os.Stdout, "go.mod", ds)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExport(t *testing.T) {
	ds := map[string]string{
		"module": "github.com/krateoplatformops/form1",
	}

	buf := bytes.Buffer{}
	err := assets.Render(&buf, "go.mod", ds)
	if err != nil {
		t.Fatal(err)
	}

	targetDir := "/private/var/folders/qb/ckvcz6s10b590xt0b808977h0000gn/T/github.com/krateoplatformops/form1"
	target := filepath.Join(targetDir, "go.mod")
	err = assets.Export(target, buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}
