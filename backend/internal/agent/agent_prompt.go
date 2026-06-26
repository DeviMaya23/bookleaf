package agent

import "github.com/anthropics/anthropic-sdk-go"

const folderSuggestionSystemPrompt = `You are helping a user organize images into folders in their personal image reference library.
You will be given an image's metadata (name and visual labels). Use the get_folder_list tool to see the user's existing folders, including each folder's name, description, and full path.
Decide whether an existing folder is a good fit for the image, using the folder's description and path — not just its name, since names alone are often too generic.
If an existing folder fits well, call submit_existing_folder with that folder's ID.
If no existing folder is a good fit, call submit_new_folder with a suggested folder name, and optionally a parent folder ID if it should be a subfolder of an existing one. Do not force a poor match just to avoid creating a new folder.
Always provide your reasoning when calling either tool.`

var folderSuggestionToolParams = []anthropic.ToolParam{
	{
		Name:        "get_folder_list",
		Description: anthropic.String("Get a list of all folders belonging to the current user."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{},
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
