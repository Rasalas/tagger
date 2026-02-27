package semver

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"v1.2.3", Version{1, 2, 3}, false},
		{"1.2.3", Version{1, 2, 3}, false},
		{"v0.0.0", Version{0, 0, 0}, false},
		{"v10.20.30", Version{10, 20, 30}, false},
		{"v1.2", Version{}, true},
		{"invalid", Version{}, true},
		{"v1.2.x", Version{}, true},
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

func TestBump(t *testing.T) {
	v := Version{1, 2, 3}
	tests := []struct {
		level BumpLevel
		want  Version
	}{
		{Major, Version{2, 0, 0}},
		{Minor, Version{1, 3, 0}},
		{Patch, Version{1, 2, 4}},
		{None, Version{1, 2, 3}},
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
	v := Version{1, 2, 3}
	if got := v.Format("v"); got != "v1.2.3" {
		t.Errorf("Format(v) = %q, want %q", got, "v1.2.3")
	}
	if got := v.Format(""); got != "1.2.3" {
		t.Errorf("Format(\"\") = %q, want %q", got, "1.2.3")
	}
}

func TestString(t *testing.T) {
	v := Version{0, 1, 0}
	if got := v.String(); got != "0.1.0" {
		t.Errorf("String() = %q, want %q", got, "0.1.0")
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
