package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// baseProductionLine business-level http error codes.
// the baseProductionLineNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	baseProductionLineNO = 4
	baseProductionLineName     = "baseProductionLine"
	baseProductionLineBaseCode = errcode.HCode(baseProductionLineNO)

	ErrCreateBaseProductionLine     = errcode.NewError(baseProductionLineBaseCode+1, "failed to create "+baseProductionLineName)
	ErrDeleteByIDBaseProductionLine = errcode.NewError(baseProductionLineBaseCode+2, "failed to delete "+baseProductionLineName)
	ErrUpdateByIDBaseProductionLine = errcode.NewError(baseProductionLineBaseCode+3, "failed to update "+baseProductionLineName)
	ErrGetByIDBaseProductionLine    = errcode.NewError(baseProductionLineBaseCode+4, "failed to get "+baseProductionLineName+" details")
	ErrListBaseProductionLine       = errcode.NewError(baseProductionLineBaseCode+5, "failed to list of "+baseProductionLineName)

	// error codes are globally unique, adding 1 to the previous error code
)
