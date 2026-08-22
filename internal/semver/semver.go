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
	v := Version{Major: maj, Minor: min, Patch: pat}
	if hasBuild {
		n, err := strconv.Atoi(build)
		if err != nil {
			return Version{}, fmt.Errorf("invalid build: %q", tag)
		}
		v.Build = n
		v.HasBuild = true
	}
	return v, nil
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

// Bump returns a new version bumped by the given level.
func (v Version) Bump(level BumpLevel) Version {
	nextBuild := v.Build
	if v.HasBuild {
		nextBuild++
	}
	switch level {
	case Major:
		return Version{Major: v.Major + 1, Build: nextBuild, HasBuild: v.HasBuild}
	case Minor:
		return Version{Major: v.Major, Minor: v.Minor + 1, Build: nextBuild, HasBuild: v.HasBuild}
	case Patch:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1, Build: nextBuild, HasBuild: v.HasBuild}
	default:
		return v
	}
}
