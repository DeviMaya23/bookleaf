package agent

import (
	"encoding/json"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawLabels(t *testing.T, labels []domain.Label) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(labels)
	require.NoError(t, err)
	return b
}

func TestExtractLabels_AboveThresholdReturned(t *testing.T) {
	input := rawLabels(t, []domain.Label{
		{Description: "Sunset", Score: 0.90},
		{Description: "Sky", Score: 0.80},
		{Description: "Cloud", Score: 0.60},
	})

	result, err := extractLabels(input, 0.75)

	require.NoError(t, err)
	assert.Equal(t, []string{"Sunset", "Sky"}, result)
}

func TestExtractLabels_AllBelowThresholdReturnsEmpty(t *testing.T) {
	input := rawLabels(t, []domain.Label{
		{Description: "Cloud", Score: 0.50},
	})

	result, err := extractLabels(input, 0.75)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestExtractLabels_InvalidJSONReturnsError(t *testing.T) {
	_, err := extractLabels(json.RawMessage(`not-json`), 0.75)

	assert.Error(t, err)
}

func TestFormatFolderTopLabels_CorrectTopCounts(t *testing.T) {
	folderID := uuid.New()
	images := []*domain.Image{
		{
			AILabels: rawLabels(t, []domain.Label{
				{Description: "Nature", Score: 0.90},
				{Description: "Sky", Score: 0.80},
			}),
			Tags: []domain.Tag{{Name: "outdoors"}, {Name: "travel"}},
		},
		{
			AILabels: rawLabels(t, []domain.Label{
				{Description: "Nature", Score: 0.85},
			}),
			Tags: []domain.Tag{{Name: "outdoors"}},
		},
	}

	result, err := formatFolderTopLabels(folderID, images, 0.75)

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
	images := []*domain.Image{
		{
			AILabels: rawLabels(t, []domain.Label{
				{Description: "Nature", Score: 0.90},
				{Description: "Fog", Score: 0.50},
			}),
			Tags: []domain.Tag{},
		},
	}

	result, err := formatFolderTopLabels(folderID, images, 0.75)

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	topLabels := out["top_vision_labels"].([]any)
	assert.Equal(t, []any{"Nature"}, topLabels)
}

func TestFormatFolderTopLabels_FewerThan5LabelsReturnsAllAvailable(t *testing.T) {
	folderID := uuid.New()
	images := []*domain.Image{
		{
			AILabels: rawLabels(t, []domain.Label{
				{Description: "Nature", Score: 0.90},
				{Description: "Sky", Score: 0.80},
			}),
			Tags: []domain.Tag{{Name: "outdoors"}},
		},
	}

	result, err := formatFolderTopLabels(folderID, images, 0.75)

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out["top_vision_labels"].([]any), 2)
	assert.Len(t, out["top_user_tags"].([]any), 1)
}

func TestFormatFolderTopLabels_NoImagesReturnsZeroCounts(t *testing.T) {
	folderID := uuid.New()

	result, err := formatFolderTopLabels(folderID, []*domain.Image{}, 0.75)

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, float64(0), out["image_count"])
	assert.Empty(t, out["top_vision_labels"].([]any))
	assert.Empty(t, out["top_user_tags"].([]any))
}

func TestFormatFolderImageSamples_MoreThan5ReturnsOnly5(t *testing.T) {
	images := make([]*domain.Image, 7)
	for i := range images {
		images[i] = &domain.Image{
			Title:    "img",
			AILabels: rawLabels(t, []domain.Label{}),
		}
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out, 5)
}

func TestFormatFolderImageSamples_FewerThan5ImagesReturnsAll(t *testing.T) {
	images := []*domain.Image{
		{Title: "a.jpg", AILabels: rawLabels(t, []domain.Label{})},
		{Title: "b.jpg", AILabels: rawLabels(t, []domain.Label{})},
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Len(t, out, 2)
}

func TestFormatFolderImageSamples_NilDescriptionAndSourceURLSerialiseAsEmpty(t *testing.T) {
	images := []*domain.Image{
		{
			Title:       "photo.jpg",
			Description: nil,
			SourceURL:   nil,
			AILabels:    rawLabels(t, []domain.Label{}),
		},
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "", out[0]["image_notes"])
	assert.Equal(t, "", out[0]["image_source_url"])
}

func TestFormatFolderImageSamples_BelowThresholdLabelsExcluded(t *testing.T) {
	images := []*domain.Image{
		{
			Title: "photo.jpg",
			AILabels: rawLabels(t, []domain.Label{
				{Description: "Nature", Score: 0.90},
				{Description: "Blur", Score: 0.40},
			}),
		},
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	labels := out[0]["image_vision_labels"].([]any)
	assert.Equal(t, []any{"Nature"}, labels)
}

func TestFormatFolderImageSamples_SearchEngineSourceURLReturnsEmpty(t *testing.T) {
	googleURL := "https://www.google.com/search?q=heather+honey&udm=2"
	images := []*domain.Image{
		{Title: "photo.jpg", SourceURL: &googleURL, AILabels: rawLabels(t, []domain.Label{})},
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "", out[0]["image_source_url"])
}

func TestFormatFolderImageSamples_NonDeniedSourceURLPassesThrough(t *testing.T) {
	sourceURL := "https://unsplash.com/photos/abc123"
	images := []*domain.Image{
		{Title: "photo.jpg", SourceURL: &sourceURL, AILabels: rawLabels(t, []domain.Label{})},
	}

	result, err := formatFolderImageSamples(images, 0.75)

	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, sourceURL, out[0]["image_source_url"])
}

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
