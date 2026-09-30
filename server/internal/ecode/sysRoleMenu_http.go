package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// sysRoleMenu business-level http error codes.
// the sysRoleMenuNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	sysRoleMenuNO = 73
	sysRoleMenuName     = "sysRoleMenu"
	sysRoleMenuBaseCode = errcode.HCode(sysRoleMenuNO)

	ErrCreateSysRoleMenu     = errcode.NewError(sysRoleMenuBaseCode+1, "failed to create "+sysRoleMenuName)
	ErrDeleteByIDSysRoleMenu = errcode.NewError(sysRoleMenuBaseCode+2, "failed to delete "+sysRoleMenuName)
	ErrUpdateByIDSysRoleMenu = errcode.NewError(sysRoleMenuBaseCode+3, "failed to update "+sysRoleMenuName)
	ErrGetByIDSysRoleMenu    = errcode.NewError(sysRoleMenuBaseCode+4, "failed to get "+sysRoleMenuName+" details")
	ErrListSysRoleMenu       = errcode.NewError(sysRoleMenuBaseCode+5, "failed to list of "+sysRoleMenuName)

	// error codes are globally unique, adding 1 to the previous error code
)
