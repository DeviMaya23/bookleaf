package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

const annotateURL = "https://vision.googleapis.com/v1/images:annotate"

type Label struct {
	Description string
	Score       float32
}

type VisionService interface {
	AnnotateImage(ctx context.Context, imageBytes []byte) ([]Label, error)
}

type visionClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewVisionClient(apiKey string) VisionService {
	return &visionClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type annotateRequest struct {
	Requests []imageRequest `json:"requests"`
}

type imageRequest struct {
	Image    imageContent  `json:"image"`
	Features []featureSpec `json:"features"`
}

type imageContent struct {
	Content string `json:"content"`
}

type featureSpec struct {
	Type string `json:"type"`
}

type annotateResponse struct {
	Responses []imageAnnotation `json:"responses"`
}

type imageAnnotation struct {
	LabelAnnotations []labelAnnotation `json:"labelAnnotations"`
}

type labelAnnotation struct {
	Description string  `json:"description"`
	Score       float32 `json:"score"`
}

func (c *visionClient) AnnotateImage(ctx context.Context, imageBytes []byte) ([]Label, error) {
	body := annotateRequest{
		Requests: []imageRequest{
			{
				Image: imageContent{Content: base64.StdEncoding.EncodeToString(imageBytes)},
				Features: []featureSpec{
					{Type: "LABEL_DETECTION"},
				},
			},
		},
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, annotateURL+"?key="+c.apiKey, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("build vision request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision api returned status %d", resp.StatusCode)
	}

	var result annotateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode vision response: %w", err)
	}

	if len(result.Responses) == 0 {
		return nil, nil
	}

	annotations := result.Responses[0].LabelAnnotations
	labels := make([]Label, len(annotations))
	for i, a := range annotations {
		labels[i] = Label{Description: a.Description, Score: a.Score}
	}

	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Score > labels[j].Score
	})

	return labels, nil
}
