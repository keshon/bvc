package file

import (
	"path/filepath"
	"testing"
)

// helper for pattern test
func match(pat, path string) bool {
	return matchPattern(pat, filepath.ToSlash(path))
}

// TestMatchPattern_Basics tests the Match function with basic patterns such as exact, wildcard, single-char ?, nested paths, double-star recursive, double-star in middle, trailing slash pattern, mixed wildcards, prefix with **/ pattern, and double wildcards at both ends, and pattern with static prefix.
func TestMatchPattern_Basics(t *testing.T) {
	cases := []struct {
		pat  string
		path string
		want bool
	}{
		// exact
		{"foo.txt", "foo.txt", true},
		{"foo.txt", "bar.txt", false},

		// wildcard *
		{"*.txt", "foo.txt", true},
		{"*.txt", "bar.log", false},
		{"foo*", "foobar", true},
		{"foo*", "barfoo", false},

		// single-char ?
		{"file?.txt", "file1.txt", true},
		{"file?.txt", "file12.txt", false},

		// nested paths
		{"dir/*.txt", "dir/foo.txt", true},
		{"dir/*.txt", "dir/sub/foo.txt", false},

		// double-star recursive
		{"dir/**", "dir/foo.txt", true},
		{"dir/**", "dir/sub/foo.txt", true},
		{"dir/**", "dir/sub/deep/foo.txt", true},
		{"dir/**", "other/foo.txt", false},

		// double-star in middle
		{"dir/**/foo.txt", "dir/foo.txt", true},
		{"dir/**/foo.txt", "dir/sub/foo.txt", true},
		{"dir/**/foo.txt", "dir/a/b/c/foo.txt", true},
		{"dir/**/foo.txt", "dir/bar/baz.txt", false},

		// trailing slash pattern
		{"dir/**/", "dir/sub", true},
		{"dir/**/", "dir/sub/deep", true},
		{"dir/**/", "other", false},

		// mixed wildcards
		{"**/*.txt", "a.txt", true},
		{"**/*.txt", "a/b/c.txt", true},
		{"**/*.txt", "a/b/c.log", false},

		// prefix with **/ pattern
		{"**/foo.txt", "foo.txt", true},
		{"**/foo.txt", "a/foo.txt", true},
		{"**/foo.txt", "a/b/c/foo.txt", true},
		{"**/foo.txt", "a/b/c/bar.txt", false},

		// double wildcards at both ends
		{"**/*.log", "foo/bar/baz.log", true},
		{"**/*.log", "foo/bar/baz.txt", false},

		// pattern with static prefix
		{"config/*.yml", "config/test.yml", true},
		{"config/*.yml", "config/sub/test.yml", false},

		// deep directory ignoring
		{"build/**", "build/output/file.txt", true},
		{"build/**", "build/file.txt", true},
		{"build/**", "docs/file.txt", false},
	}

	for _, tt := range cases {
		got := match(tt.pat, tt.path)
		if got != tt.want {
			t.Errorf("pattern %q path %q => got %v, want %v", tt.pat, tt.path, got, tt.want)
		}
	}
}

// TestIgnore_StaticAndPatterns tests the Match function with a mix of static and pattern-based ignores.
// It tests both exact matches and pattern matches, and ensures that the ignore logic works as expected.
func TestIgnore_StaticAndPatterns(t *testing.T) {
	m := &Ignore{
		static: map[string]bool{
			"exact.txt": true,
			"temp.log":  true,
		},
		pattern: []string{
			"*.bak",
			"logs/**",
			"**/*.tmp",
		},
	}

	cases := []struct {
		path string
		want bool
	}{
		// static
		{"exact.txt", true},
		{"temp.log", true},
		{"something.log", false},

		// wildcard
		{"foo.bak", true},
		{"bar.txt", false},

		// recursive logs
		{"logs/file.log", true},
		{"logs/sub/deep.txt", true},
		{"notlogs/file.log", false},

		// recursive tmp
		{"foo.tmp", true},
		{"deep/dir/file.tmp", true},
		{"deep/dir/file.txt", false},
	}

	for _, tt := range cases {
		got := m.Match(tt.path)
		if got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// TestMatchPattern_WeirdCases tests Match with unusual patterns such as empty string, double stars at start, and partial directories.
func TestMatchPattern_WeirdCases(t *testing.T) {
	cases := []struct {
		pat, path string
		want      bool
	}{
		// empty
		{"", "", true},
		{"", "foo", false},

		// double stars at start
		{"**", "foo/bar", true},
		{"**", "", true},

		// partial dirs
		{"foo/**/bar", "foo/bar", true},
		{"foo/**/bar", "foo/x/bar", true},
		{"foo/**/bar", "foo/x/y/z/bar", true},
		{"foo/**/bar", "bar/foo/bar", false},
	}

	for _, tt := range cases {
		got := match(tt.pat, tt.path)
		if got != tt.want {
			t.Errorf("pattern %q path %q => got %v, want %v", tt.pat, tt.path, got, tt.want)
		}
	}
}
