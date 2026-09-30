#!/bin/bash

# Generate DAO files for all entities

ENTITIES=(
  "Factory:factory"
  "Workshop:workshop"
  "ProductionLine:production_line"
  "Product:product"
  "ProcessRoute:process_route"
  "Operation:operation"
  "Recipe:recipe"
)

for entity in "${ENTITIES[@]}"; do
  IFS=: read -r struct_name table_name <<< "$entity"
  lower_name=$(echo "$struct_name" | awk '{print tolower(substr($0,1,1)) substr($0,2)}')
  
  cat > "../internal/dao/${lower_name}.go" << EOF
package dao

import (
	"semi-mes/server/internal/model"
	"gorm.io/gorm"
)

type ${struct_name}Dao struct {
	db *gorm.DB
}

func New${struct_name}Dao(db *gorm.DB) *${struct_name}Dao {
	return &${struct_name}Dao{db: db}
}

func (d *${struct_name}Dao) Create(entity *model.${struct_name}) error {
	return d.db.Create(entity).Error
}

func (d *${struct_name}Dao) Update(entity *model.${struct_name}) error {
	return d.db.Save(entity).Error
}

func (d *${struct_name}Dao) Delete(id int64) error {
	return d.db.Delete(&model.${struct_name}{}, id).Error
}

func (d *${struct_name}Dao) GetByID(id int64) (*model.${struct_name}, error) {
	var entity model.${struct_name}
	err := d.db.Where("id = ?", id).First(&entity).Error
	return &entity, err
}

func (d *${struct_name}Dao) List(conditions map[string]interface{}, page, pageSize int) ([]model.${struct_name}, int64, error) {
	var entities []model.${struct_name}
	var total int64

	query := d.db.Model(&model.${struct_name}{})
	
	for key, value := range conditions {
		if value != nil && value != "" {
			query = query.Where(key, value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)
	
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
EOF

  echo "Generated DAO for $struct_name"
done

echo "All DAO files generated successfully"
EOF

chmod +x /workspace/server/scripts/generate-dao.sh
