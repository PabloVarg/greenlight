package validator

import (
	"fmt"
	"testing"
)

func TestUnique(t *testing.T) {
	tbl := []struct {
		list   []string
		expect bool
	}{
		{
			list:   []string{"a", "b", "c"},
			expect: true,
		},
		{
			list:   []string{"a", "a", "b", "c"},
			expect: false,
		},
		{
			list:   []string{"a", "A", "b", "c"},
			expect: false,
		},
		{
			list:   []string{"a", "a"},
			expect: false,
		},
	}

	for i, test := range tbl {
		t.Run(fmt.Sprintf("Case %d", i+1), func(t *testing.T) {
			ok := Unique(test.list)

			if ok != test.expect {
				t.Fatal()
			}
		})
	}
}
