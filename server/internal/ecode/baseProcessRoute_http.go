package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseProcessRoute business-level http error codes.
// the baseProcessRouteNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseProcessRouteNO = 56
	baseProcessRouteName     = "baseProcessRoute"
	baseProcessRouteBaseCode = errcode.HCode(baseProcessRouteNO)

	ErrCreateBaseProcessRoute     = errcode.NewError(baseProcessRouteBaseCode+1, "failed to create "+baseProcessRouteName)
	ErrDeleteByIDBaseProcessRoute = errcode.NewError(baseProcessRouteBaseCode+2, "failed to delete "+baseProcessRouteName)
	ErrUpdateByIDBaseProcessRoute = errcode.NewError(baseProcessRouteBaseCode+3, "failed to update "+baseProcessRouteName)
	ErrGetByIDBaseProcessRoute    = errcode.NewError(baseProcessRouteBaseCode+4, "failed to get "+baseProcessRouteName+" details")
	ErrListBaseProcessRoute       = errcode.NewError(baseProcessRouteBaseCode+5, "failed to list of "+baseProcessRouteName)

	// error codes are globally unique, adding 1 to the previous error code
)
