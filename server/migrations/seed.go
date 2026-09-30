package migrations

import (
	"log"
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.UserRole{},
		&model.RoleMenu{},
		&model.Factory{},
		&model.Workshop{},
		&model.ProductionLine{},
		&model.Product{},
		&model.ProcessRoute{},
		&model.Operation{},
		&model.Recipe{},
	)
}

func SeedData(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedUsers(db); err != nil {
		return err
	}
	if err := seedMenus(db); err != nil {
		return err
	}
	if err := seedRoleMenus(db); err != nil {
		return err
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	var count int64
	db.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return nil
	}

	roles := []model.Role{
		{RoleCode: "super_admin", RoleName: "Super Administrator", Description: "System super administrator with all permissions", Status: 1},
		{RoleCode: "sys_admin", RoleName: "System Administrator", Description: "System administrator", Status: 1},
		{RoleCode: "operator", RoleName: "Operator", Description: "Production operator", Status: 1},
		{RoleCode: "viewer", RoleName: "Viewer", Description: "Read-only user", Status: 1},
	}

	return db.Create(&roles).Error
}

func seedUsers(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}

	admin := model.User{
		Username: "admin",
		Password: hashedPassword,
		RealName: "Administrator",
		Email:    "admin@example.com",
		Phone:    "13800138000",
		Status:   1,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	var superAdminRole model.Role
	if err := db.Where("role_code = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		return err
	}

	return db.Create(&model.UserRole{
		UserID: admin.ID,
		RoleID: superAdminRole.ID,
	}).Error
}

func seedMenus(db *gorm.DB) error {
	var count int64
	db.Model(&model.Menu{}).Count(&count)
	if count > 0 {
		return nil
	}

	menus := []model.Menu{
		{ID: 1, ParentID: 0, MenuType: 1, MenuName: "menu.dashboard", PermissionCode: "dashboard", RoutePath: "/dashboard", ComponentPath: "dashboard", Icon: "mdi:monitor-dashboard", SortOrder: 1, Status: 1},
		
		{ID: 2, ParentID: 0, MenuType: 1, MenuName: "menu.system", PermissionCode: "system", RoutePath: "/system", ComponentPath: "", Icon: "mdi:cog", SortOrder: 2, Status: 1},
		{ID: 3, ParentID: 2, MenuType: 2, MenuName: "menu.system.user", PermissionCode: "system:user", RoutePath: "/system/user", ComponentPath: "system/user", Icon: "mdi:account", SortOrder: 1, Status: 1},
		{ID: 4, ParentID: 3, MenuType: 3, MenuName: "button.query", PermissionCode: "system:user:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 5, ParentID: 3, MenuType: 3, MenuName: "button.add", PermissionCode: "system:user:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 6, ParentID: 3, MenuType: 3, MenuName: "button.edit", PermissionCode: "system:user:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 7, ParentID: 3, MenuType: 3, MenuName: "button.delete", PermissionCode: "system:user:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 8, ParentID: 2, MenuType: 2, MenuName: "menu.system.role", PermissionCode: "system:role", RoutePath: "/system/role", ComponentPath: "system/role", Icon: "mdi:account-group", SortOrder: 2, Status: 1},
		{ID: 9, ParentID: 8, MenuType: 3, MenuName: "button.query", PermissionCode: "system:role:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 10, ParentID: 8, MenuType: 3, MenuName: "button.add", PermissionCode: "system:role:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 11, ParentID: 8, MenuType: 3, MenuName: "button.edit", PermissionCode: "system:role:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 12, ParentID: 8, MenuType: 3, MenuName: "button.delete", PermissionCode: "system:role:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 13, ParentID: 2, MenuType: 2, MenuName: "menu.system.menu", PermissionCode: "system:menu", RoutePath: "/system/menu", ComponentPath: "system/menu", Icon: "mdi:menu", SortOrder: 3, Status: 1},
		{ID: 14, ParentID: 13, MenuType: 3, MenuName: "button.query", PermissionCode: "system:menu:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 15, ParentID: 13, MenuType: 3, MenuName: "button.add", PermissionCode: "system:menu:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 16, ParentID: 13, MenuType: 3, MenuName: "button.edit", PermissionCode: "system:menu:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 17, ParentID: 13, MenuType: 3, MenuName: "button.delete", PermissionCode: "system:menu:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},

		{ID: 100, ParentID: 0, MenuType: 1, MenuName: "menu.baseData", PermissionCode: "baseData", RoutePath: "/base-data", ComponentPath: "", Icon: "mdi:database", SortOrder: 3, Status: 1},
		{ID: 101, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.factory", PermissionCode: "base:factory", RoutePath: "/base-data/factory", ComponentPath: "base-data/factory", Icon: "mdi:factory", SortOrder: 1, Status: 1},
		{ID: 102, ParentID: 101, MenuType: 3, MenuName: "button.query", PermissionCode: "base:factory:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 103, ParentID: 101, MenuType: 3, MenuName: "button.add", PermissionCode: "base:factory:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 104, ParentID: 101, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:factory:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 105, ParentID: 101, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:factory:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 106, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.workshop", PermissionCode: "base:workshop", RoutePath: "/base-data/workshop", ComponentPath: "base-data/workshop", Icon: "mdi:warehouse", SortOrder: 2, Status: 1},
		{ID: 107, ParentID: 106, MenuType: 3, MenuName: "button.query", PermissionCode: "base:workshop:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 108, ParentID: 106, MenuType: 3, MenuName: "button.add", PermissionCode: "base:workshop:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 109, ParentID: 106, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:workshop:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 110, ParentID: 106, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:workshop:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 111, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.line", PermissionCode: "base:line", RoutePath: "/base-data/line", ComponentPath: "base-data/line", Icon: "mdi:pipe", SortOrder: 3, Status: 1},
		{ID: 112, ParentID: 111, MenuType: 3, MenuName: "button.query", PermissionCode: "base:line:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 113, ParentID: 111, MenuType: 3, MenuName: "button.add", PermissionCode: "base:line:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 114, ParentID: 111, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:line:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 115, ParentID: 111, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:line:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 116, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.product", PermissionCode: "base:product", RoutePath: "/base-data/product", ComponentPath: "base-data/product", Icon: "mdi:chip", SortOrder: 4, Status: 1},
		{ID: 117, ParentID: 116, MenuType: 3, MenuName: "button.query", PermissionCode: "base:product:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 118, ParentID: 116, MenuType: 3, MenuName: "button.add", PermissionCode: "base:product:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 119, ParentID: 116, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:product:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 120, ParentID: 116, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:product:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 121, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.route", PermissionCode: "base:route", RoutePath: "/base-data/route", ComponentPath: "base-data/route", Icon: "mdi:map-marker-path", SortOrder: 5, Status: 1},
		{ID: 122, ParentID: 121, MenuType: 3, MenuName: "button.query", PermissionCode: "base:route:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 123, ParentID: 121, MenuType: 3, MenuName: "button.add", PermissionCode: "base:route:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 124, ParentID: 121, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:route:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 125, ParentID: 121, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:route:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 126, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.operation", PermissionCode: "base:operation", RoutePath: "/base-data/operation", ComponentPath: "base-data/operation", Icon: "mdi:cog-outline", SortOrder: 6, Status: 1},
		{ID: 127, ParentID: 126, MenuType: 3, MenuName: "button.query", PermissionCode: "base:operation:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 128, ParentID: 126, MenuType: 3, MenuName: "button.add", PermissionCode: "base:operation:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 129, ParentID: 126, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:operation:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 130, ParentID: 126, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:operation:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},
		{ID: 131, ParentID: 100, MenuType: 2, MenuName: "menu.baseData.recipe", PermissionCode: "base:recipe", RoutePath: "/base-data/recipe", ComponentPath: "base-data/recipe", Icon: "mdi:file-document-outline", SortOrder: 7, Status: 1},
		{ID: 132, ParentID: 131, MenuType: 3, MenuName: "button.query", PermissionCode: "base:recipe:query", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 1, Status: 1},
		{ID: 133, ParentID: 131, MenuType: 3, MenuName: "button.add", PermissionCode: "base:recipe:add", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 2, Status: 1},
		{ID: 134, ParentID: 131, MenuType: 3, MenuName: "button.edit", PermissionCode: "base:recipe:edit", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 3, Status: 1},
		{ID: 135, ParentID: 131, MenuType: 3, MenuName: "button.delete", PermissionCode: "base:recipe:delete", RoutePath: "", ComponentPath: "", Icon: "", SortOrder: 4, Status: 1},

		{ID: 200, ParentID: 0, MenuType: 1, MenuName: "menu.workOrder", PermissionCode: "workOrder", RoutePath: "/work-order", ComponentPath: "", Icon: "mdi:clipboard-text", SortOrder: 4, Status: 1},
		{ID: 201, ParentID: 200, MenuType: 2, MenuName: "menu.workOrder.list", PermissionCode: "wo:order:query", RoutePath: "/work-order/list", ComponentPath: "work-order/list", Icon: "mdi:format-list-bulleted", SortOrder: 1, Status: 1},

		{ID: 300, ParentID: 0, MenuType: 1, MenuName: "menu.lot", PermissionCode: "lot", RoutePath: "/lot", ComponentPath: "", Icon: "mdi:package-variant", SortOrder: 5, Status: 1},
		{ID: 301, ParentID: 300, MenuType: 2, MenuName: "menu.lot.list", PermissionCode: "lot:lot:query", RoutePath: "/lot/list", ComponentPath: "lot/list", Icon: "mdi:format-list-bulleted", SortOrder: 1, Status: 1},

		{ID: 400, ParentID: 0, MenuType: 1, MenuName: "menu.wip", PermissionCode: "wip", RoutePath: "/wip", ComponentPath: "", Icon: "mdi:transit-connection-variant", SortOrder: 6, Status: 1},
		{ID: 401, ParentID: 400, MenuType: 2, MenuName: "menu.wip.move", PermissionCode: "wip:move:query", RoutePath: "/wip/move", ComponentPath: "wip/move", Icon: "mdi:transfer", SortOrder: 1, Status: 1},

		{ID: 500, ParentID: 0, MenuType: 1, MenuName: "menu.equipment", PermissionCode: "equipment", RoutePath: "/equipment", ComponentPath: "", Icon: "mdi:wrench", SortOrder: 7, Status: 1},
		{ID: 501, ParentID: 500, MenuType: 2, MenuName: "menu.equipment.list", PermissionCode: "eqp:equipment:query", RoutePath: "/equipment/list", ComponentPath: "equipment/list", Icon: "mdi:format-list-bulleted", SortOrder: 1, Status: 1},

		{ID: 600, ParentID: 0, MenuType: 1, MenuName: "menu.quality", PermissionCode: "quality", RoutePath: "/quality", ComponentPath: "", Icon: "mdi:clipboard-check", SortOrder: 8, Status: 1},
		{ID: 601, ParentID: 600, MenuType: 2, MenuName: "menu.quality.inspection", PermissionCode: "qc:inspection:query", RoutePath: "/quality/inspection", ComponentPath: "quality/inspection", Icon: "mdi:magnify", SortOrder: 1, Status: 1},
	}

	result := db.Create(&menus)
	if result.Error != nil {
		log.Printf("Error seeding menus: %v", result.Error)
		return result.Error
	}

	return nil
}

func seedRoleMenus(db *gorm.DB) error {
	var count int64
	db.Model(&model.RoleMenu{}).Count(&count)
	if count > 0 {
		return nil
	}

	var superAdminRole model.Role
	if err := db.Where("role_code = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		return err
	}

	var allMenus []model.Menu
	if err := db.Find(&allMenus).Error; err != nil {
		return err
	}

	roleMenus := make([]model.RoleMenu, len(allMenus))
	for i, menu := range allMenus {
		roleMenus[i] = model.RoleMenu{
			RoleID: superAdminRole.ID,
			MenuID: menu.ID,
		}
	}

	return db.Create(&roleMenus).Error
}
