package agent

import "github.com/anthropics/anthropic-sdk-go"

const folderSuggestionSystemPrompt = `You are helping a user organize images into folders in their personal image reference library.
You will be given an image's metadata (name and visual labels) and a list of user's existing folders, including each folder's name, description, and full path.
Decide whether an existing folder is a good fit for the image, or if a new folder should be created.
If the folder names and descriptions give you enough confidence to decide, call submit_existing_folder or submit_new_folder directly.
If you are uncertain between two or more folders, or a folder's description is missing or too vague to judge, you may call get_folder_top_labels on a specific candidate folder to see its most common labels and tags.
If the aggregated summary still isn't enough to decide — for example, if you want to see what specific images already exist in a folder — you may call get_folder_image_samples on that folder to see a few example images directly.
Use these tools only on folders you are genuinely considering, not on every folder in the list.
Always provide your reasoning when calling either tool.`

var folderSuggestionToolParams = []anthropic.ToolParam{
	{
		Name:        "get_folder_top_labels",
		Description: anthropic.String("Get aggregated vision labels and user tags for all images in a folder."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"folder_id": map[string]any{
					"type":        "string",
					"description": "The ID of the folder to aggregate labels for.",
				},
			},
			Required: []string{"folder_id"},
		},
	},
	{
		Name:        "get_folder_image_samples",
		Description: anthropic.String("Get metadata for up to 5 of the newest images in a folder."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"folder_id": map[string]any{
					"type":        "string",
					"description": "The ID of the folder to sample images from.",
				},
			},
			Required: []string{"folder_id"},
		},
	},
	{
		Name:        "submit_existing_folder",
		Description: anthropic.String("Submit your suggestion: place the image in an existing folder."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"folder_id": map[string]any{
					"type":        "string",
					"description": "The ID of the existing folder the image should be placed under.",
				},
				"reasoning": map[string]any{
					"type":        "string",
					"description": "Why this folder is the right fit.",
				},
			},
			Required: []string{"folder_id", "reasoning"},
		},
	},
	{
		Name:        "submit_new_folder",
		Description: anthropic.String("Submit your suggestion: create a new folder for the image, since no existing folder fits."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"new_folder_name": map[string]any{
					"type":        "string",
					"description": "Name of the new folder to create.",
				},
				"new_folder_parent_id": map[string]any{
					"type":        "string",
					"description": "If this should be a subfolder of an existing folder, its ID. Omit for a top-level folder.",
				},
				"reasoning": map[string]any{
					"type":        "string",
					"description": "Why no existing folder fits and why this new folder name makes sense.",
				},
			},
			Required: []string{"new_folder_name", "reasoning"},
		},
	},
}
