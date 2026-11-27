package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestCleanSubjects(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "strips think blocks",
			raw:  "<think lang=\"en\">line one\nline two</think lang=\"en\">\nAdd parser",
			want: []string{"Add parser"},
		},
		{
			name: "strips numbering and bullets",
			raw:  "1. Add parser\n2) Fix parser\n- Add tests\n* Fix tests",
			want: []string{"Add parser", "Fix parser", "Add tests", "Fix tests"},
		},
		{
			name: "strips quotes and backticks",
			raw:  "\"Add parser\"\n'Fix parser'\n`Add tests`",
			want: []string{"Add parser", "Fix parser", "Add tests"},
		},
		{
			name: "strips conventional prefixes",
			raw:  "feat: Add parser\nfix(parser)!: Fix nil pointer\nchore!: Bump dependencies\ndocs(readme): Update usage",
			want: []string{"Add parser", "Fix nil pointer", "Bump dependencies", "Update usage"},
		},
		{
			name: "removes trailing period and capitalizes",
			raw:  "add parser.\nFIX PARSER.",
			want: []string{"Add parser", "FIX PARSER"},
		},
		{
			name: "drops empty long and duplicate lines",
			raw: "\n\n-\n*\n" + strings.Repeat("a", 73) +
				"\nAdd parser.\nadd parser",
			want: []string{"Add parser"},
		},
		{
			name: "keeps exactly 72 characters",
			raw:  strings.Repeat("a", 72),
			want: []string{"A" + strings.Repeat("a", 71)},
		},
		{
			name: "capitalizes non-ascii first rune",
			raw:  "éclair support",
			want: []string{"Éclair support"},
		},
		{
			name: "validates a single user-written line",
			raw:  "  1. `feat: add parser.`  ",
			want: []string{"Add parser"},
		},
		{
			name: "empty input yields nothing",
			raw:  "",
			want: nil,
		},
		{
			name: "combined messy output",
			raw: "<think lang=\"en\">\nI should suggest:\n</think lang=\"en\">\n" +
				"1. feat(parser): add validation.\n" +
				"2) - add error handling.\n" +
				"3. `fix: nil pointer in parser.`\n" +
				"4. add validation\n",
			want: []string{"Add validation", "Add error handling", "Nil pointer in parser"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanSubjects(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("cleanSubjects(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}
