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

var _ SysRoleHandler = (*sysRoleHandler)(nil)

// SysRoleHandler defining the handler interface
type SysRoleHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type sysRoleHandler struct {
	iDao dao.SysRoleDao
}

// NewSysRoleHandler creating the handler interface
func NewSysRoleHandler() SysRoleHandler {
	return &sysRoleHandler{
		iDao: dao.NewSysRoleDao(
			database.GetDB(), // db driver is sqlite
			cache.NewSysRoleCache(database.GetCacheType()),
		),
	}
}

// Create a new sysRole
// @Summary Create a new sysRole
// @Description Creates a new sysRole entity using the provided data in the request body.
// @Tags sysRole
// @Accept json
// @Produce json
// @Param data body types.CreateSysRoleRequest true "sysRole information"
// @Success 200 {object} types.CreateSysRoleReply{}
// @Router /api/v1/sysRole [post]
// @Security BearerAuth
func (h *sysRoleHandler) Create(c *gin.Context) {
	form := &types.CreateSysRoleRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	sysRole := &model.SysRole{}
	err = copier.Copy(sysRole, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateSysRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, sysRole)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": sysRole.ID})
}

// DeleteByID delete a sysRole by id
// @Summary Delete a sysRole by id
// @Description Deletes a existing sysRole identified by the given id in the path.
// @Tags sysRole
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteSysRoleByIDReply{}
// @Router /api/v1/sysRole/{id} [delete]
// @Security BearerAuth
func (h *sysRoleHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getSysRoleIDFromPath(c)
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

// UpdateByID update a sysRole by id
// @Summary Update a sysRole by id
// @Description Updates the specified sysRole by given id in the path, support partial update.
// @Tags sysRole
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateSysRoleByIDRequest true "sysRole information"
// @Success 200 {object} types.UpdateSysRoleByIDReply{}
// @Router /api/v1/sysRole/{id} [put]
// @Security BearerAuth
func (h *sysRoleHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getSysRoleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateSysRoleByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	sysRole := &model.SysRole{}
	err = copier.Copy(sysRole, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDSysRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, sysRole)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a sysRole by id
// @Summary Get a sysRole by id
// @Description Gets detailed information of a sysRole specified by the given id in the path.
// @Tags sysRole
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetSysRoleByIDReply{}
// @Router /api/v1/sysRole/{id} [get]
// @Security BearerAuth
func (h *sysRoleHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getSysRoleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysRole, err := h.iDao.GetByID(ctx, id)
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

	data := &types.SysRoleObjDetail{}
	err = copier.Copy(data, sysRole)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDSysRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"sysRole": data})
}

// List get a paginated list of sysRoles by custom conditions
// @Summary Get a paginated list of sysRoles by custom conditions
// @Description Returns a paginated list of sysRole based on query filters, including page number and size.
// @Tags sysRole
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListSysRolesReply{}
// @Router /api/v1/sysRole/list [post]
// @Security BearerAuth
func (h *sysRoleHandler) List(c *gin.Context) {
	form := &types.ListSysRolesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysRoles, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertSysRoles(sysRoles)
	if err != nil {
		response.Error(c, ecode.ErrListSysRole)
		return
	}

	response.Success(c, gin.H{
		"sysRoles": data,
		"total":        total,
	})
}

func getSysRoleIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertSysRole(sysRole *model.SysRole) (*types.SysRoleObjDetail, error) {
	data := &types.SysRoleObjDetail{}
	err := copier.Copy(data, sysRole)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertSysRoles(fromValues []*model.SysRole) ([]*types.SysRoleObjDetail, error) {
	toValues := []*types.SysRoleObjDetail{}
	for _, v := range fromValues {
		data, err := convertSysRole(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
