import { request } from '../request';

export interface EquipmentRow {
  id: number;
  equipmentCode: string;
  equipmentName: string;
  equipmentGroup: string;
  status: string;
}

export interface InspectItemView {
  paramCode: string;
  paramName: string;
  unit: string;
  target?: number | null;
  lsl?: number | null;
  usl?: number | null;
  sampleSize: number;
  required: boolean;
}

export interface StationView {
  lot: Record<string, any>;
  nodeType: string;
  nodeName: string;
  equipmentGroup: string;
  operationID: number;
  recipeID: number;
  inspectionRequired: boolean;
  inspectPlan?: {
    id: number;
    operationID: number;
    planName: string;
    operationCode?: string;
    items: InspectItemView[];
  } | null;
  latestResult?: string;
  openMove: Record<string, any> | null;
  equipment: EquipmentRow[];
  carrierNo?: string;
}

export function fetchStation(lotNo: string) {
  return request<StationView>({ url: '/wipMove/station', params: { lotNo } });
}

export function trackIn(lotId: number, equipmentId: number, carrierNo = '') {
  return request({ url: '/wipMove/trackIn', method: 'post', data: { lotId, equipmentId, carrierNo } });
}

export function trackOut(data: Record<string, unknown>) {
  return request<{ lot: Record<string, any>; result: { action: string; nextNodeKey?: string; reason: string } }>({
    url: '/wipMove/trackOut',
    method: 'post',
    data
  });
}

export function abortTrack(lotId: number, reason: string) {
  return request({ url: '/wipMove/abort', method: 'post', data: { lotId, reason } });
}

export function passNode(data: Record<string, unknown>) {
  return request<{ lot: Record<string, any>; result: { action: string; nextNodeKey?: string; reason: string } }>({
    url: '/wipMove/pass',
    method: 'post',
    data
  });
}

export function fetchMoves(body: Record<string, unknown>) {
  return request<{ wipMoves: Array<Record<string, any>>; total: number }>({
    url: '/wipMove/list',
    method: 'post',
    data: body
  });
}

export function fetchOverview() {
  return request<{
    byStatus: Array<{ key: string; count: number; qty: number }>;
    byNode: Array<{ key: string; count: number; qty: number }>;
    byProduct: Array<{ key: string; count: number; qty: number }>;
    holds: Array<Record<string, any>>;
  }>({ url: '/wipMove/overview' });
}

export function fetchEquipment() {
  return request<{ equipment: EquipmentRow[] }>({ url: '/eqpEquipment' });
}
