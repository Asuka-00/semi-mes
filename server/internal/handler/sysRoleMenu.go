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

var _ SysRoleMenuHandler = (*sysRoleMenuHandler)(nil)

// SysRoleMenuHandler defining the handler interface
type SysRoleMenuHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type sysRoleMenuHandler struct {
	iDao dao.SysRoleMenuDao
}

// NewSysRoleMenuHandler creating the handler interface
func NewSysRoleMenuHandler() SysRoleMenuHandler {
	return &sysRoleMenuHandler{
		iDao: dao.NewSysRoleMenuDao(
			database.GetDB(), // db driver is sqlite
			cache.NewSysRoleMenuCache(database.GetCacheType()),
		),
	}
}

// Create a new sysRoleMenu
// @Summary Create a new sysRoleMenu
// @Description Creates a new sysRoleMenu entity using the provided data in the request body.
// @Tags sysRoleMenu
// @Accept json
// @Produce json
// @Param data body types.CreateSysRoleMenuRequest true "sysRoleMenu information"
// @Success 200 {object} types.CreateSysRoleMenuReply{}
// @Router /api/v1/sysRoleMenu [post]
// @Security BearerAuth
func (h *sysRoleMenuHandler) Create(c *gin.Context) {
	form := &types.CreateSysRoleMenuRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	sysRoleMenu := &model.SysRoleMenu{}
	err = copier.Copy(sysRoleMenu, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateSysRoleMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, sysRoleMenu)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": sysRoleMenu.ID})
}

// DeleteByID delete a sysRoleMenu by id
// @Summary Delete a sysRoleMenu by id
// @Description Deletes a existing sysRoleMenu identified by the given id in the path.
// @Tags sysRoleMenu
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteSysRoleMenuByIDReply{}
// @Router /api/v1/sysRoleMenu/{id} [delete]
// @Security BearerAuth
func (h *sysRoleMenuHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getSysRoleMenuIDFromPath(c)
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

// UpdateByID update a sysRoleMenu by id
// @Summary Update a sysRoleMenu by id
// @Description Updates the specified sysRoleMenu by given id in the path, support partial update.
// @Tags sysRoleMenu
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateSysRoleMenuByIDRequest true "sysRoleMenu information"
// @Success 200 {object} types.UpdateSysRoleMenuByIDReply{}
// @Router /api/v1/sysRoleMenu/{id} [put]
// @Security BearerAuth
func (h *sysRoleMenuHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getSysRoleMenuIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateSysRoleMenuByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	sysRoleMenu := &model.SysRoleMenu{}
	err = copier.Copy(sysRoleMenu, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDSysRoleMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, sysRoleMenu)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a sysRoleMenu by id
// @Summary Get a sysRoleMenu by id
// @Description Gets detailed information of a sysRoleMenu specified by the given id in the path.
// @Tags sysRoleMenu
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetSysRoleMenuByIDReply{}
// @Router /api/v1/sysRoleMenu/{id} [get]
// @Security BearerAuth
func (h *sysRoleMenuHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getSysRoleMenuIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysRoleMenu, err := h.iDao.GetByID(ctx, id)
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

	data := &types.SysRoleMenuObjDetail{}
	err = copier.Copy(data, sysRoleMenu)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDSysRoleMenu)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"sysRoleMenu": data})
}

// List get a paginated list of sysRoleMenus by custom conditions
// @Summary Get a paginated list of sysRoleMenus by custom conditions
// @Description Returns a paginated list of sysRoleMenu based on query filters, including page number and size.
// @Tags sysRoleMenu
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListSysRoleMenusReply{}
// @Router /api/v1/sysRoleMenu/list [post]
// @Security BearerAuth
func (h *sysRoleMenuHandler) List(c *gin.Context) {
	form := &types.ListSysRoleMenusRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysRoleMenus, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertSysRoleMenus(sysRoleMenus)
	if err != nil {
		response.Error(c, ecode.ErrListSysRoleMenu)
		return
	}

	response.Success(c, gin.H{
		"sysRoleMenus": data,
		"total":        total,
	})
}

func getSysRoleMenuIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertSysRoleMenu(sysRoleMenu *model.SysRoleMenu) (*types.SysRoleMenuObjDetail, error) {
	data := &types.SysRoleMenuObjDetail{}
	err := copier.Copy(data, sysRoleMenu)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertSysRoleMenus(fromValues []*model.SysRoleMenu) ([]*types.SysRoleMenuObjDetail, error) {
	toValues := []*types.SysRoleMenuObjDetail{}
	for _, v := range fromValues {
		data, err := convertSysRoleMenu(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
