import { request } from '../request';

export interface NoticeItem {
  id: number;
  kind: string;
  titleZh: string;
  titleEn: string;
  bodyZh: string;
  bodyEn: string;
  read: boolean;
  createdAt?: string;
  refType?: string;
  refId?: number;
}

export function fetchFeatures() {
  return request<{ spc: boolean }>({ url: '/features' });
}

export function fetchNotices() {
  return request<{ notices: NoticeItem[]; unread: number }>({ url: '/sysNotice' });
}

export function markNoticeRead(id: number) {
  return request({ url: `/sysNotice/${id}/read`, method: 'post' });
}

export function markAllNoticesRead() {
  return request({ url: '/sysNotice/readAll', method: 'post' });
}
