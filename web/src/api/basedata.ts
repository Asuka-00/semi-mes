import request from '@/utils/request'

export interface Factory {
  id?: number
  factory_code: string
  factory_name: string
  address?: string
  contact?: string
  phone?: string
  status: number
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function getFactories(params: any) {
  return request.get<any, PageResult<Factory>>('/api/v1/base-data/factories', { params })
}

export function createFactory(data: Factory) {
  return request.post('/api/v1/base-data/factories', data)
}

export function updateFactory(id: number, data: Factory) {
  return request.put(`/api/v1/base-data/factories/${id}`, data)
}

export function deleteFactory(id: number) {
  return request.delete(`/api/v1/base-data/factories/${id}`)
}

export interface Workshop {
  id?: number
  factory_id: number
  workshop_code: string
  workshop_name: string
  workshop_type?: string
  description?: string
  status: number
  factory?: Factory
}

export function getWorkshops(params: any) {
  return request.get<any, PageResult<Workshop>>('/api/v1/base-data/workshops', { params })
}

export function createWorkshop(data: Workshop) {
  return request.post('/api/v1/base-data/workshops', data)
}

export function updateWorkshop(id: number, data: Workshop) {
  return request.put(`/api/v1/base-data/workshops/${id}`, data)
}

export function deleteWorkshop(id: number) {
  return request.delete(`/api/v1/base-data/workshops/${id}`)
}

export interface ProductionLine {
  id?: number
  workshop_id: number
  line_code: string
  line_name: string
  capacity?: number
  description?: string
  status: number
  workshop?: Workshop
}

export function getProductionLines(params: any) {
  return request.get<any, PageResult<ProductionLine>>('/api/v1/base-data/production-lines', { params })
}

export function createProductionLine(data: ProductionLine) {
  return request.post('/api/v1/base-data/production-lines', data)
}

export function updateProductionLine(id: number, data: ProductionLine) {
  return request.put(`/api/v1/base-data/production-lines/${id}`, data)
}

export function deleteProductionLine(id: number) {
  return request.delete(`/api/v1/base-data/production-lines/${id}`)
}

export interface Product {
  id?: number
  product_code: string
  product_name: string
  product_type?: string
  version?: string
  description?: string
  status: number
}

export function getProducts(params: any) {
  return request.get<any, PageResult<Product>>('/api/v1/base-data/products', { params })
}

export function createProduct(data: Product) {
  return request.post('/api/v1/base-data/products', data)
}

export function updateProduct(id: number, data: Product) {
  return request.put(`/api/v1/base-data/products/${id}`, data)
}

export function deleteProduct(id: number) {
  return request.delete(`/api/v1/base-data/products/${id}`)
}

export interface ProcessRoute {
  id?: number
  product_id: number
  route_code: string
  route_name: string
  version?: string
  is_default: number
  description?: string
  status: number
  product?: Product
}

export function getProcessRoutes(params: any) {
  return request.get<any, PageResult<ProcessRoute>>('/api/v1/base-data/process-routes', { params })
}

export function createProcessRoute(data: ProcessRoute) {
  return request.post('/api/v1/base-data/process-routes', data)
}

export function updateProcessRoute(id: number, data: ProcessRoute) {
  return request.put(`/api/v1/base-data/process-routes/${id}`, data)
}

export function deleteProcessRoute(id: number) {
  return request.delete(`/api/v1/base-data/process-routes/${id}`)
}

export interface Operation {
  id?: number
  route_id: number
  operation_code: string
  operation_name: string
  operation_type?: string
  sequence: number
  standard_time?: number
  description?: string
  status: number
  route?: ProcessRoute
}

export function getOperations(params: any) {
  return request.get<any, PageResult<Operation>>('/api/v1/base-data/operations', { params })
}

export function createOperation(data: Operation) {
  return request.post('/api/v1/base-data/operations', data)
}

export function updateOperation(id: number, data: Operation) {
  return request.put(`/api/v1/base-data/operations/${id}`, data)
}

export function deleteOperation(id: number) {
  return request.delete(`/api/v1/base-data/operations/${id}`)
}

export interface Recipe {
  id?: number
  operation_id: number
  recipe_code: string
  recipe_name: string
  version?: string
  parameters?: string
  is_default: number
  description?: string
  status: number
  operation?: Operation
}

export function getRecipes(params: any) {
  return request.get<any, PageResult<Recipe>>('/api/v1/base-data/recipes', { params })
}

export function createRecipe(data: Recipe) {
  return request.post('/api/v1/base-data/recipes', data)
}

export function updateRecipe(id: number, data: Recipe) {
  return request.put(`/api/v1/base-data/recipes/${id}`, data)
}

export function deleteRecipe(id: number) {
  return request.delete(`/api/v1/base-data/recipes/${id}`)
}
