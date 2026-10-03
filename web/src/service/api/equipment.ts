import { request } from '../request';

export interface EquipmentRow {
  id: number;
  equipmentCode: string;
  equipmentName: string;
  equipmentGroup: string;
  equipmentType: string;
  status: string;
  lineID: number;
  modelName: string;
  manufacturer: string;
  serialNo: string;
  location: string;
  chamberCount: number;
  capacity: number;
  installDate?: string;
  resumeState?: string;
}

export interface CapabilityRow {
  id: number;
  operationID: number;
  recipeID: number;
  operationCode?: string;
  operationName?: string;
  recipeCode?: string;
}

export function fetchEquipmentPage(body: Record<string, unknown>) {
  return request<{ eqpEquipments: EquipmentRow[]; total: number }>({
    url: '/eqpEquipment/list',
    method: 'post',
    data: body
  });
}

export function fetchEquipmentDetail(id: number) {
  return request<{
    equipment: EquipmentRow;
    capabilities: CapabilityRow[];
    stateLogs: Array<Record<string, any>>;
    pmTasks: Array<Record<string, any>>;
    recentLots: Array<Record<string, any>>;
    openLots: number;
  }>({ url: `/eqpEquipment/${id}` });
}

export function saveEquipment(data: Record<string, unknown>, id?: number) {
  if (id) return request({ url: `/eqpEquipment/${id}`, method: 'put', data });
  return request<{ id: number }>({ url: '/eqpEquipment', method: 'post', data });
}

export function removeEquipment(id: number) {
  return request({ url: `/eqpEquipment/${id}`, method: 'delete' });
}

export function changeEquipmentState(id: number, data: Record<string, unknown>) {
  return request({ url: `/eqpEquipment/${id}/state`, method: 'post', data });
}

export function fetchEquipmentBoard() {
  return request<{
    byState: Array<{ key: string; count: number }>;
    tools: Array<Record<string, any>>;
    overdue: number;
  }>({ url: '/eqpEquipment/board' });
}

export function fetchPmPlans(body?: Record<string, unknown>) {
  if (body) return request<{ plans: Array<Record<string, any>>; total: number }>({ url: '/eqpPmPlan/list', method: 'post', data: body });
  return request<{ plans: Array<Record<string, any>> }>({ url: '/eqpPmPlan' });
}

export function savePmPlan(data: Record<string, unknown>, id?: number) {
  if (id) return request({ url: `/eqpPmPlan/${id}`, method: 'put', data });
  return request<{ id: number }>({ url: '/eqpPmPlan', method: 'post', data });
}

export function removePmPlan(id: number) {
  return request({ url: `/eqpPmPlan/${id}`, method: 'delete' });
}

export function fetchPmTasks(body: Record<string, unknown>) {
  return request<{ tasks: Array<Record<string, any>>; total: number }>({
    url: '/eqpPmTask/list',
    method: 'post',
    data: body
  });
}

export function startPmTask(id: number) {
  return request({ url: `/eqpPmTask/${id}/start`, method: 'post', data: {} });
}

export function completePmTask(id: number, data: Record<string, unknown>) {
  return request({ url: `/eqpPmTask/${id}/complete`, method: 'post', data });
}
