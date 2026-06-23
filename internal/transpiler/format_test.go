package transpiler_test

import (
	"testing"

	"github.com/krateoplatformops/crdgen/internal/transpiler"
	"github.com/krateoplatformops/crdgen/internal/transpiler/jsonschema"
)

// TestIntegerFormatIsHonoured guards against regressing the int64 -> int32
// downgrade (krateoplatformops/crdgen#45): an integer field declared with
// "format": "int64" must transpile to a Go int64 so the generated CRD keeps
// "format: int64" and the kube-apiserver accepts values above 2^31-1 (e.g.
// byte-sized fields such as disk size or instance memory).
func TestIntegerFormatIsHonoured(t *testing.T) {
	cases := []struct {
		name     string
		typ      string
		format   string
		wantType string
	}{
		{"int64 keeps width", "integer", "int64", "int64"},
		{"int32 stays int32", "integer", "int32", "int32"},
		{"unformatted integer is left as int", "integer", "", "int"},
		{"double is float64", "number", "double", "float64"},
		{"float is float32", "number", "float", "float32"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := jsonschema.Schema{
				SchemaType: "http://json-schema.org/draft-04/schema#",
				Title:      "TestFormat",
				Properties: map[string]*jsonschema.Schema{
					"value": {
						TypeValue: tc.typ,
						Format:    tc.format,
					},
				},
			}
			root.Init()

			structs, err := transpiler.Transpile(&root)
			if err != nil {
				t.Fatalf("transpile: %v", err)
			}

			st, ok := structs["Root"]
			if !ok {
				t.Fatalf("expected a Root struct, got %v", keys(structs))
			}
			f, ok := st.Fields["Value"]
			if !ok {
				t.Fatalf("expected a Value field, got %v", st.Fields)
			}
			if f.Type != tc.wantType {
				t.Errorf("format %q on %q: got Go type %q, want %q",
					tc.format, tc.typ, f.Type, tc.wantType)
			}
		})
	}
}

func keys(m map[string]transpiler.Struct) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
