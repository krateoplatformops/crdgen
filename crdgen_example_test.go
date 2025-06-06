//go:build integration
// +build integration

package crdgen_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/krateoplatformops/crdgen"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDuplicateStructs(t *testing.T) {
	os.Setenv("CRDGEN_CLEAN_WORKDIR", "FALSE")
	opts := crdgen.Options{
		//Verbose: true,
		Managed: true,
		WorkDir: "xapp",
		GVK: schema.GroupVersionKind{
			Group:   "example.org",
			Version: "v1alpha1",
			Kind:    "Xapp",
		},
		SpecJsonSchemaGetter: &fileJsonSchemaGetter{"./testdata/duplicate.structs.schema.json"},
		//StatusJsonSchemaGetter: &fileJsonSchemaGetter{"./testdata/hello.status.schema.json"},
	}

	res := crdgen.Generate(context.TODO(), opts)
	if res.Err != nil {
		t.Fatal(res.Err)
	}

	fmt.Println(res.WorkDir)
	fmt.Println()

	//fmt.Println("digest: ", res.Digest)
	//fmt.Println()

	fmt.Println(string(res.Manifest))
}

func TestExample(t *testing.T) {
	//os.Setenv("CRDGEN_CLEAN_WORKDIR", "FALSE")
	opts := crdgen.Options{
		//Verbose: true,
		Managed: true,
		WorkDir: "xapp",
		GVK: schema.GroupVersionKind{
			Group:   "example.org",
			Version: "v1alpha1",
			Kind:    "Xapp",
		},
		SpecJsonSchemaGetter: &fileJsonSchemaGetter{"./testdata/issue.43.hack.json"},
		//StatusJsonSchemaGetter: &fileJsonSchemaGetter{"./testdata/hello.status.schema.json"},
	}

	res := crdgen.Generate(context.TODO(), opts)
	if res.Err != nil {
		t.Fatal(res.Err)
	}

	fmt.Println(res.WorkDir)
	fmt.Println()

	//fmt.Println("digest: ", res.Digest)
	//fmt.Println()

	fmt.Println(string(res.Manifest))
}

var _ crdgen.JsonSchemaGetter = (*fileJsonSchemaGetter)(nil)

type fileJsonSchemaGetter struct {
	filename string
}

func (f *fileJsonSchemaGetter) Get() ([]byte, error) {
	fin, err := os.Open(f.filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	return io.ReadAll(fin)
}
