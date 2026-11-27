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
		{
			name: "drops preamble ending in a colon",
			raw:  "Here are 2 commit subjects:\n\nAdd parser\nFix parser",
			want: []string{"Add parser", "Fix parser"},
		},
		{
			name: "strips markdown bold and headings",
			raw:  "**Add parser**\n# Fix tests\n## \n#123 stays",
			want: []string{"Add parser", "Fix tests", "#123 stays"},
		},
		{
			name: "drops reasoning after an unclosed think tag",
			raw:  "Add parser\n<think>\nI should add a parser",
			want: []string{"Add parser"},
		},
		{
			name: "drops reasoning before a stray closing think tag",
			raw:  "I should add a parser\n</think>\nAdd parser",
			want: []string{"Add parser"},
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

func TestCleanInput(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "keeps a plain subject", in: "add parser.", want: "Add parser"},
		{name: "rejects a conventional prefix", in: "Fix: crash on empty input", wantErr: true},
		{name: "rejects a scoped prefix", in: "feat(cli): add flag", wantErr: true},
		{name: "rejects a trailing colon", in: "Update docs:", wantErr: true},
		{name: "rejects an overlong subject", in: strings.Repeat("a", maxSubjectLen+1), wantErr: true},
		{name: "keeps a word that starts like a prefix", in: "Fixture cleanup", want: "Fixture cleanup"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanInput(tt.in)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("cleanInput(%q) = %q, %v; want %q, error %v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}
