package model

import (
	"time"

	"gorm.io/gorm"
)

type Factory struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	FactoryCode string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"factory_code"`
	FactoryName string         `gorm:"type:varchar(100);not null" json:"factory_name"`
	Address     string         `gorm:"type:varchar(255)" json:"address"`
	Contact     string         `gorm:"type:varchar(50)" json:"contact"`
	Phone       string         `gorm:"type:varchar(20)" json:"phone"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Factory) TableName() string {
	return "base_factory"
}

type Workshop struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	FactoryID    int64          `gorm:"not null;index" json:"factory_id"`
	WorkshopCode string         `gorm:"type:varchar(50);not null" json:"workshop_code"`
	WorkshopName string         `gorm:"type:varchar(100);not null" json:"workshop_name"`
	WorkshopType string         `gorm:"type:varchar(20)" json:"workshop_type"`
	Description  string         `gorm:"type:varchar(255)" json:"description"`
	Status       int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Factory      *Factory       `gorm:"foreignKey:FactoryID" json:"factory,omitempty"`
}

func (Workshop) TableName() string {
	return "base_workshop"
}

type ProductionLine struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkshopID  int64          `gorm:"not null;index" json:"workshop_id"`
	LineCode    string         `gorm:"type:varchar(50);not null" json:"line_code"`
	LineName    string         `gorm:"type:varchar(100);not null" json:"line_name"`
	Capacity    int            `gorm:"default:0" json:"capacity"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Workshop    *Workshop      `gorm:"foreignKey:WorkshopID" json:"workshop,omitempty"`
}

func (ProductionLine) TableName() string {
	return "base_production_line"
}

type Product struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductCode string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"product_code"`
	ProductName string         `gorm:"type:varchar(100);not null" json:"product_name"`
	ProductType string         `gorm:"type:varchar(50)" json:"product_type"`
	Version     string         `gorm:"type:varchar(20)" json:"version"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
	return "base_product"
}

type ProcessRoute struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   int64          `gorm:"not null;index" json:"product_id"`
	RouteCode   string         `gorm:"type:varchar(50);not null" json:"route_code"`
	RouteName   string         `gorm:"type:varchar(100);not null" json:"route_name"`
	Version     string         `gorm:"type:varchar(20)" json:"version"`
	IsDefault   int8           `gorm:"type:tinyint;default:0" json:"is_default"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Product     *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Operations  []Operation    `gorm:"foreignKey:RouteID" json:"operations,omitempty"`
}

func (ProcessRoute) TableName() string {
	return "base_process_route"
}

type Operation struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RouteID       int64          `gorm:"not null;index" json:"route_id"`
	OperationCode string         `gorm:"type:varchar(50);not null" json:"operation_code"`
	OperationName string         `gorm:"type:varchar(100);not null" json:"operation_name"`
	OperationType string         `gorm:"type:varchar(50)" json:"operation_type"`
	Sequence      int            `gorm:"not null" json:"sequence"`
	StandardTime  int            `gorm:"default:0" json:"standard_time"`
	Description   string         `gorm:"type:varchar(500)" json:"description"`
	Status        int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Route         *ProcessRoute  `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	Recipes       []Recipe       `gorm:"foreignKey:OperationID" json:"recipes,omitempty"`
}

func (Operation) TableName() string {
	return "base_operation"
}

type Recipe struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OperationID int64          `gorm:"not null;index" json:"operation_id"`
	RecipeCode  string         `gorm:"type:varchar(50);not null" json:"recipe_code"`
	RecipeName  string         `gorm:"type:varchar(100);not null" json:"recipe_name"`
	Version     string         `gorm:"type:varchar(20)" json:"version"`
	Parameters  string         `gorm:"type:text" json:"parameters"`
	IsDefault   int8           `gorm:"type:tinyint;default:0" json:"is_default"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Operation   *Operation     `gorm:"foreignKey:OperationID" json:"operation,omitempty"`
}

func (Recipe) TableName() string {
	return "base_recipe"
}
