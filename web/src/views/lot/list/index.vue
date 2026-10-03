<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { NButton, NSpace, NTag } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { useDictOptions, useReasonOptions } from '@/hooks/business/dict';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, rangeColumns, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { mesBatch } from '@/service/api/mes';
import { fetchLots, holdLot, mergeLots, releaseHold, splitLot } from '@/service/api/wip';
import type { LotRow } from '@/service/api/wip';

const router = useRouter();
const { hasAuth } = useAuth();
const canEdit = computed(() => hasAuth('lot:lot:edit'));

const loading = ref(false);
const rows = ref<LotRow[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');
const statuses = ref<string[]>([]);
const productCode = ref('');
const createdRange = ref<[number, number] | null>(null);
const sort = ref('-id');
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const batchMode = ref(false);
const checked = ref<DataTableRowKey[]>([]);

const holdOpen = ref(false);
const holdId = ref(0);
const holdRelease = ref(false);
const reasonCode = ref('');
const reason = ref('');

const splitOpen = ref(false);
const splitId = ref(0);
const splitText = ref('');

function statusLabel(status: string) {
  const map: Record<string, string> = {
    waiting: 'page.mes.wip.waiting',
    running: 'page.mes.wip.running',
    hold: 'page.mes.wip.hold',
    scrapped: 'page.mes.wip.scrapped',
    completed: 'page.mes.wip.completed',
    merged: 'page.mes.wip.merged'
  };
  return $t((map[status] || 'page.mes.wip.status') as App.I18n.I18nKey);
}

const dataColumns = computed(() => [
  { key: 'lotNo', label: $t('page.mes.wip.lotNo') },
  { key: 'orderNo', label: $t('page.mes.wip.orderNo') },
  { key: 'quantity', label: $t('page.mes.wip.quantity') },
  { key: 'lotType', label: $t('page.mes.wip.lotType') },
  { key: 'nodeName', label: $t('page.mes.wip.currentNode') },
  { key: 'status', label: $t('page.mes.wip.status') }
]);

const { options: statusOptions, labelOf: dictStatus } = useDictOptions('lot_status', () =>
  ['waiting', 'running', 'hold', 'scrapped', 'completed', 'merged'].map(value => ({ label: statusLabel(value), value }))
);
const { labelOf: lotTypeLabel } = useDictOptions('lot_type', () => [
  { label: $t('page.mes.wip.production'), value: 'production' },
  { label: $t('page.mes.wip.engineering'), value: 'engineering' }
]);
const { options: holdReasons } = useReasonOptions('hold', () => [{ label: 'ENG_HOLD', value: 'ENG_HOLD' }, { label: 'QA', value: 'QA' }]);
const { options: releaseReasons } = useReasonOptions('release', () => [{ label: 'RELEASE', value: 'RELEASE' }]);

const columns = computed<DataTableColumns<LotRow>>(() => {
  const visible = new Set(orderedKeys(dataColumns.value.map(item => item.key), columnOrder.value).filter(key => !hidden.value.includes(key)));
  const all: DataTableColumns<LotRow> = [
  { type: 'selection' },
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo', minWidth: 150 },
  { title: $t('page.mes.wip.orderNo'), key: 'orderNo', minWidth: 120 },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 80 },
  {
    title: $t('page.mes.wip.lotType'),
    key: 'lotType',
    width: 100,
    render: row => lotTypeLabel(row.lotType)
  },
  { title: $t('page.mes.wip.currentNode'), key: 'nodeName', minWidth: 140, render: row => row.nodeName || row.currentNodeKey },
  {
    title: $t('page.mes.wip.status'),
    key: 'status',
    width: 110,
    render: row => h(NTag, { size: 'small', type: row.status === 'hold' ? 'warning' : 'default' }, { default: () => dictStatus(row.status) })
  },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 280,
    render: row => {
      const buttons = [
        h(
          NButton,
          { size: 'small', ghost: true, type: 'primary', onClick: () => router.push({ name: 'lot_detail', params: { id: String(row.id) } }) },
          { default: () => $t('page.mes.wip.detail') }
        )
      ];
      if (canEdit.value && row.status === 'waiting') {
        buttons.push(h(NButton, { size: 'small', ghost: true, onClick: () => openHold(row.id, false) }, { default: () => $t('page.mes.wip.holdAction') }));
        buttons.push(h(NButton, { size: 'small', ghost: true, onClick: () => openSplit(row.id) }, { default: () => $t('page.mes.wip.split') }));
      }
      if (canEdit.value && row.status === 'hold') {
        buttons.push(h(NButton, { size: 'small', ghost: true, onClick: () => openHold(row.id, true) }, { default: () => $t('page.mes.wip.releaseHold') }));
      }
      return h(NSpace, { size: 8 }, { default: () => buttons });
    }
  }
  ];
  return all.filter(column => keepColumn(column, visible));
});

function filters() {
  const query: Array<{ name: string; exp: string; value: string; logic: string }> = [];
  if (keyword.value) query.push({ name: 'lot_no', exp: 'like', value: keyword.value, logic: 'and' });
  if (statuses.value.length) query.push({ name: 'status', exp: 'in', value: statuses.value.join(','), logic: 'and' });
  if (productCode.value) query.push({ name: 'product_code', exp: 'like', value: productCode.value, logic: 'and' });
  return [...query, ...rangeColumns('created_at', createdRange.value)];
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('wipLot', {
    search: { keyword: keyword.value, statuses: statuses.value, productCode: productCode.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value,
    sort: sort.value,
    createdRange: createdRange.value
  });
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchLots({
    page: page.value - 1,
    limit: pageSize.value,
    sort: sort.value,
    columns: filters()
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipLots || [];
  total.value = Number(data.total || 0);
}

function search() {
  page.value = 1;
  persist();
  load();
}

function reset() {
  keyword.value = '';
  statuses.value = [];
  productCode.value = '';
  createdRange.value = null;
  sort.value = '-id';
  search();
}

async function exportRows() {
  const { data, error } = await fetchLots({ page: 0, limit: 2000, sort: sort.value, columns: filters() });
  if (error || !data?.wipLots?.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  const keys = orderedKeys(dataColumns.value.map(item => item.key), columnOrder.value).filter(key => !hidden.value.includes(key));
  downloadCsv(
    'lots',
    keys.map(key => dataColumns.value.find(item => item.key === key)?.label || key),
    data.wipLots.map(row => keys.map(key => String((row as unknown as Record<string, unknown>)[key] ?? '')))
  );
  window.$message?.success($t('page.mes.query.exported', { count: data.wipLots.length }));
}

function openBatch(release: boolean) {
  batchMode.value = true;
  holdRelease.value = release;
  reasonCode.value = release ? 'RELEASE' : 'HOLD';
  reason.value = '';
  holdOpen.value = true;
}

function openHold(id: number, release: boolean) {
  batchMode.value = false;
  holdId.value = id;
  holdRelease.value = release;
  reasonCode.value = release ? 'RELEASE' : 'HOLD';
  reason.value = '';
  holdOpen.value = true;
}

async function submitHold() {
  if (batchMode.value) {
    const ids = checked.value.map(item => Number(item));
    const { data, error } = await mesBatch({
      resource: 'wipLot',
      action: holdRelease.value ? 'release' : 'hold',
      ids,
      reasonCode: reasonCode.value,
      reason: reason.value
    });
    if (error || !data) return;
    holdOpen.value = false;
    checked.value = [];
    window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
    load();
    return;
  }
  const call = holdRelease.value ? releaseHold : holdLot;
  const { error } = await call(holdId.value, reasonCode.value, reason.value);
  if (error) return;
  holdOpen.value = false;
  load();
}

function openSplit(id: number) {
  splitId.value = id;
  splitText.value = '';
  splitOpen.value = true;
}

async function submitSplit() {
  const quantities = splitText.value
    .split(',')
    .map(item => Number(item.trim()))
    .filter(item => item > 0);
  const { error } = await splitLot(splitId.value, quantities);
  if (error) return;
  splitOpen.value = false;
  window.$message?.success($t('page.mes.wip.saved'));
  load();
}

async function submitMerge() {
  const ids = checked.value.map(item => Number(item));
  if (ids.length < 2) return;
  const { error } = await mergeLots(ids[0], ids.slice(1));
  if (error) return;
  checked.value = [];
  window.$message?.success($t('page.mes.wip.saved'));
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('wipLot');
  keyword.value = String(pref.search?.keyword || '');
  statuses.value = Array.isArray(pref.search?.statuses) ? (pref.search?.statuses as string[]) : [];
  productCode.value = String(pref.search?.productCode || '');
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || dataColumns.value.map(item => item.key);
  collapsed.value = Boolean(pref.collapsed);
  sort.value = pref.sort || '-id';
  createdRange.value = pref.createdRange || null;
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NSpace class="mb-12px" justify="space-between" wrap>
      <NSpace>
        <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
        <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
        <ColumnPicker
          :items="orderedKeys(dataColumns.map(item => item.key), columnOrder).map(key => dataColumns.find(item => item.key === key)!)"
          :hidden="hidden"
          @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
          @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : dataColumns.map(item => item.key), key, dir); persist(); }"
        />
        <NButton v-if="canEdit" :disabled="!checked.length" @click="openBatch(false)">{{ $t('page.mes.query.batchHold') }}</NButton>
        <NButton v-if="canEdit" :disabled="!checked.length" @click="openBatch(true)">{{ $t('page.mes.query.batchRelease') }}</NButton>
        <NButton v-if="canEdit" :disabled="checked.length < 2" type="primary" @click="submitMerge">{{ $t('page.mes.wip.merge') }}</NButton>
      </NSpace>
      <span v-if="checked.length">{{ $t('page.mes.query.selected', { count: checked.length }) }}</span>
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.wip.lotNo')" clearable class="w-180px" @keyup.enter="search" />
      <NInput v-model:value="productCode" :placeholder="$t('page.mes.wip.product')" clearable class="w-160px" @keyup.enter="search" />
      <NSelect v-model:value="statuses" multiple clearable class="w-220px" :options="statusOptions" :placeholder="$t('page.mes.wip.status')" />
      <NDatePicker v-model:value="createdRange" type="daterange" clearable />
      <NSelect v-model:value="sort" class="w-140px" :options="[{ label: 'ID ↓', value: '-id' }, { label: 'ID ↑', value: 'id' }, { label: 'Lot ↑', value: 'lot_no' }, { label: 'Lot ↓', value: '-lot_no' }]" />
      <NButton type="primary" @click="search">{{ $t('common.search') }}</NButton>
      <NButton @click="reset">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable
      v-model:checked-row-keys="checked"
      remote
      :row-key="(row: LotRow) => row.id"
      :loading="loading"
      :columns="columns"
      :data="rows"
      :pagination="{ page, pageSize, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }"
    />
    <NModal v-model:show="holdOpen" preset="card" :title="holdRelease ? $t('page.mes.wip.releaseHold') : $t('page.mes.wip.holdAction')" class="w-460px">
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="$t('page.mes.wip.reasonCode')">
          <NSelect v-model:value="reasonCode" :options="holdRelease ? releaseReasons : holdReasons" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.reason')">
          <NInput v-model:value="reason" type="textarea" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="holdOpen = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitHold">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="splitOpen" preset="card" :title="$t('page.mes.wip.split')" class="w-460px">
      <p class="mb-8px">{{ $t('page.mes.wip.splitHint') }}</p>
      <NInput v-model:value="splitText" placeholder="2, 1" />
      <template #footer>
        <NSpace justify="end">
          <NButton @click="splitOpen = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitSplit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
