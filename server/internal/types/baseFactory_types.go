package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseFactoryRequest request params
type CreateBaseFactoryRequest struct {
	FactoryCode  string `json:"factoryCode" binding:""`
	FactoryName  string `json:"factoryName" binding:""`
	Address  string `json:"address" binding:""`
	Contact  string `json:"contact" binding:""`
	Phone  string `json:"phone" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseFactoryByIDRequest request params
type UpdateBaseFactoryByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	FactoryCode  string `json:"factoryCode" binding:""`
	FactoryName  string `json:"factoryName" binding:""`
	Address  string `json:"address" binding:""`
	Contact  string `json:"contact" binding:""`
	Phone  string `json:"phone" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseFactoryObjDetail detail
type BaseFactoryObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	FactoryCode  string `json:"factoryCode"`
	FactoryName  string `json:"factoryName"`
	Address  string `json:"address"`
	Contact  string `json:"contact"`
	Phone  string `json:"phone"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseFactoryReply only for api docs
type CreateBaseFactoryReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseFactoryByIDReply only for api docs
type DeleteBaseFactoryByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseFactoryByIDReply only for api docs
type UpdateBaseFactoryByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseFactoryByIDReply only for api docs
type GetBaseFactoryByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseFactory BaseFactoryObjDetail `json:"baseFactory"`
	} `json:"data"` // return data
}

// ListBaseFactorysRequest request params
type ListBaseFactorysRequest struct {
	query.Params
}

// ListBaseFactorysReply only for api docs
type ListBaseFactorysReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseFactorys []BaseFactoryObjDetail `json:"baseFactorys"`
	} `json:"data"` // return data
}
