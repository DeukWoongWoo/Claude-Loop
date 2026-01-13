package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       *SemVer
		wantErr    bool
		errMessage string
	}{
		{
			name:  "simple version with v prefix",
			input: "v1.2.3",
			want:  &SemVer{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "simple version without v prefix",
			input: "1.2.3",
			want:  &SemVer{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "version with prerelease",
			input: "v1.0.0-beta.1",
			want:  &SemVer{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta.1"},
		},
		{
			name:  "version with build metadata",
			input: "v1.0.0+20240101",
			want:  &SemVer{Major: 1, Minor: 0, Patch: 0, Build: "20240101"},
		},
		{
			name:  "version with prerelease and build",
			input: "v1.0.0-rc.2+build.123",
			want:  &SemVer{Major: 1, Minor: 0, Patch: 0, Prerelease: "rc.2", Build: "build.123"},
		},
		{
			name:  "zero version",
			input: "v0.0.0",
			want:  &SemVer{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "large numbers",
			input: "v100.200.300",
			want:  &SemVer{Major: 100, Minor: 200, Patch: 300},
		},
		{
			name:  "version with alpha prerelease",
			input: "v1.0.0-alpha",
			want:  &SemVer{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha"},
		},
		{
			name:  "version with whitespace",
			input: "  v1.2.3  ",
			want:  &SemVer{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:       "empty string",
			input:      "",
			wantErr:    true,
			errMessage: "empty version string",
		},
		{
			name:       "invalid - only major",
			input:      "v1",
			wantErr:    true,
			errMessage: "invalid format",
		},
		{
			name:       "invalid - major.minor only",
			input:      "v1.2",
			wantErr:    true,
			errMessage: "invalid format",
		},
		{
			name:       "invalid - four parts",
			input:      "v1.2.3.4",
			wantErr:    true,
			errMessage: "invalid format",
		},
		{
			name:       "invalid - letters in version",
			input:      "vabc.def.ghi",
			wantErr:    true,
			errMessage: "invalid format",
		},
		{
			name:       "invalid - random string",
			input:      "not-a-version",
			wantErr:    true,
			errMessage: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Major, got.Major)
			assert.Equal(t, tt.want.Minor, got.Minor)
			assert.Equal(t, tt.want.Patch, got.Patch)
			assert.Equal(t, tt.want.Prerelease, got.Prerelease)
			assert.Equal(t, tt.want.Build, got.Build)
		})
	}
}

func TestSemVer_String(t *testing.T) {
	tests := []struct {
		name string
		v    *SemVer
		want string
	}{
		{
			name: "simple version",
			v:    &SemVer{Major: 1, Minor: 2, Patch: 3},
			want: "v1.2.3",
		},
		{
			name: "with prerelease",
			v:    &SemVer{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta.1"},
			want: "v1.0.0-beta.1",
		},
		{
			name: "with build",
			v:    &SemVer{Major: 1, Minor: 0, Patch: 0, Build: "build.123"},
			want: "v1.0.0+build.123",
		},
		{
			name: "with both",
			v:    &SemVer{Major: 2, Minor: 1, Patch: 0, Prerelease: "rc.1", Build: "20240101"},
			want: "v2.1.0-rc.1+20240101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.v.String())
		})
	}
}

func TestSemVer_Compare(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		// Basic comparisons
		{"equal versions", "v1.0.0", "v1.0.0", 0},
		{"major difference", "v2.0.0", "v1.0.0", 1},
		{"major difference reverse", "v1.0.0", "v2.0.0", -1},
		{"minor difference", "v1.1.0", "v1.0.0", 1},
		{"minor difference reverse", "v1.0.0", "v1.1.0", -1},
		{"patch difference", "v1.0.1", "v1.0.0", 1},
		{"patch difference reverse", "v1.0.0", "v1.0.1", -1},

		// Complex comparisons
		{"1.0.0 < 1.0.1", "v1.0.0", "v1.0.1", -1},
		{"1.0.1 < 1.1.0", "v1.0.1", "v1.1.0", -1},
		{"1.1.0 < 2.0.0", "v1.1.0", "v2.0.0", -1},

		// Prerelease comparisons
		{"release > prerelease", "v1.0.0", "v1.0.0-alpha", 1},
		{"prerelease < release", "v1.0.0-alpha", "v1.0.0", -1},
		{"alpha < beta", "v1.0.0-alpha", "v1.0.0-beta", -1},
		{"beta > alpha", "v1.0.0-beta", "v1.0.0-alpha", 1},
		{"alpha.1 < alpha.2", "v1.0.0-alpha.1", "v1.0.0-alpha.2", -1},
		{"alpha.2 < alpha.10", "v1.0.0-alpha.2", "v1.0.0-alpha.10", -1},
		{"rc.1 < rc.2", "v1.0.0-rc.1", "v1.0.0-rc.2", -1},
		{"equal prereleases", "v1.0.0-beta.1", "v1.0.0-beta.1", 0},

		// Build metadata ignored
		{"build ignored equal", "v1.0.0+build1", "v1.0.0+build2", 0},
		{"build ignored with prerelease", "v1.0.0-alpha+build1", "v1.0.0-alpha+build2", 0},

		// Numeric vs alphanumeric in prerelease
		{"numeric < alphanumeric", "v1.0.0-1", "v1.0.0-alpha", -1},
		{"alphanumeric > numeric", "v1.0.0-alpha", "v1.0.0-1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			require.NoError(t, err)
			v2, err := ParseVersion(tt.v2)
			require.NoError(t, err)
			assert.Equal(t, tt.want, v1.Compare(v2))
		})
	}
}

func TestSemVer_IsNewerThan(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{"newer major", "v2.0.0", "v1.0.0", true},
		{"newer minor", "v1.1.0", "v1.0.0", true},
		{"newer patch", "v1.0.1", "v1.0.0", true},
		{"release newer than prerelease", "v1.0.0", "v1.0.0-alpha", true},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"older version", "v1.0.0", "v2.0.0", false},
		{"prerelease not newer than release", "v1.0.0-alpha", "v1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			require.NoError(t, err)
			v2, err := ParseVersion(tt.v2)
			require.NoError(t, err)
			assert.Equal(t, tt.want, v1.IsNewerThan(v2))
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		v1      string
		v2      string
		want    int
		wantErr bool
	}{
		{"equal", "v1.0.0", "v1.0.0", 0, false},
		{"v1 greater", "v2.0.0", "v1.0.0", 1, false},
		{"v1 less", "v1.0.0", "v2.0.0", -1, false},
		{"invalid v1", "invalid", "v1.0.0", 0, true},
		{"invalid v2", "v1.0.0", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.v1, tt.v2)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		newVersion     string
		want           bool
		wantErr        bool
	}{
		{"newer version available", "v1.0.0", "v1.1.0", true, false},
		{"same version", "v1.0.0", "v1.0.0", false, false},
		{"older version", "v2.0.0", "v1.0.0", false, false},
		{"major update", "v1.0.0", "v2.0.0", true, false},
		{"prerelease to release", "v1.0.0-beta", "v1.0.0", true, false},
		{"invalid current", "invalid", "v1.0.0", false, true},
		{"invalid new", "v1.0.0", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsNewer(tt.currentVersion, tt.newVersion)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComparePrerelease(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "alpha", "alpha", 0},
		{"alpha < beta", "alpha", "beta", -1},
		{"beta > alpha", "beta", "alpha", 1},
		{"1 < 2", "1", "2", -1},
		{"10 > 2", "10", "2", 1},
		{"alpha.1 < alpha.2", "alpha.1", "alpha.2", -1},
		{"shorter < longer with same prefix", "alpha", "alpha.1", -1},
		{"numeric < alphanumeric", "1", "alpha", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, comparePrerelease(tt.a, tt.b))
		})
	}
}
