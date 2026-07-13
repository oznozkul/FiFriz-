package ai

import (
	"gorm.io/gorm"

	"fifriz/internal/ai/pipeline"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&ChatSession{}, &ChatMessage{}, &AIActionSuggestion{}); err != nil {
		return err
	}
	return pipeline.Migrate(db)
}
