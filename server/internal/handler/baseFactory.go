package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/cache"
	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
	"semi-mes/server/internal/types"
)

var _ BaseFactoryHandler = (*baseFactoryHandler)(nil)

// BaseFactoryHandler defining the handler interface
type BaseFactoryHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseFactoryHandler struct {
	iDao dao.BaseFactoryDao
}

// NewBaseFactoryHandler creating the handler interface
func NewBaseFactoryHandler() BaseFactoryHandler {
	return &baseFactoryHandler{
		iDao: dao.NewBaseFactoryDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseFactoryCache(database.GetCacheType()),
		),
	}
}

// Create a new baseFactory
// @Summary Create a new baseFactory
// @Description Creates a new baseFactory entity using the provided data in the request body.
// @Tags baseFactory
// @Accept json
// @Produce json
// @Param data body types.CreateBaseFactoryRequest true "baseFactory information"
// @Success 200 {object} types.CreateBaseFactoryReply{}
// @Router /api/v1/baseFactory [post]
// @Security BearerAuth
func (h *baseFactoryHandler) Create(c *gin.Context) {
	form := &types.CreateBaseFactoryRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseFactory := &model.BaseFactory{}
	err = copier.Copy(baseFactory, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseFactory)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseFactory)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseFactory.ID})
}

// DeleteByID delete a baseFactory by id
// @Summary Delete a baseFactory by id
// @Description Deletes a existing baseFactory identified by the given id in the path.
// @Tags baseFactory
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseFactoryByIDReply{}
// @Router /api/v1/baseFactory/{id} [delete]
// @Security BearerAuth
func (h *baseFactoryHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseFactoryIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.DeleteByID(ctx, id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// UpdateByID update a baseFactory by id
// @Summary Update a baseFactory by id
// @Description Updates the specified baseFactory by given id in the path, support partial update.
// @Tags baseFactory
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseFactoryByIDRequest true "baseFactory information"
// @Success 200 {object} types.UpdateBaseFactoryByIDReply{}
// @Router /api/v1/baseFactory/{id} [put]
// @Security BearerAuth
func (h *baseFactoryHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseFactoryIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseFactoryByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseFactory := &model.BaseFactory{}
	err = copier.Copy(baseFactory, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseFactory)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseFactory)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseFactory by id
// @Summary Get a baseFactory by id
// @Description Gets detailed information of a baseFactory specified by the given id in the path.
// @Tags baseFactory
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseFactoryByIDReply{}
// @Router /api/v1/baseFactory/{id} [get]
// @Security BearerAuth
func (h *baseFactoryHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseFactoryIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseFactory, err := h.iDao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByID not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Error(c, ecode.NotFound)
		} else {
			logger.Error("GetByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Output(c, ecode.InternalServerError.ToHTTPCode())
		}
		return
	}

	data := &types.BaseFactoryObjDetail{}
	err = copier.Copy(data, baseFactory)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseFactory)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseFactory": data})
}

// List get a paginated list of baseFactorys by custom conditions
// @Summary Get a paginated list of baseFactorys by custom conditions
// @Description Returns a paginated list of baseFactory based on query filters, including page number and size.
// @Tags baseFactory
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseFactorysReply{}
// @Router /api/v1/baseFactory/list [post]
// @Security BearerAuth
func (h *baseFactoryHandler) List(c *gin.Context) {
	form := &types.ListBaseFactorysRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseFactorys, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseFactorys(baseFactorys)
	if err != nil {
		response.Error(c, ecode.ErrListBaseFactory)
		return
	}

	response.Success(c, gin.H{
		"baseFactorys": data,
		"total":        total,
	})
}

func getBaseFactoryIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseFactory(baseFactory *model.BaseFactory) (*types.BaseFactoryObjDetail, error) {
	data := &types.BaseFactoryObjDetail{}
	err := copier.Copy(data, baseFactory)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseFactorys(fromValues []*model.BaseFactory) ([]*types.BaseFactoryObjDetail, error) {
	toValues := []*types.BaseFactoryObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseFactory(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
