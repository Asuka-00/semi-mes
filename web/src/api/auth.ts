import request from '@/utils/request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  user_info: {
    id: number
    username: string
    real_name: string
    email: string
    phone: string
  }
}

export interface UserInfo {
  id: number
  username: string
  real_name: string
  email: string
  phone: string
  roles: any[]
  permissions: string[]
}

export function login(data: LoginParams) {
  return request.post<any, LoginResult>('/api/v1/auth/login', data)
}

export function logout() {
  return request.post('/api/v1/auth/logout')
}

export function getUserInfo() {
  return request.get<any, UserInfo>('/api/v1/auth/user-info')
}

export function getMenus() {
  return request.get<any, any[]>('/api/v1/auth/menus')
}
