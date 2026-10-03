import { request } from '../request';

export function fetchAuditList(data: {
  page: number;
  limit: number;
  sort?: string;
  columns?: Array<{ name: string; exp: string; value: string; logic: string }>;
}) {
  return request<{ sysAudits: Array<Record<string, any>>; total: number }>({
    url: '/sysAudit/list',
    method: 'post',
    data
  });
}
