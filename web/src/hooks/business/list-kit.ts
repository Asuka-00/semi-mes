import { useAuthStore } from '@/store/modules/auth';
import { request } from '@/service/request';

export interface ListPref {
  search?: Record<string, unknown>;
  hidden?: string[];
  order?: string[];
  collapsed?: boolean;
  sort?: string;
  createdRange?: [number, number] | null;
}

export function loadPagePref(pageKey: string) {
  const auth = useAuthStore();
  const storageKey = `mes-list:${auth.userInfo.userId || '0'}:${pageKey}`;
  return request<{ payload: string }>({ url: `/sysPref/${encodeURIComponent(pageKey)}` }).then(({ data, error }) => {
    let saved: ListPref = {};
    const raw = localStorage.getItem(storageKey);
    if (raw) {
      try {
        saved = JSON.parse(raw) as ListPref;
      } catch {
        saved = {};
      }
    }
    if (!error && data?.payload) {
      try {
        saved = JSON.parse(data.payload) as ListPref;
      } catch {
        /* keep the local copy */
      }
    }
    return saved || {};
  });
}

export function savePagePref(pageKey: string, payload: ListPref) {
  const auth = useAuthStore();
  const storageKey = `mes-list:${auth.userInfo.userId || '0'}:${pageKey}`;
  const text = JSON.stringify(payload);
  localStorage.setItem(storageKey, text);
  return request({ url: `/sysPref/${encodeURIComponent(pageKey)}`, method: 'put', data: { payload: text } });
}

export function downloadCsv(filename: string, headers: string[], rows: Array<Array<string | number | null | undefined>>) {
  const esc = (value: string | number | null | undefined) => `"${String(value ?? '').replace(/"/g, '""')}"`;
  const body = [headers.map(esc).join(','), ...rows.map(row => row.map(esc).join(','))].join('\n');
  const blob = new Blob([`\uFEFF${body}`], { type: 'text/csv;charset=utf-8' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename.endsWith('.csv') ? filename : `${filename}.csv`;
  link.click();
  URL.revokeObjectURL(link.href);
}

export function formatDay(value: number) {
  const date = new Date(value);
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${month}-${day}`;
}

export function nextDay(ymd: string) {
  const [year, month, day] = ymd.split('-').map(Number);
  const date = new Date(year, (month || 1) - 1, day || 1);
  date.setDate(date.getDate() + 1);
  const mm = String(date.getMonth() + 1).padStart(2, '0');
  const dd = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${mm}-${dd}`;
}

export function rangeColumns(name: string, range: [number, number] | null | undefined) {
  if (!range || range.length < 2 || !range[0] || !range[1]) return [];
  const start = formatDay(range[0]);
  const end = formatDay(range[1]);
  return [
    { name, exp: 'gte', value: start, logic: 'and' },
    { name, exp: 'lt', value: nextDay(end), logic: 'and' }
  ];
}

export function snake(key: string) {
  return key.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
}

export function orderedKeys(keys: string[], order: string[]) {
  const rank = new Map(order.map((key, index) => [key, index]));
  return [...keys].sort((a, b) => (rank.get(a) ?? 1000) - (rank.get(b) ?? 1000));
}

export function keepColumn(column: object, visible: Set<string>, always: string[] = ['actions', 'action']) {
  const typed = column as { type?: string; key?: unknown };
  if (typed.type === 'selection') return true;
  const key = String(typed.key ?? '');
  return always.includes(key) || visible.has(key);
}

export function moveKey(order: string[], key: string, direction: number) {
  const keys = order.includes(key) ? [...order] : [...order, key];
  const index = keys.indexOf(key);
  const target = index + direction;
  if (index < 0 || target < 0 || target >= keys.length) return keys;
  const next = [...keys];
  const [item] = next.splice(index, 1);
  next.splice(target, 0, item);
  return next;
}
