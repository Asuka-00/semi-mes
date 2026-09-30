package handler

import (
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseDataHandler struct {
	db *gorm.DB
}

func NewBaseDataHandler(db *gorm.DB) *BaseDataHandler {
	return &BaseDataHandler{db: db}
}

type FactoryRequest struct {
	FactoryCode string `json:"factory_code" binding:"required"`
	FactoryName string `json:"factory_name" binding:"required"`
	Address     string `json:"address"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
	Status      int8   `json:"status"`
}

func (h *BaseDataHandler) ListFactories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := h.db.Model(&model.Factory{})
	if keyword != "" {
		query = query.Where("factory_code LIKE ? OR factory_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var factories []model.Factory
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&factories).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, factories, total, page, pageSize)
}

func (h *BaseDataHandler) CreateFactory(c *gin.Context) {
	var req FactoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	factory := model.Factory{
		FactoryCode: req.FactoryCode,
		FactoryName: req.FactoryName,
		Address:     req.Address,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Status:      req.Status,
	}

	if err := h.db.Create(&factory).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, factory)
}

func (h *BaseDataHandler) UpdateFactory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req FactoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var factory model.Factory
	if err := h.db.Where("id = ?", id).First(&factory).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"factory_name": req.FactoryName,
		"address":      req.Address,
		"contact":      req.Contact,
		"phone":        req.Phone,
		"status":       req.Status,
	}

	if err := h.db.Model(&factory).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteFactory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Factory{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type WorkshopRequest struct {
	FactoryID    int64  `json:"factory_id" binding:"required"`
	WorkshopCode string `json:"workshop_code" binding:"required"`
	WorkshopName string `json:"workshop_name" binding:"required"`
	WorkshopType string `json:"workshop_type"`
	Description  string `json:"description"`
	Status       int8   `json:"status"`
}

func (h *BaseDataHandler) ListWorkshops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	factoryID := c.Query("factory_id")

	query := h.db.Model(&model.Workshop{})
	if keyword != "" {
		query = query.Where("workshop_code LIKE ? OR workshop_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if factoryID != "" {
		query = query.Where("factory_id = ?", factoryID)
	}

	var total int64
	query.Count(&total)

	var workshops []model.Workshop
	offset := (page - 1) * pageSize
	if err := query.Preload("Factory").Offset(offset).Limit(pageSize).Find(&workshops).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, workshops, total, page, pageSize)
}

func (h *BaseDataHandler) CreateWorkshop(c *gin.Context) {
	var req WorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	workshop := model.Workshop{
		FactoryID:    req.FactoryID,
		WorkshopCode: req.WorkshopCode,
		WorkshopName: req.WorkshopName,
		WorkshopType: req.WorkshopType,
		Description:  req.Description,
		Status:       req.Status,
	}

	if err := h.db.Create(&workshop).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, workshop)
}

func (h *BaseDataHandler) UpdateWorkshop(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req WorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var workshop model.Workshop
	if err := h.db.Where("id = ?", id).First(&workshop).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"factory_id":    req.FactoryID,
		"workshop_name": req.WorkshopName,
		"workshop_type": req.WorkshopType,
		"description":   req.Description,
		"status":        req.Status,
	}

	if err := h.db.Model(&workshop).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteWorkshop(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Workshop{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type ProductionLineRequest struct {
	WorkshopID  int64  `json:"workshop_id" binding:"required"`
	LineCode    string `json:"line_code" binding:"required"`
	LineName    string `json:"line_name" binding:"required"`
	Capacity    int    `json:"capacity"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (h *BaseDataHandler) ListProductionLines(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	workshopID := c.Query("workshop_id")

	query := h.db.Model(&model.ProductionLine{})
	if keyword != "" {
		query = query.Where("line_code LIKE ? OR line_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if workshopID != "" {
		query = query.Where("workshop_id = ?", workshopID)
	}

	var total int64
	query.Count(&total)

	var lines []model.ProductionLine
	offset := (page - 1) * pageSize
	if err := query.Preload("Workshop.Factory").Offset(offset).Limit(pageSize).Find(&lines).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, lines, total, page, pageSize)
}

func (h *BaseDataHandler) CreateProductionLine(c *gin.Context) {
	var req ProductionLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	line := model.ProductionLine{
		WorkshopID:  req.WorkshopID,
		LineCode:    req.LineCode,
		LineName:    req.LineName,
		Capacity:    req.Capacity,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&line).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, line)
}

func (h *BaseDataHandler) UpdateProductionLine(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req ProductionLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var line model.ProductionLine
	if err := h.db.Where("id = ?", id).First(&line).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"workshop_id": req.WorkshopID,
		"line_name":   req.LineName,
		"capacity":    req.Capacity,
		"description": req.Description,
		"status":      req.Status,
	}

	if err := h.db.Model(&line).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteProductionLine(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.ProductionLine{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type ProductRequest struct {
	ProductCode string `json:"product_code" binding:"required"`
	ProductName string `json:"product_name" binding:"required"`
	ProductType string `json:"product_type"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (h *BaseDataHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := h.db.Model(&model.Product{})
	if keyword != "" {
		query = query.Where("product_code LIKE ? OR product_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var products []model.Product
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, products, total, page, pageSize)
}

func (h *BaseDataHandler) CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	product := model.Product{
		ProductCode: req.ProductCode,
		ProductName: req.ProductName,
		ProductType: req.ProductType,
		Version:     req.Version,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&product).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, product)
}

func (h *BaseDataHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var product model.Product
	if err := h.db.Where("id = ?", id).First(&product).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"product_name": req.ProductName,
		"product_type": req.ProductType,
		"version":      req.Version,
		"description":  req.Description,
		"status":       req.Status,
	}

	if err := h.db.Model(&product).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Product{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type ProcessRouteRequest struct {
	ProductID   int64  `json:"product_id" binding:"required"`
	RouteCode   string `json:"route_code" binding:"required"`
	RouteName   string `json:"route_name" binding:"required"`
	Version     string `json:"version"`
	IsDefault   int8   `json:"is_default"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (h *BaseDataHandler) ListProcessRoutes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	productID := c.Query("product_id")

	query := h.db.Model(&model.ProcessRoute{})
	if keyword != "" {
		query = query.Where("route_code LIKE ? OR route_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	var total int64
	query.Count(&total)

	var routes []model.ProcessRoute
	offset := (page - 1) * pageSize
	if err := query.Preload("Product").Offset(offset).Limit(pageSize).Find(&routes).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, routes, total, page, pageSize)
}

func (h *BaseDataHandler) CreateProcessRoute(c *gin.Context) {
	var req ProcessRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	route := model.ProcessRoute{
		ProductID:   req.ProductID,
		RouteCode:   req.RouteCode,
		RouteName:   req.RouteName,
		Version:     req.Version,
		IsDefault:   req.IsDefault,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&route).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, route)
}

func (h *BaseDataHandler) UpdateProcessRoute(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req ProcessRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var route model.ProcessRoute
	if err := h.db.Where("id = ?", id).First(&route).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"product_id":  req.ProductID,
		"route_name":  req.RouteName,
		"version":     req.Version,
		"is_default":  req.IsDefault,
		"description": req.Description,
		"status":      req.Status,
	}

	if err := h.db.Model(&route).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteProcessRoute(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.ProcessRoute{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type OperationRequest struct {
	RouteID       int64  `json:"route_id" binding:"required"`
	OperationCode string `json:"operation_code" binding:"required"`
	OperationName string `json:"operation_name" binding:"required"`
	OperationType string `json:"operation_type"`
	Sequence      int    `json:"sequence" binding:"required"`
	StandardTime  int    `json:"standard_time"`
	Description   string `json:"description"`
	Status        int8   `json:"status"`
}

func (h *BaseDataHandler) ListOperations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	routeID := c.Query("route_id")

	query := h.db.Model(&model.Operation{})
	if keyword != "" {
		query = query.Where("operation_code LIKE ? OR operation_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if routeID != "" {
		query = query.Where("route_id = ?", routeID)
	}

	var total int64
	query.Count(&total)

	var operations []model.Operation
	offset := (page - 1) * pageSize
	if err := query.Preload("Route.Product").Order("sequence ASC").Offset(offset).Limit(pageSize).Find(&operations).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, operations, total, page, pageSize)
}

func (h *BaseDataHandler) CreateOperation(c *gin.Context) {
	var req OperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	operation := model.Operation{
		RouteID:       req.RouteID,
		OperationCode: req.OperationCode,
		OperationName: req.OperationName,
		OperationType: req.OperationType,
		Sequence:      req.Sequence,
		StandardTime:  req.StandardTime,
		Description:   req.Description,
		Status:        req.Status,
	}

	if err := h.db.Create(&operation).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, operation)
}

func (h *BaseDataHandler) UpdateOperation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req OperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var operation model.Operation
	if err := h.db.Where("id = ?", id).First(&operation).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"route_id":       req.RouteID,
		"operation_name": req.OperationName,
		"operation_type": req.OperationType,
		"sequence":       req.Sequence,
		"standard_time":  req.StandardTime,
		"description":    req.Description,
		"status":         req.Status,
	}

	if err := h.db.Model(&operation).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteOperation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Operation{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

type RecipeRequest struct {
	OperationID int64  `json:"operation_id" binding:"required"`
	RecipeCode  string `json:"recipe_code" binding:"required"`
	RecipeName  string `json:"recipe_name" binding:"required"`
	Version     string `json:"version"`
	Parameters  string `json:"parameters"`
	IsDefault   int8   `json:"is_default"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (h *BaseDataHandler) ListRecipes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	operationID := c.Query("operation_id")

	query := h.db.Model(&model.Recipe{})
	if keyword != "" {
		query = query.Where("recipe_code LIKE ? OR recipe_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if operationID != "" {
		query = query.Where("operation_id = ?", operationID)
	}

	var total int64
	query.Count(&total)

	var recipes []model.Recipe
	offset := (page - 1) * pageSize
	if err := query.Preload("Operation").Offset(offset).Limit(pageSize).Find(&recipes).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, recipes, total, page, pageSize)
}

func (h *BaseDataHandler) CreateRecipe(c *gin.Context) {
	var req RecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	recipe := model.Recipe{
		OperationID: req.OperationID,
		RecipeCode:  req.RecipeCode,
		RecipeName:  req.RecipeName,
		Version:     req.Version,
		Parameters:  req.Parameters,
		IsDefault:   req.IsDefault,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&recipe).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, recipe)
}

func (h *BaseDataHandler) UpdateRecipe(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req RecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var recipe model.Recipe
	if err := h.db.Where("id = ?", id).First(&recipe).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"operation_id": req.OperationID,
		"recipe_name":  req.RecipeName,
		"version":      req.Version,
		"parameters":   req.Parameters,
		"is_default":   req.IsDefault,
		"description":  req.Description,
		"status":       req.Status,
	}

	if err := h.db.Model(&recipe).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *BaseDataHandler) DeleteRecipe(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Recipe{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}
