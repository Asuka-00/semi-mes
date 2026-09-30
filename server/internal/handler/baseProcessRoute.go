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

var _ BaseProcessRouteHandler = (*baseProcessRouteHandler)(nil)

// BaseProcessRouteHandler defining the handler interface
type BaseProcessRouteHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseProcessRouteHandler struct {
	iDao dao.BaseProcessRouteDao
}

// NewBaseProcessRouteHandler creating the handler interface
func NewBaseProcessRouteHandler() BaseProcessRouteHandler {
	return &baseProcessRouteHandler{
		iDao: dao.NewBaseProcessRouteDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseProcessRouteCache(database.GetCacheType()),
		),
	}
}

// Create a new baseProcessRoute
// @Summary Create a new baseProcessRoute
// @Description Creates a new baseProcessRoute entity using the provided data in the request body.
// @Tags baseProcessRoute
// @Accept json
// @Produce json
// @Param data body types.CreateBaseProcessRouteRequest true "baseProcessRoute information"
// @Success 200 {object} types.CreateBaseProcessRouteReply{}
// @Router /api/v1/baseProcessRoute [post]
// @Security BearerAuth
func (h *baseProcessRouteHandler) Create(c *gin.Context) {
	form := &types.CreateBaseProcessRouteRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseProcessRoute := &model.BaseProcessRoute{}
	err = copier.Copy(baseProcessRoute, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseProcessRoute)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseProcessRoute)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseProcessRoute.ID})
}

// DeleteByID delete a baseProcessRoute by id
// @Summary Delete a baseProcessRoute by id
// @Description Deletes a existing baseProcessRoute identified by the given id in the path.
// @Tags baseProcessRoute
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseProcessRouteByIDReply{}
// @Router /api/v1/baseProcessRoute/{id} [delete]
// @Security BearerAuth
func (h *baseProcessRouteHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseProcessRouteIDFromPath(c)
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

// UpdateByID update a baseProcessRoute by id
// @Summary Update a baseProcessRoute by id
// @Description Updates the specified baseProcessRoute by given id in the path, support partial update.
// @Tags baseProcessRoute
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseProcessRouteByIDRequest true "baseProcessRoute information"
// @Success 200 {object} types.UpdateBaseProcessRouteByIDReply{}
// @Router /api/v1/baseProcessRoute/{id} [put]
// @Security BearerAuth
func (h *baseProcessRouteHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseProcessRouteIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseProcessRouteByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseProcessRoute := &model.BaseProcessRoute{}
	err = copier.Copy(baseProcessRoute, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseProcessRoute)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseProcessRoute)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseProcessRoute by id
// @Summary Get a baseProcessRoute by id
// @Description Gets detailed information of a baseProcessRoute specified by the given id in the path.
// @Tags baseProcessRoute
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseProcessRouteByIDReply{}
// @Router /api/v1/baseProcessRoute/{id} [get]
// @Security BearerAuth
func (h *baseProcessRouteHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseProcessRouteIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProcessRoute, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseProcessRouteObjDetail{}
	err = copier.Copy(data, baseProcessRoute)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseProcessRoute)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseProcessRoute": data})
}

// List get a paginated list of baseProcessRoutes by custom conditions
// @Summary Get a paginated list of baseProcessRoutes by custom conditions
// @Description Returns a paginated list of baseProcessRoute based on query filters, including page number and size.
// @Tags baseProcessRoute
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseProcessRoutesReply{}
// @Router /api/v1/baseProcessRoute/list [post]
// @Security BearerAuth
func (h *baseProcessRouteHandler) List(c *gin.Context) {
	form := &types.ListBaseProcessRoutesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProcessRoutes, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseProcessRoutes(baseProcessRoutes)
	if err != nil {
		response.Error(c, ecode.ErrListBaseProcessRoute)
		return
	}

	response.Success(c, gin.H{
		"baseProcessRoutes": data,
		"total":        total,
	})
}

func getBaseProcessRouteIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseProcessRoute(baseProcessRoute *model.BaseProcessRoute) (*types.BaseProcessRouteObjDetail, error) {
	data := &types.BaseProcessRouteObjDetail{}
	err := copier.Copy(data, baseProcessRoute)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseProcessRoutes(fromValues []*model.BaseProcessRoute) ([]*types.BaseProcessRouteObjDetail, error) {
	toValues := []*types.BaseProcessRouteObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseProcessRoute(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
