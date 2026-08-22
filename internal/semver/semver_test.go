package semver

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"v1.2.3", Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"1.2.3", Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"v0.0.0", Version{}, false},
		{"v10.20.30", Version{Major: 10, Minor: 20, Patch: 30}, false},
		{"v3.4.0-58", Version{Major: 3, Minor: 4, Patch: 0, Build: 58, HasBuild: true}, false},
		{"3.4.0-58", Version{Major: 3, Minor: 4, Patch: 0, Build: 58, HasBuild: true}, false},
		{"v1.2", Version{}, true},
		{"invalid", Version{}, true},
		{"v1.2.x", Version{}, true},
		{"v1.2.3-x", Version{}, true},
		{"v1.2.3-", Version{}, true},
		{"v1.2.3-1-2", Version{}, true},
		{"", Version{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWithPrefix(t *testing.T) {
	got, err := ParseWithPrefix("release-1.2.3-4", "release-")
	if err != nil {
		t.Fatalf("ParseWithPrefix() error = %v", err)
	}
	want := Version{Major: 1, Minor: 2, Patch: 3, Build: 4, HasBuild: true}
	if got != want {
		t.Errorf("ParseWithPrefix() = %v, want %v", got, want)
	}

	if _, err := ParseWithPrefix("v1.2.3", "release-"); err == nil {
		t.Fatal("expected prefix error")
	}
}

func TestBump(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	tests := []struct {
		level BumpLevel
		want  Version
	}{
		{Major, Version{Major: 2}},
		{Minor, Version{Major: 1, Minor: 3}},
		{Patch, Version{Major: 1, Minor: 2, Patch: 4}},
		{None, Version{Major: 1, Minor: 2, Patch: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got := v.Bump(tt.level)
			if got != tt.want {
				t.Errorf("Bump(%v) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestBumpWithBuild(t *testing.T) {
	v := Version{Major: 3, Minor: 4, Patch: 0, Build: 58, HasBuild: true}
	tests := []struct {
		level BumpLevel
		want  Version
	}{
		{Major, Version{Major: 4, Build: 59, HasBuild: true}},
		{Minor, Version{Major: 3, Minor: 5, Build: 59, HasBuild: true}},
		{Patch, Version{Major: 3, Minor: 4, Patch: 1, Build: 59, HasBuild: true}},
		{None, Version{Major: 3, Minor: 4, Patch: 0, Build: 58, HasBuild: true}},
	}
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got := v.Bump(tt.level)
			if got != tt.want {
				t.Errorf("Bump(%v) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	if got := v.Format("v"); got != "v1.2.3" {
		t.Errorf("Format(v) = %q, want %q", got, "v1.2.3")
	}
	if got := v.Format(""); got != "1.2.3" {
		t.Errorf("Format(\"\") = %q, want %q", got, "1.2.3")
	}

	v = Version{Major: 1, Minor: 2, Patch: 3, Build: 4, HasBuild: true}
	if got := v.Format("v"); got != "v1.2.3-4" {
		t.Errorf("Format(v) = %q, want %q", got, "v1.2.3-4")
	}
}

func TestString(t *testing.T) {
	v := Version{Minor: 1}
	if got := v.String(); got != "0.1.0" {
		t.Errorf("String() = %q, want %q", got, "0.1.0")
	}

	v = Version{Major: 3, Minor: 4, Build: 58, HasBuild: true}
	if got := v.String(); got != "3.4.0-58" {
		t.Errorf("String() = %q, want %q", got, "3.4.0-58")
	}
}

func TestBumpLevelString(t *testing.T) {
	tests := []struct {
		level BumpLevel
		want  string
	}{
		{None, "none"},
		{Patch, "patch"},
		{Minor, "minor"},
		{Major, "major"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("BumpLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}
