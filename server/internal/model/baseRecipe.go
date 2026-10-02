package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseRecipe struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperationID int            `gorm:"column:operation_id;type:integer;not null" json:"operationID"`
	RecipeCode  string         `gorm:"column:recipe_code;type:text;not null" json:"recipeCode"`
	RecipeName  string         `gorm:"column:recipe_name;type:text;not null" json:"recipeName"`
	Version     string         `gorm:"column:version;type:text" json:"version"`
	Parameters  string         `gorm:"column:parameters;type:text" json:"parameters"`
	IsDefault   int            `gorm:"column:is_default;type:integer;not null" json:"isDefault"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Status      int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseRecipe) TableName() string {
	return "base_recipe"
}

// BaseRecipeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseRecipeColumnNames = map[string]bool{
	"id":           true,
	"operation_id": true,
	"recipe_code":  true,
	"recipe_name":  true,
	"version":      true,
	"parameters":   true,
	"is_default":   true,
	"description":  true,
	"status":       true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
}
