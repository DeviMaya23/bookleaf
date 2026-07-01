package agent

import (
	"encoding/json"
	"net/url"
	"sort"
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

func formatImageLabels(image *domain.Image, threshold float64) (string, error) {
	labels, err := extractLabels(image.AILabels, threshold)
	if err != nil {
		return "", err
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

func findFolder(folders []*domain.Folder, id uuid.UUID) *domain.Folder {
	for _, f := range folders {
		if f.ID == id {
			return f
		}
	}
	return nil
}

func formatFolderTopLabels(folderID uuid.UUID, folder *domain.Folder, images []*domain.Image, threshold float64) (string, error) {
	labelCount := map[string]int{}
	tagCount := map[string]int{}

	for _, img := range images {
		labels, err := extractLabels(img.AILabels, threshold)
		if err != nil {
			return "", err
		}
		for _, l := range labels {
			labelCount[l]++
		}
		for _, t := range img.Tags {
			tagCount[t.Name]++
		}
	}

	type entry struct {
		name  string
		count int
	}
	topN := func(counts map[string]int, n int) []string {
		entries := make([]entry, 0, len(counts))
		for name, count := range counts {
			entries = append(entries, entry{name, count})
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].count != entries[j].count {
				return entries[i].count > entries[j].count
			}
			return entries[i].name < entries[j].name
		})
		if len(entries) > n {
			entries = entries[:n]
		}
		result := make([]string, len(entries))
		for i, e := range entries {
			result[i] = e.name
		}
		return result
	}

	name, desc := "", ""
	if folder != nil {
		name = folder.Name
		desc = util.DerefOr(folder.Description, "")
	}
	out := map[string]interface{}{
		"folder_id":          folderID.String(),
		"folder_name":        name,
		"folder_description": desc,
		"image_count":        len(images),
		"top_vision_labels":  topN(labelCount, 5),
		"top_user_tags":      topN(tagCount, 5),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func formatFolderImageSamples(images []*domain.Image, threshold float64) (string, error) {
	if len(images) > 5 {
		images = images[:5]
	}
	type imageSample struct {
		Title        string   `json:"image_name"`
		Notes        string   `json:"image_notes"`
		SourceURL    string   `json:"image_source_url"`
		VisionLabels []string `json:"image_vision_labels"`
	}
	samples := make([]imageSample, 0, len(images))
	for _, img := range images {
		labels, err := extractLabels(img.AILabels, threshold)
		if err != nil {
			return "", err
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

func extractLabels(aiLabels json.RawMessage, threshold float64) ([]string, error) {
	var labels []domain.Label
	if err := json.Unmarshal(aiLabels, &labels); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(labels))
	for _, l := range labels {
		if float64(l.Score) >= threshold {
			result = append(result, l.Description)
		}
	}
	return result, nil
}
