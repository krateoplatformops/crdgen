package crdgen

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/krateoplatformops/crdgen/internal/coder"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type JsonSchemaGetter interface {
	Get() ([]byte, error)
}

type Options struct {
	WorkDir          string
	GVK              schema.GroupVersionKind
	Categories       []string
	JsonSchemaGetter JsonSchemaGetter
}

type Result struct {
	Manifest []byte
	Digest   string
	Err      error
}

func Generate(ctx context.Context, opts Options) (res Result) {
	dat, err := opts.JsonSchemaGetter.Get()
	if err != nil {
		res.Err = err
		return
	}

	nfo := coder.Resource{
		Group:      opts.GVK.Group,
		Version:    opts.GVK.Version,
		Kind:       opts.GVK.Kind,
		Schema:     dat,
		Categories: opts.Categories,
	}

	cfg, err := defaultCodeGeneratorOptions(opts.WorkDir)
	if err != nil {
		res.Err = err
		return
	}

	clean := len(os.Getenv("CODEGEN_CLEAN_WORKDIR")) == 0
	if clean {
		defer os.RemoveAll(cfg.Workdir)
	}

	if err := coder.Do(&nfo, cfg); err != nil {
		res.Err = err
		return
	}

	cmd := exec.Command("go", "mod", "init", cfg.Module)
	cmd.Dir = cfg.Workdir
	if err := cmd.Run(); err != nil {
		res.Err = fmt.Errorf("%s: performing 'go mod init' (workdir: %s, module: %s, gvk: %s/%s,%s)",
			err.Error(), cfg.Workdir, cfg.Module, nfo.Group, nfo.Version, nfo.Kind)
		return
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = cfg.Workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			res.Err = fmt.Errorf("%s: performing 'go mod tidy' (workdir: %s, module: %s, gvk: %s/%s,%s)",
				string(out), cfg.Workdir, cfg.Module, nfo.Group, nfo.Version, nfo.Kind)
			return
		}
		res.Err = fmt.Errorf("%s: performing 'go mod tidy' (workdir: %s, module: %s, gvk: %s/%s,%s)",
			err.Error(), cfg.Workdir, cfg.Module, nfo.Group, nfo.Version, nfo.Kind)
		return
	}

	cmd = exec.Command("go",
		"run",
		"--tags",
		"generate",
		"sigs.k8s.io/controller-tools/cmd/controller-gen",
		"object:headerFile=./hack/boilerplate.go.txt",
		"paths=./...", "crd:crdVersions=v1",
		"output:artifacts:config=./crds",
	)
	cmd.Dir = cfg.Workdir
	out, err = cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			res.Err = fmt.Errorf("%s: performing 'go run --tags generate...' (workdir: %s, module: %s, gvk: %s/%s,%s)",
				string(out), cfg.Workdir, cfg.Module, nfo.Group, nfo.Version, nfo.Kind)
			return
		}
		res.Err = fmt.Errorf("%s: performing 'go run --tags generate...' (workdir: %s, module: %s, gvk: %s/%s,%s)",
			err.Error(), cfg.Workdir, cfg.Module, nfo.Group, nfo.Version, nfo.Kind)
		return
	}

	fsys := os.DirFS(cfg.Workdir)
	all, err := fs.ReadDir(fsys, "crds")
	if err != nil {
		res.Err = err
		return
	}

	fp, err := fsys.Open(filepath.Join("crds", all[0].Name()))
	if err != nil {
		res.Err = err
		return
	}
	defer fp.Close()

	res.Manifest, res.Err = io.ReadAll(fp)
	if res.Err != nil {
		return
	}

	h := sha256.New()
	if _, res.Err = h.Write(dat); res.Err != nil {
		return
	}
	res.Digest = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func defaultCodeGeneratorOptions(rootDir string) (opts coder.Options, err error) {
	opts.Module = fmt.Sprintf("github.com/krateoplatformops/%s", rootDir)
	opts.Workdir = filepath.Join(os.TempDir(), opts.Module)
	err = os.MkdirAll(opts.Workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return opts, err
		}
	}

	return opts, nil
}
