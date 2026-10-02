<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { $t } from '@/locales';
import { fetchEquipment } from '@/service/api/track';
import type { EquipmentRow } from '@/service/api/track';

const rows = ref<EquipmentRow[]>([]);

const columns: DataTableColumns<EquipmentRow> = [
  { title: $t('page.mes.track.code'), key: 'equipmentCode' },
  { title: $t('page.mes.track.name'), key: 'equipmentName' },
  { title: $t('page.mes.track.group'), key: 'equipmentGroup' },
  {
    title: $t('page.mes.wip.status'),
    key: 'status',
    render: row => (row.status === 'down' ? $t('page.mes.track.down') : $t('page.mes.track.idle'))
  }
];

onMounted(async () => {
  const { data, error } = await fetchEquipment();
  if (error || !data) return;
  rows.value = data.equipment || [];
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('route.equipment_list')">
    <NDataTable :columns="columns" :data="rows" size="small" />
  </NCard>
</template>
