package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"roci.dev/fracdex"
)

type imageFolderRow struct {
	ImageID  uuid.UUID `gorm:"column:image_id"`
	FolderID uuid.UUID `gorm:"column:folder_id"`
}

func main() {
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "postgres connection string (or set DATABASE_URL)")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("database connection string required: set DATABASE_URL or pass -dsn <url>")
	}

	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	ctx := context.Background()

	// Collect all (folder_id, image_id) pairs ordered by folder then current position.
	var rows []imageFolderRow
	if err := db.WithContext(ctx).
		Table("image_folders").
		Select("folder_id, image_id").
		Order("folder_id ASC, position ASC").
		Scan(&rows).Error; err != nil {
		log.Fatalf("query image_folders: %v", err)
	}

	if len(rows) == 0 {
		log.Println("no rows to rebalance")
		os.Exit(0)
	}

	// Group by folder.
	type folderGroup struct {
		folderID uuid.UUID
		imageIDs []uuid.UUID
	}
	var groups []folderGroup
	for _, row := range rows {
		if len(groups) == 0 || groups[len(groups)-1].folderID != row.FolderID {
			groups = append(groups, folderGroup{folderID: row.FolderID})
		}
		groups[len(groups)-1].imageIDs = append(groups[len(groups)-1].imageIDs, row.ImageID)
	}

	totalUpdated := 0
	for _, g := range groups {
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			prev := ""
			for _, imageID := range g.imageIDs {
				key, err := fracdex.KeyBetween(prev, "")
				if err != nil {
					return fmt.Errorf("generate key after %q: %w", prev, err)
				}
				if err := tx.Exec(
					"UPDATE image_folders SET position = ? WHERE image_id = ? AND folder_id = ?",
					key, imageID, g.folderID,
				).Error; err != nil {
					return fmt.Errorf("update image %s: %w", imageID, err)
				}
				prev = key
			}
			return nil
		})
		if err != nil {
			log.Fatalf("rebalance folder %s: %v", g.folderID, err)
		}
		log.Printf("rebalanced folder %s: %d images", g.folderID, len(g.imageIDs))
		totalUpdated += len(g.imageIDs)
	}

	log.Printf("done: %d images across %d folders rebalanced", totalUpdated, len(groups))
}
