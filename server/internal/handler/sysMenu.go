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

var _ SysMenuHandler = (*sysMenuHandler)(nil)

// SysMenuHandler defining the handler interface
type SysMenuHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type sysMenuHandler struct {
	iDao dao.SysMenuDao
}

// NewSysMenuHandler creating the handler interface
func NewSysMenuHandler() SysMenuHandler {
	return &sysMenuHandler{
		iDao: dao.NewSysMenuDao(
			database.GetDB(), // db driver is sqlite
			cache.NewSysMenuCache(database.GetCacheType()),
		),
	}
}

// Create a new sysMenu
// @Summary Create a new sysMenu
// @Description Creates a new sysMenu entity using the provided data in the request body.
// @Tags sysMenu
// @Accept json
// @Produce json
// @Param data body types.CreateSysMenuRequest true "sysMenu information"
// @Success 200 {object} types.CreateSysMenuReply{}
// @Router /api/v1/sysMenu [post]
// @Security BearerAuth
func (h *sysMenuHandler) Create(c *gin.Context) {
	form := &types.CreateSysMenuRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	sysMenu := &model.SysMenu{}
	err = copier.Copy(sysMenu, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateSysMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, sysMenu)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": sysMenu.ID})
}

// DeleteByID delete a sysMenu by id
// @Summary Delete a sysMenu by id
// @Description Deletes a existing sysMenu identified by the given id in the path.
// @Tags sysMenu
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteSysMenuByIDReply{}
// @Router /api/v1/sysMenu/{id} [delete]
// @Security BearerAuth
func (h *sysMenuHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getSysMenuIDFromPath(c)
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

// UpdateByID update a sysMenu by id
// @Summary Update a sysMenu by id
// @Description Updates the specified sysMenu by given id in the path, support partial update.
// @Tags sysMenu
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateSysMenuByIDRequest true "sysMenu information"
// @Success 200 {object} types.UpdateSysMenuByIDReply{}
// @Router /api/v1/sysMenu/{id} [put]
// @Security BearerAuth
func (h *sysMenuHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getSysMenuIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateSysMenuByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	sysMenu := &model.SysMenu{}
	err = copier.Copy(sysMenu, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDSysMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, sysMenu)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a sysMenu by id
// @Summary Get a sysMenu by id
// @Description Gets detailed information of a sysMenu specified by the given id in the path.
// @Tags sysMenu
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetSysMenuByIDReply{}
// @Router /api/v1/sysMenu/{id} [get]
// @Security BearerAuth
func (h *sysMenuHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getSysMenuIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysMenu, err := h.iDao.GetByID(ctx, id)
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

	data := &types.SysMenuObjDetail{}
	err = copier.Copy(data, sysMenu)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDSysMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"sysMenu": data})
}

// List get a paginated list of sysMenus by custom conditions
// @Summary Get a paginated list of sysMenus by custom conditions
// @Description Returns a paginated list of sysMenu based on query filters, including page number and size.
// @Tags sysMenu
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListSysMenusReply{}
// @Router /api/v1/sysMenu/list [post]
// @Security BearerAuth
func (h *sysMenuHandler) List(c *gin.Context) {
	form := &types.ListSysMenusRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysMenus, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertSysMenus(sysMenus)
	if err != nil {
		response.Error(c, ecode.ErrListSysMenu)
		return
	}

	response.Success(c, gin.H{
		"sysMenus": data,
		"total":        total,
	})
}

func getSysMenuIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertSysMenu(sysMenu *model.SysMenu) (*types.SysMenuObjDetail, error) {
	data := &types.SysMenuObjDetail{}
	err := copier.Copy(data, sysMenu)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertSysMenus(fromValues []*model.SysMenu) ([]*types.SysMenuObjDetail, error) {
	toValues := []*types.SysMenuObjDetail{}
	for _, v := range fromValues {
		data, err := convertSysMenu(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
