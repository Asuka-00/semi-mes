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

var _ BaseProductionLineHandler = (*baseProductionLineHandler)(nil)

// BaseProductionLineHandler defining the handler interface
type BaseProductionLineHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseProductionLineHandler struct {
	iDao dao.BaseProductionLineDao
}

// NewBaseProductionLineHandler creating the handler interface
func NewBaseProductionLineHandler() BaseProductionLineHandler {
	return &baseProductionLineHandler{
		iDao: dao.NewBaseProductionLineDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseProductionLineCache(database.GetCacheType()),
		),
	}
}

// Create a new baseProductionLine
// @Summary Create a new baseProductionLine
// @Description Creates a new baseProductionLine entity using the provided data in the request body.
// @Tags baseProductionLine
// @Accept json
// @Produce json
// @Param data body types.CreateBaseProductionLineRequest true "baseProductionLine information"
// @Success 200 {object} types.CreateBaseProductionLineReply{}
// @Router /api/v1/baseProductionLine [post]
// @Security BearerAuth
func (h *baseProductionLineHandler) Create(c *gin.Context) {
	form := &types.CreateBaseProductionLineRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseProductionLine := &model.BaseProductionLine{}
	err = copier.Copy(baseProductionLine, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseProductionLine)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseProductionLine)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseProductionLine.ID})
}

// DeleteByID delete a baseProductionLine by id
// @Summary Delete a baseProductionLine by id
// @Description Deletes a existing baseProductionLine identified by the given id in the path.
// @Tags baseProductionLine
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseProductionLineByIDReply{}
// @Router /api/v1/baseProductionLine/{id} [delete]
// @Security BearerAuth
func (h *baseProductionLineHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseProductionLineIDFromPath(c)
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

// UpdateByID update a baseProductionLine by id
// @Summary Update a baseProductionLine by id
// @Description Updates the specified baseProductionLine by given id in the path, support partial update.
// @Tags baseProductionLine
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseProductionLineByIDRequest true "baseProductionLine information"
// @Success 200 {object} types.UpdateBaseProductionLineByIDReply{}
// @Router /api/v1/baseProductionLine/{id} [put]
// @Security BearerAuth
func (h *baseProductionLineHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseProductionLineIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseProductionLineByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseProductionLine := &model.BaseProductionLine{}
	err = copier.Copy(baseProductionLine, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseProductionLine)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseProductionLine)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseProductionLine by id
// @Summary Get a baseProductionLine by id
// @Description Gets detailed information of a baseProductionLine specified by the given id in the path.
// @Tags baseProductionLine
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseProductionLineByIDReply{}
// @Router /api/v1/baseProductionLine/{id} [get]
// @Security BearerAuth
func (h *baseProductionLineHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseProductionLineIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProductionLine, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseProductionLineObjDetail{}
	err = copier.Copy(data, baseProductionLine)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseProductionLine)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseProductionLine": data})
}

// List get a paginated list of baseProductionLines by custom conditions
// @Summary Get a paginated list of baseProductionLines by custom conditions
// @Description Returns a paginated list of baseProductionLine based on query filters, including page number and size.
// @Tags baseProductionLine
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseProductionLinesReply{}
// @Router /api/v1/baseProductionLine/list [post]
// @Security BearerAuth
func (h *baseProductionLineHandler) List(c *gin.Context) {
	form := &types.ListBaseProductionLinesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseProductionLines, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseProductionLines(baseProductionLines)
	if err != nil {
		response.Error(c, ecode.ErrListBaseProductionLine)
		return
	}

	response.Success(c, gin.H{
		"baseProductionLines": data,
		"total":        total,
	})
}

func getBaseProductionLineIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseProductionLine(baseProductionLine *model.BaseProductionLine) (*types.BaseProductionLineObjDetail, error) {
	data := &types.BaseProductionLineObjDetail{}
	err := copier.Copy(data, baseProductionLine)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseProductionLines(fromValues []*model.BaseProductionLine) ([]*types.BaseProductionLineObjDetail, error) {
	toValues := []*types.BaseProductionLineObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseProductionLine(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
