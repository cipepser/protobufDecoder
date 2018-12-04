package decoder

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func atob(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func TestUnmarshalPerson(t *testing.T) {
	tests := []struct {
		b      []byte
		expect Person
	}{
		{
			b: atob("0a070a05416c69636512020814"),
			expect: Person{
				Name: &Name{Value: "Alice"},
				Age:  &Age{Value: 20},
			},
		},
		{
			b:      atob(""),
			expect: Person{},
		},
		{
			b: atob("0a070a05416c696365"),
			expect: Person{
				Name: &Name{Value: "Alice"},
			},
		},
		{
			b: atob("12020814"),
			expect: Person{
				Age:  &Age{Value: 20},
			},
		},
	}

	for i, tt := range tests {
		p := Person{}
		if err := p.Unmarshal(tt.b); err != nil {
			t.Fatalf("test[%d - failed to Unmarshal. got err:%q", i, err)
		}
		if diff := cmp.Diff(p, tt.expect); diff != "" {
			t.Fatalf("test[%d - failed to Unmarshal. expected=%q, got=%q", i, tt.expect, p)
		}
	}
}
