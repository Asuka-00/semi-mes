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

var _ BaseFactoryDao = (*baseFactoryDao)(nil)

// BaseFactoryDao defining the dao interface
type BaseFactoryDao interface {
	Create(ctx context.Context, table *model.BaseFactory) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.BaseFactory) error
	GetByID(ctx context.Context, id uint64) (*model.BaseFactory, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.BaseFactory, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseFactory) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseFactory) error
}

type baseFactoryDao struct {
	db    *gorm.DB
	cache cache.BaseFactoryCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewBaseFactoryDao creating the dao interface
func NewBaseFactoryDao(db *gorm.DB, xCache cache.BaseFactoryCache) BaseFactoryDao {
	if xCache == nil {
		return &baseFactoryDao{db: db}
	}
	return &baseFactoryDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *baseFactoryDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new baseFactory, insert the record and the id value is written back to the table
func (d *baseFactoryDao) Create(ctx context.Context, table *model.BaseFactory) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a baseFactory by id
func (d *baseFactoryDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BaseFactory{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a baseFactory by id, support partial update
func (d *baseFactoryDao) UpdateByID(ctx context.Context, table *model.BaseFactory) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *baseFactoryDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.BaseFactory) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.FactoryCode != "" {
		update["factory_code"] = table.FactoryCode
	}
	if table.FactoryName != "" {
		update["factory_name"] = table.FactoryName
	}
	if table.Address != "" {
		update["address"] = table.Address
	}
	if table.Contact != "" {
		update["contact"] = table.Contact
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a baseFactory by id
func (d *baseFactoryDao) GetByID(ctx context.Context, id uint64) (*model.BaseFactory, error) {
	// no cache
	if d.cache == nil {
		record := &model.BaseFactory{}
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
			table := &model.BaseFactory{}
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
			if err = d.cache.Set(ctx, id, table, cache.BaseFactoryExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.BaseFactory)
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

// GetByColumns get a paginated list of baseFactorys by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *baseFactoryDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.BaseFactory, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.BaseFactoryColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.BaseFactory{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.BaseFactory{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *baseFactoryDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseFactory) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *baseFactoryDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.BaseFactory{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *baseFactoryDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.BaseFactory) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
