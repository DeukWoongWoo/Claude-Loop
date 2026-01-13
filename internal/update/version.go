package update

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVer represents a parsed semantic version.
type SemVer struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string // e.g., "beta.1", "rc.2"
	Build      string // e.g., "20240101"
}

// Pre-compiled regex for semantic version parsing.
// Matches: v1.2.3, 1.2.3, v1.2.3-beta.1, v1.2.3+build
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$`)

// ParseVersion parses a semantic version string.
func ParseVersion(v string) (*SemVer, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, NewVersionError(v, "empty version string", nil)
	}

	matches := semverRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, NewVersionError(v, "invalid format", nil)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &SemVer{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// String returns the version as a string.
func (v *SemVer) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare compares two versions.
// Returns: -1 if v < other, 0 if equal, 1 if v > other.
func (v *SemVer) Compare(other *SemVer) int {
	if v.Major != other.Major {
		return compareInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return compareInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return compareInt(v.Patch, other.Patch)
	}

	// Pre-release versions have lower precedence than release versions
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1 // release > prerelease
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1 // prerelease < release
	}
	if v.Prerelease != other.Prerelease {
		return comparePrerelease(v.Prerelease, other.Prerelease)
	}

	// Build metadata is ignored in precedence
	return 0
}

// IsNewerThan returns true if v is newer than other.
func (v *SemVer) IsNewerThan(other *SemVer) bool {
	return v.Compare(other) > 0
}

// CompareVersions compares two version strings.
// Returns: -1 if v1 < v2, 0 if equal, 1 if v1 > v2.
func CompareVersions(v1, v2 string) (int, error) {
	sv1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}
	sv2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}
	return sv1.Compare(sv2), nil
}

// IsNewer returns true if newVersion is newer than currentVersion.
func IsNewer(currentVersion, newVersion string) (bool, error) {
	cmp, err := CompareVersions(newVersion, currentVersion)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}

// compareInt compares two integers.
func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// comparePrerelease compares pre-release identifiers according to semver spec.
func comparePrerelease(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	maxLen := max(len(partsA), len(partsB))

	for i := 0; i < maxLen; i++ {
		var partA, partB string
		if i < len(partsA) {
			partA = partsA[i]
		}
		if i < len(partsB) {
			partB = partsB[i]
		}

		if partA == "" && partB != "" {
			return -1
		}
		if partA != "" && partB == "" {
			return 1
		}

		numA, errA := strconv.Atoi(partA)
		numB, errB := strconv.Atoi(partB)

		switch {
		case errA == nil && errB == nil:
			if cmp := compareInt(numA, numB); cmp != 0 {
				return cmp
			}
		case errA == nil:
			return -1
		case errB == nil:
			return 1
		default:
			if cmp := strings.Compare(partA, partB); cmp != 0 {
				return cmp
			}
		}
	}

	return 0
}
