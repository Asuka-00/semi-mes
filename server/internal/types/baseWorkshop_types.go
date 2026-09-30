package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseWorkshopRequest request params
type CreateBaseWorkshopRequest struct {
	FactoryID  int `json:"factoryID" binding:""`
	WorkshopCode  string `json:"workshopCode" binding:""`
	WorkshopName  string `json:"workshopName" binding:""`
	WorkshopType  string `json:"workshopType" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseWorkshopByIDRequest request params
type UpdateBaseWorkshopByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	FactoryID  int `json:"factoryID" binding:""`
	WorkshopCode  string `json:"workshopCode" binding:""`
	WorkshopName  string `json:"workshopName" binding:""`
	WorkshopType  string `json:"workshopType" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseWorkshopObjDetail detail
type BaseWorkshopObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	FactoryID  int `json:"factoryID"`
	WorkshopCode  string `json:"workshopCode"`
	WorkshopName  string `json:"workshopName"`
	WorkshopType  string `json:"workshopType"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseWorkshopReply only for api docs
type CreateBaseWorkshopReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseWorkshopByIDReply only for api docs
type DeleteBaseWorkshopByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseWorkshopByIDReply only for api docs
type UpdateBaseWorkshopByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseWorkshopByIDReply only for api docs
type GetBaseWorkshopByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseWorkshop BaseWorkshopObjDetail `json:"baseWorkshop"`
	} `json:"data"` // return data
}

// ListBaseWorkshopsRequest request params
type ListBaseWorkshopsRequest struct {
	query.Params
}

// ListBaseWorkshopsReply only for api docs
type ListBaseWorkshopsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseWorkshops []BaseWorkshopObjDetail `json:"baseWorkshops"`
	} `json:"data"` // return data
}
