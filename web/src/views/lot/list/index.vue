<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { NButton, NSpace, NTag } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
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
    hold: 'page.mes.wip.hold',
    completed: 'page.mes.wip.completed',
    merged: 'page.mes.wip.merged'
  };
  return $t((map[status] || 'page.mes.wip.status') as App.I18n.I18nKey);
}

const columns = computed<DataTableColumns<LotRow>>(() => [
  { type: 'selection' },
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo', minWidth: 150 },
  { title: $t('page.mes.wip.orderNo'), key: 'orderNo', minWidth: 120 },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 80 },
  {
    title: $t('page.mes.wip.lotType'),
    key: 'lotType',
    width: 100,
    render: row => (row.lotType === 'engineering' ? $t('page.mes.wip.engineering') : $t('page.mes.wip.production'))
  },
  { title: $t('page.mes.wip.currentNode'), key: 'nodeName', minWidth: 140, render: row => row.nodeName || row.currentNodeKey },
  {
    title: $t('page.mes.wip.status'),
    key: 'status',
    width: 110,
    render: row => h(NTag, { size: 'small', type: row.status === 'hold' ? 'warning' : 'default' }, { default: () => statusLabel(row.status) })
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
]);

async function load() {
  loading.value = true;
  const { data, error } = await fetchLots({
    page: page.value - 1,
    limit: pageSize.value,
    sort: '-id',
    columns: keyword.value ? [{ name: 'lot_no', exp: 'like', value: keyword.value, logic: 'and' }] : undefined
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipLots || [];
  total.value = Number(data.total || 0);
}

function search() {
  page.value = 1;
  load();
}

function openHold(id: number, release: boolean) {
  holdId.value = id;
  holdRelease.value = release;
  reasonCode.value = release ? 'RELEASE' : 'HOLD';
  reason.value = '';
  holdOpen.value = true;
}

async function submitHold() {
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

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NSpace class="mb-12px" justify="space-between">
      <NSpace>
        <NInput v-model:value="keyword" :placeholder="$t('page.mes.wip.lotNo')" clearable class="w-220px" @keyup.enter="search" />
        <NButton @click="search">{{ $t('common.search') }}</NButton>
      </NSpace>
      <NButton v-if="canEdit" :disabled="checked.length < 2" type="primary" @click="submitMerge">{{ $t('page.mes.wip.merge') }}</NButton>
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
          <NInput v-model:value="reasonCode" />
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
