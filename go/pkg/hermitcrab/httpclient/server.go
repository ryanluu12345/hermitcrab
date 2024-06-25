package httpclient

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/gorilla/mux"
	"github.com/lassenordahl/hermitcrab/pkg/hermitcrab/bucket"
	"github.com/lassenordahl/hermitcrab/pkg/hermitcrab/version"
)

type Server struct {
	router        *mux.Router
	bucketManager bucket.BucketManager
	cacheDir      string
	logger        *log.Logger
}

func NewServer(bm bucket.BucketManager, cacheDir string, logger *log.Logger) *Server {
	s := &Server{
		router:        mux.NewRouter(),
		bucketManager: bm,
		cacheDir:      cacheDir,
		logger:        logger,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.serveLatestVersion).Methods("GET")
	s.router.HandleFunc("/version/{version}", s.serveSpecificVersion).Methods("GET")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) serveLatestVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	latestVersion, err := s.bucketManager.GetLatestPatchVersion(ctx, "24.1") // Hardcoded for demo
	if err != nil {
		http.Error(w, "Failed to get latest version", http.StatusInternalServerError)
		return
	}
	s.serveVersion(w, r, latestVersion)
}

func (s *Server) serveSpecificVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionStr := vars["version"]
	v, err := version.ParseVersion(versionStr)
	if err != nil {
		http.Error(w, "Invalid version", http.StatusBadRequest)
		return
	}
	s.serveVersion(w, r, v)
}

func (s *Server) serveVersion(w http.ResponseWriter, r *http.Request, v *semver.Version) {
	ctx := r.Context()
	versionDir := filepath.Join(s.cacheDir, v.String())

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		if err := s.downloadAndExtract(ctx, v, versionDir); err != nil {
			http.Error(w, "Failed to prepare version", http.StatusInternalServerError)
			return
		}
	}

	http.FileServer(http.Dir(versionDir)).ServeHTTP(w, r)
}

func (s *Server) downloadAndExtract(ctx context.Context, v *semver.Version, destDir string) error {
	reader, err := s.bucketManager.DownloadPatchVersion(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to download version: %w", err)
	}
	defer reader.Close()

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file contents: %w", err)
			}
			f.Close()
		}
	}

	return nil
}
