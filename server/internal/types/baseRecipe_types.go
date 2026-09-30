package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateBaseRecipeRequest request params
type CreateBaseRecipeRequest struct {
	OperationID  int `json:"operationID" binding:""`
	RecipeCode  string `json:"recipeCode" binding:""`
	RecipeName  string `json:"recipeName" binding:""`
	Version  string `json:"version" binding:""`
	Parameters  string `json:"parameters" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// UpdateBaseRecipeByIDRequest request params
type UpdateBaseRecipeByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	OperationID  int `json:"operationID" binding:""`
	RecipeCode  string `json:"recipeCode" binding:""`
	RecipeName  string `json:"recipeName" binding:""`
	Version  string `json:"version" binding:""`
	Parameters  string `json:"parameters" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""`
}

// BaseRecipeObjDetail detail
type BaseRecipeObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	OperationID  int `json:"operationID"`
	RecipeCode  string `json:"recipeCode"`
	RecipeName  string `json:"recipeName"`
	Version  string `json:"version"`
	Parameters  string `json:"parameters"`
	IsDefault  int `json:"isDefault"`
	Description  string `json:"description"`
	Status  int `json:"status"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateBaseRecipeReply only for api docs
type CreateBaseRecipeReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteBaseRecipeByIDReply only for api docs
type DeleteBaseRecipeByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateBaseRecipeByIDReply only for api docs
type UpdateBaseRecipeByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetBaseRecipeByIDReply only for api docs
type GetBaseRecipeByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseRecipe BaseRecipeObjDetail `json:"baseRecipe"`
	} `json:"data"` // return data
}

// ListBaseRecipesRequest request params
type ListBaseRecipesRequest struct {
	query.Params
}

// ListBaseRecipesReply only for api docs
type ListBaseRecipesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		BaseRecipes []BaseRecipeObjDetail `json:"baseRecipes"`
	} `json:"data"` // return data
}
