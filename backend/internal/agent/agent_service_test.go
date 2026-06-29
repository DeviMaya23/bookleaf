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
	image           *domain.Image
	listByFolderErr error
}

func (s *spyAgentImageRepo) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return s.image, nil
}

func (s *spyAgentImageRepo) ListByFolder(_ context.Context, _ string, _ uuid.UUID, _ *string, _ *string) ([]*domain.Image, error) {
	if s.listByFolderErr != nil {
		return nil, s.listByFolderErr
	}
	return []*domain.Image{}, nil
}

type spyAgentFolderRepo struct{}

func (s *spyAgentFolderRepo) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return []*domain.Folder{{ID: uuid.New(), Name: "Nature"}}, nil
}

// --- helpers ---

func agentNoopTel() *observability.Telemetry {
	return observability.NewTelemetry(nil, nil, nil)
}

func agentTestImage(t *testing.T) *domain.Image {
	t.Helper()
	labels, err := json.Marshal([]domain.Label{})
	require.NoError(t, err)
	return &domain.Image{ID: uuid.New(), Title: "test.jpg", AILabels: labels}
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
		name            string
		toolName        string
		folderInput     map[string]any
		listByFolderErr error
	}{
		{
			name:        "get_folder_top_labels invalid UUID",
			toolName:    "get_folder_top_labels",
			folderInput: map[string]any{"folder_id": "not-a-uuid"},
		},
		{
			name:        "get_folder_image_samples invalid UUID",
			toolName:    "get_folder_image_samples",
			folderInput: map[string]any{"folder_id": "not-a-uuid"},
		},
		{
			name:            "get_folder_top_labels DB error",
			toolName:        "get_folder_top_labels",
			folderInput:     map[string]any{"folder_id": uuid.New().String()},
			listByFolderErr: errors.New("db failure"),
		},
		{
			name:            "get_folder_image_samples DB error",
			toolName:        "get_folder_image_samples",
			folderInput:     map[string]any{"folder_id": uuid.New().String()},
			listByFolderErr: errors.New("db failure"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := sequentialServer(t, []string{
				anthropicMsg(toolUseJSON("toolu_001", tc.toolName, tc.folderInput)),
				submitFolderMsg(existingFolderID),
			})
			imageRepo := &spyAgentImageRepo{image: agentTestImage(t), listByFolderErr: tc.listByFolderErr}
			svc := newTestAgentService(t, imageRepo, &spyAgentFolderRepo{}, srv.URL)

			suggestion, err := svc.GetFolderSuggestion(context.Background(), "kp_user", uuid.New())

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
	imageRepo := &spyAgentImageRepo{image: agentTestImage(t)}
	svc := newTestAgentService(t, imageRepo, &spyAgentFolderRepo{}, srv.URL)

	_, err := svc.GetFolderSuggestion(context.Background(), "kp_user", uuid.New())

	require.Error(t, err)
	assert.ErrorContains(t, err, "agent exceeded invalid input cap")
}
