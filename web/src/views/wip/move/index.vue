<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, rangeColumns, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { fetchMoves } from '@/service/api/track';

const keyword = ref('');
const state = ref<string[]>([]);
const equipment = ref('');
const createdRange = ref<[number, number] | null>(null);
const sort = ref('-id');
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const rows = ref<Array<Record<string, any>>>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const keys = ['lotNo', 'nodeName', 'equipmentCode', 'operatorName', 'qtyIn', 'qtyOut', 'qtyScrap', 'queueSeconds', 'processSeconds', 'state', 'resolveReason'];

const allColumns: DataTableColumns<Record<string, any>> = [
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

const columns = computed(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  return allColumns.filter(column => keepColumn(column, visible, []));
});

function filters() {
  const query: Array<{ name: string; exp: string; value: string; logic: string }> = [];
  if (keyword.value) query.push({ name: 'lot_no', exp: 'like', value: keyword.value, logic: 'and' });
  if (state.value.length) query.push({ name: 'state', exp: 'in', value: state.value.join(','), logic: 'and' });
  if (equipment.value) query.push({ name: 'equipment_code', exp: 'like', value: equipment.value, logic: 'and' });
  return [...query, ...rangeColumns('created_at', createdRange.value)];
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('wipMove', {
    search: { keyword: keyword.value, state: state.value, equipment: equipment.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value,
    sort: sort.value,
    createdRange: createdRange.value
  });
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchMoves({
    page: page.value - 1,
    limit: 10,
    sort: sort.value,
    columns: filters()
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipMoves || [];
  total.value = Number(data.total || 0);
}

async function exportRows() {
  const { data, error } = await fetchMoves({ page: 0, limit: 2000, sort: sort.value, columns: filters() });
  const list = data?.wipMoves || [];
  if (error || !list.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('moves', ['lotNo', 'state', 'equipmentCode'], list.map((row: Record<string, any>) => [row.lotNo, row.state, row.equipmentCode]));
}

function reset() {
  keyword.value = '';
  state.value = [];
  equipment.value = '';
  createdRange.value = null;
  sort.value = '-id';
  page.value = 1;
  persist();
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('wipMove');
  keyword.value = String(pref.search?.keyword || '');
  state.value = Array.isArray(pref.search?.state) ? (pref.search.state as string[]) : [];
  equipment.value = String(pref.search?.equipment || '');
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  sort.value = pref.sort || '-id';
  createdRange.value = pref.createdRange || null;
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.wip.moves')">
    <NSpace class="mb-12px" wrap>
      <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
      <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
      <ColumnPicker
        :items="orderedKeys(keys, columnOrder).map(key => ({ key, label: key }))"
        :hidden="hidden"
        @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
        @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : [...keys], key, dir); persist(); }"
      />
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.track.lotNo')" class="w-180px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="equipment" :placeholder="$t('page.mes.track.equipment')" class="w-160px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NSelect v-model:value="state" multiple clearable class="w-200px" :placeholder="$t('page.mes.track.state')" :options="['running', 'completed', 'aborted'].map(value => ({ label: value, value }))" />
      <NDatePicker v-model:value="createdRange" type="daterange" clearable />
      <NSelect v-model:value="sort" class="w-120px" :options="[{ label: 'ID ↓', value: '-id' }, { label: 'ID ↑', value: 'id' }]" />
      <NButton type="primary" @click="page = 1; persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="reset">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable
      remote
      :loading="loading"
      :columns="columns"
      :data="rows"
      :scroll-x="1280"
      :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }"
    />
  </NCard>
</template>
