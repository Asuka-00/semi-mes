import { request } from '../request';

export interface TraceReport {
  lot: Record<string, any>;
  backward: Array<Record<string, any>>;
  forward: Array<Record<string, any>>;
  history: Array<Record<string, any>>;
  moves: Array<Record<string, any>>;
  defects: Array<Record<string, any>>;
  measurements: Array<Record<string, any>>;
  wafers: Array<Record<string, any>>;
}

export function fetchLotTrace(lotNo: string) {
  return request<TraceReport>({
    url: '/wipLot/trace',
    method: 'post',
    data: { lotNo }
  });
}

export function fetchTraceReverse(data: { equipmentCode?: string; nodeKey?: string; from?: string; to?: string }) {
  return request<{ lots: Array<Record<string, any>> }>({
    url: '/wipLot/trace/reverse',
    method: 'post',
    data
  });
}
