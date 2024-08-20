package versions

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestWriteManifestToFile(t *testing.T) {
	type args struct {
		manifest Manifest
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "properly writes the file to a json",
			args: args{
				manifest: Manifest{
					Name:        "Sample Manifest",
					Description: "This is a sample manifest",
					Meta: map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					},
					Versions: []Version{
						{
							Version:     &semver.Version{},
							SemVersion:  "1.0.0",
							BuildNumber: 1,
							FilePath:    "/path/to/1.0.0.tar.gz",
						},
						{
							Version:     &semver.Version{},
							SemVersion:  "1.0.1",
							BuildNumber: 2,
							FilePath:    "/path/to/molt-v1.0.1-beta+2",
						},
					},
				},
				filename: "test.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteManifestToFile(tt.args.manifest, tt.args.filename)
			require.NoError(t, err)
		})
	}
}

func TestParseManifestFromFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *Manifest
		wantErr error
	}{
		{
			name: "can read manifest file",
			args: args{
				filename: "test.json",
			},
			want:    nil,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseManifestFromFile(tt.args.filename)
			for _, v := range got.Versions {
				fmt.Println(v, v.FilePath, v.BuildNumber, v.SemVersion)
			}
			require.NoError(t, err)
		})
	}
}
