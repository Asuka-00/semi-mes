package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseOperation business-level http error codes.
// the baseOperationNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseOperationNO = 8
	baseOperationName     = "baseOperation"
	baseOperationBaseCode = errcode.HCode(baseOperationNO)

	ErrCreateBaseOperation     = errcode.NewError(baseOperationBaseCode+1, "failed to create "+baseOperationName)
	ErrDeleteByIDBaseOperation = errcode.NewError(baseOperationBaseCode+2, "failed to delete "+baseOperationName)
	ErrUpdateByIDBaseOperation = errcode.NewError(baseOperationBaseCode+3, "failed to update "+baseOperationName)
	ErrGetByIDBaseOperation    = errcode.NewError(baseOperationBaseCode+4, "failed to get "+baseOperationName+" details")
	ErrListBaseOperation       = errcode.NewError(baseOperationBaseCode+5, "failed to list of "+baseOperationName)

	// error codes are globally unique, adding 1 to the previous error code
)
