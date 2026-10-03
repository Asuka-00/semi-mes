<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { NButton, NTag } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { request } from '@/service/request';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('wip:carrier:add'));
const canEdit = computed(() => hasAuth('wip:carrier:edit'));
const canDelete = computed(() => hasAuth('wip:carrier:delete'));
const keyword = ref('');
const status = ref<string | null>(null);
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const rows = ref<Array<Record<string, any>>>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const modal = ref(false);
const editingId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);
const selected = ref<Record<string, any> | null>(null);
const slots = ref<Array<Record<string, any>>>([]);
const lotNo = ref('');
const bindLotId = ref<number | null>(null);
const keys = ['carrierNo', 'carrierType', 'capacity', 'status', 'location', 'cleanCount', 'cleanLimit', 'lotNo'];
const form = reactive({
  carrierNo: '',
  carrierType: 'FOUP',
  capacity: 25,
  status: 'empty',
  location: '',
  cleanCount: 0,
  cleanLimit: 50,
  note: ''
});

const typeOptions = computed(() => [
  { label: $t('page.mes.carrier.foup'), value: 'FOUP' },
  { label: $t('page.mes.carrier.cassette'), value: 'Cassette' }
]);
const statusOptions = computed(() => [
  { label: $t('page.mes.carrier.empty'), value: 'empty' },
  { label: $t('page.mes.carrier.inUse'), value: 'in_use' },
  { label: $t('page.mes.carrier.cleaning'), value: 'cleaning' },
  { label: $t('page.mes.carrier.down'), value: 'down' }
]);
const rules: FormRules = {
  carrierType: { required: true, message: $t('form.required'), trigger: 'change' },
  capacity: { required: true, type: 'number', message: $t('form.required'), trigger: 'blur' }
};

function statusLabel(value: string) {
  return statusOptions.value.find(item => item.value === value)?.label || value;
}

const allColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.carrier.no'), key: 'carrierNo', minWidth: 140 },
  { title: $t('page.mes.carrier.type'), key: 'carrierType', width: 110 },
  { title: $t('page.mes.carrier.capacity'), key: 'capacity', width: 80 },
  { title: $t('page.mes.carrier.status'), key: 'status', width: 110, render: row => statusLabel(String(row.status)) },
  { title: $t('page.mes.carrier.location'), key: 'location', minWidth: 120 },
  { title: $t('page.mes.carrier.cleanCount'), key: 'cleanCount', width: 100 },
  { title: $t('page.mes.carrier.cleanLimit'), key: 'cleanLimit', width: 100 },
  { title: $t('page.mes.track.lotNo'), key: 'lotNo', minWidth: 140 },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 160,
    render: row =>
      h('div', { class: 'flex gap-4px' }, [
        canEdit.value ? h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => $t('common.edit') }) : null,
        canDelete.value
          ? h(NButton, { size: 'tiny', type: 'error', onClick: () => remove(Number(row.id)) }, { default: () => $t('common.delete') })
          : null
      ])
  }
]);
const columns = computed(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  return allColumns.value.filter(column => keepColumn(column, visible, ['actions']));
});

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('wipCarrier', {
    search: { keyword: keyword.value, status: status.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value
  });
}

async function load() {
  loading.value = true;
  const columnsQuery = [];
  if (keyword.value) columnsQuery.push({ name: 'carrier_no', exp: 'like', value: keyword.value, logic: 'and' });
  if (status.value) columnsQuery.push({ name: 'status', exp: '=', value: status.value, logic: 'and' });
  const { data, error } = await request<{ wipCarriers: Array<Record<string, any>>; total: number }>({
    url: '/wipCarrier/list',
    method: 'post',
    data: { page: page.value - 1, limit: 10, sort: '-id', columns: columnsQuery }
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipCarriers || [];
  total.value = Number(data.total || 0);
}

async function openSlots(row: Record<string, any>) {
  selected.value = row;
  const { data, error } = await request<{ slots: Array<Record<string, any>>; lotNo: string }>({ url: `/wipCarrier/${row.id}/slots` });
  if (error || !data) return;
  slots.value = data.slots || [];
  lotNo.value = data.lotNo || '';
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, { carrierNo: '', carrierType: 'FOUP', capacity: 25, status: 'empty', location: '', cleanCount: 0, cleanLimit: 50, note: '' });
  modal.value = true;
}

function openEdit(row: Record<string, any>) {
  editingId.value = Number(row.id);
  Object.assign(form, row);
  modal.value = true;
}

async function save() {
  await formRef.value?.validate();
  const { error } = await request({
    url: editingId.value ? `/wipCarrier/${editingId.value}` : '/wipCarrier',
    method: editingId.value ? 'put' : 'post',
    data: { ...form }
  });
  if (error) return;
  modal.value = false;
  load();
}

async function remove(id: number) {
  const { error } = await request({ url: `/wipCarrier/${id}`, method: 'delete' });
  if (error) return;
  if (selected.value?.id === id) {
    selected.value = null;
    slots.value = [];
  }
  load();
}

async function bind() {
  if (!selected.value || !bindLotId.value) return;
  const { error } = await request({ url: `/wipCarrier/${selected.value.id}/bind`, method: 'post', data: { lotId: bindLotId.value } });
  if (error) return;
  window.$message?.success($t('page.mes.carrier.bind'));
  await load();
  const row = rows.value.find(item => item.id === selected.value?.id) || selected.value;
  openSlots(row);
}

async function unbind() {
  if (!selected.value) return;
  const { error } = await request({ url: `/wipCarrier/${selected.value.id}/unbind`, method: 'post', data: {} });
  if (error) return;
  await load();
  const row = rows.value.find(item => item.id === selected.value?.id) || selected.value;
  openSlots(row);
}

function exportRows() {
  if (!rows.value.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('carriers', keys, rows.value.map(row => keys.map(key => row[key])));
}

onMounted(async () => {
  const pref = await loadPagePref('wipCarrier');
  keyword.value = String(pref.search?.keyword || '');
  status.value = pref.search?.status ? String(pref.search.status) : null;
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  prefsReady.value = true;
  load();
});
</script>

<template>
  <div class="flex-col gap-12px">
    <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.carrier.title')">
      <NSpace class="mb-12px" wrap>
        <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
        <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
        <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
        <ColumnPicker
          :items="orderedKeys(keys, columnOrder).map(key => ({ key, label: key }))"
          :hidden="hidden"
          @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
          @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : [...keys], key, dir); persist(); }"
        />
      </NSpace>
      <NSpace v-show="!collapsed" class="mb-12px" wrap>
        <NInput v-model:value="keyword" :placeholder="$t('page.mes.carrier.no')" class="w-180px" clearable @keyup.enter="page = 1; persist(); load()" />
        <NSelect v-model:value="status" :options="statusOptions" clearable class="w-140px" :placeholder="$t('page.mes.carrier.status')" />
        <NButton type="primary" @click="page = 1; persist(); load()">{{ $t('common.search') }}</NButton>
      </NSpace>
      <NDataTable
        remote
        :loading="loading"
        :columns="columns"
        :data="rows"
        :row-props="(row: Record<string, any>) => ({ style: 'cursor: pointer', onClick: () => openSlots(row) })"
        :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }"
      />
    </NCard>
    <NCard v-if="selected" :bordered="false" class="card-wrapper" :title="`${$t('page.mes.carrier.slots')} ${selected.carrierNo}`">
      <NSpace class="mb-12px" wrap>
        <NTag v-if="lotNo">{{ lotNo }}</NTag>
        <NInputNumber v-if="canEdit && !lotNo" v-model:value="bindLotId" :placeholder="$t('page.mes.carrier.lotId')" class="w-160px" />
        <NButton v-if="canEdit && !lotNo" type="primary" :disabled="!bindLotId" @click="bind">{{ $t('page.mes.carrier.bind') }}</NButton>
        <NButton v-if="canEdit && lotNo" @click="unbind">{{ $t('page.mes.carrier.unbind') }}</NButton>
      </NSpace>
      <div class="grid grid-cols-5 gap-8px">
        <div v-for="cell in slots" :key="cell.slot" class="border border-#e5e7eb rounded-6px px-8px py-6px" :class="cell.waferNo ? 'bg-#e8f7ee' : ''">
          <div class="text-12px text-#888">{{ cell.slot }}</div>
          <div class="truncate text-13px">{{ cell.waferNo || $t('page.mes.carrier.emptySlot') }}</div>
        </div>
      </div>
    </NCard>
    <NModal v-model:show="modal" preset="card" :title="editingId ? $t('common.edit') : $t('common.add')" class="w-520px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="110">
        <NFormItem :label="$t('page.mes.carrier.no')"><NInput v-model:value="form.carrierNo" :placeholder="$t('page.mes.number.auto')" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.type')" path="carrierType"><NSelect v-model:value="form.carrierType" :options="typeOptions" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.capacity')" path="capacity"><NInputNumber v-model:value="form.capacity" :min="1" :max="25" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.status')"><NSelect v-model:value="form.status" :options="statusOptions" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.location')"><NInput v-model:value="form.location" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.cleanCount')"><NInputNumber v-model:value="form.cleanCount" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.cleanLimit')"><NInputNumber v-model:value="form.cleanLimit" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.carrier.note')"><NInput v-model:value="form.note" /></NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="save">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
