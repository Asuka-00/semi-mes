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

var _ BaseWorkshopHandler = (*baseWorkshopHandler)(nil)

// BaseWorkshopHandler defining the handler interface
type BaseWorkshopHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseWorkshopHandler struct {
	iDao dao.BaseWorkshopDao
}

// NewBaseWorkshopHandler creating the handler interface
func NewBaseWorkshopHandler() BaseWorkshopHandler {
	return &baseWorkshopHandler{
		iDao: dao.NewBaseWorkshopDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseWorkshopCache(database.GetCacheType()),
		),
	}
}

// Create a new baseWorkshop
// @Summary Create a new baseWorkshop
// @Description Creates a new baseWorkshop entity using the provided data in the request body.
// @Tags baseWorkshop
// @Accept json
// @Produce json
// @Param data body types.CreateBaseWorkshopRequest true "baseWorkshop information"
// @Success 200 {object} types.CreateBaseWorkshopReply{}
// @Router /api/v1/baseWorkshop [post]
// @Security BearerAuth
func (h *baseWorkshopHandler) Create(c *gin.Context) {
	form := &types.CreateBaseWorkshopRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseWorkshop := &model.BaseWorkshop{}
	err = copier.Copy(baseWorkshop, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseWorkshop)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseWorkshop)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseWorkshop.ID})
}

// DeleteByID delete a baseWorkshop by id
// @Summary Delete a baseWorkshop by id
// @Description Deletes a existing baseWorkshop identified by the given id in the path.
// @Tags baseWorkshop
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseWorkshopByIDReply{}
// @Router /api/v1/baseWorkshop/{id} [delete]
// @Security BearerAuth
func (h *baseWorkshopHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseWorkshopIDFromPath(c)
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

// UpdateByID update a baseWorkshop by id
// @Summary Update a baseWorkshop by id
// @Description Updates the specified baseWorkshop by given id in the path, support partial update.
// @Tags baseWorkshop
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseWorkshopByIDRequest true "baseWorkshop information"
// @Success 200 {object} types.UpdateBaseWorkshopByIDReply{}
// @Router /api/v1/baseWorkshop/{id} [put]
// @Security BearerAuth
func (h *baseWorkshopHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseWorkshopIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseWorkshopByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseWorkshop := &model.BaseWorkshop{}
	err = copier.Copy(baseWorkshop, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseWorkshop)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseWorkshop)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseWorkshop by id
// @Summary Get a baseWorkshop by id
// @Description Gets detailed information of a baseWorkshop specified by the given id in the path.
// @Tags baseWorkshop
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseWorkshopByIDReply{}
// @Router /api/v1/baseWorkshop/{id} [get]
// @Security BearerAuth
func (h *baseWorkshopHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseWorkshopIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseWorkshop, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseWorkshopObjDetail{}
	err = copier.Copy(data, baseWorkshop)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseWorkshop)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseWorkshop": data})
}

// List get a paginated list of baseWorkshops by custom conditions
// @Summary Get a paginated list of baseWorkshops by custom conditions
// @Description Returns a paginated list of baseWorkshop based on query filters, including page number and size.
// @Tags baseWorkshop
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseWorkshopsReply{}
// @Router /api/v1/baseWorkshop/list [post]
// @Security BearerAuth
func (h *baseWorkshopHandler) List(c *gin.Context) {
	form := &types.ListBaseWorkshopsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseWorkshops, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseWorkshops(baseWorkshops)
	if err != nil {
		response.Error(c, ecode.ErrListBaseWorkshop)
		return
	}

	response.Success(c, gin.H{
		"baseWorkshops": data,
		"total":         total,
	})
}

func getBaseWorkshopIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseWorkshop(baseWorkshop *model.BaseWorkshop) (*types.BaseWorkshopObjDetail, error) {
	data := &types.BaseWorkshopObjDetail{}
	err := copier.Copy(data, baseWorkshop)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseWorkshops(fromValues []*model.BaseWorkshop) ([]*types.BaseWorkshopObjDetail, error) {
	toValues := []*types.BaseWorkshopObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseWorkshop(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
