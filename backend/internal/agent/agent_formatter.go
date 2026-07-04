package agent

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/util"
	"github.com/google/uuid"
)

var sourceURLDenyList = []string{
	"google.com",
	"bing.com",
	"duckduckgo.com",
	"yahoo.com",
	"yandex.com",
	"yandex.ru",
	"baidu.com",
}

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

func formatImageLabels(title string, labels []string) (string, error) {
	metadata := map[string]interface{}{
		"image_name":     title,
		"vision_results": labels,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(metadataJSON), nil
}

func findFolder(folders []*domain.Folder, id uuid.UUID) *domain.Folder {
	for _, f := range folders {
		if f.ID == id {
			return f
		}
	}
	return nil
}

func formatFolderTopLabels(folderID uuid.UUID, folder *domain.Folder, imageCount int, topVisionLabels []string, topUserTags []string) (string, error) {
	name, desc := "", ""
	if folder != nil {
		name = folder.Name
		desc = util.DerefOr(folder.Description, "")
	}
	out := map[string]interface{}{
		"folder_id":          folderID.String(),
		"folder_name":        name,
		"folder_description": desc,
		"image_count":        imageCount,
		"top_vision_labels":  topVisionLabels,
		"top_user_tags":      topUserTags,
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func formatFolderImageSamples(images []*domain.Image, labelMap map[uuid.UUID][]string) (string, error) {
	type imageSample struct {
		Title        string   `json:"image_name"`
		Notes        string   `json:"image_notes"`
		SourceURL    string   `json:"image_source_url"`
		VisionLabels []string `json:"image_vision_labels"`
	}
	samples := make([]imageSample, 0, len(images))
	for _, img := range images {
		labels := labelMap[img.ID]
		if labels == nil {
			labels = []string{}
		}
		sourceURL := util.DerefOr(img.SourceURL, "")
		if isDeniedSourceURL(sourceURL) {
			sourceURL = ""
		}
		samples = append(samples, imageSample{
			Title:        img.Title,
			Notes:        util.DerefOr(img.Description, ""),
			SourceURL:    sourceURL,
			VisionLabels: labels,
		})
	}
	b, err := json.Marshal(samples)
	if err != nil {
		return "", err
	}
	return string(b), nil
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

func isDeniedSourceURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, denied := range sourceURLDenyList {
		if strings.Contains(host, denied) {
			return true
		}
	}
	return false
}
