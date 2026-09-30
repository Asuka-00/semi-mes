package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseProductRequest request params
type CreateBaseProductRequest struct {
	ProductCode  string `json:"productCode" binding:""`
	ProductName  string `json:"productName" binding:""`
	ProductType  string `json:"productType" binding:""`
	Version  string `json:"version" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseProductByIDRequest request params
type UpdateBaseProductByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ProductCode  string `json:"productCode" binding:""`
	ProductName  string `json:"productName" binding:""`
	ProductType  string `json:"productType" binding:""`
	Version  string `json:"version" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseProductObjDetail detail
type BaseProductObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ProductCode  string `json:"productCode"`
	ProductName  string `json:"productName"`
	ProductType  string `json:"productType"`
	Version  string `json:"version"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseProductReply only for api docs
type CreateBaseProductReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseProductByIDReply only for api docs
type DeleteBaseProductByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseProductByIDReply only for api docs
type UpdateBaseProductByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseProductByIDReply only for api docs
type GetBaseProductByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProduct BaseProductObjDetail `json:"baseProduct"`
	} `json:"data"` // return data
}

// ListBaseProductsRequest request params
type ListBaseProductsRequest struct {
	query.Params
}

// ListBaseProductsReply only for api docs
type ListBaseProductsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseProducts []BaseProductObjDetail `json:"baseProducts"`
	} `json:"data"` // return data
}
