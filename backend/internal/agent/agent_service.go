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
	"go.uber.org/zap"
)

const VISION_LABEL_SCORE_THRESHOLD = 0.75

type AgentImageRepository interface {
	GetImageWithLabels(ctx context.Context, id uuid.UUID, userID uuid.UUID, threshold float64) (*domain.Image, []string, error)
	GetFolderTopLabels(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, threshold float64, topN int) (*domain.FolderAggregate, error)
	GetFolderImageSamples(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, threshold float64, limit int) ([]*domain.Image, map[uuid.UUID][]string, error)
}

type AgentFolderRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]*domain.Folder, error)
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

func (u *AgentService) GetFolderSuggestion(ctx context.Context, userID uuid.UUID, imageID uuid.UUID) (Suggestion, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "agent.GetFolderSuggestion")
	defer span.End()

	res := Suggestion{}
	toolsUsed := []string{}
	defer func() {
		observability.LoggerFromContext(ctx, u.tel.Logger).Info("folder suggestion complete",
			zap.String("user_id", userID.String()),
			zap.String("image_id", imageID.String()),
			zap.Int("tool_call_count", len(toolsUsed)),
			zap.Strings("tools_used", toolsUsed),
		)
	}()

	tools := make([]anthropic.ToolUnionParam, len(folderSuggestionToolParams))
	for i, t := range folderSuggestionToolParams {
		tp := t
		tools[i] = anthropic.ToolUnionParam{OfTool: &tp}
	}

	img, labels, err := u.imageRepo.GetImageWithLabels(ctx, imageID, userID, VISION_LABEL_SCORE_THRESHOLD)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}

	imageMetadata, err := formatImageLabels(img.Title, labels)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}

	folders, err := u.folderRepo.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}
	folderList, err := formatFolderList(folders)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}

	userPrompt := fmt.Sprintf("Image metadata: %s\nFolder list: %s", imageMetadata, folderList)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
	}

	invalidInputCount := 0

	handleFolderIDTool := func(rawInput json.RawMessage, fetchFn func(context.Context, uuid.UUID, uuid.UUID) (string, error)) (string, error) {
		var input struct {
			FolderID string `json:"folder_id"`
		}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			invalidInputCount++
			if invalidInputCount >= 3 {
				capErr := fmt.Errorf("agent exceeded invalid input cap")
				span.RecordError(capErr)
				span.SetStatus(codes.Error, capErr.Error())
				return "", capErr
			}
			return "invalid folder ID format", nil
		}
		folderID, err := uuid.Parse(input.FolderID)
		if err != nil {
			invalidInputCount++
			if invalidInputCount >= 3 {
				capErr := fmt.Errorf("agent exceeded invalid input cap")
				span.RecordError(capErr)
				span.SetStatus(codes.Error, capErr.Error())
				return "", capErr
			}
			return "invalid folder ID format", nil
		}
		result, err := fetchFn(ctx, userID, folderID)
		if err != nil {
			invalidInputCount++
			if invalidInputCount >= 3 {
				capErr := fmt.Errorf("agent exceeded invalid input cap")
				span.RecordError(capErr)
				span.SetStatus(codes.Error, capErr.Error())
				return "", capErr
			}
			return "could not retrieve folder data for the given ID", nil
		}
		return result, nil
	}

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
				var toolResultText string
				switch variant.Name {
				case "get_folder_top_labels":
					toolsUsed = append(toolsUsed, variant.Name)
					result, err := handleFolderIDTool(variant.Input, func(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) (string, error) {
						return u.getFolderTopLabels(ctx, userID, folderID, findFolder(folders, folderID))
					})
					if err != nil {
						return res, err
					}
					toolResultText = result

				case "get_folder_image_samples":
					toolsUsed = append(toolsUsed, variant.Name)
					result, err := handleFolderIDTool(variant.Input, u.getFolderImageSamples)
					if err != nil {
						return res, err
					}
					toolResultText = result

				case "submit_existing_folder", "submit_new_folder":
					toolsUsed = append(toolsUsed, variant.Name)
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


func (a *AgentService) getFolderTopLabels(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, folder *domain.Folder) (string, error) {
	agg, err := a.imageRepo.GetFolderTopLabels(ctx, userID, folderID, VISION_LABEL_SCORE_THRESHOLD, 5)
	if err != nil {
		return "", err
	}

	return formatFolderTopLabels(folderID, folder, agg.ImageCount, agg.TopVisionLabels, agg.TopUserTags)
}

func (a *AgentService) getFolderImageSamples(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) (string, error) {
	images, labelMap, err := a.imageRepo.GetFolderImageSamples(ctx, userID, folderID, VISION_LABEL_SCORE_THRESHOLD, 5)
	if err != nil {
		return "", err
	}
	return formatFolderImageSamples(images, labelMap)
}
