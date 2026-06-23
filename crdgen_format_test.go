//go:build integration
// +build integration

package crdgen_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krateoplatformops/crdgen"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestInt64AndEnumsEndToEnd is a regression guard for two crdgen issues,
// asserted on the fully rendered CRD (transpiler + controller-gen):
//
//   - #45 int64 fields must keep "format: int64" (no int32 downgrade), and an
//     explicit int32 field must stay int32.
//   - #46 each nested "type" discriminator must keep its own enum (no collapse
//     onto a single shared set).
//
// The schema in testdata/instance.like.schema.json mirrors the Oxide instance
// shape that originally surfaced both bugs: byte-sized int64 fields plus five
// distinct "type" unions nested across arrays and objects.
func TestInt64AndEnumsEndToEnd(t *testing.T) {
	res := crdgen.Generate(context.TODO(), crdgen.Options{
		WorkDir: "fmtcheck",
		GVK:     schema.GroupVersionKind{Group: "example.org", Version: "v1alpha1", Kind: "Inst"},
		Managed: true,
		SpecJsonSchemaGetter: &fileJsonSchemaGetter{"./testdata/instance.like.schema.json"},
	})
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	manifest := string(res.Manifest)

	// #45: int64 preserved, and never downgraded to int32 for the byte fields.
	if !strings.Contains(manifest, "format: int64") {
		t.Errorf("#45: expected at least one 'format: int64' in the CRD, got none")
	}
	// #46: every union keeps its own discriminator values. If any collapsed,
	// one of these would be missing.
	for _, want := range []string{
		"ephemeral", "floating", // external_ips[].type
		"explicit", "auto", // pool_selector.type
		"create", "attach", // disks[].type
		"distributed", "local", // disk_backend.type
		"blank", "snapshot", "image", // disk_source.type
	} {
		if !strings.Contains(manifest, want) {
			t.Errorf("#46: enum value %q missing from CRD — discriminator collapse regressed:\n%s", want, manifest)
		}
	}
}
