<template>
  <n-layout has-sider style="height: 100vh">
    <n-layout-sider
      bordered
      show-trigger
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="appStore.collapsed"
      @collapse="appStore.toggleSidebar"
      @expand="appStore.toggleSidebar"
    >
      <div class="logo">
        <span v-if="!appStore.collapsed">MES System</span>
        <span v-else>MES</span>
      </div>
      <n-menu
        :collapsed="appStore.collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="activeKey"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered style="height: 64px; padding: 0 24px; display: flex; align-items: center; justify-content: space-between;">
        <div>{{ currentTitle }}</div>
        <n-space>
          <n-dropdown :options="languageOptions" @select="handleLanguageSelect">
            <n-button text>
              {{ appStore.locale === 'zh-CN' ? '中文' : 'English' }}
            </n-button>
          </n-dropdown>
          <n-dropdown :options="userMenuOptions" @select="handleUserMenuSelect">
            <n-button text>
              {{ userStore.userInfo?.real_name || userStore.userInfo?.username }}
            </n-button>
          </n-dropdown>
        </n-space>
      </n-layout-header>
      <n-layout-content content-style="padding: 24px;">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import {
  HomeOutline,
  SettingsOutline,
  PeopleOutline,
  BusinessOutline,
  CubeOutline,
  LogOutOutline
} from '@vicons/ionicons5'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()
const { t, locale } = useI18n()

const activeKey = computed(() => route.path)
const currentTitle = computed(() => {
  const meta = route.meta as any
  return meta?.title ? t(meta.title) : ''
})

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = computed<MenuOption[]>(() => {
  const menus: MenuOption[] = [
    {
      label: t('menu.dashboard'),
      key: '/dashboard',
      icon: renderIcon(HomeOutline)
    },
    {
      label: t('menu.system'),
      key: '/system',
      icon: renderIcon(SettingsOutline),
      children: [
        {
          label: t('menu.system.user'),
          key: '/system/user'
        },
        {
          label: t('menu.system.role'),
          key: '/system/role'
        },
        {
          label: t('menu.system.menu'),
          key: '/system/menu'
        }
      ]
    },
    {
      label: t('menu.baseData'),
      key: '/base-data',
      icon: renderIcon(BusinessOutline),
      children: [
        {
          label: t('menu.baseData.factory'),
          key: '/base-data/factory'
        },
        {
          label: t('menu.baseData.workshop'),
          key: '/base-data/workshop'
        },
        {
          label: t('menu.baseData.line'),
          key: '/base-data/line'
        },
        {
          label: t('menu.baseData.product'),
          key: '/base-data/product'
        },
        {
          label: t('menu.baseData.route'),
          key: '/base-data/route'
        },
        {
          label: t('menu.baseData.operation'),
          key: '/base-data/operation'
        },
        {
          label: t('menu.baseData.recipe'),
          key: '/base-data/recipe'
        }
      ]
    }
  ]
  return menus
})

const languageOptions = [
  {
    label: '简体中文',
    key: 'zh-CN'
  },
  {
    label: 'English',
    key: 'en-US'
  }
]

const userMenuOptions = [
  {
    label: t('auth.logout'),
    key: 'logout',
    icon: renderIcon(LogOutOutline)
  }
]

function handleMenuSelect(key: string) {
  router.push(key)
}

function handleLanguageSelect(key: string) {
  appStore.setLocale(key)
  locale.value = key
}

async function handleUserMenuSelect(key: string) {
  if (key === 'logout') {
    await userStore.logout()
    router.push('/login')
  }
}

onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.fetchUserInfo()
  }
})
</script>

<style scoped>
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid #f0f0f0;
}
</style>
