package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseProcessRouteRequest request params
type CreateBaseProcessRouteRequest struct {
	ProductID  int `json:"productID" binding:""`
	RouteCode  string `json:"routeCode" binding:""`
	RouteName  string `json:"routeName" binding:""`
	Version  string `json:"version" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseProcessRouteByIDRequest request params
type UpdateBaseProcessRouteByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ProductID  int `json:"productID" binding:""`
	RouteCode  string `json:"routeCode" binding:""`
	RouteName  string `json:"routeName" binding:""`
	Version  string `json:"version" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseProcessRouteObjDetail detail
type BaseProcessRouteObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ProductID  int `json:"productID"`
	RouteCode  string `json:"routeCode"`
	RouteName  string `json:"routeName"`
	Version  string `json:"version"`
	IsDefault  int `json:"isDefault"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseProcessRouteReply only for api docs
type CreateBaseProcessRouteReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseProcessRouteByIDReply only for api docs
type DeleteBaseProcessRouteByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseProcessRouteByIDReply only for api docs
type UpdateBaseProcessRouteByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseProcessRouteByIDReply only for api docs
type GetBaseProcessRouteByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProcessRoute BaseProcessRouteObjDetail `json:"baseProcessRoute"`
	} `json:"data"` // return data
}

// ListBaseProcessRoutesRequest request params
type ListBaseProcessRoutesRequest struct {
	query.Params
}

// ListBaseProcessRoutesReply only for api docs
type ListBaseProcessRoutesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProcessRoutes []BaseProcessRouteObjDetail `json:"baseProcessRoutes"`
	} `json:"data"` // return data
}
