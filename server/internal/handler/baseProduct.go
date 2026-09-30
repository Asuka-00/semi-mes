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

var _ BaseProductHandler = (*baseProductHandler)(nil)

// BaseProductHandler defining the handler interface
type BaseProductHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseProductHandler struct {
	iDao dao.BaseProductDao
}

// NewBaseProductHandler creating the handler interface
func NewBaseProductHandler() BaseProductHandler {
	return &baseProductHandler{
		iDao: dao.NewBaseProductDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseProductCache(database.GetCacheType()),
		),
	}
}

// Create a new baseProduct
// @Summary Create a new baseProduct
// @Description Creates a new baseProduct entity using the provided data in the request body.
// @Tags baseProduct
// @Accept json
// @Produce json
// @Param data body types.CreateBaseProductRequest true "baseProduct information"
// @Success 200 {object} types.CreateBaseProductReply{}
// @Router /api/v1/baseProduct [post]
// @Security BearerAuth
func (h *baseProductHandler) Create(c *gin.Context) {
	form := &types.CreateBaseProductRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseProduct := &model.BaseProduct{}
	err = copier.Copy(baseProduct, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseProduct)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseProduct)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseProduct.ID})
}

// DeleteByID delete a baseProduct by id
// @Summary Delete a baseProduct by id
// @Description Deletes a existing baseProduct identified by the given id in the path.
// @Tags baseProduct
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseProductByIDReply{}
// @Router /api/v1/baseProduct/{id} [delete]
// @Security BearerAuth
func (h *baseProductHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseProductIDFromPath(c)
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

// UpdateByID update a baseProduct by id
// @Summary Update a baseProduct by id
// @Description Updates the specified baseProduct by given id in the path, support partial update.
// @Tags baseProduct
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseProductByIDRequest true "baseProduct information"
// @Success 200 {object} types.UpdateBaseProductByIDReply{}
// @Router /api/v1/baseProduct/{id} [put]
// @Security BearerAuth
func (h *baseProductHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseProductIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseProductByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseProduct := &model.BaseProduct{}
	err = copier.Copy(baseProduct, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseProduct)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseProduct)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseProduct by id
// @Summary Get a baseProduct by id
// @Description Gets detailed information of a baseProduct specified by the given id in the path.
// @Tags baseProduct
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseProductByIDReply{}
// @Router /api/v1/baseProduct/{id} [get]
// @Security BearerAuth
func (h *baseProductHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseProductIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProduct, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseProductObjDetail{}
	err = copier.Copy(data, baseProduct)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseProduct)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseProduct": data})
}

// List get a paginated list of baseProducts by custom conditions
// @Summary Get a paginated list of baseProducts by custom conditions
// @Description Returns a paginated list of baseProduct based on query filters, including page number and size.
// @Tags baseProduct
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseProductsReply{}
// @Router /api/v1/baseProduct/list [post]
// @Security BearerAuth
func (h *baseProductHandler) List(c *gin.Context) {
	form := &types.ListBaseProductsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProducts, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseProducts(baseProducts)
	if err != nil {
		response.Error(c, ecode.ErrListBaseProduct)
		return
	}

	response.Success(c, gin.H{
		"baseProducts": data,
		"total":        total,
	})
}

func getBaseProductIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseProduct(baseProduct *model.BaseProduct) (*types.BaseProductObjDetail, error) {
	data := &types.BaseProductObjDetail{}
	err := copier.Copy(data, baseProduct)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseProducts(fromValues []*model.BaseProduct) ([]*types.BaseProductObjDetail, error) {
	toValues := []*types.BaseProductObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseProduct(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
