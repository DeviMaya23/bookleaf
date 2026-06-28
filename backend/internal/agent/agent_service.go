package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

type AgentImageRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
}

type AgentFolderRepository interface {
	List(ctx context.Context, userID string) ([]*domain.Folder, error)
}

type Suggestion struct {
	FolderID          string `json:"folder_id,omitempty"`
	NewFolderName     string `json:"new_folder_name,omitempty"`
	NewFolderParentID string `json:"new_folder_parent_id,omitempty"`
	Reasoning         string `json:"reasoning"`
}

type AgentService struct {
	imageRepo  AgentImageRepository
	folderRepo AgentFolderRepository
	aiClient   *anthropic.Client
	tel        *observability.Telemetry
	model      string
}

func NewAgentService(imageRepo AgentImageRepository, folderRepo AgentFolderRepository, aiClient *anthropic.Client, tel *observability.Telemetry, model string) *AgentService {
	return &AgentService{
		imageRepo:  imageRepo,
		folderRepo: folderRepo,
		aiClient:   aiClient,
		tel:        tel,
		model:      model,
	}
}

func (a *AgentService) listFolders(ctx context.Context, userID string) (string, error) {
	folders, err := a.folderRepo.List(ctx, userID)
	if err != nil {
		return "", err
	}
	return formatFolderList(folders)
}

func (u *AgentService) GetFolderSuggestion(ctx context.Context, userID string, imageID uuid.UUID) (Suggestion, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "agent.GetFolderSuggestion")
	defer span.End()

	res := Suggestion{}

	tools := make([]anthropic.ToolUnionParam, len(folderSuggestionToolParams))
	for i, t := range folderSuggestionToolParams {
		tp := t
		tools[i] = anthropic.ToolUnionParam{OfTool: &tp}
	}

	img, err := u.imageRepo.GetByID(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}

	imageMetadata, err := formatImageLabels(img)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}
	userPrompt := fmt.Sprintf("\nImage metadata: %s", imageMetadata)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
	}

	var toolResultText string

	for {
		response, err := u.aiClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     u.model,
			MaxTokens: 1024,
			Tools:     tools,
			Messages:  messages,
			System: []anthropic.TextBlockParam{
				{Text: folderSuggestionSystemPrompt},
			},
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return res, fmt.Errorf("call anthropic api: %w", err)
		}

		messages = append(messages, response.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}

		for _, block := range response.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.ToolUseBlock:
				switch variant.Name {
				case "get_folder_list":
					folders, err := u.listFolders(ctx, userID)
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, err.Error())
						return res, err
					}
					toolResultText = fmt.Sprintf("Folder list: %s", folders)
				case "submit_existing_folder", "submit_new_folder":
					if err := json.Unmarshal([]byte(variant.Input), &res); err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, err.Error())
						return res, fmt.Errorf("unmarshal tool input: %w", err)
					}
					return res, nil
				default:
					toolResultText = "Tool not found"
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(variant.ID, toolResultText, false))
			}
		}
		// No tool calls
		if len(toolResults) == 0 {
			var fallbackText string
			for _, block := range response.Content {
				if text, ok := block.AsAny().(anthropic.TextBlock); ok {
					fallbackText = text.Text
				}
			}
			err := fmt.Errorf("agent did not call a submit tool, got text instead: %s", fallbackText)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return res, err
		}

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}
