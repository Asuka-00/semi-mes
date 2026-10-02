import { request } from '../request';
import type { NoticeItem } from './notice';

export interface DashBucket {
  key: string;
  label: string;
  count: number;
  qty: number;
}

export interface DashHold {
  id: number;
  lotNo: string;
  productCode: string;
  nodeKey: string;
  nodeName: string;
  quantity: number;
  holdReasonCode: string;
  holdReason: string;
  heldAt?: string;
  holdSeconds: number;
}

export interface DashDay {
  day: string;
  moves: number;
  scrap: number;
}

export interface DashEvent {
  id: number;
  lotId: number;
  lotNo: string;
  eventType: string;
  reason: string;
  createdAt?: string;
}

export interface ShopDashboard {
  updatedAt: string;
  wip: boolean;
  equipment: boolean;
  wipLots: number;
  wipQty: number;
  holdLots: number;
  runningLots: number;
  todayMoves: number;
  todayCompleted: number;
  todayScrap: number;
  byStep: DashBucket[];
  byProduct: DashBucket[];
  byStatus: DashBucket[];
  holds: DashHold[];
  trend: DashDay[];
  events: DashEvent[];
  byEquipment: DashBucket[];
  overduePm: number;
  notices: NoticeItem[];
  unread: number;
}

export function fetchDashboard() {
  return request<ShopDashboard>({ url: '/dashboard' });
}
