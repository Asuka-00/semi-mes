import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const locale = ref<string>(localStorage.getItem('locale') || 'zh-CN')
  const collapsed = ref(false)

  function setLocale(lang: string) {
    locale.value = lang
    localStorage.setItem('locale', lang)
  }

  function toggleSidebar() {
    collapsed.value = !collapsed.value
  }

  return {
    locale,
    collapsed,
    setLocale,
    toggleSidebar
  }
})
