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

var _ BaseRecipeHandler = (*baseRecipeHandler)(nil)

// BaseRecipeHandler defining the handler interface
type BaseRecipeHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type baseRecipeHandler struct {
	iDao dao.BaseRecipeDao
}

// NewBaseRecipeHandler creating the handler interface
func NewBaseRecipeHandler() BaseRecipeHandler {
	return &baseRecipeHandler{
		iDao: dao.NewBaseRecipeDao(
			database.GetDB(), // db driver is sqlite
			cache.NewBaseRecipeCache(database.GetCacheType()),
		),
	}
}

// Create a new baseRecipe
// @Summary Create a new baseRecipe
// @Description Creates a new baseRecipe entity using the provided data in the request body.
// @Tags baseRecipe
// @Accept json
// @Produce json
// @Param data body types.CreateBaseRecipeRequest true "baseRecipe information"
// @Success 200 {object} types.CreateBaseRecipeReply{}
// @Router /api/v1/baseRecipe [post]
// @Security BearerAuth
func (h *baseRecipeHandler) Create(c *gin.Context) {
	form := &types.CreateBaseRecipeRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	baseRecipe := &model.BaseRecipe{}
	err = copier.Copy(baseRecipe, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateBaseRecipe)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, baseRecipe)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": baseRecipe.ID})
}

// DeleteByID delete a baseRecipe by id
// @Summary Delete a baseRecipe by id
// @Description Deletes a existing baseRecipe identified by the given id in the path.
// @Tags baseRecipe
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteBaseRecipeByIDReply{}
// @Router /api/v1/baseRecipe/{id} [delete]
// @Security BearerAuth
func (h *baseRecipeHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getBaseRecipeIDFromPath(c)
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

// UpdateByID update a baseRecipe by id
// @Summary Update a baseRecipe by id
// @Description Updates the specified baseRecipe by given id in the path, support partial update.
// @Tags baseRecipe
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateBaseRecipeByIDRequest true "baseRecipe information"
// @Success 200 {object} types.UpdateBaseRecipeByIDReply{}
// @Router /api/v1/baseRecipe/{id} [put]
// @Security BearerAuth
func (h *baseRecipeHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getBaseRecipeIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateBaseRecipeByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	baseRecipe := &model.BaseRecipe{}
	err = copier.Copy(baseRecipe, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDBaseRecipe)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, baseRecipe)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a baseRecipe by id
// @Summary Get a baseRecipe by id
// @Description Gets detailed information of a baseRecipe specified by the given id in the path.
// @Tags baseRecipe
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBaseRecipeByIDReply{}
// @Router /api/v1/baseRecipe/{id} [get]
// @Security BearerAuth
func (h *baseRecipeHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getBaseRecipeIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseRecipe, err := h.iDao.GetByID(ctx, id)
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

	data := &types.BaseRecipeObjDetail{}
	err = copier.Copy(data, baseRecipe)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDBaseRecipe)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"baseRecipe": data})
}

// List get a paginated list of baseRecipes by custom conditions
// @Summary Get a paginated list of baseRecipes by custom conditions
// @Description Returns a paginated list of baseRecipe based on query filters, including page number and size.
// @Tags baseRecipe
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListBaseRecipesReply{}
// @Router /api/v1/baseRecipe/list [post]
// @Security BearerAuth
func (h *baseRecipeHandler) List(c *gin.Context) {
	form := &types.ListBaseRecipesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	baseRecipes, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertBaseRecipes(baseRecipes)
	if err != nil {
		response.Error(c, ecode.ErrListBaseRecipe)
		return
	}

	response.Success(c, gin.H{
		"baseRecipes": data,
		"total":        total,
	})
}

func getBaseRecipeIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertBaseRecipe(baseRecipe *model.BaseRecipe) (*types.BaseRecipeObjDetail, error) {
	data := &types.BaseRecipeObjDetail{}
	err := copier.Copy(data, baseRecipe)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertBaseRecipes(fromValues []*model.BaseRecipe) ([]*types.BaseRecipeObjDetail, error) {
	toValues := []*types.BaseRecipeObjDetail{}
	for _, v := range fromValues {
		data, err := convertBaseRecipe(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
