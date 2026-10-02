package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// sysUserRole business-level http error codes.
// the sysUserRoleNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	sysUserRoleNO = 10
	sysUserRoleName     = "sysUserRole"
	sysUserRoleBaseCode = errcode.HCode(sysUserRoleNO)

	ErrCreateSysUserRole     = errcode.NewError(sysUserRoleBaseCode+1, "failed to create "+sysUserRoleName)
	ErrDeleteByIDSysUserRole = errcode.NewError(sysUserRoleBaseCode+2, "failed to delete "+sysUserRoleName)
	ErrUpdateByIDSysUserRole = errcode.NewError(sysUserRoleBaseCode+3, "failed to update "+sysUserRoleName)
	ErrGetByIDSysUserRole    = errcode.NewError(sysUserRoleBaseCode+4, "failed to get "+sysUserRoleName+" details")
	ErrListSysUserRole       = errcode.NewError(sysUserRoleBaseCode+5, "failed to list of "+sysUserRoleName)

	// error codes are globally unique, adding 1 to the previous error code
)
