import { request } from '../request';

export interface InspectItem {
  id?: number;
  paramCode: string;
  paramName: string;
  unit: string;
  target?: number | null;
  lsl?: number | null;
  usl?: number | null;
  lcl?: number | null;
  ucl?: number | null;
  sampleSize: number;
  required: boolean;
}

export interface InspectPlan {
  id: number;
  operationID: number;
  productID: number;
  planName: string;
  enabled: boolean;
  operationCode?: string;
  items: InspectItem[];
}

export interface ChartPoint {
  index: number;
  value: number;
  at: string;
  lotID: number;
  violations: number[] | null;
  oos: boolean;
}

export interface ChartView {
  paramCode: string;
  chartType: string;
  center: number;
  lcl: number;
  ucl: number;
  points: ChartPoint[];
  events: Array<Record<string, any>>;
}

export function fetchPlans(body?: Record<string, unknown>) {
  if (body) return request<{ plans: InspectPlan[]; total: number }>({ url: '/qcInspectPlan/list', method: 'post', data: body });
  return request<{ plans: InspectPlan[] }>({ url: '/qcInspectPlan' });
}

export function savePlan(data: Record<string, unknown>, id?: number) {
  return request({ url: id ? `/qcInspectPlan/${id}` : '/qcInspectPlan', method: id ? 'put' : 'post', data });
}

export function removePlan(id: number) {
  return request({ url: `/qcInspectPlan/${id}`, method: 'delete' });
}

export function recordMeasurement(data: Record<string, unknown>) {
  return request<{ result: string }>({ url: '/qcMeasurement', method: 'post', data });
}

export function fetchDefectCodes(body?: Record<string, unknown>) {
  if (body) return request<{ codes: Array<Record<string, any>>; total: number }>({ url: '/qcDefectCode/list', method: 'post', data: body });
  return request<{ codes: Array<Record<string, any>> }>({ url: '/qcDefectCode' });
}

export function saveDefectCode(data: Record<string, unknown>, id?: number) {
  return request({ url: id ? `/qcDefectCode/${id}` : '/qcDefectCode', method: id ? 'put' : 'post', data });
}

export function fetchDefects(lotId?: number, body?: Record<string, unknown>) {
  if (body) return request<{ defects: Array<Record<string, any>>; total: number }>({ url: '/qcDefect/list', method: 'post', data: body });
  return request<{ defects: Array<Record<string, any>> }>({
    url: '/qcDefect',
    params: lotId ? { lotId } : undefined
  });
}

export function recordDefect(data: Record<string, unknown>) {
  return request({ url: '/qcDefect', method: 'post', data });
}

export function fetchPareto() {
  return request<{ rows: Array<{ defectCode: string; defectName: string; severity: string; quantity: number }> }>({
    url: '/qcDefect/pareto'
  });
}

export function fetchChart(params: Record<string, unknown>) {
  return request<ChartView>({ url: '/qcSpc/chart', params });
}

export function fetchPolicies() {
  return request<{ policies: Array<Record<string, any>> }>({ url: '/qcSpc/policy' });
}

export function savePolicy(data: Record<string, unknown>) {
  return request({ url: '/qcSpc/policy', method: 'put', data });
}
