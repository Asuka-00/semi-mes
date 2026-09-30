<template>
  <div>
    <n-card :title="$t('menu.dashboard')">
      <n-space vertical>
        <n-alert type="success">
          {{ $t('auth.welcomeBack') }}, {{ userStore.userInfo?.real_name || userStore.userInfo?.username }}!
        </n-alert>
        <n-statistic label="System Version">
          v1.0.0
        </n-statistic>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>
