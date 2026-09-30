package dao

import (
	"semi-mes/server/internal/model"

	"gorm.io/gorm"
)

type FactoryDao struct {
	db *gorm.DB
}

func NewFactoryDao(db *gorm.DB) *FactoryDao {
	return &FactoryDao{db: db}
}

func (d *FactoryDao) Create(factory *model.Factory) error {
	return d.db.Create(factory).Error
}

func (d *FactoryDao) Update(factory *model.Factory) error {
	return d.db.Save(factory).Error
}

func (d *FactoryDao) Delete(id int64) error {
	return d.db.Delete(&model.Factory{}, id).Error
}

func (d *FactoryDao) GetByID(id int64) (*model.Factory, error) {
	var factory model.Factory
	err := d.db.Where("id = ?", id).First(&factory).Error
	return &factory, err
}

func (d *FactoryDao) List(keyword string, page, pageSize int) ([]model.Factory, int64, error) {
	var factories []model.Factory
	var total int64

	query := d.db.Model(&model.Factory{})
	if keyword != "" {
		query = query.Where("factory_code LIKE ? OR factory_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&factories).Error; err != nil {
		return nil, 0, err
	}

	return factories, total, nil
}
