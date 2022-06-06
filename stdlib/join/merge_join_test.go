package join

import "testing"

func TestMergeJoin(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "inner",
		},
		{
			name: "left",
		},
		{
			name: "right",
		},
		{
			name: "full",
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
	}
}
