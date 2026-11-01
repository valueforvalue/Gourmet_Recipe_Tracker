package main

import "testing"

func TestSplitLineLinks(t *testing.T) {
	cases := []struct {
		in   string
		want []linkSegment
	}{
		{
			"Source: https://example.com/x",
			[]linkSegment{{text: "Source: ", isURL: false}, {text: "https://example.com/x", isURL: true}},
		},
		{
			"just plain text",
			[]linkSegment{{text: "just plain text", isURL: false}},
		},
		{
			"See https://a.com and https://b.com",
			[]linkSegment{
				{text: "See ", isURL: false},
				{text: "https://a.com", isURL: true},
				{text: " and ", isURL: false},
				{text: "https://b.com", isURL: true},
			},
		},
	}

	for _, c := range cases {
		got := splitLineLinks(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("splitLineLinks(%q) = %d segments, want %d", c.in, len(got), len(c.want))
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("splitLineLinks(%q)[%d] = %+v, want %+v", c.in, i, got[i], c.want[i])
			}
		}
	}
}
