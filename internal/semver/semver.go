package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// BumpLevel indicates the type of version bump.
type BumpLevel int

const (
	None  BumpLevel = iota
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
	Major int
	Minor int
	Patch int
}

// Parse parses a version string, stripping an optional prefix (e.g. "v").
func Parse(tag string) (Version, error) {
	s := strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version: %q", tag)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major: %q", tag)
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor: %q", tag)
	}
	pat, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch: %q", tag)
	}
	return Version{Major: maj, Minor: min, Patch: pat}, nil
}

// String returns the version without prefix, e.g. "1.2.3".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Format returns the version with the given prefix, e.g. "v1.2.3".
func (v Version) Format(prefix string) string {
	return prefix + v.String()
}

// Bump returns a new version bumped by the given level.
func (v Version) Bump(level BumpLevel) Version {
	switch level {
	case Major:
		return Version{Major: v.Major + 1}
	case Minor:
		return Version{Major: v.Major, Minor: v.Minor + 1}
	case Patch:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	default:
		return v
	}
}
