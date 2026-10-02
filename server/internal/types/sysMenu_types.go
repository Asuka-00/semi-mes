package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateSysMenuRequest request params
type CreateSysMenuRequest struct {
	ParentID  int `json:"parentID" binding:""`
	MenuType  int `json:"menuType" binding:""`
	MenuName  string `json:"menuName" binding:""`
	PermissionCode  string `json:"permissionCode" binding:""`
	RouteName  string `json:"routeName" binding:""`
	RoutePath  string `json:"routePath" binding:""`
	ComponentPath  string `json:"componentPath" binding:""`
	Icon  string `json:"icon" binding:""`
	SortOrder  int `json:"sortOrder" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateSysMenuByIDRequest request params
type UpdateSysMenuByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ParentID  int `json:"parentID" binding:""`
	MenuType  int `json:"menuType" binding:""`
	MenuName  string `json:"menuName" binding:""`
	PermissionCode  string `json:"permissionCode" binding:""`
	RouteName  string `json:"routeName" binding:""`
	RoutePath  string `json:"routePath" binding:""`
	ComponentPath  string `json:"componentPath" binding:""`
	Icon  string `json:"icon" binding:""`
	SortOrder  int `json:"sortOrder" binding:""`
	Status  int `json:"status" binding:""`
}

// SysMenuObjDetail detail
type SysMenuObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ParentID  int `json:"parentID"`
	MenuType  int `json:"menuType"`
	MenuName  string `json:"menuName"`
	PermissionCode  string `json:"permissionCode"`
	RouteName  string `json:"routeName"`
	RoutePath  string `json:"routePath"`
	ComponentPath  string `json:"componentPath"`
	Icon  string `json:"icon"`
	SortOrder  int `json:"sortOrder"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateSysMenuReply only for api docs
type CreateSysMenuReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteSysMenuByIDReply only for api docs
type DeleteSysMenuByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateSysMenuByIDReply only for api docs
type UpdateSysMenuByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetSysMenuByIDReply only for api docs
type GetSysMenuByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysMenu SysMenuObjDetail `json:"sysMenu"`
	} `json:"data"` // return data
}

// ListSysMenusRequest request params
type ListSysMenusRequest struct {
	query.Params
}

// ListSysMenusReply only for api docs
type ListSysMenusReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		SysMenus []SysMenuObjDetail `json:"sysMenus"`
	} `json:"data"` // return data
}
