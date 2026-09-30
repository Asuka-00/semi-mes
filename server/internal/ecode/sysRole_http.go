package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// sysRole business-level http error codes.
// the sysRoleNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	sysRoleNO = 11
	sysRoleName     = "sysRole"
	sysRoleBaseCode = errcode.HCode(sysRoleNO)

	ErrCreateSysRole     = errcode.NewError(sysRoleBaseCode+1, "failed to create "+sysRoleName)
	ErrDeleteByIDSysRole = errcode.NewError(sysRoleBaseCode+2, "failed to delete "+sysRoleName)
	ErrUpdateByIDSysRole = errcode.NewError(sysRoleBaseCode+3, "failed to update "+sysRoleName)
	ErrGetByIDSysRole    = errcode.NewError(sysRoleBaseCode+4, "failed to get "+sysRoleName+" details")
	ErrListSysRole       = errcode.NewError(sysRoleBaseCode+5, "failed to list of "+sysRoleName)

	// error codes are globally unique, adding 1 to the previous error code
)
