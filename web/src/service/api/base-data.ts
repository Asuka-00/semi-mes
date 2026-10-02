import { request } from '../request';

export namespace BaseData {
  export interface Factory {
    id?: number;
    factory_code: string;
    factory_name: string;
    address?: string;
    contact?: string;
    phone?: string;
    status: number;
  }

  export interface Workshop {
    id?: number;
    factory_id: number;
    workshop_code: string;
    workshop_name: string;
    workshop_type?: string;
    description?: string;
    status: number;
    factory?: Factory;
  }

  export interface ProductionLine {
    id?: number;
    workshop_id: number;
    line_code: string;
    line_name: string;
    capacity?: number;
    description?: string;
    status: number;
    workshop?: Workshop;
  }

  export interface Product {
    id?: number;
    product_code: string;
    product_name: string;
    product_type?: string;
    version?: string;
    description?: string;
    status: number;
  }

  export interface ProcessRoute {
    id?: number;
    product_id: number;
    route_code: string;
    route_name: string;
    version?: string;
    is_default: number;
    description?: string;
    status: number;
    product?: Product;
  }

  export interface Operation {
    id?: number;
    route_id: number;
    operation_code: string;
    operation_name: string;
    operation_type?: string;
    sequence: number;
    standard_time?: number;
    description?: string;
    status: number;
    route?: ProcessRoute;
  }

  export interface Recipe {
    id?: number;
    operation_id: number;
    recipe_code: string;
    recipe_name: string;
    version?: string;
    parameters?: string;
    is_default: number;
    description?: string;
    status: number;
    operation?: Operation;
  }

  export interface PageResult<T> {
    list: T[];
    total: number;
    page: number;
    page_size: number;
  }
}

// Factory API
export function fetchFactories(params: any) {
  return request<BaseData.PageResult<BaseData.Factory>>({
    url: '/base-data/factories',
    params
  });
}

export function createFactory(data: BaseData.Factory) {
  return request({
    url: '/base-data/factories',
    method: 'post',
    data
  });
}

export function updateFactory(id: number, data: BaseData.Factory) {
  return request({
    url: `/base-data/factories/${id}`,
    method: 'put',
    data
  });
}

export function deleteFactory(id: number) {
  return request({
    url: `/base-data/factories/${id}`,
    method: 'delete'
  });
}

// Workshop API
export function fetchWorkshops(params: any) {
  return request<BaseData.PageResult<BaseData.Workshop>>({
    url: '/base-data/workshops',
    params
  });
}

export function createWorkshop(data: BaseData.Workshop) {
  return request({
    url: '/base-data/workshops',
    method: 'post',
    data
  });
}

export function updateWorkshop(id: number, data: BaseData.Workshop) {
  return request({
    url: `/base-data/workshops/${id}`,
    method: 'put',
    data
  });
}

export function deleteWorkshop(id: number) {
  return request({
    url: `/base-data/workshops/${id}`,
    method: 'delete'
  });
}

// ProductionLine API
export function fetchProductionLines(params: any) {
  return request<BaseData.PageResult<BaseData.ProductionLine>>({
    url: '/base-data/production-lines',
    params
  });
}

export function createProductionLine(data: BaseData.ProductionLine) {
  return request({
    url: '/base-data/production-lines',
    method: 'post',
    data
  });
}

export function updateProductionLine(id: number, data: BaseData.ProductionLine) {
  return request({
    url: `/base-data/production-lines/${id}`,
    method: 'put',
    data
  });
}

export function deleteProductionLine(id: number) {
  return request({
    url: `/base-data/production-lines/${id}`,
    method: 'delete'
  });
}

// Product API
export function fetchProducts(params: any) {
  return request<BaseData.PageResult<BaseData.Product>>({
    url: '/base-data/products',
    params
  });
}

export function createProduct(data: BaseData.Product) {
  return request({
    url: '/base-data/products',
    method: 'post',
    data
  });
}

export function updateProduct(id: number, data: BaseData.Product) {
  return request({
    url: `/base-data/products/${id}`,
    method: 'put',
    data
  });
}

export function deleteProduct(id: number) {
  return request({
    url: `/base-data/products/${id}`,
    method: 'delete'
  });
}

// ProcessRoute API
export function fetchProcessRoutes(params: any) {
  return request<BaseData.PageResult<BaseData.ProcessRoute>>({
    url: '/base-data/process-routes',
    params
  });
}

export function createProcessRoute(data: BaseData.ProcessRoute) {
  return request({
    url: '/base-data/process-routes',
    method: 'post',
    data
  });
}

export function updateProcessRoute(id: number, data: BaseData.ProcessRoute) {
  return request({
    url: `/base-data/process-routes/${id}`,
    method: 'put',
    data
  });
}

export function deleteProcessRoute(id: number) {
  return request({
    url: `/base-data/process-routes/${id}`,
    method: 'delete'
  });
}

// Operation API
export function fetchOperations(params: any) {
  return request<BaseData.PageResult<BaseData.Operation>>({
    url: '/base-data/operations',
    params
  });
}

export function createOperation(data: BaseData.Operation) {
  return request({
    url: '/base-data/operations',
    method: 'post',
    data
  });
}

export function updateOperation(id: number, data: BaseData.Operation) {
  return request({
    url: `/base-data/operations/${id}`,
    method: 'put',
    data
  });
}

export function deleteOperation(id: number) {
  return request({
    url: `/base-data/operations/${id}`,
    method: 'delete'
  });
}

// Recipe API
export function fetchRecipes(params: any) {
  return request<BaseData.PageResult<BaseData.Recipe>>({
    url: '/base-data/recipes',
    params
  });
}

export function createRecipe(data: BaseData.Recipe) {
  return request({
    url: '/base-data/recipes',
    method: 'post',
    data
  });
}

export function updateRecipe(id: number, data: BaseData.Recipe) {
  return request({
    url: `/base-data/recipes/${id}`,
    method: 'put',
    data
  });
}

export function deleteRecipe(id: number) {
  return request({
    url: `/base-data/recipes/${id}`,
    method: 'delete'
  });
}
