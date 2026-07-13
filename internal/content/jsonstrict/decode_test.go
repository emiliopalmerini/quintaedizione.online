package jsonstrict

import "testing"

func TestDecodeAcceptsOneStrictObject(t *testing.T) {
	var target struct {
		Version int `json:"version"`
	}
	if err := Decode([]byte(`{"version":1}`), &target); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if target.Version != 1 {
		t.Fatalf("Version = %d, want 1", target.Version)
	}
}

func TestDecodeRejectsAmbiguousJSON(t *testing.T) {
	tests := []string{
		`{"version":1,"version":2}`,
		`{"version":1,"Version":2}`,
		`{"version":1,"nested":{"path":"a","path":"b"}}`,
		`{"version":1,"unknown":true}`,
		`{"version":1} {}`,
	}

	for _, data := range tests {
		t.Run(data, func(t *testing.T) {
			var target struct {
				Version int `json:"version"`
			}
			if err := Decode([]byte(data), &target); err == nil {
				t.Fatal("Decode() succeeded with ambiguous JSON")
			}
		})
	}
}

func TestValidateAllowsUnknownFieldsButRejectsDuplicateKeys(t *testing.T) {
	if err := Validate([]byte(`{"future":true}`)); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := Validate([]byte(`{"future":true,"future":false}`)); err == nil {
		t.Fatal("Validate() succeeded with duplicate keys")
	}
}
