<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { $t } from '@/locales';
import { fetchPareto } from '@/service/api/quality';

const rows = ref<Array<{ defectCode: string; defectName: string; severity: string; quantity: number }>>([]);
const maxQty = computed(() => Math.max(1, ...rows.value.map(row => row.quantity || 0)));

async function load() {
  const { data, error } = await fetchPareto();
  if (error || !data) return;
  rows.value = data.rows || [];
}

const columns = computed<DataTableColumns<(typeof rows.value)[number]>>(() => [
  { title: $t('page.mes.qc.code'), key: 'defectCode', width: 140 },
  { title: $t('page.mes.qc.name'), key: 'defectName', width: 160 },
  { title: $t('page.mes.qc.severity'), key: 'severity', width: 120 },
  { title: $t('page.mes.track.qty'), key: 'quantity', width: 80 },
  {
    title: $t('page.mes.qc.pareto'),
    key: 'bar',
    render: row => `${Math.round((row.quantity / maxQty.value) * 100)}%`
  }
]);

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.pareto')">
    <div v-for="row in rows" :key="row.defectCode" class="mb-8px">
      <div class="mb-4px">{{ row.defectCode }} {{ row.defectName }} · {{ row.quantity }}</div>
      <div class="h-16px rd-4px bg-#efeff5">
        <div class="h-16px rd-4px bg-#2080f0" :style="{ width: `${(row.quantity / maxQty) * 100}%` }" />
      </div>
    </div>
    <NEmpty v-if="!rows.length" :description="$t('common.noData')" />
    <NDataTable :columns="columns" :data="rows" class="mt-16px" />
  </NCard>
</template>
