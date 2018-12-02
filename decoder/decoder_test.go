package decoder

import (
	"testing"
)

func TestUnmarshalPerson(t *testing.T) {

	tests := []struct {
		b      []byte
		expect Person
	}{
		{
			b: []byte{10, 7, 10, 5, 65, 108, 105, 99, 101, 18, 2, 8, 20}, // "0a070a05416c69636512020814"
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
		p := Person{}
		if err := p.Unmarshal(tt.b); err != nil {
			t.Fatalf("test[%d - failed to Unmarshal. got err:%q", i, err)
		}
		if p.Name.Value != tt.expect.Name.Value || p.Age.Value != tt.expect.Age.Value {
			t.Fatalf("test[%d - failed to Unmarshal. expected=%q, got=%q", i, tt.expect, p)
		}
	}
}
