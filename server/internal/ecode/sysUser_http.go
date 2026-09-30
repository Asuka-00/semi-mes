package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// sysUser business-level http error codes.
// the sysUserNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	sysUserNO = 44
	sysUserName     = "sysUser"
	sysUserBaseCode = errcode.HCode(sysUserNO)

	ErrCreateSysUser     = errcode.NewError(sysUserBaseCode+1, "failed to create "+sysUserName)
	ErrDeleteByIDSysUser = errcode.NewError(sysUserBaseCode+2, "failed to delete "+sysUserName)
	ErrUpdateByIDSysUser = errcode.NewError(sysUserBaseCode+3, "failed to update "+sysUserName)
	ErrGetByIDSysUser    = errcode.NewError(sysUserBaseCode+4, "failed to get "+sysUserName+" details")
	ErrListSysUser       = errcode.NewError(sysUserBaseCode+5, "failed to list of "+sysUserName)

	// error codes are globally unique, adding 1 to the previous error code
)
