import { request } from '../request';

export interface ReleasedRoute {
  versionID: number;
  routeID: number;
  versionNo: number;
  routeCode: string;
  routeName: string;
  productID: number;
  productCode: string;
  productName: string;
}

export interface WorkOrderRow {
  id: number;
  orderNo: string;
  productID: number;
  routeVersionID: number;
  plannedQty: number;
  releasedQty: number;
  completedQty: number;
  priority: number;
  dueDateText: string;
  status: string;
  note: string;
  productCode: string;
  productName: string;
  routeCode: string;
  routeName: string;
  versionNo: number;
}

export interface LotRow {
  id: number;
  lotNo: string;
  orderID: number;
  orderNo: string;
  quantity: number;
  lotType: string;
  status: string;
  currentNodeKey: string;
  nodeName: string;
  holdReasonCode: string;
  holdReason: string;
  priority: number;
  routeCode: string;
  versionNo: number;
}

export function fetchReleasedRoutes() {
  return request<{ versions: ReleasedRoute[] }>({ url: '/wipWorkOrder/releasedVersions' });
}

export function fetchWorkOrders(body: Record<string, unknown>) {
  return request<{ wipWorkOrders: WorkOrderRow[]; total: number }>({
    url: '/wipWorkOrder/list',
    method: 'post',
    data: body
  });
}

export function saveWorkOrder(data: Record<string, unknown>, id?: number) {
  if (id) {
    return request({ url: `/wipWorkOrder/${id}`, method: 'put', data });
  }
  return request<{ id: number }>({ url: '/wipWorkOrder', method: 'post', data });
}

export function removeWorkOrder(id: number) {
  return request({ url: `/wipWorkOrder/${id}`, method: 'delete' });
}

export function releaseWorkOrder(id: number) {
  return request({ url: `/wipWorkOrder/${id}/release`, method: 'post', data: {} });
}

export function closeWorkOrder(id: number) {
  return request({ url: `/wipWorkOrder/${id}/close`, method: 'post', data: {} });
}

export function startLot(orderId: number, quantity: number, lotType: string) {
  return request<{ id: number; lotNo: string }>({
    url: `/wipWorkOrder/${orderId}/start`,
    method: 'post',
    data: { quantity, lotType }
  });
}

export function fetchLots(body: Record<string, unknown>) {
  return request<{ wipLots: LotRow[]; total: number }>({ url: '/wipLot/list', method: 'post', data: body });
}

export function fetchLot(id: number) {
  return request<{
    lot: LotRow;
    history: Array<Record<string, any>>;
    links: Array<Record<string, any>>;
    moves: Array<Record<string, any>>;
    nodes: Array<Record<string, any>>;
    edges: Array<Record<string, any>>;
  }>({ url: `/wipLot/${id}` });
}

export function holdLot(id: number, reasonCode: string, reason: string) {
  return request({ url: `/wipLot/${id}/hold`, method: 'post', data: { reasonCode, reason } });
}

export function releaseHold(id: number, reasonCode: string, reason: string) {
  return request({ url: `/wipLot/${id}/releaseHold`, method: 'post', data: { reasonCode, reason } });
}

export function splitLot(id: number, quantities: number[]) {
  return request({ url: `/wipLot/${id}/split`, method: 'post', data: { quantities } });
}

export function mergeLots(targetId: number, sourceIds: number[]) {
  return request({ url: '/wipLot/merge', method: 'post', data: { targetId, sourceIds } });
}

export function advanceLot(id: number, inspectionResult: string, defectCode: string) {
  return request<{ lot: LotRow; result: { action: string; nextNodeKey: string; reason: string } }>({
    url: `/wipLot/${id}/advance`,
    method: 'post',
    data: { inspectionResult, defectCode }
  });
}
