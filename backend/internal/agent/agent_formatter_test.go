package agent

import (
	"encoding/json"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatImageLabels_CorrectOutput(t *testing.T) {
	rawLabels, err := json.Marshal([]domain.Label{
		{Description: "Nature", Score: 0.98},
		{Description: "Forest", Score: 0.91},
	})
	require.NoError(t, err)
	img := &domain.Image{Title: "sunset.jpg", AILabels: rawLabels}

	result, err := formatImageLabels(img, 0.9)

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "sunset.jpg", out["image_name"])
	vr, ok := out["vision_results"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"Nature", "Forest"}, vr)
}

func TestGetFolderPath_FlatFolderReturnsName(t *testing.T) {
	folder := &domain.Folder{ID: uuid.New(), Name: "Nature"}

	path := getFolderPath([]*domain.Folder{folder}, folder)

	assert.Equal(t, "Nature", path)
}

func TestGetFolderPath_NestedFolderReturnsFullPath(t *testing.T) {
	parentID := uuid.New()
	parent := &domain.Folder{ID: parentID, Name: "Outdoors"}
	child := &domain.Folder{ID: uuid.New(), Name: "Nature", ParentID: &parentID}

	path := getFolderPath([]*domain.Folder{parent, child}, child)

	assert.Equal(t, "Outdoors > Nature", path)
}
