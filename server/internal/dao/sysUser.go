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

var _ SysUserDao = (*sysUserDao)(nil)

// SysUserDao defining the dao interface
type SysUserDao interface {
	Create(ctx context.Context, table *model.SysUser) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.SysUser) error
	GetByID(ctx context.Context, id uint64) (*model.SysUser, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.SysUser, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.SysUser) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.SysUser) error
}

type sysUserDao struct {
	db    *gorm.DB
	cache cache.SysUserCache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// NewSysUserDao creating the dao interface
func NewSysUserDao(db *gorm.DB, xCache cache.SysUserCache) SysUserDao {
	if xCache == nil {
		return &sysUserDao{db: db}
	}
	return &sysUserDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *sysUserDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new sysUser, insert the record and the id value is written back to the table
func (d *sysUserDao) Create(ctx context.Context, table *model.SysUser) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a sysUser by id
func (d *sysUserDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SysUser{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a sysUser by id, support partial update
func (d *sysUserDao) UpdateByID(ctx context.Context, table *model.SysUser) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *sysUserDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.SysUser) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}
	
	if table.Username != "" {
		update["username"] = table.Username
	}
	if table.Password != "" {
		update["password"] = table.Password
	}
	if table.RealName != "" {
		update["real_name"] = table.RealName
	}
	if table.Email != "" {
		update["email"] = table.Email
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.Status != 0 {
		update["status"] = table.Status
	}
	

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a sysUser by id
func (d *sysUserDao) GetByID(ctx context.Context, id uint64) (*model.SysUser, error) {
	// no cache
	if d.cache == nil {
		record := &model.SysUser{}
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
			table := &model.SysUser{}
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
			if err = d.cache.Set(ctx, id, table, cache.SysUserExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.SysUser)
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

// GetByColumns get a paginated list of sysUsers by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *sysUserDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.SysUser, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.SysUserColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.SysUser{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.SysUser{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *sysUserDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.SysUser) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *sysUserDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.SysUser{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *sysUserDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.SysUser) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
