package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicOption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- spies ---

type spyAgentImageRepo struct {
	image                 *domain.Image
	imageLabels           []string
	getImageWithLabelsErr error
	folderTopLabelsErr    error
	folderImageSamplesErr error
	listByFolderErr       error
}

func (s *spyAgentImageRepo) GetImageWithLabels(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64) (*domain.Image, []string, error) {
	if s.getImageWithLabelsErr != nil {
		return nil, nil, s.getImageWithLabelsErr
	}
	return s.image, s.imageLabels, nil
}

func (s *spyAgentImageRepo) GetFolderTopLabels(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) (*domain.FolderAggregate, error) {
	if s.folderTopLabelsErr != nil {
		return nil, s.folderTopLabelsErr
	}
	return &domain.FolderAggregate{}, nil
}

func (s *spyAgentImageRepo) GetFolderImageSamples(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ float64, _ int) ([]*domain.Image, map[uuid.UUID][]string, error) {
	if s.folderImageSamplesErr != nil {
		return nil, nil, s.folderImageSamplesErr
	}
	return []*domain.Image{}, map[uuid.UUID][]string{}, nil
}

func (s *spyAgentImageRepo) ListByFolder(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ *string, _ *string) ([]*domain.Image, error) {
	if s.listByFolderErr != nil {
		return nil, s.listByFolderErr
	}
	return []*domain.Image{}, nil
}

type spyAgentFolderRepo struct{}

func (s *spyAgentFolderRepo) List(_ context.Context, _ uuid.UUID) ([]*domain.Folder, error) {
	return []*domain.Folder{{ID: uuid.New(), Name: "Nature"}}, nil
}

// --- helpers ---

func agentNoopTel() *observability.Telemetry {
	return observability.NewTelemetry(nil, nil, nil)
}

func agentTestImage() *domain.Image {
	return &domain.Image{ID: uuid.New(), Title: "test.jpg"}
}

func anthropicMsg(contentBlocks ...string) string {
	content := ""
	for i, b := range contentBlocks {
		if i > 0 {
			content += ","
		}
		content += b
	}
	return fmt.Sprintf(`{"id":"msg_test","type":"message","role":"assistant","model":"claude-test","content":[%s],"stop_reason":"tool_use","usage":{"input_tokens":10,"output_tokens":10}}`, content)
}

func toolUseJSON(id, name string, input map[string]any) string {
	inputJSON, _ := json.Marshal(input)
	return fmt.Sprintf(`{"type":"tool_use","id":%q,"name":%q,"input":%s}`, id, name, string(inputJSON))
}

func submitFolderMsg(folderID string) string {
	return anthropicMsg(toolUseJSON("toolu_submit", "submit_existing_folder", map[string]any{
		"folder_id": folderID,
		"reasoning": "good fit",
	}))
}

func sequentialServer(t *testing.T, responses []string) *httptest.Server {
	t.Helper()
	var idx atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(idx.Add(1)) - 1
		if i >= len(responses) {
			t.Errorf("unexpected Anthropic API call #%d", i+1)
			http.Error(w, "unexpected", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, responses[i])
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestAgentService(t *testing.T, imageRepo AgentImageRepository, folderRepo AgentFolderRepository, serverURL string) *AgentService {
	t.Helper()
	client := anthropic.NewClient(
		anthropicOption.WithBaseURL(serverURL+"/"),
		anthropicOption.WithAPIKey("test-key"),
	)
	return NewAgentService(imageRepo, folderRepo, &client, agentNoopTel(), "claude-test")
}

// --- tests ---

func TestGetFolderSuggestion_InvalidToolInputReturnsLenientError(t *testing.T) {
	existingFolderID := uuid.New().String()

	cases := []struct {
		name      string
		toolName  string
		folderInput map[string]any
		imageRepo *spyAgentImageRepo
	}{
		{
			name:        "get_folder_top_labels invalid UUID",
			toolName:    "get_folder_top_labels",
			folderInput: map[string]any{"folder_id": "not-a-uuid"},
			imageRepo:   &spyAgentImageRepo{image: agentTestImage()},
		},
		{
			name:        "get_folder_image_samples invalid UUID",
			toolName:    "get_folder_image_samples",
			folderInput: map[string]any{"folder_id": "not-a-uuid"},
			imageRepo:   &spyAgentImageRepo{image: agentTestImage()},
		},
		{
			name:        "get_folder_top_labels DB error",
			toolName:    "get_folder_top_labels",
			folderInput: map[string]any{"folder_id": uuid.New().String()},
			imageRepo:   &spyAgentImageRepo{image: agentTestImage(), folderTopLabelsErr: errors.New("db failure")},
		},
		{
			name:        "get_folder_image_samples DB error",
			toolName:    "get_folder_image_samples",
			folderInput: map[string]any{"folder_id": uuid.New().String()},
			imageRepo:   &spyAgentImageRepo{image: agentTestImage(), folderImageSamplesErr: errors.New("db failure")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := sequentialServer(t, []string{
				anthropicMsg(toolUseJSON("toolu_001", tc.toolName, tc.folderInput)),
				submitFolderMsg(existingFolderID),
			})
			svc := newTestAgentService(t, tc.imageRepo, &spyAgentFolderRepo{}, srv.URL)

			suggestion, err := svc.GetFolderSuggestion(context.Background(), uuid.New(), uuid.New())

			require.NoError(t, err)
			assert.Equal(t, existingFolderID, suggestion.FolderID)
		})
	}
}

func TestGetFolderSuggestion_ExceedsInvalidInputCapReturnsError(t *testing.T) {
	badInput := map[string]any{"folder_id": "not-a-uuid"}
	srv := sequentialServer(t, []string{
		anthropicMsg(toolUseJSON("toolu_001", "get_folder_top_labels", badInput)),
		anthropicMsg(toolUseJSON("toolu_002", "get_folder_top_labels", badInput)),
		anthropicMsg(toolUseJSON("toolu_003", "get_folder_top_labels", badInput)),
	})
	imageRepo := &spyAgentImageRepo{image: agentTestImage()}
	svc := newTestAgentService(t, imageRepo, &spyAgentFolderRepo{}, srv.URL)

	_, err := svc.GetFolderSuggestion(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	assert.ErrorContains(t, err, "agent exceeded invalid input cap")
}
