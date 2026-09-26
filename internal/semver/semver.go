package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// BumpLevel indicates the type of version bump.
type BumpLevel int

const (
	None BumpLevel = iota
	Patch
	Minor
	Major
)

func (b BumpLevel) String() string {
	switch b {
	case Patch:
		return "patch"
	case Minor:
		return "minor"
	case Major:
		return "major"
	default:
		return "none"
	}
}

// Version represents a semantic version.
type Version struct {
	Major    int
	Minor    int
	Patch    int
	Build    int
	HasBuild bool
}

// Parse parses a version string, stripping an optional "v" prefix.
func Parse(tag string) (Version, error) {
	return parse(strings.TrimPrefix(tag, "v"), tag)
}

// ParseWithPrefix parses a version string, stripping the configured tag prefix.
func ParseWithPrefix(tag, prefix string) (Version, error) {
	s := tag
	if prefix != "" {
		if !strings.HasPrefix(tag, prefix) {
			return Version{}, fmt.Errorf("invalid prefix: %q", tag)
		}
		s = strings.TrimPrefix(tag, prefix)
	}
	return parse(s, tag)
}

func parse(s, tag string) (Version, error) {
	base, build, hasBuild := strings.Cut(s, "-")
	if hasBuild {
		if build == "" {
			return Version{}, fmt.Errorf("invalid build: %q", tag)
		}
		if strings.Contains(build, "-") {
			return Version{}, fmt.Errorf("invalid build: %q", tag)
		}
	}
	parts := strings.SplitN(base, ".", 3)
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version: %q", tag)
	}
	maj, err := parseDecimal(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major: %q", tag)
	}
	min, err := parseDecimal(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor: %q", tag)
	}
	pat, err := parseDecimal(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch: %q", tag)
	}
	v := Version{Major: maj, Minor: min, Patch: pat}
	if hasBuild {
		n, err := parseDecimal(build)
		if err != nil {
			return Version{}, fmt.Errorf("invalid build: %q", tag)
		}
		v.Build = n
		v.HasBuild = true
	}
	return v, nil
}

func parseDecimal(s string) (int, error) {
	if s == "" || len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("noncanonical decimal %q", s)
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("non-decimal character in %q", s)
		}
	}
	return strconv.Atoi(s)
}

// String returns the version without prefix, e.g. "1.2.3".
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.HasBuild {
		return fmt.Sprintf("%s-%d", s, v.Build)
	}
	return s
}

// Format returns the version with the given prefix, e.g. "v1.2.3".
func (v Version) Format(prefix string) string {
	return prefix + v.String()
}

// Bump returns a new version or an error if an increment would overflow.
func (v Version) Bump(level BumpLevel) (Version, error) {
	if level == None {
		return v, nil
	}
	maxInt := int(^uint(0) >> 1)
	nextBuild := v.Build
	if v.HasBuild {
		if v.Build == maxInt {
			return Version{}, fmt.Errorf("build counter overflow at %d", v.Build)
		}
		nextBuild++
	}
	switch level {
	case Major:
		if v.Major == maxInt {
			return Version{}, fmt.Errorf("major version overflow at %d", v.Major)
		}
		return Version{Major: v.Major + 1, Build: nextBuild, HasBuild: v.HasBuild}, nil
	case Minor:
		if v.Minor == maxInt {
			return Version{}, fmt.Errorf("minor version overflow at %d", v.Minor)
		}
		return Version{Major: v.Major, Minor: v.Minor + 1, Build: nextBuild, HasBuild: v.HasBuild}, nil
	case Patch:
		if v.Patch == maxInt {
			return Version{}, fmt.Errorf("patch version overflow at %d", v.Patch)
		}
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1, Build: nextBuild, HasBuild: v.HasBuild}, nil
	default:
		return v, nil
	}
}
