package agent

import (
	"encoding/json"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFolderTopLabels_CorrectTopCounts(t *testing.T) {
	folderID := uuid.New()
	topVisionLabels := []string{"Nature", "Sky"}
	topUserTags := []string{"outdoors", "travel"}

	result, err := formatFolderTopLabels(folderID, nil, 2, topVisionLabels, topUserTags)

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, folderID.String(), out["folder_id"])
	assert.Equal(t, float64(2), out["image_count"])
	topLabels := out["top_vision_labels"].([]any)
	assert.Equal(t, "Nature", topLabels[0])
	topTags := out["top_user_tags"].([]any)
	assert.Equal(t, "outdoors", topTags[0])
}

func TestFormatFolderTopLabels_BelowThresholdExcluded(t *testing.T) {
	folderID := uuid.New()
	topVisionLabels := []string{"Nature"}

	result, err := formatFolderTopLabels(folderID, nil, 1, topVisionLabels, []string{})

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	topLabels := out["top_vision_labels"].([]any)
	assert.Equal(t, []any{"Nature"}, topLabels)
}

func TestFormatFolderTopLabels_FewerThan5LabelsReturnsAllAvailable(t *testing.T) {
	folderID := uuid.New()
	topVisionLabels := []string{"Nature", "Sky"}

	result, err := formatFolderTopLabels(folderID, nil, 1, topVisionLabels, []string{"outdoors"})

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out["top_vision_labels"].([]any), 2)
	assert.Len(t, out["top_user_tags"].([]any), 1)
}

func TestFormatFolderTopLabels_NoImagesReturnsZeroCounts(t *testing.T) {
	folderID := uuid.New()

	result, err := formatFolderTopLabels(folderID, nil, 0, []string{}, []string{})

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, float64(0), out["image_count"])
	assert.Empty(t, out["top_vision_labels"].([]any))
	assert.Empty(t, out["top_user_tags"].([]any))
}

func TestFormatFolderTopLabels_IncludesFolderNameAndDescription(t *testing.T) {
	folderID := uuid.New()
	desc := "mood shots"
	folder := &domain.Folder{ID: folderID, Name: "Photography", Description: &desc}

	result, err := formatFolderTopLabels(folderID, folder, 0, []string{}, []string{})

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "Photography", out["folder_name"])
	assert.Equal(t, "mood shots", out["folder_description"])
}

func TestFormatFolderImageSamples_ReturnsAllPassedImages(t *testing.T) {
	images := make([]*domain.Image, 3)
	labelMap := map[uuid.UUID][]string{}
	for i := range images {
		images[i] = &domain.Image{ID: uuid.New(), Title: "img"}
		labelMap[images[i].ID] = []string{}
	}

	result, err := formatFolderImageSamples(images, labelMap)

	require.NoError(t, err)
	var out []any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out, 3)
}

func TestFormatFolderImageSamples_FewerThan5ImagesReturnsAll(t *testing.T) {
	img1 := &domain.Image{ID: uuid.New(), Title: "a.jpg"}
	img2 := &domain.Image{ID: uuid.New(), Title: "b.jpg"}
	labelMap := map[uuid.UUID][]string{img1.ID: {}, img2.ID: {}}

	result, err := formatFolderImageSamples([]*domain.Image{img1, img2}, labelMap)

	require.NoError(t, err)
	var out []any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out, 2)
}

func TestFormatFolderImageSamples_NilDescriptionAndSourceURLSerialiseAsEmpty(t *testing.T) {
	img := &domain.Image{ID: uuid.New(), Title: "photo.jpg", Description: nil, SourceURL: nil}
	labelMap := map[uuid.UUID][]string{img.ID: {}}

	result, err := formatFolderImageSamples([]*domain.Image{img}, labelMap)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "", out[0]["image_notes"])
	assert.Equal(t, "", out[0]["image_source_url"])
}

func TestFormatFolderImageSamples_LabelsFromMapUsedPerImage(t *testing.T) {
	img := &domain.Image{ID: uuid.New(), Title: "photo.jpg"}
	labelMap := map[uuid.UUID][]string{img.ID: {"Nature"}}

	result, err := formatFolderImageSamples([]*domain.Image{img}, labelMap)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	labels := out[0]["image_vision_labels"].([]any)
	assert.Equal(t, []any{"Nature"}, labels)
}

func TestFormatFolderImageSamples_ImageAbsentFromLabelMapReturnsEmpty(t *testing.T) {
	img := &domain.Image{ID: uuid.New(), Title: "photo.jpg"}

	result, err := formatFolderImageSamples([]*domain.Image{img}, map[uuid.UUID][]string{})

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Empty(t, out[0]["image_vision_labels"].([]any))
}

func TestFormatFolderImageSamples_SearchEngineSourceURLReturnsEmpty(t *testing.T) {
	googleURL := "https://www.google.com/search?q=heather+honey&udm=2"
	img := &domain.Image{ID: uuid.New(), Title: "photo.jpg", SourceURL: &googleURL}
	labelMap := map[uuid.UUID][]string{img.ID: {}}

	result, err := formatFolderImageSamples([]*domain.Image{img}, labelMap)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "", out[0]["image_source_url"])
}

func TestFormatFolderImageSamples_NonDeniedSourceURLPassesThrough(t *testing.T) {
	sourceURL := "https://unsplash.com/photos/abc123"
	img := &domain.Image{ID: uuid.New(), Title: "photo.jpg", SourceURL: &sourceURL}
	labelMap := map[uuid.UUID][]string{img.ID: {}}

	result, err := formatFolderImageSamples([]*domain.Image{img}, labelMap)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, sourceURL, out[0]["image_source_url"])
}

func TestFormatImageLabels_CorrectOutput(t *testing.T) {
	result, err := formatImageLabels("sunset.jpg", []string{"Nature", "Forest"})

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
