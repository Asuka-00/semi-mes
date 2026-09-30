package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseWorkshop business-level http error codes.
// the baseWorkshopNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseWorkshopNO = 21
	baseWorkshopName     = "baseWorkshop"
	baseWorkshopBaseCode = errcode.HCode(baseWorkshopNO)

	ErrCreateBaseWorkshop     = errcode.NewError(baseWorkshopBaseCode+1, "failed to create "+baseWorkshopName)
	ErrDeleteByIDBaseWorkshop = errcode.NewError(baseWorkshopBaseCode+2, "failed to delete "+baseWorkshopName)
	ErrUpdateByIDBaseWorkshop = errcode.NewError(baseWorkshopBaseCode+3, "failed to update "+baseWorkshopName)
	ErrGetByIDBaseWorkshop    = errcode.NewError(baseWorkshopBaseCode+4, "failed to get "+baseWorkshopName+" details")
	ErrListBaseWorkshop       = errcode.NewError(baseWorkshopBaseCode+5, "failed to list of "+baseWorkshopName)

	// error codes are globally unique, adding 1 to the previous error code
)
