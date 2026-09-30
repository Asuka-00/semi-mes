package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateSysUserRoleRequest request params
type CreateSysUserRoleRequest struct {
	UserID  int `json:"userID" binding:""`
	RoleID  int `json:"roleID" binding:""`
}

// UpdateSysUserRoleByIDRequest request params
type UpdateSysUserRoleByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  int `json:"userID" binding:""`
	RoleID  int `json:"roleID" binding:""`
}

// SysUserRoleObjDetail detail
type SysUserRoleObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  int `json:"userID"`
	RoleID  int `json:"roleID"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateSysUserRoleReply only for api docs
type CreateSysUserRoleReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteSysUserRoleByIDReply only for api docs
type DeleteSysUserRoleByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateSysUserRoleByIDReply only for api docs
type UpdateSysUserRoleByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetSysUserRoleByIDReply only for api docs
type GetSysUserRoleByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysUserRole SysUserRoleObjDetail `json:"sysUserRole"`
	} `json:"data"` // return data
}

// ListSysUserRolesRequest request params
type ListSysUserRolesRequest struct {
	query.Params
}

// ListSysUserRolesReply only for api docs
type ListSysUserRolesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysUserRoles []SysUserRoleObjDetail `json:"sysUserRoles"`
	} `json:"data"` // return data
}
