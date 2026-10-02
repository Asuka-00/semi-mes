import { request } from '../request';

export interface MesColumn {
  name: string;
  exp: string;
  value: string | number;
  logic: string;
}

export interface MesListQuery {
  page: number;
  limit: number;
  sort?: string;
  columns?: MesColumn[];
}

export function mesList(resource: string, body: MesListQuery) {
  return request<Record<string, any>>({
    url: `/${resource}/list`,
    method: 'post',
    data: body
  });
}

export function mesCreate(resource: string, data: Record<string, unknown>) {
  return request<{ id: number }>({
    url: `/${resource}`,
    method: 'post',
    data
  });
}

export function mesUpdate(resource: string, id: number, data: Record<string, unknown>) {
  return request<unknown>({
    url: `/${resource}/${id}`,
    method: 'put',
    data
  });
}

export function mesDelete(resource: string, id: number) {
  return request<unknown>({
    url: `/${resource}/${id}`,
    method: 'delete'
  });
}

export function mesGetUserRoles(id: number) {
  return request<{ roleIds: number[] }>({ url: `/sysUser/${id}/roles` });
}

export function mesSetUserRoles(id: number, roleIds: number[]) {
  return request<unknown>({ url: `/sysUser/${id}/roles`, method: 'put', data: { roleIds } });
}

export function mesGetRoleMenus(id: number) {
  return request<{ menuIds: number[] }>({ url: `/sysRole/${id}/menus` });
}

export function mesSetRoleMenus(id: number, menuIds: number[]) {
  return request<unknown>({ url: `/sysRole/${id}/menus`, method: 'put', data: { menuIds } });
}
