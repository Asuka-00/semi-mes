package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseFactory business-level http error codes.
// the baseFactoryNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseFactoryNO = 5
	baseFactoryName     = "baseFactory"
	baseFactoryBaseCode = errcode.HCode(baseFactoryNO)

	ErrCreateBaseFactory     = errcode.NewError(baseFactoryBaseCode+1, "failed to create "+baseFactoryName)
	ErrDeleteByIDBaseFactory = errcode.NewError(baseFactoryBaseCode+2, "failed to delete "+baseFactoryName)
	ErrUpdateByIDBaseFactory = errcode.NewError(baseFactoryBaseCode+3, "failed to update "+baseFactoryName)
	ErrGetByIDBaseFactory    = errcode.NewError(baseFactoryBaseCode+4, "failed to get "+baseFactoryName+" details")
	ErrListBaseFactory       = errcode.NewError(baseFactoryBaseCode+5, "failed to list of "+baseFactoryName)

	// error codes are globally unique, adding 1 to the previous error code
)
