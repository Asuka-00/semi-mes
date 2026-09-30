package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseRecipe business-level http error codes.
// the baseRecipeNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseRecipeNO = 41
	baseRecipeName     = "baseRecipe"
	baseRecipeBaseCode = errcode.HCode(baseRecipeNO)

	ErrCreateBaseRecipe     = errcode.NewError(baseRecipeBaseCode+1, "failed to create "+baseRecipeName)
	ErrDeleteByIDBaseRecipe = errcode.NewError(baseRecipeBaseCode+2, "failed to delete "+baseRecipeName)
	ErrUpdateByIDBaseRecipe = errcode.NewError(baseRecipeBaseCode+3, "failed to update "+baseRecipeName)
	ErrGetByIDBaseRecipe    = errcode.NewError(baseRecipeBaseCode+4, "failed to get "+baseRecipeName+" details")
	ErrListBaseRecipe       = errcode.NewError(baseRecipeBaseCode+5, "failed to list of "+baseRecipeName)

	// error codes are globally unique, adding 1 to the previous error code
)
