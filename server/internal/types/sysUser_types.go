package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateSysUserRequest request params
type CreateSysUserRequest struct {
	Username  string `json:"username" binding:""`
	Password  string `json:"password" binding:""`
	RealName  string `json:"realName" binding:""`
	Email  string `json:"email" binding:""`
	Phone  string `json:"phone" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateSysUserByIDRequest request params
type UpdateSysUserByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	Username  string `json:"username" binding:""`
	Password  string `json:"password" binding:""`
	RealName  string `json:"realName" binding:""`
	Email  string `json:"email" binding:""`
	Phone  string `json:"phone" binding:""`
	Status  int `json:"status" binding:""`
}

// SysUserObjDetail detail
type SysUserObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	Username  string `json:"username"`
	Password  string `json:"password"`
	RealName  string `json:"realName"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateSysUserReply only for api docs
type CreateSysUserReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteSysUserByIDReply only for api docs
type DeleteSysUserByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateSysUserByIDReply only for api docs
type UpdateSysUserByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetSysUserByIDReply only for api docs
type GetSysUserByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysUser SysUserObjDetail `json:"sysUser"`
	} `json:"data"` // return data
}

// ListSysUsersRequest request params
type ListSysUsersRequest struct {
	query.Params
}

// ListSysUsersReply only for api docs
type ListSysUsersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysUsers []SysUserObjDetail `json:"sysUsers"`
	} `json:"data"` // return data
}
