package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseProductionLineRequest request params
type CreateBaseProductionLineRequest struct {
	WorkshopID  int `json:"workshopID" binding:""`
	LineCode  string `json:"lineCode" binding:""`
	LineName  string `json:"lineName" binding:""`
	Capacity  int `json:"capacity" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseProductionLineByIDRequest request params
type UpdateBaseProductionLineByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	WorkshopID  int `json:"workshopID" binding:""`
	LineCode  string `json:"lineCode" binding:""`
	LineName  string `json:"lineName" binding:""`
	Capacity  int `json:"capacity" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseProductionLineObjDetail detail
type BaseProductionLineObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	WorkshopID  int `json:"workshopID"`
	LineCode  string `json:"lineCode"`
	LineName  string `json:"lineName"`
	Capacity  int `json:"capacity"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseProductionLineReply only for api docs
type CreateBaseProductionLineReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseProductionLineByIDReply only for api docs
type DeleteBaseProductionLineByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseProductionLineByIDReply only for api docs
type UpdateBaseProductionLineByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseProductionLineByIDReply only for api docs
type GetBaseProductionLineByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProductionLine BaseProductionLineObjDetail `json:"baseProductionLine"`
	} `json:"data"` // return data
}

// ListBaseProductionLinesRequest request params
type ListBaseProductionLinesRequest struct {
	query.Params
}

// ListBaseProductionLinesReply only for api docs
type ListBaseProductionLinesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProductionLines []BaseProductionLineObjDetail `json:"baseProductionLines"`
	} `json:"data"` // return data
}
