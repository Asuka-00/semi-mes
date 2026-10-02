package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseOperationRequest request params
type CreateBaseOperationRequest struct {
	RouteID  int `json:"routeID" binding:""`
	OperationCode  string `json:"operationCode" binding:""`
	OperationName  string `json:"operationName" binding:""`
	OperationType  string `json:"operationType" binding:""`
	Sequence  int `json:"sequence" binding:""`
	StandardTime  int `json:"standardTime" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseOperationByIDRequest request params
type UpdateBaseOperationByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RouteID  int `json:"routeID" binding:""`
	OperationCode  string `json:"operationCode" binding:""`
	OperationName  string `json:"operationName" binding:""`
	OperationType  string `json:"operationType" binding:""`
	Sequence  int `json:"sequence" binding:""`
	StandardTime  int `json:"standardTime" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseOperationObjDetail detail
type BaseOperationObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RouteID  int `json:"routeID"`
	OperationCode  string `json:"operationCode"`
	OperationName  string `json:"operationName"`
	OperationType  string `json:"operationType"`
	Sequence  int `json:"sequence"`
	StandardTime  int `json:"standardTime"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseOperationReply only for api docs
type CreateBaseOperationReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseOperationByIDReply only for api docs
type DeleteBaseOperationByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseOperationByIDReply only for api docs
type UpdateBaseOperationByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseOperationByIDReply only for api docs
type GetBaseOperationByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseOperation BaseOperationObjDetail `json:"baseOperation"`
	} `json:"data"` // return data
}

// ListBaseOperationsRequest request params
type ListBaseOperationsRequest struct {
	query.Params
}

// ListBaseOperationsReply only for api docs
type ListBaseOperationsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseOperations []BaseOperationObjDetail `json:"baseOperations"`
	} `json:"data"` // return data
}
