package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// sysMenu business-level http error codes.
// the sysMenuNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	sysMenuNO = 47
	sysMenuName     = "sysMenu"
	sysMenuBaseCode = errcode.HCode(sysMenuNO)

	ErrCreateSysMenu     = errcode.NewError(sysMenuBaseCode+1, "failed to create "+sysMenuName)
	ErrDeleteByIDSysMenu = errcode.NewError(sysMenuBaseCode+2, "failed to delete "+sysMenuName)
	ErrUpdateByIDSysMenu = errcode.NewError(sysMenuBaseCode+3, "failed to update "+sysMenuName)
	ErrGetByIDSysMenu    = errcode.NewError(sysMenuBaseCode+4, "failed to get "+sysMenuName+" details")
	ErrListSysMenu       = errcode.NewError(sysMenuBaseCode+5, "failed to list of "+sysMenuName)

	// error codes are globally unique, adding 1 to the previous error code
)
