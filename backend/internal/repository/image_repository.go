package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"roci.dev/fracdex"
)

const labelSearchScoreThreshold float32 = 0.75

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *imageRepository {
	return &imageRepository{
		db: db,
	}
}

func (r *imageRepository) Create(ctx context.Context, image *domain.Image) (*domain.Image, error) {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return nil, fmt.Errorf("insert image: %w", err)
	}

	return image, nil
}

func (r *imageRepository) List(ctx context.Context, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, searchLabels bool, sortField *string, direction *string, cursor *usecase.ImageCursor, limit int) ([]*domain.Image, error) {
	var images []*domain.Image
	dispatch := usecase.ResolveSort(sortField, direction)

	query := r.db.WithContext(ctx).
		Where("images.user_id = ?", userID).
		Order(dispatch.OrderClause).
		Limit(limit + 1).
		Preload("Tags").
		Preload("ImageFolders")

	if unfiled {
		query = query.
			Joins("LEFT JOIN image_folders ON image_folders.image_id = images.id").
			Where("image_folders.image_id IS NULL")
	}

	if len(folderIDs) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM image_folders WHERE image_folders.image_id = images.id AND image_folders.folder_id IN (?))", folderIDs)
	}

	if len(tagIDs) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND image_tags.tag_id IN (?))", tagIDs)
	}

	if len(mimeTypes) > 0 {
		query = query.Where("images.mime_type IN (?)", mimeTypes)
	}

	if name != nil && *name != "" {
		if searchLabels {
			query = query.Where("(images.title ILIKE ? OR EXISTS (SELECT 1 FROM image_labels WHERE image_id = images.id AND label ILIKE ? AND score >= ?))", "%"+*name+"%", "%"+*name+"%", labelSearchScoreThreshold)
		} else {
			query = query.Where("images.title ILIKE ?", "%"+*name+"%")
		}
	}

	if cursor != nil {
		if dispatch.Column == "title" && cursor.Title != nil {
			query = query.Where(fmt.Sprintf("(images.title, images.id) %s (?, ?)", dispatch.WhereOperator), *cursor.Title, cursor.ID)
		} else {
			query = query.Where(fmt.Sprintf("(images.created_at, images.id) %s (?, ?)", dispatch.WhereOperator), cursor.CreatedAt, cursor.ID)
		}
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	return images, nil
}

func (r *imageRepository) ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error) {
	var images []*domain.Image
	dispatch := usecase.ResolveSort(sortField, direction)

	query := r.db.WithContext(ctx).
		Where("images.user_id = ?", userID).
		Joins("JOIN image_folders ON image_folders.image_id = images.id AND image_folders.folder_id = ?", folderID).
		Preload("Tags").
		Preload("ImageFolders")

	if sortField != nil {
		query = query.Order(dispatch.OrderClause)
	} else {
		query = query.Order("image_folders.position ASC")
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list images by folder: %w", err)
	}

	return images, nil
}

func (r *imageRepository) SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error {
	if folderID == nil {
		if err := r.db.WithContext(ctx).
			Where("image_id = ?", imageID).
			Delete(&domain.ImageFolder{}).Error; err != nil {
			return fmt.Errorf("delete image folder: %w", err)
		}
		return nil
	}

	var maxPos string
	r.db.WithContext(ctx).
		Model(&domain.ImageFolder{}).
		Where("folder_id = ?", *folderID).
		Select("COALESCE(MAX(position), '')").
		Scan(&maxPos)
	position, err := fracdex.KeyBetween(maxPos, "")
	if err != nil {
		return fmt.Errorf("generate position key: %w", err)
	}

	row := domain.ImageFolder{
		ImageID:  imageID,
		FolderID: *folderID,
		Position: position,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("insert image folder: %w", err)
	}
	return nil
}

func (r *imageRepository) AddImageToFolder(ctx context.Context, imageID, folderID uuid.UUID) error {
	var maxPos string
	r.db.WithContext(ctx).
		Model(&domain.ImageFolder{}).
		Where("folder_id = ?", folderID).
		Select("COALESCE(MAX(position), '')").
		Scan(&maxPos)
	position, err := fracdex.KeyBetween(maxPos, "")
	if err != nil {
		return fmt.Errorf("generate position key: %w", err)
	}

	row := domain.ImageFolder{
		ImageID:  imageID,
		FolderID: folderID,
		Position: position,
	}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "image_id"}, {Name: "folder_id"}},
			DoNothing: true,
		}).
		Create(&row).Error; err != nil {
		return fmt.Errorf("insert image folder: %w", err)
	}
	return nil
}

func (r *imageRepository) FilterOwnedImageIDs(ctx context.Context, ids []uuid.UUID, userID string) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var owned []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id IN (?) AND user_id = ?", ids, userID).
		Pluck("id", &owned).Error; err != nil {
		return nil, fmt.Errorf("filter owned image ids: %w", err)
	}

	return owned, nil
}

func (r *imageRepository) SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current []domain.ImageFolder
		if err := tx.Where("image_id = ?", imageID).Find(&current).Error; err != nil {
			return fmt.Errorf("list image folders: %w", err)
		}

		currentSet := make(map[uuid.UUID]struct{}, len(current))
		for _, row := range current {
			currentSet[row.FolderID] = struct{}{}
		}

		desiredSet := make(map[uuid.UUID]struct{}, len(folderIDs))
		toAdd := make([]uuid.UUID, 0)
		for _, folderID := range folderIDs {
			if _, seen := desiredSet[folderID]; seen {
				continue
			}
			desiredSet[folderID] = struct{}{}
			if _, exists := currentSet[folderID]; !exists {
				toAdd = append(toAdd, folderID)
			}
		}

		toDelete := make([]uuid.UUID, 0)
		for folderID := range currentSet {
			if _, keep := desiredSet[folderID]; !keep {
				toDelete = append(toDelete, folderID)
			}
		}

		if len(toDelete) > 0 {
			if err := tx.
				Where("image_id = ? AND folder_id IN ?", imageID, toDelete).
				Delete(&domain.ImageFolder{}).Error; err != nil {
				return fmt.Errorf("delete image folders: %w", err)
			}
		}

		for _, folderID := range toAdd {
			var maxPos string
			if err := tx.Model(&domain.ImageFolder{}).
				Where("folder_id = ?", folderID).
				Select("COALESCE(MAX(position), '')").
				Scan(&maxPos).Error; err != nil {
				return fmt.Errorf("select image folder position: %w", err)
			}
			position, err := fracdex.KeyBetween(maxPos, "")
			if err != nil {
				return fmt.Errorf("generate position key: %w", err)
			}
			row := domain.ImageFolder{
				ImageID:  imageID,
				FolderID: folderID,
				Position: position,
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("insert image folder: %w", err)
			}
		}

		return nil
	})
}

func (r *imageRepository) MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if fromFolderID != nil {
			if err := tx.
				Where("image_id = ? AND folder_id = ?", imageID, *fromFolderID).
				Delete(&domain.ImageFolder{}).Error; err != nil {
				return fmt.Errorf("delete image folder: %w", err)
			}
		}

		if toFolderID != nil {
			var maxPos string
			if err := tx.Model(&domain.ImageFolder{}).
				Where("folder_id = ?", *toFolderID).
				Select("COALESCE(MAX(position), '')").
				Scan(&maxPos).Error; err != nil {
				return fmt.Errorf("select image folder position: %w", err)
			}
			position, err := fracdex.KeyBetween(maxPos, "")
			if err != nil {
				return fmt.Errorf("generate position key: %w", err)
			}
			row := domain.ImageFolder{
				ImageID:  imageID,
				FolderID: *toFolderID,
				Position: position,
			}
			if err := tx.
				Where(domain.ImageFolder{ImageID: imageID, FolderID: *toFolderID}).
				FirstOrCreate(&row).Error; err != nil {
				return fmt.Errorf("insert image folder: %w", err)
			}
		}

		return nil
	})
}

func (r *imageRepository) UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.ImageFolder{}).
		Where("image_id = ? AND folder_id = ?", imageID, folderID).
		Update("position", position)
	if result.Error != nil {
		return fmt.Errorf("update image folder position: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image folder position: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	var image domain.Image
	if err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("ImageFolders").
		Where("id = ? AND user_id = ?", id, userID).
		First(&image).Error; err != nil {
		return nil, fmt.Errorf("select image: %w", err)
	}

	return &image, nil
}

func (r *imageRepository) GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	var image domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Preload("ImageFolders").
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		First(&image).Error; err != nil {
		return nil, fmt.Errorf("select deleted image: %w", err)
	}

	return &image, nil
}

func (r *imageRepository) UpdateThumbnailPath(ctx context.Context, id uuid.UUID, thumbnailPath string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ?", id).
		Update("thumbnail_path", thumbnailPath)
	if result.Error != nil {
		return fmt.Errorf("update image thumbnail_path: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image thumbnail_path: %w", gorm.ErrRecordNotFound)
	}

	return nil
}


func (r *imageRepository) UpdateLabels(ctx context.Context, id uuid.UUID, rawJSON json.RawMessage, labels []domain.ImageLabel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domain.Image{}).Where("id = ?", id).Update("ai_labels", rawJSON)
		if result.Error != nil {
			return fmt.Errorf("update image ai_labels: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("update image ai_labels: %w", gorm.ErrRecordNotFound)
		}

		if err := tx.Where("image_id = ?", id).Delete(&domain.ImageLabel{}).Error; err != nil {
			return fmt.Errorf("delete image labels: %w", err)
		}

		if len(labels) > 0 {
			if err := tx.Create(&labels).Error; err != nil {
				return fmt.Errorf("insert image labels: %w", err)
			}
		}

		return nil
	})
}

func (r *imageRepository) Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields)
	if result.Error != nil {
		return nil, fmt.Errorf("update image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update image: %w", gorm.ErrRecordNotFound)
	}

	return r.GetByID(ctx, id, userID) // includes Preload("Tags")
}

func (r *imageRepository) SoftDelete(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Image{})
	if result.Error != nil {
		return fmt.Errorf("soft delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("soft delete image: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) Restore(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Model(&domain.Image{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("restore image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("restore image: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) ListTrashed(ctx context.Context, userID string, name *string, sortField *string, direction *string, cursor *usecase.ImageCursor, limit int) ([]*domain.Image, error) {
	var images []*domain.Image
	dispatch := usecase.ResolveSort(sortField, direction)

	query := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND user_id = ?", userID).
		Order(dispatch.OrderClause).
		Limit(limit + 1)

	if name != nil && *name != "" {
		query = query.Where("images.title ILIKE ?", "%"+*name+"%")
	}

	if cursor != nil {
		switch {
		case dispatch.Column == "title" && cursor.Title != nil:
			query = query.Where(fmt.Sprintf("(images.title, images.id) %s (?, ?)", dispatch.WhereOperator), *cursor.Title, cursor.ID)
		case dispatch.Column == "deleted_at" && cursor.DeletedAt != nil:
			query = query.Where(fmt.Sprintf("(images.deleted_at, images.id) %s (?, ?)", dispatch.WhereOperator), *cursor.DeletedAt, cursor.ID)
		default:
			query = query.Where(fmt.Sprintf("(images.created_at, images.id) %s (?, ?)", dispatch.WhereOperator), cursor.CreatedAt, cursor.ID)
		}
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list trashed images: %w", err)
	}

	return images, nil
}

func (r *imageRepository) ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND user_id = ?", userID).
		Order("deleted_at ASC, id ASC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list all trashed images: %w", err)
	}
	return images, nil
}

func (r *imageRepository) CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Joins("JOIN image_folders ON image_folders.image_id = images.id").
		Where("image_folders.folder_id = ?", folderID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count images by folder: %w", err)
	}

	return count, nil
}

func (r *imageRepository) ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", olderThan).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list expired trash: %w", err)
	}
	return images, nil
}

func (r *imageRepository) HardDelete(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Image{})
	if result.Error != nil {
		return fmt.Errorf("hard delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("hard delete image: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *imageRepository) ListAllByUserID(ctx context.Context, userID string) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ?", userID).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list all images by user: %w", err)
	}
	return images, nil
}

func (r *imageRepository) HardDeleteAllByUserID(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ?", userID).
		Delete(&domain.Image{}).Error; err != nil {
		return fmt.Errorf("hard delete all images by user: %w", err)
	}
	return nil
}

func (r *imageRepository) FindDuplicates(ctx context.Context, userID string, phash string, excludeID uuid.UUID, threshold int) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND id != ? AND phash IS NOT NULL AND bit_count(phash # ?::bit(64)) <= ?", userID, excludeID, phash, threshold).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("find duplicate images: %w", err)
	}
	return images, nil
}

func (r *imageRepository) ListUnhashed(ctx context.Context, limit int) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("phash IS NULL AND deleted_at IS NULL").
		Limit(limit).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list unhashed images: %w", err)
	}
	return images, nil
}

func (r *imageRepository) ListUnlabelled(ctx context.Context, userID string) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND ai_labels IS NULL AND deleted_at IS NULL", userID).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list unlabelled images: %w", err)
	}
	return images, nil
}

func (r *imageRepository) UpdatePHash(ctx context.Context, id uuid.UUID, phash string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ?", id).
		Update("phash", phash)
	if result.Error != nil {
		return fmt.Errorf("update image phash: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image phash: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *imageRepository) GetImageWithLabels(ctx context.Context, id uuid.UUID, userID string, threshold float64) (*domain.Image, []string, error) {
	type row struct {
		domain.Image
		LabelText *string  `gorm:"column:il_label"`
		LabelScore *float32 `gorm:"column:il_score"`
	}

	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT i.*, il.label AS il_label, il.score AS il_score
		FROM images i
		LEFT JOIN image_labels il ON il.image_id = i.id AND il.score >= ?
		WHERE i.id = ? AND i.user_id = ? AND i.deleted_at IS NULL
		ORDER BY il.score DESC NULLS LAST
	`, threshold, id, userID).Scan(&rows).Error
	if err != nil {
		return nil, nil, fmt.Errorf("get image with labels: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("get image with labels: %w", gorm.ErrRecordNotFound)
	}

	img := rows[0].Image
	labels := make([]string, 0)
	for _, row := range rows {
		if row.LabelText != nil {
			labels = append(labels, *row.LabelText)
		}
	}

	return &img, labels, nil
}

func (r *imageRepository) GetFolderTopLabels(ctx context.Context, userID string, folderID uuid.UUID, threshold float64, topN int) (*domain.FolderAggregate, error) {
	type nameRow struct {
		Name string `gorm:"column:name"`
	}

	db := r.db.WithContext(ctx)

	var countResult struct {
		Count int `gorm:"column:count"`
	}
	if err := db.Raw(`
		SELECT COUNT(DISTINCT i.id) AS count
		FROM images i
		JOIN image_folders imf ON imf.image_id = i.id AND imf.folder_id = ?
		WHERE i.user_id = ? AND i.deleted_at IS NULL
	`, folderID, userID).Scan(&countResult).Error; err != nil {
		return nil, fmt.Errorf("get folder image count: %w", err)
	}

	var labelRows []nameRow
	if err := db.Raw(`
		SELECT il.label AS name
		FROM image_labels il
		JOIN image_folders imf ON imf.image_id = il.image_id AND imf.folder_id = ?
		JOIN images i ON i.id = il.image_id AND i.user_id = ? AND i.deleted_at IS NULL
		WHERE il.score >= ?
		GROUP BY il.label
		ORDER BY COUNT(*) DESC, il.label ASC
		LIMIT ?
	`, folderID, userID, threshold, topN).Scan(&labelRows).Error; err != nil {
		return nil, fmt.Errorf("get folder top vision labels: %w", err)
	}

	var tagRows []nameRow
	if err := db.Raw(`
		SELECT t.name AS name
		FROM tags t
		JOIN image_tags it ON it.tag_id = t.id
		JOIN image_folders imf ON imf.image_id = it.image_id AND imf.folder_id = ?
		JOIN images i ON i.id = it.image_id AND i.user_id = ? AND i.deleted_at IS NULL
		GROUP BY t.name
		ORDER BY COUNT(*) DESC, t.name ASC
		LIMIT ?
	`, folderID, userID, topN).Scan(&tagRows).Error; err != nil {
		return nil, fmt.Errorf("get folder top user tags: %w", err)
	}

	agg := &domain.FolderAggregate{
		ImageCount:      countResult.Count,
		TopVisionLabels: make([]string, len(labelRows)),
		TopUserTags:     make([]string, len(tagRows)),
	}
	for i, row := range labelRows {
		agg.TopVisionLabels[i] = row.Name
	}
	for i, row := range tagRows {
		agg.TopUserTags[i] = row.Name
	}
	return agg, nil
}

func (r *imageRepository) GetFolderImageSamples(ctx context.Context, userID string, folderID uuid.UUID, threshold float64, limit int) ([]*domain.Image, map[uuid.UUID][]string, error) {
	var images []*domain.Image
	err := r.db.WithContext(ctx).
		Joins("JOIN image_folders imf ON imf.image_id = images.id AND imf.folder_id = ?", folderID).
		Where("images.user_id = ? AND images.deleted_at IS NULL", userID).
		Order("images.created_at DESC").
		Limit(limit).
		Find(&images).Error
	if err != nil {
		return nil, nil, fmt.Errorf("get folder image samples: %w", err)
	}

	labelMap := make(map[uuid.UUID][]string, len(images))
	for _, img := range images {
		labelMap[img.ID] = []string{}
	}

	if len(images) == 0 {
		return images, labelMap, nil
	}

	ids := make([]uuid.UUID, len(images))
	for i, img := range images {
		ids[i] = img.ID
	}

	var labelRows []domain.ImageLabel
	err = r.db.WithContext(ctx).
		Where("image_id IN ? AND score >= ?", ids, threshold).
		Find(&labelRows).Error
	if err != nil {
		return nil, nil, fmt.Errorf("get folder image sample labels: %w", err)
	}

	for _, l := range labelRows {
		labelMap[l.ImageID] = append(labelMap[l.ImageID], l.Label)
	}

	return images, labelMap, nil
}

var _ usecase.ImageRepository = (*imageRepository)(nil)
var _ usecase.ImageHashRepository = (*imageRepository)(nil)
