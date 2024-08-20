package versions

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	semver "github.com/Masterminds/semver/v3"
)

type Version struct {
	*semver.Version
	BuildNumber int
	FileName    string
}

var trimPatterns = []string{"tar.gz", "tar.zst"}

const semverPattern string = `v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

var semverRegex = regexp.MustCompile(semverPattern)

func parseSemanticVersion(filename string) (string, error) {
	// Define the semantic versioning regex pattern

	// Find the first match in the filename
	match := semverRegex.FindString(filename)
	if match == "" {
		return "", fmt.Errorf("no semantic version found in filename")
	}

	return match, nil
}

// ParseString parses a version filename into a Version struct.
func ParseFileName(filePath string) (*Version, error) {
	// Only get the base, which is the file name.
	fileName := path.Base(filePath)
	semVer, err := parseSemanticVersion(fileName)
	if err != nil {
		return nil, err
	}

	// Trim the file extensions, if relevant.
	// We want to remove the file information because
	// in can mess up how the version presents as.
	for _, pattern := range trimPatterns {
		semVer = strings.TrimRight(semVer, pattern)
	}

	v, err := semver.NewVersion(semVer)
	if err != nil {
		return nil, err
	}

	buildNum := 0
	if v.Metadata() != "" {
		parsedNum, err := strconv.Atoi(v.Metadata())
		if err != nil {
			return nil, err
		}
		buildNum = parsedNum
	}

	return &Version{
		Version:     v,
		FileName:    filePath,
		BuildNumber: buildNum,
	}, err
}
