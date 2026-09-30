package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseProduct business-level http error codes.
// the baseProductNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseProductNO = 54
	baseProductName     = "baseProduct"
	baseProductBaseCode = errcode.HCode(baseProductNO)

	ErrCreateBaseProduct     = errcode.NewError(baseProductBaseCode+1, "failed to create "+baseProductName)
	ErrDeleteByIDBaseProduct = errcode.NewError(baseProductBaseCode+2, "failed to delete "+baseProductName)
	ErrUpdateByIDBaseProduct = errcode.NewError(baseProductBaseCode+3, "failed to update "+baseProductName)
	ErrGetByIDBaseProduct    = errcode.NewError(baseProductBaseCode+4, "failed to get "+baseProductName+" details")
	ErrListBaseProduct       = errcode.NewError(baseProductBaseCode+5, "failed to list of "+baseProductName)

	// error codes are globally unique, adding 1 to the previous error code
)
