package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateSysRoleMenuRequest request params
type CreateSysRoleMenuRequest struct {
	RoleID  int `json:"roleID" binding:""`
	MenuID  int `json:"menuID" binding:""`
}

// UpdateSysRoleMenuByIDRequest request params
type UpdateSysRoleMenuByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RoleID  int `json:"roleID" binding:""`
	MenuID  int `json:"menuID" binding:""`
}

// SysRoleMenuObjDetail detail
type SysRoleMenuObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RoleID  int `json:"roleID"`
	MenuID  int `json:"menuID"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateSysRoleMenuReply only for api docs
type CreateSysRoleMenuReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteSysRoleMenuByIDReply only for api docs
type DeleteSysRoleMenuByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateSysRoleMenuByIDReply only for api docs
type UpdateSysRoleMenuByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetSysRoleMenuByIDReply only for api docs
type GetSysRoleMenuByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysRoleMenu SysRoleMenuObjDetail `json:"sysRoleMenu"`
	} `json:"data"` // return data
}

// ListSysRoleMenusRequest request params
type ListSysRoleMenusRequest struct {
	query.Params
}

// ListSysRoleMenusReply only for api docs
type ListSysRoleMenusReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysRoleMenus []SysRoleMenuObjDetail `json:"sysRoleMenus"`
	} `json:"data"` // return data
}
