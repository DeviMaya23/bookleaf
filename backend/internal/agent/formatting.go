package agent

import (
	"encoding/json"

	"github.com/devi/bookleaf/internal/domain"
)

// O(n^2) lookup, maybe fix later if we have a lot of folders
func getFolderPath(folders []*domain.Folder, folder *domain.Folder) string {

	if folder.ParentID == nil {
		return folder.Name
	}
	var parentFolder *domain.Folder
	for _, f := range folders {
		if f.ID == *folder.ParentID {
			parentFolder = f
			break
		}
	}
	if parentFolder == nil {
		return folder.Name
	}

	return getFolderPath(folders, parentFolder) + " > " + folder.Name
}

func getImageMetadata(image *domain.Image) (string, error) {
	var results []struct {
		Description string  `json:"Description"`
		Score       float64 `json:"Score"`
	}
	err := json.Unmarshal(image.AILabels, &results)
	if err != nil {
		return "", err
	}

	labels := make([]string, 0, len(results))
	for _, result := range results {
		labels = append(labels, result.Description)
	}
	metadata := map[string]interface{}{
		"image_name":     image.Title,
		"vision_results": labels,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(metadataJSON), nil
}
