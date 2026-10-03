package bootstrap

import (
	"fmt"
	"time"

	"semi-mes/server/internal/config"
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
	defer func() { _ = syncSequences(db) }()
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
		&model.QcInspectPlan{},
		&model.QcInspectItem{},
		&model.QcMeasurement{},
		&model.QcJudgement{},
		&model.QcDefectCode{},
		&model.QcDefect{},
		&model.QcSpcPolicy{},
		&model.QcSpcLimit{},
		&model.QcSpcEvent{},
		&model.SysNotification{},
		&model.SysUserPref{},
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
	_ = db.Model(&model.SysMenu{}).Where("id = ? AND permission_code = ?", 601, "qc:inspection:query").
		Update("permission_code", "qc:measure:query").Error
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
	status := 0
	if config.SPCEnabled() {
		status = 1
	}
	return db.Model(&model.SysMenu{}).Where("id IN ?", []uint64{630, 631}).Update("status", status).Error
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
		if err := seedQuality(db, productID); err != nil {
			return err
		}
		return seedDashboard(db, productID, versionID)
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
	if err := db.Create(&model.EqpStateLog{
		EquipmentID: wet.ID, FromState: model.EqpStandby, ToState: model.EqpProductive, ReasonCode: "PROD_START",
		OperatorID: 1, StartedAt: trackIn,
	}).Error; err != nil {
		return err
	}
	if err := seedQuality(db, productID); err != nil {
		return err
	}
	return seedDashboard(db, productID, versionID)
}

// seedDashboard adds lots, holds, and a week of track-outs so the home board is not empty.
// It is skipped once DEMO-DASH-DONE exists.
func seedDashboard(db *gorm.DB, productID, versionID uint64) error {
	var exists int64
	if err := db.Model(&model.WipLot{}).Where("lot_no = ?", "DEMO-DASH-DONE").Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	var order model.WipWorkOrder
	if err := db.Where("order_no = ?", "DEMO-WIP").First(&order).Error; err != nil {
		return nil
	}
	now := time.Now()
	held := now.Add(-8 * time.Hour)
	olderHold := now.Add(-3 * time.Hour)
	sensor := model.BaseProduct{ProductCode: "SENSOR-DEMO", ProductName: "Sensor Demo", ProductType: "IC", Version: "A", Status: 1}
	if err := db.Where(model.BaseProduct{ProductCode: "SENSOR-DEMO"}).FirstOrCreate(&sensor).Error; err != nil {
		return err
	}
	arrived := now.Add(-90 * time.Minute)
	lots := []model.WipLot{
		{LotNo: "DEMO-DASH-ETCH", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "etch", Quantity: 15, Priority: 3, LotType: model.LotTypeProduction, Status: model.LotWaiting, ReworkJSON: "{}", ArrivedAt: &arrived},
		{LotNo: "DEMO-DASH-HOLD", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "etch", Quantity: 8, Priority: 2, LotType: model.LotTypeProduction, Status: model.LotHold, HoldReasonCode: "YIELD", HoldReason: "良率偏低，等待复测", ReworkJSON: "{}", ArrivedAt: &held},
		{LotNo: "DEMO-SENSOR", OrderID: order.ID, ProductID: sensor.ID, RouteVersionID: versionID, CurrentNodeKey: "photo", Quantity: 9, Priority: 4, LotType: model.LotTypeEngineering, Status: model.LotWaiting, ReworkJSON: "{}", ArrivedAt: &arrived},
		{LotNo: "DEMO-DASH-DONE", OrderID: order.ID, ProductID: productID, RouteVersionID: versionID, CurrentNodeKey: "end_main", Quantity: 25, Priority: 3, LotType: model.LotTypeProduction, Status: model.LotCompleted, ReworkJSON: "{}", ArrivedAt: &held},
	}
	for i := range lots {
		if err := db.Create(&lots[i]).Error; err != nil {
			return err
		}
	}
	var holdLot, doneLot model.WipLot
	if err := db.Where("lot_no = ?", "DEMO-DASH-HOLD").First(&holdLot).Error; err != nil {
		return err
	}
	if err := db.Where("lot_no = ?", "DEMO-DASH-DONE").First(&doneLot).Error; err != nil {
		return err
	}
	if err := db.Create(&model.WipLotHistory{
		LotID: holdLot.ID, EventType: model.EventHold, FromNodeKey: "etch", ToNodeKey: "etch",
		ReasonCode: "YIELD", Reason: "良率偏低，等待复测", Quantity: 8, CreatedAt: &held,
	}).Error; err != nil {
		return err
	}
	var demoHold model.WipLot
	if err := db.Where("lot_no = ?", "DEMO-WIP-002").First(&demoHold).Error; err == nil {
		var n int64
		if err := db.Model(&model.WipLotHistory{}).Where("lot_id = ? AND event_type = ?", demoHold.ID, model.EventHold).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&model.WipLotHistory{
				LotID: demoHold.ID, EventType: model.EventHold, FromNodeKey: "inspect", ToNodeKey: "inspect",
				ReasonCode: demoHold.HoldReasonCode, Reason: demoHold.HoldReason, Quantity: demoHold.Quantity, CreatedAt: &olderHold,
			}).Error; err != nil {
				return err
			}
		}
	}
	doneAt := now.Add(-2 * time.Hour)
	if err := db.Create(&model.WipLotHistory{
		LotID: doneLot.ID, EventType: model.EventComplete, FromNodeKey: "etch", ToNodeKey: "end_main",
		Quantity: 25, CreatedAt: &doneAt,
	}).Error; err != nil {
		return err
	}
	if err := db.Create(&model.WipLotHistory{
		LotID: doneLot.ID, EventType: model.EventTrackOut, FromNodeKey: "etch", ToNodeKey: "end_main",
		Quantity: 25, CreatedAt: &doneAt,
	}).Error; err != nil {
		return err
	}
	var photo model.EqpEquipment
	_ = db.Where("equipment_code = ?", "PHOTO-01").First(&photo).Error
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	counts := []int{3, 5, 4, 7, 6, 8, 4}
	for i, n := range counts {
		at := start.AddDate(0, 0, -6+i).Add(12 * time.Hour)
		for j := 0; j < n; j++ {
			out := at.Add(time.Duration(j) * time.Minute)
			in := out.Add(-30 * time.Minute)
			scrap := 0
			if i == 5 && j == 0 {
				scrap = 1
			}
			if i == 6 && j == 0 {
				scrap = 2
			}
			move := model.WipMove{
				LotID: doneLot.ID, NodeKey: "photo", EquipmentID: photo.ID, OperatorID: 1,
				QtyIn: 10, QtyOut: 10 - scrap, QtyScrap: scrap, State: model.MoveCompleted,
				TrackInAt: &in, TrackOutAt: &out, FromNodeKey: "photo", ToNodeKey: "inspect",
			}
			if err := db.Create(&move).Error; err != nil {
				return err
			}
		}
	}
	return nil
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
	links := []struct{ eqp, op string }{
		{"WET-01", "CLEAN"},
		{"PHOTO-01", "PHOTO"},
		{"PHOTO-09", "PHOTO"},
		{"METRO-01", "INSPECT"},
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

func seedQuality(db *gorm.DB, productID uint64) error {
	var inspect model.BaseOperation
	if err := db.Where("operation_code = ?", "INSPECT").First(&inspect).Error; err != nil {
		return nil
	}
	var plans int64
	if err := db.Model(&model.QcInspectPlan{}).Count(&plans).Error; err != nil {
		return err
	}
	if plans == 0 {
		target, lsl, usl := 500.0, 470.0, 530.0
		if _, err := dao.SaveInspectPlan(db, 0, dao.PlanInput{
			OperationID: inspect.ID, PlanName: "线宽 CD", Enabled: true,
			Items: []dao.ItemInput{{
				ParamCode: "CD", ParamName: "关键尺寸", Unit: "nm",
				Target: &target, LSL: &lsl, USL: &usl, SampleSize: 1, Required: true,
			}},
		}); err != nil {
			return err
		}
	}
	for _, item := range []struct{ code, name, cat, sev string }{
		{"PARTICLE", "颗粒", "particle", "major"},
		{"SCRATCH", "划伤", "scratch", "critical"},
		{"PATTERN", "图形异常", "pattern", "minor"},
	} {
		var n int64
		if err := db.Model(&model.QcDefectCode{}).Where("defect_code = ?", item.code).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if _, err := dao.SaveDefectCode(db, 0, item.code, item.name, item.cat, item.sev, 1); err != nil {
				return err
			}
		}
	}
	var policies int64
	if err := db.Model(&model.QcSpcPolicy{}).Where("param_code = ? AND operation_id = ?", "CD", inspect.ID).Count(&policies).Error; err != nil {
		return err
	}
	if policies == 0 {
		if _, err := dao.SaveSpcPolicy(db, model.QcSpcPolicy{
			ParamCode: "CD", OperationID: inspect.ID, OnOOC: model.ReactHold, OnOOS: model.ReactHold,
		}); err != nil {
			return err
		}
	}
	center, lcl, ucl := 500.0, 480.0, 520.0
	if _, err := dao.SaveSpcLimit(db, model.QcSpcLimit{
		ParamCode: "CD", OperationID: inspect.ID, ChartType: "imr", Center: &center, LCL: &lcl, UCL: &ucl, UseManual: true,
	}); err != nil {
		return err
	}
	var measured int64
	if err := db.Model(&model.QcMeasurement{}).Count(&measured).Error; err != nil {
		return err
	}
	if measured > 0 {
		return nil
	}
	var order model.WipWorkOrder
	if err := db.Where("order_no = ?", "DEMO-WIP").First(&order).Error; err != nil {
		return nil
	}
	now := time.Now()
	lot := model.WipLot{
		LotNo: "DEMO-QC", OrderID: order.ID, ProductID: productID, RouteVersionID: order.RouteVersionID,
		CurrentNodeKey: "inspect", Quantity: 6, Priority: 3, LotType: model.LotTypeProduction,
		Status: model.LotWaiting, ReworkJSON: "{}", ArrivedAt: &now,
	}
	if err := db.Create(&lot).Error; err != nil {
		return err
	}
	var metro model.EqpEquipment
	_ = db.Where("equipment_code = ?", "METRO-01").First(&metro).Error
	values := []float64{498, 502, 501, 499, 503, 497, 500, 504, 496, 501, 499, 620}
	base := now.Add(-time.Duration(len(values)) * time.Hour)
	for i, value := range values {
		if _, err := dao.RecordMeasurements(db, dao.MeasureInput{
			LotID: lot.ID, EquipmentID: metro.ID, OperationID: inspect.ID, OperatorID: 1,
			Samples: []dao.SampleInput{{ParamCode: "CD", Values: []float64{value}}},
		}); err != nil {
			return err
		}
		var last model.QcMeasurement
		if err := db.Where("lot_id = ?", lot.ID).Order("id desc").First(&last).Error; err == nil {
			_ = db.Model(&last).Update("measured_at", base.Add(time.Duration(i)*time.Hour)).Error
		}
	}
	var demo model.WipLot
	if err := db.Where("lot_no = ?", "DEMO-WIP-001").First(&demo).Error; err != nil {
		return nil
	}
	for _, item := range []struct {
		code string
		qty  int
	}{{"PARTICLE", 12}, {"SCRATCH", 5}, {"PATTERN", 2}} {
		if _, err := dao.RecordDefect(db, dao.DefectInput{
			LotID: demo.ID, EquipmentID: metro.ID, OperationID: inspect.ID, NodeKey: "photo",
			DefectCode: item.code, Quantity: item.qty, Disposition: model.DispositionUseAsIs, Note: "演示", OperatorID: 1,
		}); err != nil {
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

func isQualityMenu(menu model.SysMenu) bool {
	if menu.RouteName == "quality" || hasPrefix(menu.RouteName, "quality_") {
		return true
	}
	return hasPrefix(menu.PermissionCode, "qc:")
}

func isModuleMenu(menu model.SysMenu) bool {
	return isBaseMenu(menu) || isWipMenu(menu) || isQualityMenu(menu)
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

// syncSequences moves Postgres identity sequences past explicitly inserted ids.
// MySQL updates AUTO_INCREMENT on those inserts; Postgres does not.
func syncSequences(db *gorm.DB) error {
	driver := config.Get().Database.Driver
	if driver != "postgresql" && driver != "postgres" {
		return nil
	}
	var tables []string
	if err := db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`).Scan(&tables).Error; err != nil {
		return err
	}
	for _, table := range tables {
		err := db.Exec(`DO $$ BEGIN
			IF pg_get_serial_sequence('` + table + `', 'id') IS NOT NULL THEN
				PERFORM setval(pg_get_serial_sequence('` + table + `', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM ` + table + `), 1), 1));
			END IF;
		END $$;`).Error
		if err != nil {
			return err
		}
	}
	return nil
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
		{601, 600, 2, "route.quality_inspection", "qc:measure:query", "quality_inspection", "/quality/inspection", "view.quality_inspection", "mdi:magnify", 1},
		{602, 601, 3, "route.quality_inspection", "qc:measure:add", "", "", "", "", 1},
		{603, 600, 2, "route.quality_plan", "qc:plan:query", "quality_plan", "/quality/plan", "view.quality_plan", "mdi:clipboard-list", 2},
		{604, 603, 3, "route.quality_plan", "qc:plan:add", "", "", "", "", 1},
		{605, 603, 3, "route.quality_plan", "qc:plan:edit", "", "", "", "", 2},
		{606, 603, 3, "route.quality_plan", "qc:plan:delete", "", "", "", "", 3},
		{610, 600, 2, "route.quality_defect", "qc:defect:query", "quality_defect", "/quality/defect", "view.quality_defect", "mdi:alert-circle-outline", 3},
		{611, 610, 3, "route.quality_defect", "qc:defect:add", "", "", "", "", 1},
		{620, 600, 2, "route.quality_pareto", "qc:defect:query", "quality_pareto", "/quality/pareto", "view.quality_pareto", "mdi:chart-bar", 4},
		{630, 600, 2, "route.quality_spc", "qc:spc:query", "quality_spc", "/quality/spc", "view.quality_spc", "mdi:chart-bell-curve", 5},
		{631, 630, 3, "route.quality_spc", "qc:spc:edit", "", "", "", "", 1},
	}
	menus := make([]model.SysMenu, 0, len(raw))
	for _, item := range raw {
		if !config.SPCEnabled() && (item.ID == 630 || item.ID == 631) {
			continue
		}
		menus = append(menus, model.SysMenu{
			ID: item.ID, ParentID: item.ParentID, MenuType: item.MenuType, MenuName: item.MenuName,
			PermissionCode: item.PermissionCode, RouteName: item.RouteName, RoutePath: item.RoutePath,
			ComponentPath: item.ComponentPath, Icon: item.Icon, SortOrder: item.SortOrder, Status: 1,
		})
	}
	return menus
}
