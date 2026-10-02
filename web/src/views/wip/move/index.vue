<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { $t } from '@/locales';
import { fetchMoves } from '@/service/api/track';

const keyword = ref('');
const rows = ref<Array<Record<string, any>>>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);

const columns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo', minWidth: 140 },
  { title: $t('page.mes.track.step'), key: 'nodeName', minWidth: 120, render: row => row.nodeName || row.nodeKey },
  { title: $t('page.mes.track.equipment'), key: 'equipmentCode', minWidth: 120, render: row => row.equipmentCode || '' },
  { title: $t('page.mes.track.operator'), key: 'operatorName', width: 120 },
  { title: $t('page.mes.track.qty'), key: 'qtyIn', width: 80 },
  { title: $t('page.mes.track.qtyOut'), key: 'qtyOut', width: 90 },
  { title: $t('page.mes.track.qtyScrap'), key: 'qtyScrap', width: 90 },
  { title: $t('page.mes.track.queue'), key: 'queueSeconds', width: 100 },
  { title: $t('page.mes.track.process'), key: 'processSeconds', width: 100 },
  { title: $t('page.mes.track.state'), key: 'state', width: 100 },
  { title: $t('page.mes.wip.result'), key: 'resolveReason', minWidth: 120 }
];

async function load() {
  loading.value = true;
  const { data, error } = await fetchMoves({
    page: page.value - 1,
    limit: 10,
    columns: keyword.value ? [{ name: 'lot_no', exp: 'like', value: keyword.value, logic: 'and' }] : undefined
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipMoves || [];
  total.value = Number(data.total || 0);
}

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.wip.moves')">
    <NSpace class="mb-12px">
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.track.lotNo')" class="w-220px" clearable @keyup.enter="page = 1; load()" />
      <NButton @click="page = 1; load()">{{ $t('common.search') }}</NButton>
    </NSpace>
    <NDataTable
      remote
      :loading="loading"
      :columns="columns"
      :data="rows"
      :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }"
    />
  </NCard>
</template>
