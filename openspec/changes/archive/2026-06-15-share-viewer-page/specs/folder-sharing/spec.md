## MODIFIED Requirements

### Requirement: ShareUsecase — GetSharedFolder (public read)

`ShareUsecase` SHALL define:

```go
type SharedImage struct {
    Title        string
    ThumbnailURL *string
    FullResURL   string
    DownloadURL  string
    Width        *int
    Height       *int
}

type SharedFolder struct {
    Name   string
    Notes  *string
    Images []SharedImage
}

GetSharedFolder(ctx context.Context, token string) (*SharedFolder, error)
```

Behavior:
1. Call `FolderShareRepository.GetByToken(token)`. Return `gorm.ErrRecordNotFound` if the token is unknown.
2. Call `ShareImageRepository.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)` to get the folder's direct images, ordered by `image_folders.position` (same default ordering as the owner's folder view).
3. For each image, generate `ThumbnailURL` (nil if `ThumbnailPath` is nil) and `FullResURL` via `StorageService.GeneratePresignedGetURL` with the existing `presignedGetTTL`. Generate `DownloadURL` via `StorageService.GeneratePresignedDownloadURL(ctx, img.R2Path, filename, presignedGetTTL)`, where `filename` is `img.Title` plus the extension from `downloadFileExtension(img.MIMEType)`. Set `Width` and `Height` directly from `domain.Image.Width` and `domain.Image.Height`.
4. Return `SharedFolder{Name: share.Folder.Name, Notes: share.Folder.Description, Images: ...}`.

#### Scenario: Returns folder name, notes, and ordered images

- **WHEN** `GetSharedFolder` is called with a valid token for a folder with a description and multiple images
- **THEN** it returns the folder's `Name` and `Notes` (from `Description`)
- **AND** `Images` are ordered the same as `image_folders.position`
- **AND** each image has a non-empty `FullResURL` and a non-empty `DownloadURL`

#### Scenario: Image without a thumbnail has a nil ThumbnailURL

- **WHEN** `GetSharedFolder` is called for a folder containing an image with `ThumbnailPath == nil`
- **THEN** that image's `SharedImage.ThumbnailURL` is `nil`

#### Scenario: Image dimensions are passed through

- **WHEN** `GetSharedFolder` is called for a folder containing an image with non-nil `domain.Image.Width` and `Height`
- **THEN** that image's `SharedImage.Width` and `SharedImage.Height` equal the source image's `Width` and `Height`

#### Scenario: Image with no recorded dimensions has nil Width and Height

- **WHEN** `GetSharedFolder` is called for a folder containing an image with `Width == nil` and `Height == nil`
- **THEN** that image's `SharedImage.Width` and `SharedImage.Height` are `nil`

#### Scenario: Unknown token returns not-found

- **WHEN** `GetSharedFolder` is called with a token that does not match any `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: Folder with no images returns an empty list

- **WHEN** `GetSharedFolder` is called for a shared folder with zero direct images
- **THEN** it returns `SharedFolder.Images` as an empty slice and no error

---

### Requirement: Public Shared Folder Endpoint

The system SHALL expose `GET /share/:token` as a public (unauthenticated) route, registered outside the `protected` group but still behind the global CORS and recovery middleware. The handler SHALL call `ShareUsecase.GetSharedFolder(ctx, token)` and respond:

- `200 OK` with body:
  ```json
  {
    "folder": { "name": "...", "notes": "...|null" },
    "images": [
      { "title": "...", "thumbnail_url": "...|null", "full_res_url": "...", "download_url": "...", "width": 0, "height": 0 }
    ]
  }
  ```
  `width` and `height` SHALL be `null` when the corresponding `SharedImage.Width`/`Height` is `nil`. `full_res_url` SHALL be a presigned URL suitable for inline display (e.g. `<img src>`), while `download_url` SHALL be a presigned URL with `Content-Disposition: attachment; filename="<image title and extension>"` set, suitable for triggering a file download.
- `404 Not Found` if `GetSharedFolder` returns `gorm.ErrRecordNotFound`

#### Scenario: Valid token returns folder and images

- **WHEN** `GET /share/:token` is called with a token for a shared folder
- **THEN** the response is `200 OK`
- **AND** the body contains `folder.name`, `folder.notes`, and an `images` array with `title`, `thumbnail_url`, `full_res_url`, `download_url`, `width`, and `height` per image

#### Scenario: Unknown or revoked token returns 404

- **WHEN** `GET /share/:token` is called with a token that does not exist (including a previously-valid token that has since been revoked via `DELETE /folders/:id/share`)
- **THEN** the response is `404 Not Found`

#### Scenario: No authentication required

- **WHEN** `GET /share/:token` is called without any Authorization header
- **THEN** the request is not rejected by auth middleware
