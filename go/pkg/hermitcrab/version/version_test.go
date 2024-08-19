package version

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{"Valid version", "24.1+ui.1", "24.1.0+ui.1", false},
		{"Valid version with v prefix", "v24.1+ui.1", "24.1.0+ui.1", false},
		{"Valid version with patch", "24.1.5+ui.1", "24.1.5+ui.1", false},
		{"Valid version with single digit", "24+ui.1", "24.0.0+ui.1", false},
		{"Invalid format", "24.1.ui.1", "", true},
		{"Invalid UI version", "24.1+ui.a", "", true},
		{"Missing UI version", "24.1+ui", "", true},
		{"Extra segments", "24.1.2+ui.1.extra", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.String())

				// Additional checks for CustomVersion fields
				customV, ok := result.(*UIVersion)
				assert.True(t, ok, "Result should be of type *CustomVersion")
				assert.Equal(t, tt.expected[:strings.Index(tt.expected, "+")], customV.SemVer.String())
				expectedUI, _ := strconv.ParseUint(tt.expected[strings.LastIndex(tt.expected, ".")+1:], 10, 64)
				assert.Equal(t, expectedUI, customV.UI)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"v1 > v2", "24.1.0+ui.2", "24.1.0+ui.1", 1},
		{"v1 < v2", "24.1.0+ui.1", "24.1.0+ui.2", -1},
		{"v1 == v2", "24.1.0+ui.1", "24.1.0+ui.1", 0},
		{"Major version difference", "25.0.0+ui.1", "24.1.0+ui.2", 1},
		{"Minor version difference", "24.2.0+ui.1", "24.1.0+ui.2", 1},
		{"Patch version difference", "24.1.1+ui.1", "24.1.0+ui.2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			assert.NoError(t, err)
			v2, err := ParseVersion(tt.v2)
			assert.NoError(t, err)
			result := v1.Compare(v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionMethods(t *testing.T) {
	v, err := ParseVersion("24.1.5+ui.10")
	assert.NoError(t, err)

	assert.Equal(t, uint64(24), v.Major())
	assert.Equal(t, uint64(1), v.Minor())
	assert.Equal(t, uint64(5), v.Patch())
	assert.Equal(t, uint64(10), v.UI())
	assert.Equal(t, "24.1.5+ui.10", v.String())

	// Test direct field access
	customV, ok := v.(*UIVersion)
	assert.True(t, ok)
	assert.Equal(t, uint64(10), customV.UI())
}
