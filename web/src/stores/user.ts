import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as apiLogin, logout as apiLogout, getUserInfo, getMenus } from '@/api/auth'
import type { LoginParams } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const permissions = ref<string[]>([])
  const menus = ref<any[]>([])

  async function login(params: LoginParams) {
    const result = await apiLogin(params)
    token.value = result.token
    localStorage.setItem('token', result.token)
    return result
  }

  async function logout() {
    await apiLogout()
    token.value = ''
    userInfo.value = null
    permissions.value = []
    menus.value = []
    localStorage.removeItem('token')
  }

  async function fetchUserInfo() {
    const info = await getUserInfo()
    userInfo.value = info
    permissions.value = info.permissions || []
    return info
  }

  async function fetchMenus() {
    const menuList = await getMenus()
    menus.value = menuList
    return menuList
  }

  function hasPermission(permission: string): boolean {
    return permissions.value.includes(permission)
  }

  return {
    token,
    userInfo,
    permissions,
    menus,
    login,
    logout,
    fetchUserInfo,
    fetchMenus,
    hasPermission
  }
})
