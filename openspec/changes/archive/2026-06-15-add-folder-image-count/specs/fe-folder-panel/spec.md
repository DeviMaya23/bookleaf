## ADDED Requirements

### Requirement: Folder image count subtitle

`FolderPanelContent` SHALL display the folder's image count as a subtitle below the folder name input, using `folderDetail.image_count` from the existing folder detail query. The label SHALL use singular form for a count of `1` and plural form otherwise (including `0`): "1 image" or "{n} images".

While the folder detail query is loading (`folderDetail` is `undefined`), the subtitle SHALL NOT be rendered.

#### Scenario: Folder with multiple images

- **WHEN** the folder detail query resolves with `image_count: 12`
- **THEN** the panel header shows "12 images" below the folder name

#### Scenario: Folder with exactly one image

- **WHEN** the folder detail query resolves with `image_count: 1`
- **THEN** the panel header shows "1 image" below the folder name

#### Scenario: Empty folder

- **WHEN** the folder detail query resolves with `image_count: 0`
- **THEN** the panel header shows "0 images" below the folder name

#### Scenario: Subtitle hidden while loading

- **WHEN** the folder detail query has not yet resolved (`folderDetail` is `undefined`)
- **THEN** no image count subtitle is rendered
