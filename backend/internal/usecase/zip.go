package usecase

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
)

// writeImagesToZip streams the given images as a zip archive to w.
// Entry names are derived from each image's title and MIME type, with
// path separators sanitized and colliding names deduplicated.
func writeImagesToZip(ctx context.Context, images []*domain.Image, store StorageService, w io.Writer) error {
	zw := zip.NewWriter(w)

	nameCounts := make(map[string]int)
	for _, image := range images {
		name := exportEntryName(image, nameCounts)

		reader, err := store.GetObject(ctx, image.R2Path)
		if err != nil {
			return fmt.Errorf("get object: %w", err)
		}

		entry, err := zw.Create(name)
		if err != nil {
			reader.Close()
			return fmt.Errorf("create zip entry: %w", err)
		}

		if _, err := io.Copy(entry, reader); err != nil {
			reader.Close()
			return fmt.Errorf("copy object to zip entry: %w", err)
		}
		reader.Close()
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}

	return nil
}

// exportEntryName derives a zip entry name for an image, sanitizing its title
// to remove path separators and disambiguating collisions with nameCounts.
func exportEntryName(image *domain.Image, nameCounts map[string]int) string {
	title := sanitizePathSegment(image.Title)
	ext := downloadFileExtension(image.MIMEType)
	base := title + "." + ext

	count := nameCounts[base]
	nameCounts[base] = count + 1
	if count == 0 {
		return base
	}
	return fmt.Sprintf("%s (%d).%s", title, count, ext)
}

// sanitizePathSegment replaces path-separator characters so a title cannot
// introduce nested paths inside a zip archive.
func sanitizePathSegment(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}
