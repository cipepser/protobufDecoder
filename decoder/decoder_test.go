package decoder

import (
	"reflect"
	"testing"
)

func TestUnmarshalPerson(t *testing.T) {

	tests := []struct {
		b      []byte
		expect Person
	}{
		{
			b: []byte("0a070a05416c69636512020814"),
			expect: Person{
				Name: &Name{
					Value: "Alice",
				},
				Age: &Age{
					Value: 20,
				},
			},
		},
	}

	for i, tt := range tests {
		p := Person{}.Unmarshal(tt.b)
		if reflect.DeepEqual(p, tt.expect) {
			t.Fatalf("test[%d - failed to Unmarshal. expected=%q, got=%q", i, tt.expect, p)
		}
	}
}
