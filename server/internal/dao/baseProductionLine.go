package dao

import (
	"context"
	"errors"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/cache"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

var _ BaseProductionLineDao = (*baseProductionLineDao)(nil)

// BaseProductionLineDao defining the dao interface
type BaseProductionLineDao interface {
	Create(ctx context.Context, table *model.BaseProductionLine) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.BaseProductionLine) error
	GetByID(ctx context.Context, id uint64) (*model.BaseProductionLine, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.BaseProductionLine, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseProductionLine) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseProductionLine) error
}

type baseProductionLineDao struct {
	db    *gorm.DB
	cache cache.BaseProductionLineCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewBaseProductionLineDao creating the dao interface
func NewBaseProductionLineDao(db *gorm.DB, xCache cache.BaseProductionLineCache) BaseProductionLineDao {
	if xCache == nil {
		return &baseProductionLineDao{db: db}
	}
	return &baseProductionLineDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *baseProductionLineDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new baseProductionLine, insert the record and the id value is written back to the table
func (d *baseProductionLineDao) Create(ctx context.Context, table *model.BaseProductionLine) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a baseProductionLine by id
func (d *baseProductionLineDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BaseProductionLine{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a baseProductionLine by id, support partial update
func (d *baseProductionLineDao) UpdateByID(ctx context.Context, table *model.BaseProductionLine) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *baseProductionLineDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.BaseProductionLine) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}
	
	if table.WorkshopID != 0 {
		update["workshop_id"] = table.WorkshopID
	}
	if table.LineCode != "" {
		update["line_code"] = table.LineCode
	}
	if table.LineName != "" {
		update["line_name"] = table.LineName
	}
	update["capacity"] = table.Capacity
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a baseProductionLine by id
func (d *baseProductionLineDao) GetByID(ctx context.Context, id uint64) (*model.BaseProductionLine, error) {
	// no cache
	if d.cache == nil {
		record := &model.BaseProductionLine{}
		err := d.db.WithContext(ctx).Where("id = ?", id).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) { //nolint
			table := &model.BaseProductionLine{}
			err = d.db.WithContext(ctx).Where("id = ?", id).First(table).Error
			if err != nil {
				if errors.Is(err, database.ErrRecordNotFound) {
					// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, id, table, cache.BaseProductionLineExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.BaseProductionLine)
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}

// GetByColumns get a paginated list of baseProductionLines by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *baseProductionLineDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.BaseProductionLine, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.BaseProductionLineColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.BaseProductionLine{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.BaseProductionLine{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *baseProductionLineDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseProductionLine) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *baseProductionLineDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.BaseProductionLine{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *baseProductionLineDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseProductionLine) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
