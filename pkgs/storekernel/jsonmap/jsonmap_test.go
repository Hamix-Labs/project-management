package jsonmap

import (
	"testing"
)

func TestJSONStringSlice_emptyWritesArray(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		in   []string
	}{
		{name: "nil", in: nil},
		{name: "empty", in: []string{}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			j := JSONStringSlice(tc.in)
			v, err := j.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}
			s, ok := v.(string)
			if !ok {
				if b, ok := v.([]byte); ok {
					s = string(b)
				} else {
					t.Fatalf("Value type %T = %#v, want string or []byte", v, v)
				}
			}
			if s != "[]" {
				t.Fatalf("Value = %q, want %q", s, "[]")
			}
			back := StringSliceFromJSON(j)
			if back == nil || len(back) != 0 {
				t.Fatalf("StringSliceFromJSON = %#v, want empty non-nil", back)
			}
		})
	}
}

func TestJSONStringSlice_roundTrip(t *testing.T) {
	t.Parallel()
	j := JSONStringSlice([]string{"a", "b"})
	back := StringSliceFromJSON(j)
	if len(back) != 2 || back[0] != "a" || back[1] != "b" {
		t.Fatalf("got %#v", back)
	}
}
