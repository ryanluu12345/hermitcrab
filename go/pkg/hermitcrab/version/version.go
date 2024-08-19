package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Version interface defines the methods our custom version type should implement
type Version interface {
	Major() uint64
	Minor() uint64
	Patch() uint64
	String() string
	UI() uint64
	Compare(other Version) int
}

// UIVersion implements the Version interface
type UIVersion struct {
	SemVer  *semver.Version
	UIBuild uint64
}

// Major returns the major version
func (v *UIVersion) Major() uint64 {
	return v.SemVer.Major()
}

// Minor returns the minor version
func (v *UIVersion) Minor() uint64 {
	return v.SemVer.Minor()
}

// Patch returns the patch version
func (v *UIVersion) Patch() uint64 {
	return v.SemVer.Patch()
}

// UI returns the UI build number
func (v *UIVersion) UI() uint64 {
	return v.UIBuild
}

// ParseVersion parses a version string into our custom Version type
func ParseVersion(v string) (Version, error) {
	semVer, uiBuild, err := parseSemVerAndUIBuild(v)
	if err != nil {
		return nil, err
	}

	return &UIVersion{
		SemVer:  semVer,
		UIBuild: uiBuild,
	}, nil
}

// parseSemVerAndUIBuild is a helper function to parse the version string
func parseSemVerAndUIBuild(v string) (*semver.Version, uint64, error) {
	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	// Split the version string into parts
	parts := strings.Split(v, "+")
	if len(parts) != 2 || !strings.HasPrefix(parts[1], "ui.") {
		return nil, 0, fmt.Errorf("invalid version format: %s", v)
	}

	// Parse the semver part
	sv, err := semver.NewVersion(parts[0])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid semver: %w", err)
	}

	// Parse the UI build number
	uiBuildStr := strings.TrimPrefix(parts[1], "ui.")
	uiBuild, err := strconv.Atoi(uiBuildStr)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid UI build number: %s", uiBuildStr)
	}

	return sv, uint64(uiBuild), nil
}

// Compare compares two Version objects
func (v *UIVersion) Compare(other Version) int {
	otherCustom, ok := other.(*UIVersion)
	if !ok {
		panic("Cannot compare with a non-UIVersion type")
	}

	// First, compare the semver parts
	cmp := v.SemVer.Compare(otherCustom.SemVer)
	if cmp != 0 {
		return cmp
	}

	// If semver parts are equal, compare UI build numbers
	return compareInt(v.UIBuild, otherCustom.UIBuild)
}

// compareInt is a helper function to compare two integers
func compareInt(a, b uint64) int {
	if a > b {
		return 1
	} else if a < b {
		return -1
	}
	return 0
}

// String returns a string representation of the version
func (v *UIVersion) String() string {
	return fmt.Sprintf("%s+ui.%d", v.SemVer.String(), v.UIBuild)
}
