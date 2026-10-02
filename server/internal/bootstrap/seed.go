package bootstrap

import (
	"fmt"

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
		return nil
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
		{401, 400, 2, "route.wip_move", "wip:move:query", "wip_move", "/wip/move", "view.wip_move", "mdi:transfer", 1},
		{500, 0, 1, "route.equipment", "equipment", "equipment", "/equipment", "layout.base", "mdi:wrench", 7},
		{501, 500, 2, "route.equipment_list", "eqp:equipment:query", "equipment_list", "/equipment/list", "view.equipment_list", "mdi:format-list-bulleted", 1},
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
