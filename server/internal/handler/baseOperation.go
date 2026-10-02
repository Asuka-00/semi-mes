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

var _ BaseOperationHandler = (*baseOperationHandler)(nil)

// BaseOperationHandler defining the handler interface
type BaseOperationHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseOperationHandler struct {
	iDao dao.BaseOperationDao
}

// NewBaseOperationHandler creating the handler interface
func NewBaseOperationHandler() BaseOperationHandler {
	return &baseOperationHandler{
		iDao: dao.NewBaseOperationDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseOperationCache(database.GetCacheType()),
		),
	}
}

// Create a new baseOperation
// @Summary Create a new baseOperation
// @Description Creates a new baseOperation entity using the provided data in the request body.
// @Tags baseOperation
// @Accept json
// @Produce json
// @Param data body types.CreateBaseOperationRequest true "baseOperation information"
// @Success 200 {object} types.CreateBaseOperationReply{}
// @Router /api/v1/baseOperation [post]
// @Security BearerAuth
func (h *baseOperationHandler) Create(c *gin.Context) {
	form := &types.CreateBaseOperationRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseOperation := &model.BaseOperation{}
	err = copier.Copy(baseOperation, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseOperation)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseOperation)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseOperation.ID})
}

// DeleteByID delete a baseOperation by id
// @Summary Delete a baseOperation by id
// @Description Deletes a existing baseOperation identified by the given id in the path.
// @Tags baseOperation
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseOperationByIDReply{}
// @Router /api/v1/baseOperation/{id} [delete]
// @Security BearerAuth
func (h *baseOperationHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseOperationIDFromPath(c)
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

// UpdateByID update a baseOperation by id
// @Summary Update a baseOperation by id
// @Description Updates the specified baseOperation by given id in the path, support partial update.
// @Tags baseOperation
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseOperationByIDRequest true "baseOperation information"
// @Success 200 {object} types.UpdateBaseOperationByIDReply{}
// @Router /api/v1/baseOperation/{id} [put]
// @Security BearerAuth
func (h *baseOperationHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseOperationIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseOperationByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseOperation := &model.BaseOperation{}
	err = copier.Copy(baseOperation, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseOperation)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseOperation)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseOperation by id
// @Summary Get a baseOperation by id
// @Description Gets detailed information of a baseOperation specified by the given id in the path.
// @Tags baseOperation
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseOperationByIDReply{}
// @Router /api/v1/baseOperation/{id} [get]
// @Security BearerAuth
func (h *baseOperationHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseOperationIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseOperation, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseOperationObjDetail{}
	err = copier.Copy(data, baseOperation)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseOperation)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseOperation": data})
}

// List get a paginated list of baseOperations by custom conditions
// @Summary Get a paginated list of baseOperations by custom conditions
// @Description Returns a paginated list of baseOperation based on query filters, including page number and size.
// @Tags baseOperation
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseOperationsReply{}
// @Router /api/v1/baseOperation/list [post]
// @Security BearerAuth
func (h *baseOperationHandler) List(c *gin.Context) {
	form := &types.ListBaseOperationsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseOperations, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseOperations(baseOperations)
	if err != nil {
		response.Error(c, ecode.ErrListBaseOperation)
		return
	}

	response.Success(c, gin.H{
		"baseOperations": data,
		"total":        total,
	})
}

func getBaseOperationIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseOperation(baseOperation *model.BaseOperation) (*types.BaseOperationObjDetail, error) {
	data := &types.BaseOperationObjDetail{}
	err := copier.Copy(data, baseOperation)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseOperations(fromValues []*model.BaseOperation) ([]*types.BaseOperationObjDetail, error) {
	toValues := []*types.BaseOperationObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseOperation(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
