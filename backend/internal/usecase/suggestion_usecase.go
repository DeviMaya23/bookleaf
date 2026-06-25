package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type suggestionUsecase struct {
	aiClient *anthropic.Client
}

type suggestionImageRepository interface {
}

type suggestionFolderRepository interface {
}

func NewSuggestionUsecase(aiClient *anthropic.Client) *suggestionUsecase {
	return &suggestionUsecase{
		aiClient: aiClient,
	}
}

func (u *suggestionUsecase) Test2(ctx context.Context) string {
	return list_folder()
}

func (u *suggestionUsecase) Test(ctx context.Context, userID string, input string) (string, error) {

	systemPrompt := "You are helping a user organise their image into folders in a personal reference library. Given an image's metadata and a list of existing folders, give a suggesetion on which folder does the image belong to, or suggest a new folder."
	systemPrompt += "\nUse the folder's name, description, and path to determine the best fit for the image."
	systemPrompt += "\nIf no folder is a good fit, suggest a new folder name and description. If you suggest a new folder, also think about if it should be a subfolder of an existing folder, and if so which one."
	systemPrompt += "\nProvide exactly one of: folder_id (existing folder fits), or both new_folder_name and optionally new_folder_parent_id (no existing folder fits)."

	tools := make([]anthropic.ToolUnionParam, len(toolParams))
	for i, t := range toolParams {
		tp := t
		tools[i] = anthropic.ToolUnionParam{OfTool: &tp}
	}

	// lily of the valley
	// imageMetadata := map[string]any{
	// 	"image_name":     "Convallaria majalis (Lily-of-the-Valley) _ K. van Bourgondien",
	// 	"vision_results": []string{"Flower", "Lily of the valley", "Flowering plant", "Plant stem", "Perennial plant", "Dicotyledon", "Jasmine", "Cornales", "Cheesewoods"},
	// }

	// game screenshot
	// imageMetadata := map[string]any{
	// 	"image_name": "ffxiv_19042025_150825_451",
	// 	// [{"Score": 0.9239293, "Description": "Game"}, {"Score": 0.905271, "Description": "Fictional character"}, {"Score": 0.83129525, "Description": "Animation"}, {"Score": 0.82382756, "Description": "CG artwork"}, {"Score": 0.8117272, "Description": "Video Game Software"}, {"Score": 0.76375175, "Description": "PC game"}, {"Score": 0.75776386, "Description": "Action-adventure game"}, {"Score": 0.68590426, "Description": "Strategy video game"}, {"Score": 0.6469653, "Description": "Fiction"}, {"Score": 0.6184202, "Description": "Animated cartoon"}]
	// 	"vision_results": []string{"Game", "Fictional character", "Animation", "CG artwork", "Video Game Software", "PC game", "Action-adventure game", "Strategy video game", "Fiction", "Animated cartoon"},
	// }
	// metadataJSON, _ := json.Marshal(imageMetadata)
	// userPrompt := fmt.Sprintf("\nImage metadata: %s", string(metadataJSON))

	// food 1
	// imageMetadata := map[string]any{
	// 	"image_name":     "Nasi Goreng",
	// 	"vision_results": []string{"Food", "Ingredient", "Cooking", "Rice", "Vegetable", "Recipe", "Staple food", "Produce", "Fried rice", "Cooked rice"},
	// }

	// food 2
	imageMetadata := map[string]any{
		"image_name":     "Greek_lentil_soup_SQ",
		"vision_results": []string{"Food", "Ingredient", "Produce", "Stew", "Legume", "Soup", "Recipe", "Cooking", "Vegetable", "Curry"},
	}
	metadataJSON, _ := json.Marshal(imageMetadata)
	userPrompt := fmt.Sprintf("\nImage metadata: %s", string(metadataJSON))

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
	}

	var result string

	for {
		response, err := u.aiClient.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_6,
			MaxTokens: 1024,
			Tools:     tools,
			Messages:  messages,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
		})
		if err != nil {
			fmt.Printf("API error: %v\n", err)
			return "", err
		}

		messages = append(messages, response.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}

		for _, block := range response.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.ToolUseBlock:
				switch variant.Name {
				case "get_folder_list":
					result = list_folder()
				case "submit_suggestion":
					var suggestion Suggestion
					if err := json.Unmarshal([]byte(variant.Input), &suggestion); err != nil {
						result = fmt.Sprintf("invalid input: %v", err)
					} else {
						result = fmt.Sprintf("Suggestion received: folder_id=%s, new_folder_name=%s, new_folder_parent_id=%s, reasoning=%s",
							suggestion.FolderID, suggestion.NewFolderName, suggestion.NewFolderParentID, suggestion.Reasoning)
					}
					return result, nil
				default:
					result = "Tool not found"
				}

				fmt.Printf("Tool result: %s\n", result)
				toolResults = append(toolResults, anthropic.NewToolResultBlock(variant.ID, result, false))
			}
		}

		// No tool calls — we have the final answer
		if len(toolResults) == 0 {
			for _, block := range response.Content {
				if text, ok := block.AsAny().(anthropic.TextBlock); ok {
					fmt.Printf("\nAgent: %s\n", text.Text)
				}
			}
			break
		}

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	return result, nil
}

type Suggestion struct {
	FolderID          string `json:"folder_id,omitempty"`
	NewFolderName     string `json:"new_folder_name,omitempty"`
	NewFolderParentID string `json:"new_folder_parent_id,omitempty"`
	Reasoning         string `json:"reasoning"`
}

var toolParams = []anthropic.ToolParam{
	{
		Name:        "get_folder_list",
		Description: anthropic.String("Get a list of all folders belonging to the current user."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{},
		},
	},
	{
		Name:        "submit_suggestion",
		Description: anthropic.String("Submit your final suggestion for the image."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"folder_id": map[string]any{
					"type":        "string",
					"description": "The ID of the folder the image should be under. Empty string if suggesting a new folder.",
				},
				"new_folder_name": map[string]any{
					"type":        "string",
					"description": "The name of the new folder being suggested. Empty string if not suggesting a new folder.",
				},
				"new_folder_parent_id": map[string]any{
					"type":        "string",
					"description": "If suggesting a new folder, the ID of the existing folder it should be created under. Empty string for a top-level folder.",
				},
				"reasoning": map[string]any{
					"type":        "string",
					"description": "Your reasoning for why the image belongs in the suggested folder.",
				},
			},
			Required: []string{"reasoning"},
		},
	},
}

func list_folder() string {

	// dummy data
	dummyListFolder := []map[string]interface{}{
		{
			"folderID":          0,
			"folderName":        "North",
			"folderPath":        "North",
			"folderDescription": "Visual aesthetics for the northern lands. Cold, mountainous, rural. Agricultural heavy,",
		},
		{
			"folderID":          1,
			"folderName":        "Food",
			"folderPath":        "North > Food",
			"folderDescription": "Food inspirations for the north",
		},
		{
			"folderID":          2,
			"folderName":        "Estate",
			"folderPath":        "North > Estate",
			"folderDescription": "Aesthetic/inspiration for gentry manor in the north",
		},
		{
			"folderID":          3,
			"folderName":        "Plants and Forages",
			"folderPath":        "North > Plants and Forages",
			"folderDescription": "Medicinal usage/herbs/foraged/mushroom and the like.",
		},
		{
			"folderID":          4,
			"folderName":        "Food",
			"folderPath":        "Food",
			"folderDescription": "random food pics",
		},
		{
			"folderID":          5,
			"folderName":        "FFXIV",
			"folderPath":        "FFXIV",
			"folderDescription": "",
		},
		{
			"folderID":          6,
			"folderName":        "Dairy",
			"folderPath":        "North > Food > Dairy",
			"folderDescription": "cheese, milk",
		},
		{
			"folderID":          7,
			"folderName":        "Eastern",
			"folderPath":        "Eastern",
			"folderDescription": "General aesthetic moodboard for eastern lands (heavy medieval Chinese inspiration)",
		},
		{
			"folderID":          8,
			"folderName":        "Buildings",
			"folderPath":        "Eastern > Buildings",
			"folderDescription": "Building/estates moodboard for the east.",
		},
	}

	// change to string

	dummyListFolderJSON, _ := json.Marshal(dummyListFolder)
	return string(dummyListFolderJSON)

}
