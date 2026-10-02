package bootstrap

import (
	"fmt"
	"time"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type menuSeed struct {
	ID             uint64
	ParentID       int
	MenuType       int
	MenuName       string
	PermissionCode string
	RouteName      string
	RoutePath      string
	ComponentPath  string
	Icon           string
	SortOrder      int
}

// Seed creates tables and the default admin, roles, and menu tree when empty.
func Seed(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.SysUser{},
		&model.SysRole{},
		&model.SysMenu{},
		&model.SysUserRole{},
		&model.SysRoleMenu{},
		&model.BaseFactory{},
		&model.BaseWorkshop{},
		&model.BaseProductionLine{},
		&model.BaseProduct{},
		&model.BaseProcessRoute{},
		&model.BaseOperation{},
		&model.BaseRecipe{},
		&model.BaseRouteVersion{},
		&model.BaseRouteNode{},
		&model.BaseRouteEdge{},
		&model.WipWorkOrder{},
		&model.WipLot{},
		&model.WipLotHistory{},
		&model.WipLotLink{},
		&model.EqpEquipment{},
		&model.EqpCapability{},
		&model.EqpStateLog{},
		&model.EqpPmPlan{},
		&model.EqpPmTask{},
		&model.WipMove{},
	); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&model.SysUser{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		if err := ensureMenus(db); err != nil {
			return err
		}
		return seedSampleRoute(db)
	}

	roles := []model.SysRole{
		{ID: 1, RoleCode: "super_admin", RoleName: "Super Administrator", Description: "All permissions", Status: 1},
		{ID: 2, RoleCode: "sys_admin", RoleName: "System Administrator", Description: "System management", Status: 1},
		{ID: 3, RoleCode: "operator", RoleName: "Operator", Description: "Base data maintenance", Status: 1},
		{ID: 4, RoleCode: "viewer", RoleName: "Viewer", Description: "Read only", Status: 1},
	}
	if err := db.Create(&roles).Error; err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin := model.SysUser{
		ID: 1, Username: "admin", Password: string(hashed), RealName: "Administrator",
		Email: "admin@example.com", Phone: "13800138000", Status: 1,
	}
	if err = db.Create(&admin).Error; err != nil {
		return err
	}
	if err = db.Create(&model.SysUserRole{UserID: 1, RoleID: 1}).Error; err != nil {
		return err
	}

	menus := defaultMenus()
	if err = db.Create(&menus).Error; err != nil {
		return err
	}

	var roleMenus []model.SysRoleMenu
	for _, menu := range menus {
		roleMenus = append(roleMenus, model.SysRoleMenu{RoleID: 1, MenuID: int(menu.ID)})
		if isViewerMenu(menu) {
			roleMenus = append(roleMenus, model.SysRoleMenu{RoleID: 4, MenuID: int(menu.ID)})
		}
		if isOperatorMenu(menu) {
			roleMenus = append(roleMenus, model.SysRoleMenu{RoleID: 3, MenuID: int(menu.ID)})
		}
		if isSysAdminMenu(menu) {
			roleMenus = append(roleMenus, model.SysRoleMenu{RoleID: 2, MenuID: int(menu.ID)})
		}
	}
	if err = db.Create(&roleMenus).Error; err != nil {
		return err
	}
	return seedSampleRoute(db)
}

// ensureMenus inserts menu rows added after the first seed and grants them to the matching roles.
func ensureMenus(db *gorm.DB) error {
	for _, menu := range defaultMenus() {
		var count int64
		if err := db.Model(&model.SysMenu{}).Where("id = ?", menu.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
		links := []model.SysRoleMenu{{RoleID: 1, MenuID: int(menu.ID)}}
		if isViewerMenu(menu) {
			links = append(links, model.SysRoleMenu{RoleID: 4, MenuID: int(menu.ID)})
		}
		if isOperatorMenu(menu) {
			links = append(links, model.SysRoleMenu{RoleID: 3, MenuID: int(menu.ID)})
		}
		if isSysAdminMenu(menu) {
			links = append(links, model.SysRoleMenu{RoleID: 2, MenuID: int(menu.ID)})
		}
		if err := db.Create(&links).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSampleRoute(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.BaseRouteVersion{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return seedShopfloorFromExisting(db)
	}
	product := model.BaseProduct{ProductCode: "CMOS-DEMO", ProductName: "CMOS Demo", ProductType: "IC", Version: "A", Status: 1}
	if err := db.Where(model.BaseProduct{ProductCode: "CMOS-DEMO"}).FirstOrCreate(&product).Error; err != nil {
		return err
	}
	ops := []model.BaseOperation{
		{OperationCode: "CLEAN", OperationName: "清洗 Clean", OperationType: "WET", Status: 1},
		{OperationCode: "PHOTO", OperationName: "光刻 Photo", OperationType: "PHOTO", Status: 1},
		{OperationCode: "INSPECT", OperationName: "检测 Inspect", OperationType: "METRO", Status: 1},
		{OperationCode: "ETCH", OperationName: "刻蚀 Etch", OperationType: "ETCH", Status: 1},
		{OperationCode: "ENG_REVIEW", OperationName: "工程评审 Eng Review", OperationType: "ENG", Status: 1},
	}
	ids := map[string]uint64{}
	for _, op := range ops {
		row := op
		if err := db.Where(model.BaseOperation{OperationCode: op.OperationCode}).FirstOrCreate(&row).Error; err != nil {
			return err
		}
		ids[op.OperationCode] = row.ID
	}
	route := model.BaseProcessRoute{
		ProductID: int(product.ID), RouteCode: "ROUTE-CMOS", RouteName: "CMOS 主流程", Version: "1",
		IsDefault: 1, Description: "检测后按结果返工或进入工程分支", Status: 1,
	}
	if err := db.Where(model.BaseProcessRoute{RouteCode: "ROUTE-CMOS"}).FirstOrCreate(&route).Error; err != nil {
		return err
	}
	ver := model.BaseRouteVersion{RouteID: route.ID, VersionNo: 1, State: routegraph.StateDraft, Note: "示例：光刻返工与工程分支"}
	if err := db.Create(&ver).Error; err != nil {
		return err
	}
	if err := dao.SaveRouteGraph(db, route.ID, ver.ID, routegraph.Sample(ids)); err != nil {
		return err
	}
	_, issues, err := dao.ReleaseRouteVersion(db, route.ID, ver.ID)
	if err != nil {
		return err
	}
	if len(issues) > 0 {
		return fmt.Errorf("sample route invalid: %s", issues[0].Code)
	}
	return seedShopfloor(db, product.ID, ver.ID)
}

func seedShopfloorFromExisting(db *gorm.DB) error {
	var product model.BaseProduct
	if err := db.Where("product_code = ?", "CMOS-DEMO").First(&product).Error; err != nil {
		return nil
	}
	var ver model.BaseRouteVersion
	if err := db.Where("state = ?", routegraph.StateReleased).Order("id desc").First(&ver).Error; err != nil {
		return nil
	}
	return seedShopfloor(db, product.ID, ver.ID)
}

func seedShopfloor(db *gorm.DB, productID, versionID uint64) error {
	tools := []model.EqpEquipment{
		{EquipmentCode: "WET-01", EquipmentName: "清洗槽 Wet Bench", EquipmentGroup: "WET", EquipmentType: "WET", Status: model.EqpStandby, Capacity: 2, ChamberCount: 2, Manufacturer: "Screen", ModelName: "SU-3200", SerialNo: "WET-2401", Location: "Bay A1"},
		{EquipmentCode: "PHOTO-01", EquipmentName: "光刻机 Stepper", EquipmentGroup: "PHOTO", EquipmentType: "PHOTO", Status: model.EqpStandby, Capacity: 1, ChamberCount: 1, Manufacturer: "ASML", ModelName: "XT-860", SerialNo: "PH-1108", Location: "Bay B2"},
		{EquipmentCode: "PHOTO-09", EquipmentName: "光刻机（停机）", EquipmentGroup: "PHOTO", EquipmentType: "PHOTO", Status: model.EqpUnscheduledDown, Capacity: 1, ChamberCount: 1, Manufacturer: "ASML", ModelName: "XT-400", SerialNo: "PH-0902", Location: "Bay B3"},
		{EquipmentCode: "METRO-01", EquipmentName: "检测机 Metrology", EquipmentGroup: "METRO", EquipmentType: "METRO", Status: model.EqpEngineering, Capacity: 1, ChamberCount: 1, Manufacturer: "KLA", ModelName: "29xx", SerialNo: "MT-331", Location: "Bay C1"},
		{EquipmentCode: "ETCH-01", EquipmentName: "刻蚀机 Etcher", EquipmentGroup: "ETCH", EquipmentType: "ETCH", Status: model.EqpStandby, Capacity: 1, ChamberCount: 2, Manufacturer: "Lam", ModelName: "Kiyo", SerialNo: "ET-778", Location: "Bay D1"},
		{EquipmentCode: "ENG-01", EquipmentName: "工程台 Eng Bench", EquipmentGroup: "ENG", EquipmentType: "ENG", Status: model.EqpNonScheduled, Capacity: 1, ChamberCount: 1, Manufacturer: "Local", ModelName: "Bench", SerialNo: "EN-001", Location: "Lab"},
	}
	for _, tool := range tools {
		row := tool
		if err := db.Where(model.EqpEquipment{EquipmentCode: tool.EquipmentCode}).Attrs(tool).FirstOrCreate(&row).Error; err != nil {
			return err
		}
	}
	if err := seedEquipmentExtras(db); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&model.WipWorkOrder{}).Where("order_no = ?", "DEMO-WIP").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	arrived := now.Add(-2 * time.Hour)
	order := model.WipWorkOrder{
		OrderNo: "DEMO-WIP", ProductID: productID, RouteVersionID: versionID, PlannedQty: 30,
		ReleasedQty: 16, NextLotSeq: 3, Priority: 3, Status: model.OrderInProgress, Note: "演示在制品",
	}
	if err := db.Create(&order).Error; err != nil {
		return err
	}
	photoAt := arrived
	lots := []model.WipLot{
		{LotNo: "DEMO-WIP-001", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "photo", Quantity: 10, Priority: 3, LotType: model.LotTypeProduction, Status: model.LotWaiting, ReworkJSON: "{}", ArrivedAt: &photoAt},
		{LotNo: "DEMO-WIP-002", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "inspect", Quantity: 4, Priority: 3, LotType: model.LotTypeEngineering, Status: model.LotHold, HoldReasonCode: "ENG_HOLD", HoldReason: "等待工程确认", ReworkJSON: "{}", ArrivedAt: &photoAt},
	}
	for i := range lots {
		if err := db.Create(&lots[i]).Error; err != nil {
			return err
		}
	}
	trackIn := now.Add(-20 * time.Minute)
	running := model.WipLot{
		LotNo: "DEMO-WIP-003", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "clean",
		Quantity: 2, Priority: 3, LotType: model.LotTypeProduction, Status: model.LotRunning, ReworkJSON: "{}", ArrivedAt: &arrived,
	}
	if err := db.Create(&running).Error; err != nil {
		return err
	}
	var wet model.EqpEquipment
	if err := db.Where("equipment_code = ?", "WET-01").First(&wet).Error; err != nil {
		return err
	}
	if err := db.Create(&model.WipMove{
		LotID: running.ID, NodeKey: "clean", EquipmentID: wet.ID, OperatorID: 1, QtyIn: 2, State: model.MoveOpen,
		TrackInAt: &trackIn, QueueSeconds: int(trackIn.Sub(arrived).Seconds()), FromNodeKey: "clean",
	}).Error; err != nil {
		return err
	}
	if err := db.Model(&wet).Updates(map[string]any{"status": model.EqpProductive, "resume_state": model.EqpStandby}).Error; err != nil {
		return err
	}
	if err := db.Model(&model.EqpStateLog{}).Where("equipment_id = ? AND ended_at IS NULL", wet.ID).
		Updates(map[string]any{"ended_at": trackIn, "duration_seconds": int(trackIn.Sub(arrived).Seconds())}).Error; err != nil {
		return err
	}
	return db.Create(&model.EqpStateLog{
		EquipmentID: wet.ID, FromState: model.EqpStandby, ToState: model.EqpProductive, ReasonCode: "PROD_START",
		OperatorID: 1, StartedAt: trackIn,
	}).Error
}

func seedEquipmentExtras(db *gorm.DB) error {
	if err := db.Model(&model.EqpEquipment{}).Where("status = ?", model.EqpIdle).Update("status", model.EqpStandby).Error; err != nil {
		return err
	}
	if err := db.Model(&model.EqpEquipment{}).Where("status = ?", model.EqpDown).Update("status", model.EqpUnscheduledDown).Error; err != nil {
		return err
	}
	if err := db.Model(&model.EqpEquipment{}).Where("capacity = 0").Update("capacity", 1).Error; err != nil {
		return err
	}
	var tools []model.EqpEquipment
	if err := db.Find(&tools).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, tool := range tools {
		var logs int64
		if err := db.Model(&model.EqpStateLog{}).Where("equipment_id = ?", tool.ID).Count(&logs).Error; err != nil {
			return err
		}
		if logs == 0 {
			started := now.Add(-48 * time.Hour)
			if err := db.Create(&model.EqpStateLog{
				EquipmentID: tool.ID, FromState: "", ToState: tool.Status, ReasonCode: "SHIFT_START", StartedAt: started,
			}).Error; err != nil {
				return err
			}
		}
	}
	var plans int64
	if err := db.Model(&model.EqpPmPlan{}).Count(&plans).Error; err != nil {
		return err
	}
	if plans == 0 {
		var photo, etch model.EqpEquipment
		_ = db.Where("equipment_code = ?", "PHOTO-01").First(&photo).Error
		_ = db.Where("equipment_code = ?", "ETCH-01").First(&etch).Error
		soon := now.Add(10 * 24 * time.Hour)
		overdue := now.Add(-24 * time.Hour)
		var eng model.EqpEquipment
		_ = db.Where("equipment_code = ?", "ENG-01").First(&eng).Error
		rows := []model.EqpPmPlan{
			{EquipmentID: photo.ID, EquipmentGroup: "PHOTO", PlanName: "光刻灯管检查", TriggerType: model.PmTime, IntervalDays: 30, ChecklistJSON: `["灯管外观","能量均匀性","颗粒"]`, Enabled: true, NextDueAt: &soon},
			{EquipmentID: etch.ID, EquipmentGroup: "ETCH", PlanName: "刻蚀腔体清洁", TriggerType: model.PmTime, IntervalDays: 7, ChecklistJSON: `["腔体开盖","部件更换","漏率"]`, Enabled: true, BlockTrackIn: true, NextDueAt: &overdue},
			{EquipmentGroup: "WET", PlanName: "清洗槽换液", TriggerType: model.PmCount, IntervalCount: 40, ChecklistJSON: `["药液浓度","温度"]`, Enabled: true},
			{EquipmentID: eng.ID, EquipmentGroup: "ENG", PlanName: "工程台点检", TriggerType: model.PmTime, IntervalDays: 14, ChecklistJSON: `["接地","照明"]`, Enabled: true, NextDueAt: &overdue},
		}
		if err := db.Create(&rows).Error; err != nil {
			return err
		}
	}
	if err := seedCapabilities(db); err != nil {
		return err
	}
	if err := dao.EnsurePmTasks(db); err != nil {
		return err
	}
	var etch model.EqpEquipment
	if err := db.Where("equipment_code = ?", "ETCH-01").First(&etch).Error; err != nil {
		return nil
	}
	if etch.Status == model.EqpStandby {
		var task model.EqpPmTask
		err := db.Where("equipment_id = ? AND status IN ?", etch.ID, []string{model.PmDue, model.PmOverdue}).Order("id").First(&task).Error
		if err == nil {
			if _, err := dao.StartPm(db, task.ID, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func seedCapabilities(db *gorm.DB) error {
	recipes := []struct{ op, code, name string }{
		{"CLEAN", "RCP-CLEAN", "SC1 清洗"},
		{"PHOTO", "RCP-PHOTO", "i-line 曝光"},
		{"INSPECT", "RCP-INSPECT", "CD 量测"},
		{"ETCH", "RCP-ETCH", "多晶硅刻蚀"},
	}
	recipeID := map[string]uint64{}
	for _, item := range recipes {
		var op model.BaseOperation
		if err := db.Where("operation_code = ?", item.op).First(&op).Error; err != nil {
			continue
		}
		row := model.BaseRecipe{}
		err := db.Where("recipe_code = ?", item.code).Attrs(model.BaseRecipe{
			OperationID: int(op.ID), RecipeCode: item.code, RecipeName: item.name, Version: "1",
			IsDefault: 1, Status: 1, Parameters: "{}",
		}).FirstOrCreate(&row).Error
		if err != nil {
			return err
		}
		recipeID[item.op] = row.ID
	}
	// WET-01 / PHOTO-01 / METRO-01 stay group-only. Tests track those tools with a second
	// released graph whose operation ids differ from the CMOS seed.
	links := []struct{ eqp, op string }{
		{"PHOTO-09", "PHOTO"},
		{"ETCH-01", "ETCH"},
		{"ENG-01", "ENG_REVIEW"},
	}
	for _, link := range links {
		var eqp model.EqpEquipment
		if err := db.Where("equipment_code = ?", link.eqp).First(&eqp).Error; err != nil {
			continue
		}
		var n int64
		if err := db.Model(&model.EqpCapability{}).Where("equipment_id = ?", eqp.ID).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		var op model.BaseOperation
		if err := db.Where("operation_code = ?", link.op).First(&op).Error; err != nil {
			continue
		}
		cap := model.EqpCapability{EquipmentID: eqp.ID, OperationID: op.ID}
		if link.op == "ETCH" {
			cap.RecipeID = recipeID["ETCH"]
		}
		if err := db.Create(&cap).Error; err != nil {
			return err
		}
	}
	return nil
}

func isBaseMenu(menu model.SysMenu) bool {
	if menu.RouteName == "home" || menu.RouteName == "base-data" || hasPrefix(menu.RouteName, "base-data_") {
		return true
	}
	// Button rows keep an empty route name and carry the permission code only.
	return hasPrefix(menu.PermissionCode, "base:")
}

func isWipMenu(menu model.SysMenu) bool {
	if menu.RouteName == "work-order" || menu.RouteName == "lot" || hasPrefix(menu.RouteName, "work-order_") || hasPrefix(menu.RouteName, "lot_") {
		return true
	}
	if menu.RouteName == "wip" || hasPrefix(menu.RouteName, "wip_") || hasPrefix(menu.PermissionCode, "wip:") {
		return true
	}
	if menu.RouteName == "equipment" || hasPrefix(menu.RouteName, "equipment_") || hasPrefix(menu.PermissionCode, "eqp:") {
		return true
	}
	return hasPrefix(menu.PermissionCode, "wo:") || hasPrefix(menu.PermissionCode, "lot:")
}

func isModuleMenu(menu model.SysMenu) bool {
	return isBaseMenu(menu) || isWipMenu(menu)
}

func isViewerMenu(menu model.SysMenu) bool {
	if !isModuleMenu(menu) {
		return false
	}
	return menu.MenuType != 3 || hasSuffix(menu.PermissionCode, ":query")
}

func isOperatorMenu(menu model.SysMenu) bool {
	if !isModuleMenu(menu) {
		return false
	}
	return menu.MenuType != 3 || !hasSuffix(menu.PermissionCode, ":delete")
}

func isSysAdminMenu(menu model.SysMenu) bool {
	return menu.RouteName == "home" || menu.RouteName == "system" || hasPrefix(menu.RouteName, "system_") || hasPrefix(menu.PermissionCode, "system:")
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func defaultMenus() []model.SysMenu {
	raw := []menuSeed{
		{1, 0, 2, "route.home", "dashboard", "home", "/home", "layout.base$view.home", "mdi:monitor-dashboard", 1},
		{2, 0, 1, "route.system", "system", "system", "/system", "layout.base", "mdi:cog", 2},
		{3, 2, 2, "route.system_user", "system:user", "system_user", "/system/user", "view.system_user", "mdi:account", 1},
		{4, 3, 3, "route.system_user", "system:user:query", "", "", "", "", 1},
		{5, 3, 3, "route.system_user", "system:user:add", "", "", "", "", 2},
		{6, 3, 3, "route.system_user", "system:user:edit", "", "", "", "", 3},
		{7, 3, 3, "route.system_user", "system:user:delete", "", "", "", "", 4},
		{8, 2, 2, "route.system_role", "system:role", "system_role", "/system/role", "view.system_role", "mdi:account-group", 2},
		{9, 8, 3, "route.system_role", "system:role:query", "", "", "", "", 1},
		{10, 8, 3, "route.system_role", "system:role:add", "", "", "", "", 2},
		{11, 8, 3, "route.system_role", "system:role:edit", "", "", "", "", 3},
		{12, 8, 3, "route.system_role", "system:role:delete", "", "", "", "", 4},
		{13, 2, 2, "route.system_menu", "system:menu", "system_menu", "/system/menu", "view.system_menu", "mdi:menu", 3},
		{14, 13, 3, "route.system_menu", "system:menu:query", "", "", "", "", 1},
		{15, 13, 3, "route.system_menu", "system:menu:add", "", "", "", "", 2},
		{16, 13, 3, "route.system_menu", "system:menu:edit", "", "", "", "", 3},
		{17, 13, 3, "route.system_menu", "system:menu:delete", "", "", "", "", 4},

		{100, 0, 1, "route.base-data", "baseData", "base-data", "/base-data", "layout.base", "mdi:database", 3},
		{101, 100, 2, "route.base-data_factory", "base:factory", "base-data_factory", "/base-data/factory", "view.base-data_factory", "mdi:factory", 1},
		{102, 101, 3, "route.base-data_factory", "base:factory:query", "", "", "", "", 1},
		{103, 101, 3, "route.base-data_factory", "base:factory:add", "", "", "", "", 2},
		{104, 101, 3, "route.base-data_factory", "base:factory:edit", "", "", "", "", 3},
		{105, 101, 3, "route.base-data_factory", "base:factory:delete", "", "", "", "", 4},
		{106, 100, 2, "route.base-data_workshop", "base:workshop", "base-data_workshop", "/base-data/workshop", "view.base-data_workshop", "mdi:warehouse", 2},
		{107, 106, 3, "route.base-data_workshop", "base:workshop:query", "", "", "", "", 1},
		{108, 106, 3, "route.base-data_workshop", "base:workshop:add", "", "", "", "", 2},
		{109, 106, 3, "route.base-data_workshop", "base:workshop:edit", "", "", "", "", 3},
		{110, 106, 3, "route.base-data_workshop", "base:workshop:delete", "", "", "", "", 4},
		{111, 100, 2, "route.base-data_line", "base:line", "base-data_line", "/base-data/line", "view.base-data_line", "mdi:pipe", 3},
		{112, 111, 3, "route.base-data_line", "base:line:query", "", "", "", "", 1},
		{113, 111, 3, "route.base-data_line", "base:line:add", "", "", "", "", 2},
		{114, 111, 3, "route.base-data_line", "base:line:edit", "", "", "", "", 3},
		{115, 111, 3, "route.base-data_line", "base:line:delete", "", "", "", "", 4},
		{116, 100, 2, "route.base-data_product", "base:product", "base-data_product", "/base-data/product", "view.base-data_product", "mdi:chip", 4},
		{117, 116, 3, "route.base-data_product", "base:product:query", "", "", "", "", 1},
		{118, 116, 3, "route.base-data_product", "base:product:add", "", "", "", "", 2},
		{119, 116, 3, "route.base-data_product", "base:product:edit", "", "", "", "", 3},
		{120, 116, 3, "route.base-data_product", "base:product:delete", "", "", "", "", 4},
		{121, 100, 2, "route.base-data_route", "base:route", "base-data_route", "/base-data/route", "view.base-data_route", "mdi:map-marker-path", 5},
		{122, 121, 3, "route.base-data_route", "base:route:query", "", "", "", "", 1},
		{123, 121, 3, "route.base-data_route", "base:route:add", "", "", "", "", 2},
		{124, 121, 3, "route.base-data_route", "base:route:edit", "", "", "", "", 3},
		{125, 121, 3, "route.base-data_route", "base:route:delete", "", "", "", "", 4},
		{126, 100, 2, "route.base-data_operation", "base:operation", "base-data_operation", "/base-data/operation", "view.base-data_operation", "mdi:cog-outline", 6},
		{127, 126, 3, "route.base-data_operation", "base:operation:query", "", "", "", "", 1},
		{128, 126, 3, "route.base-data_operation", "base:operation:add", "", "", "", "", 2},
		{129, 126, 3, "route.base-data_operation", "base:operation:edit", "", "", "", "", 3},
		{130, 126, 3, "route.base-data_operation", "base:operation:delete", "", "", "", "", 4},
		{131, 100, 2, "route.base-data_recipe", "base:recipe", "base-data_recipe", "/base-data/recipe", "view.base-data_recipe", "mdi:file-document-outline", 7},
		{132, 131, 3, "route.base-data_recipe", "base:recipe:query", "", "", "", "", 1},
		{133, 131, 3, "route.base-data_recipe", "base:recipe:add", "", "", "", "", 2},
		{134, 131, 3, "route.base-data_recipe", "base:recipe:edit", "", "", "", "", 3},
		{135, 131, 3, "route.base-data_recipe", "base:recipe:delete", "", "", "", "", 4},

		{200, 0, 1, "route.work-order", "workOrder", "work-order", "/work-order", "layout.base", "mdi:clipboard-text", 4},
		{201, 200, 2, "route.work-order_list", "wo:order", "work-order_list", "/work-order/list", "view.work-order_list", "mdi:format-list-bulleted", 1},
		{202, 201, 3, "route.work-order_list", "wo:order:query", "", "", "", "", 1},
		{203, 201, 3, "route.work-order_list", "wo:order:add", "", "", "", "", 2},
		{204, 201, 3, "route.work-order_list", "wo:order:edit", "", "", "", "", 3},
		{205, 201, 3, "route.work-order_list", "wo:order:delete", "", "", "", "", 4},
		{300, 0, 1, "route.lot", "lot", "lot", "/lot", "layout.base", "mdi:package-variant", 5},
		{301, 300, 2, "route.lot_list", "lot:lot", "lot_list", "/lot/list", "view.lot_list", "mdi:format-list-bulleted", 1},
		{302, 300, 2, "route.lot_detail", "lot:lot:query", "lot_detail", "/lot/detail/:id", "view.lot_detail", "mdi:package-variant-closed", 2},
		{303, 301, 3, "route.lot_list", "lot:lot:query", "", "", "", "", 1},
		{304, 301, 3, "route.lot_list", "lot:lot:add", "", "", "", "", 2},
		{305, 301, 3, "route.lot_list", "lot:lot:edit", "", "", "", "", 3},
		{400, 0, 1, "route.wip", "wip", "wip", "/wip", "layout.base", "mdi:transit-connection-variant", 6},
		{401, 400, 2, "route.wip_move", "wip:move:query", "wip_move", "/wip/move", "view.wip_move", "mdi:transfer", 3},
		{410, 400, 2, "route.wip_station", "wip:move:query", "wip_station", "/wip/station", "view.wip_station", "mdi:barcode-scan", 1},
		{411, 400, 2, "route.wip_overview", "wip:move:query", "wip_overview", "/wip/overview", "view.wip_overview", "mdi:view-dashboard", 2},
		{413, 401, 3, "route.wip_move", "wip:move:track", "", "", "", "", 1},
		{500, 0, 1, "route.equipment", "equipment", "equipment", "/equipment", "layout.base", "mdi:wrench", 7},
		{501, 500, 2, "route.equipment_list", "eqp:equipment:query", "equipment_list", "/equipment/list", "view.equipment_list", "mdi:format-list-bulleted", 1},
		{502, 500, 2, "route.equipment_detail", "eqp:equipment:query", "equipment_detail", "/equipment/detail/:id", "view.equipment_detail", "mdi:information-outline", 2},
		{503, 501, 3, "route.equipment_list", "eqp:equipment:add", "", "", "", "", 1},
		{504, 501, 3, "route.equipment_list", "eqp:equipment:edit", "", "", "", "", 2},
		{505, 501, 3, "route.equipment_list", "eqp:equipment:delete", "", "", "", "", 3},
		{510, 500, 2, "route.equipment_board", "eqp:equipment:query", "equipment_board", "/equipment/board", "view.equipment_board", "mdi:view-dashboard-variant", 3},
		{511, 500, 2, "route.equipment_pm-plan", "eqp:pm:query", "equipment_pm-plan", "/equipment/pm-plan", "view.equipment_pm-plan", "mdi:calendar-check", 4},
		{512, 500, 2, "route.equipment_pm-task", "eqp:pm:query", "equipment_pm-task", "/equipment/pm-task", "view.equipment_pm-task", "mdi:clipboard-list-outline", 5},
		{513, 511, 3, "route.equipment_pm-plan", "eqp:pm:add", "", "", "", "", 1},
		{514, 511, 3, "route.equipment_pm-plan", "eqp:pm:edit", "", "", "", "", 2},
		{515, 511, 3, "route.equipment_pm-plan", "eqp:pm:delete", "", "", "", "", 3},
		{600, 0, 1, "route.quality", "quality", "quality", "/quality", "layout.base", "mdi:clipboard-check", 8},
		{601, 600, 2, "route.quality_inspection", "qc:inspection:query", "quality_inspection", "/quality/inspection", "view.quality_inspection", "mdi:magnify", 1},
	}
	menus := make([]model.SysMenu, 0, len(raw))
	for _, item := range raw {
		menus = append(menus, model.SysMenu{
			ID: item.ID, ParentID: item.ParentID, MenuType: item.MenuType, MenuName: item.MenuName,
			PermissionCode: item.PermissionCode, RouteName: item.RouteName, RoutePath: item.RoutePath,
			ComponentPath: item.ComponentPath, Icon: item.Icon, SortOrder: item.SortOrder, Status: 1,
		})
	}
	return menus
}
