<script setup lang="ts">
import { $t } from '@/locales';

defineProps<{
  items: Array<{ key: string; label: string }>;
  hidden: string[];
}>();

const emit = defineEmits<{
  toggle: [key: string, shown: boolean];
  reorder: [key: string, direction: number];
}>();
</script>

<template>
  <NPopover trigger="click" placement="bottom-end">
    <template #trigger>
      <NButton>{{ $t('page.mes.query.columns') }}</NButton>
    </template>
    <div class="max-h-360px w-280px overflow-auto">
      <div v-for="(item, index) in items" :key="item.key" class="flex items-center gap-8px py-4px">
        <NCheckbox :checked="!hidden.includes(item.key)" @update:checked="(shown: boolean) => emit('toggle', item.key, shown)" />
        <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
        <NButton size="tiny" quaternary :disabled="index === 0" @click="emit('reorder', item.key, -1)">↑</NButton>
        <NButton size="tiny" quaternary :disabled="index === items.length - 1" @click="emit('reorder', item.key, 1)">↓</NButton>
      </div>
    </div>
  </NPopover>
</template>
