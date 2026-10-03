package bootstrap

import (
	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

func ensureCatalog(db *gorm.DB) error {
	dicts := []struct {
		code, name, remark string
		items              []model.SysDictItem
	}{
		{"lot_status", "批次状态", "在制批次状态", []model.SysDictItem{
			{ItemCode: "waiting", LabelZh: "待进站", LabelEn: "Waiting", Color: "#64748b", Tag: "default", SortOrder: 1, Status: 1},
			{ItemCode: "running", LabelZh: "加工中", LabelEn: "Running", Color: "#2080f0", Tag: "info", SortOrder: 2, Status: 1},
			{ItemCode: "hold", LabelZh: "Hold", LabelEn: "Hold", Color: "#f0a020", Tag: "warning", SortOrder: 3, Status: 1},
			{ItemCode: "completed", LabelZh: "已完成", LabelEn: "Completed", Color: "#18a058", Tag: "success", SortOrder: 4, Status: 1},
			{ItemCode: "scrapped", LabelZh: "已报废", LabelEn: "Scrapped", Color: "#d03050", Tag: "error", SortOrder: 5, Status: 1},
			{ItemCode: "merged", LabelZh: "已合并", LabelEn: "Merged", Color: "#8b5cf6", Tag: "default", SortOrder: 6, Status: 1},
		}},
		{"lot_type", "批次类型", "", []model.SysDictItem{
			{ItemCode: "production", LabelZh: "量产", LabelEn: "Production", Color: "#2080f0", Tag: "info", SortOrder: 1, Status: 1},
			{ItemCode: "engineering", LabelZh: "工程", LabelEn: "Engineering", Color: "#f0a020", Tag: "warning", SortOrder: 2, Status: 1},
		}},
		{"order_status", "工单状态", "", []model.SysDictItem{
			{ItemCode: "created", LabelZh: "已创建", LabelEn: "Created", Color: "#64748b", SortOrder: 1, Status: 1},
			{ItemCode: "released", LabelZh: "已下达", LabelEn: "Released", Color: "#2080f0", SortOrder: 2, Status: 1},
			{ItemCode: "in_progress", LabelZh: "生产中", LabelEn: "In progress", Color: "#18a058", SortOrder: 3, Status: 1},
			{ItemCode: "completed", LabelZh: "已完成", LabelEn: "Completed", Color: "#18a058", SortOrder: 4, Status: 1},
			{ItemCode: "closed", LabelZh: "已关闭", LabelEn: "Closed", Color: "#d03050", SortOrder: 5, Status: 1},
		}},
		{"eqp_state", "设备状态", "E10 状态显示", []model.SysDictItem{
			{ItemCode: "standby", LabelZh: "待机", LabelEn: "Standby", Color: "#64748b", SortOrder: 1, Status: 1},
			{ItemCode: "productive", LabelZh: "生产", LabelEn: "Productive", Color: "#18a058", SortOrder: 2, Status: 1},
			{ItemCode: "engineering", LabelZh: "工程", LabelEn: "Engineering", Color: "#2080f0", SortOrder: 3, Status: 1},
			{ItemCode: "scheduled_down", LabelZh: "计划停机", LabelEn: "Scheduled down", Color: "#f0a020", SortOrder: 4, Status: 1},
			{ItemCode: "unscheduled_down", LabelZh: "非计划停机", LabelEn: "Unscheduled down", Color: "#d03050", SortOrder: 5, Status: 1},
			{ItemCode: "non_scheduled", LabelZh: "非计划时间", LabelEn: "Non-scheduled", Color: "#8b5cf6", SortOrder: 6, Status: 1},
		}},
		{"disposition", "缺陷处置", "", []model.SysDictItem{
			{ItemCode: "use_as_is", LabelZh: "特采", LabelEn: "Use as is", Color: "#18a058", SortOrder: 1, Status: 1},
			{ItemCode: "rework", LabelZh: "返工", LabelEn: "Rework", Color: "#f0a020", SortOrder: 2, Status: 1},
			{ItemCode: "scrap", LabelZh: "报废", LabelEn: "Scrap", Color: "#d03050", SortOrder: 3, Status: 1},
			{ItemCode: "hold", LabelZh: "Hold", LabelEn: "Hold", Color: "#2080f0", SortOrder: 4, Status: 1},
		}},
	}
	for _, item := range dicts {
		var count int64
		if err := db.Unscoped().Model(&model.SysDictType{}).Where("type_code = ?", item.code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		row := model.SysDictType{TypeCode: item.code, TypeName: item.name, Remark: item.remark, Status: 1}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
		for i := range item.items {
			item.items[i].TypeID = row.ID
		}
		if err := db.Create(&item.items).Error; err != nil {
			return err
		}
	}
	reasons := []model.MesReasonCode{
		{Category: "hold", ReasonCode: "ENG_HOLD", NameZh: "工程 Hold", NameEn: "Engineering hold", SortOrder: 1, Status: 1},
		{Category: "hold", ReasonCode: "YIELD", NameZh: "良率", NameEn: "Yield", SortOrder: 2, Status: 1},
		{Category: "hold", ReasonCode: "QUALITY", NameZh: "质量", NameEn: "Quality", SortOrder: 3, Status: 1},
		{Category: "hold", ReasonCode: "HOLD", NameZh: "一般 Hold", NameEn: "Hold", SortOrder: 4, Status: 1},
		{Category: "hold", ReasonCode: "QA", NameZh: "质量抽检", NameEn: "QA", SortOrder: 5, Status: 1},
		{Category: "release", ReasonCode: "RELEASE", NameZh: "解除", NameEn: "Release", SortOrder: 1, Status: 1},
		{Category: "release", ReasonCode: "ENG_OK", NameZh: "工程确认", NameEn: "Engineering OK", SortOrder: 2, Status: 1},
		{Category: "scrap", ReasonCode: "BROKEN", NameZh: "破片", NameEn: "Broken", SortOrder: 1, Status: 1},
		{Category: "scrap", ReasonCode: "PARTICLE", NameZh: "颗粒", NameEn: "Particle", SortOrder: 2, Status: 1},
		{Category: "scrap", ReasonCode: "SCRATCH", NameZh: "划伤", NameEn: "Scratch", SortOrder: 3, Status: 1},
		{Category: "scrap", ReasonCode: "OTHER", NameZh: "其他", NameEn: "Other", SortOrder: 4, Status: 1},
		{Category: "rework", ReasonCode: "REWORK", NameZh: "返工", NameEn: "Rework", SortOrder: 1, Status: 1},
		{Category: "rework", ReasonCode: "RC_PHOTO", NameZh: "光刻返工", NameEn: "Photo rework", SortOrder: 2, Status: 1},
		{Category: "eqp_down", ReasonCode: "BREAKDOWN", NameZh: "故障", NameEn: "Breakdown", SortOrder: 1, Status: 1},
		{Category: "eqp_down", ReasonCode: "PM_START", NameZh: "开始 PM", NameEn: "PM start", SortOrder: 2, Status: 1},
		{Category: "eqp", ReasonCode: "ENG_SETUP", NameZh: "工程准备", NameEn: "Engineering setup", SortOrder: 1, Status: 1},
		{Category: "eqp", ReasonCode: "ENG_DONE", NameZh: "工程结束", NameEn: "Engineering done", SortOrder: 2, Status: 1},
		{Category: "eqp", ReasonCode: "PM_START", NameZh: "开始 PM", NameEn: "PM start", SortOrder: 3, Status: 1},
		{Category: "eqp", ReasonCode: "PM_DONE", NameZh: "PM 完成", NameEn: "PM done", SortOrder: 4, Status: 1},
		{Category: "eqp", ReasonCode: "BREAKDOWN", NameZh: "故障", NameEn: "Breakdown", SortOrder: 5, Status: 1},
		{Category: "eqp", ReasonCode: "REPAIR_DONE", NameZh: "修复完成", NameEn: "Repair done", SortOrder: 6, Status: 1},
		{Category: "eqp", ReasonCode: "NO_WIP", NameZh: "无在制", NameEn: "No WIP", SortOrder: 7, Status: 1},
		{Category: "eqp", ReasonCode: "SHIFT_END", NameZh: "交班", NameEn: "Shift end", SortOrder: 8, Status: 1},
		{Category: "eqp", ReasonCode: "SHIFT_START", NameZh: "接班", NameEn: "Shift start", SortOrder: 9, Status: 1},
		{Category: "eqp", ReasonCode: "OTHER", NameZh: "其他", NameEn: "Other", SortOrder: 10, Status: 1},
		{Category: "eqp", ReasonCode: "PROD_START", NameZh: "开始生产", NameEn: "Production start", SortOrder: 11, Status: 1},
		{Category: "eqp", ReasonCode: "PROD_END", NameZh: "结束生产", NameEn: "Production end", SortOrder: 12, Status: 1},
	}
	for _, row := range reasons {
		var count int64
		if err := db.Unscoped().Model(&model.MesReasonCode{}).Where("category = ? AND reason_code = ?", row.Category, row.ReasonCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	}
	rules := []model.SysNumberRule{
		{RuleCode: model.RuleWorkOrder, RuleName: "工单号", Prefix: "WO", DatePart: "yyyyMMdd", SeqLength: 3, ResetPeriod: "daily", Separator: "-", Status: 1},
		{RuleCode: model.RuleLot, RuleName: "批次号", Prefix: "LOT", DatePart: "yyyyMMdd", SeqLength: 4, ResetPeriod: "daily", Separator: "-", Status: 1},
	}
	for _, row := range rules {
		var count int64
		if err := db.Unscoped().Model(&model.SysNumberRule{}).Where("rule_code = ?", row.RuleCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
