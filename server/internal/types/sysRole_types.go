package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateSysRoleRequest request params
type CreateSysRoleRequest struct {
	RoleCode  string `json:"roleCode" binding:""`
	RoleName  string `json:"roleName" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateSysRoleByIDRequest request params
type UpdateSysRoleByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RoleCode  string `json:"roleCode" binding:""`
	RoleName  string `json:"roleName" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// SysRoleObjDetail detail
type SysRoleObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RoleCode  string `json:"roleCode"`
	RoleName  string `json:"roleName"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateSysRoleReply only for api docs
type CreateSysRoleReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteSysRoleByIDReply only for api docs
type DeleteSysRoleByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateSysRoleByIDReply only for api docs
type UpdateSysRoleByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetSysRoleByIDReply only for api docs
type GetSysRoleByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysRole SysRoleObjDetail `json:"sysRole"`
	} `json:"data"` // return data
}

// ListSysRolesRequest request params
type ListSysRolesRequest struct {
	query.Params
}

// ListSysRolesReply only for api docs
type ListSysRolesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysRoles []SysRoleObjDetail `json:"sysRoles"`
	} `json:"data"` // return data
}
