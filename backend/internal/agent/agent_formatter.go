package agent

import (
	"encoding/json"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/util"
)

func formatFolderList(folders []*domain.Folder) (string, error) {
	transformedFolders := make([]map[string]interface{}, 0, len(folders))
	for _, folder := range folders {
		transformedFolder := map[string]interface{}{
			"folder_id":          folder.ID.String(),
			"folder_name":        folder.Name,
			"folder_description": util.DerefOr(folder.Description, ""),
			"folder_path":        getFolderPath(folders, folder),
		}
		transformedFolders = append(transformedFolders, transformedFolder)
	}
	str, err := json.Marshal(transformedFolders)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func formatImageLabels(image *domain.Image, threshold float64) (string, error) {
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
		if result.Score >= threshold {
			labels = append(labels, result.Description)
		}
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
