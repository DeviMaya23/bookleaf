package thumbnail

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/disintegration/imaging"
)

type ThumbnailService interface {
	Generate(ctx context.Context, src io.Reader) (io.Reader, error)
}

type imagingThumbnailService struct{}

func NewThumbnailService() ThumbnailService {
	return &imagingThumbnailService{}
}

func (s *imagingThumbnailService) Generate(ctx context.Context, src io.Reader) (io.Reader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	img, err := imaging.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("decode source image: %w", err)
	}

	thumb := imaging.Fit(img, 600, 600, imaging.Lanczos)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var output bytes.Buffer
	if err := imaging.Encode(&output, thumb, imaging.JPEG); err != nil {
		return nil, fmt.Errorf("encode thumbnail jpeg: %w", err)
	}

	return bytes.NewReader(output.Bytes()), nil
}

var _ ThumbnailService = (*imagingThumbnailService)(nil)
