<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { DataTableColumns } from 'naive-ui';
import { NButton, NTag } from 'naive-ui';
import { $t } from '@/locales';
import { fetchEquipmentBoard } from '@/service/api/equipment';

const router = useRouter();
const byState = ref<Array<{ key: string; count: number }>>([]);
const tools = ref<Array<Record<string, any>>>([]);
const overdue = ref(0);

function stateLabel(value: string) {
  if (!value) return '';
  return $t(`page.mes.eqp.state_${value}` as App.I18n.I18nKey);
}

const columns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.eqp.code'), key: 'equipmentCode' },
  { title: $t('page.mes.eqp.name'), key: 'equipmentName' },
  { title: $t('page.mes.eqp.group'), key: 'equipmentGroup', width: 100 },
  { title: $t('page.mes.eqp.state'), key: 'status', width: 150, render: row => h(NTag, { size: 'small' }, { default: () => stateLabel(row.status) }) },
  { title: $t('page.mes.eqp.openLots'), key: 'openLots', width: 90 },
  { title: $t('page.mes.eqp.capacity'), key: 'capacity', width: 80 },
  { title: $t('page.mes.eqp.utilization'), key: 'utilization', width: 110, render: row => `${Math.round(Number(row.utilization || 0) * 100)}%` },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 90,
    render: row => h(NButton, { size: 'small', onClick: () => router.push({ name: 'equipment_detail', params: { id: String(row.id) } }) }, { default: () => $t('page.mes.eqp.detail') })
  }
];

onMounted(async () => {
  const { data, error } = await fetchEquipmentBoard();
  if (error || !data) return;
  byState.value = data.byState || [];
  tools.value = data.tools || [];
  overdue.value = data.overdue || 0;
});
</script>

<template>
  <div class="flex flex-col gap-12px">
    <NCard :title="$t('route.equipment_board')" size="small">
      <NSpace>
        <NTag v-for="item in byState" :key="item.key">{{ stateLabel(item.key) }} {{ item.count }}</NTag>
        <NTag type="warning">{{ $t('page.mes.eqp.overdue') }} {{ overdue }}</NTag>
      </NSpace>
    </NCard>
    <NCard size="small">
      <NDataTable :columns="columns" :data="tools" size="small" />
    </NCard>
  </div>
</template>
