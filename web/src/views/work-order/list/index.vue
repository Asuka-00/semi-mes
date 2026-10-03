<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { VNode } from 'vue';
import type { DataTableColumns, DataTableRowKey, FormInst, FormRules } from 'naive-ui';
import { NButton, NPopconfirm, NSpace, NTag } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, rangeColumns, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { mesBatch } from '@/service/api/mes';
import {
  closeWorkOrder,
  fetchReleasedRoutes,
  fetchWorkOrders,
  releaseWorkOrder,
  removeWorkOrder,
  saveWorkOrder,
  startLot
} from '@/service/api/wip';
import type { ReleasedRoute, WorkOrderRow } from '@/service/api/wip';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('wo:order:add'));
const canEdit = computed(() => hasAuth('wo:order:edit'));
const canDelete = computed(() => hasAuth('wo:order:delete'));

const loading = ref(false);
const rows = ref<WorkOrderRow[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');
const statuses = ref<string[]>([]);
const productCode = ref('');
const dueRange = ref<[number, number] | null>(null);
const sort = ref('-id');
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const checked = ref<DataTableRowKey[]>([]);
const versions = ref<ReleasedRoute[]>([]);

const modal = ref(false);
const editingId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);
const form = reactive({
  orderNo: '',
  productID: null as number | null,
  routeVersionID: null as number | null,
  plannedQty: 25,
  priority: 2,
  dueDate: '',
  note: ''
});

const startOpen = ref(false);
const startOrder = ref<WorkOrderRow | null>(null);
const startQty = ref(1);
const startType = ref('production');

const rules = computed<FormRules>(() => ({
  orderNo: { required: true, message: $t('form.required'), trigger: 'blur' },
  routeVersionID: { required: true, type: 'number', message: $t('page.mes.wip.selectVersion'), trigger: 'change' },
  plannedQty: { required: true, type: 'number', message: $t('form.required'), trigger: 'blur' }
}));

const versionOptions = computed(() =>
  versions.value.map(item => ({
    label: `${item.routeCode} v${item.versionNo} · ${item.productCode}`,
    value: item.versionID
  }))
);

const priorityOptions = computed(() => [
  { label: $t('page.mes.wip.priorityLow'), value: 1 },
  { label: $t('page.mes.wip.priorityNormal'), value: 2 },
  { label: $t('page.mes.wip.priorityHigh'), value: 3 },
  { label: $t('page.mes.wip.priorityUrgent'), value: 4 }
]);

const lotTypeOptions = computed(() => [
  { label: $t('page.mes.wip.production'), value: 'production' },
  { label: $t('page.mes.wip.engineering'), value: 'engineering' }
]);

function statusLabel(status: string) {
  const map: Record<string, string> = {
    created: 'page.mes.wip.created',
    released: 'page.mes.wip.released',
    in_progress: 'page.mes.wip.inProgress',
    completed: 'page.mes.wip.completed',
    closed: 'page.mes.wip.closed'
  };
  return $t((map[status] || 'page.mes.wip.status') as App.I18n.I18nKey);
}

function priorityLabel(value: number) {
  return priorityOptions.value.find(item => item.value === value)?.label || String(value);
}

function actionButton(label: string, onClick: () => void, type: 'default' | 'primary' | 'warning' | 'error' = 'default') {
  return h(NButton, { size: 'small', ghost: true, type, onClick }, { default: () => label });
}

const dataColumnKeys = ['orderNo', 'productCode', 'routeCode', 'plannedQty', 'releasedQty', 'completedQty', 'priority', 'dueDateText', 'status'];

const columns = computed<DataTableColumns<WorkOrderRow>>(() => {
  const visible = new Set(orderedKeys(dataColumnKeys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  const all: DataTableColumns<WorkOrderRow> = [
  { type: 'selection' },
  { title: $t('page.mes.wip.orderNo'), key: 'orderNo', minWidth: 140 },
  {
    title: $t('page.mes.wip.product'),
    key: 'productCode',
    minWidth: 140,
    render: row => `${row.productCode || ''} ${row.productName || ''}`.trim()
  },
  {
    title: $t('page.mes.wip.routeVersion'),
    key: 'routeCode',
    minWidth: 160,
    render: row => (row.routeCode ? `${row.routeCode} v${row.versionNo}` : '')
  },
  { title: $t('page.mes.wip.plannedQty'), key: 'plannedQty', width: 100 },
  { title: $t('page.mes.wip.releasedQty'), key: 'releasedQty', width: 90 },
  { title: $t('page.mes.wip.completedQty'), key: 'completedQty', width: 90 },
  { title: $t('page.mes.wip.priority'), key: 'priority', width: 80, render: row => priorityLabel(row.priority) },
  { title: $t('page.mes.wip.dueDate'), key: 'dueDateText', width: 120 },
  {
    title: $t('page.mes.wip.status'),
    key: 'status',
    width: 110,
    render: row => h(NTag, { size: 'small' }, { default: () => statusLabel(row.status) })
  },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 300,
    render: row => {
      const buttons: VNode[] = [];
      if (canEdit.value && row.status === 'created') {
        buttons.push(actionButton($t('common.edit'), () => openEdit(row)));
        buttons.push(actionButton($t('page.mes.wip.releaseOrder'), () => release(row), 'primary'));
      }
      if (canEdit.value && (row.status === 'released' || row.status === 'in_progress')) {
        buttons.push(actionButton($t('page.mes.wip.startLot'), () => openStart(row)));
      }
      if (canEdit.value && row.status === 'completed') {
        buttons.push(actionButton($t('page.mes.wip.closeOrder'), () => close(row), 'warning'));
      }
      if (canDelete.value && row.status === 'created') {
        buttons.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => remove(row) },
            {
              trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => $t('common.delete') }),
              default: () => $t('common.confirmDelete')
            }
          )
        );
      }
      return h(NSpace, { size: 8 }, { default: () => buttons });
    }
  }
  ];
  return all.filter(column => keepColumn(column, visible));
});

function filters() {
  const query: Array<{ name: string; exp: string; value: string; logic: string }> = [];
  if (keyword.value) query.push({ name: 'order_no', exp: 'like', value: keyword.value, logic: 'and' });
  if (statuses.value.length) query.push({ name: 'status', exp: 'in', value: statuses.value.join(','), logic: 'and' });
  if (productCode.value) query.push({ name: 'product_code', exp: 'like', value: productCode.value, logic: 'and' });
  return [...query, ...rangeColumns('due_date', dueRange.value)];
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('wipWorkOrder', {
    search: { keyword: keyword.value, statuses: statuses.value, productCode: productCode.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value,
    sort: sort.value,
    createdRange: dueRange.value
  });
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchWorkOrders({
    page: page.value - 1,
    limit: pageSize.value,
    sort: sort.value,
    columns: filters()
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipWorkOrders || [];
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
  dueRange.value = null;
  sort.value = '-id';
  search();
}

async function exportRows() {
  const { data, error } = await fetchWorkOrders({ page: 0, limit: 2000, sort: sort.value, columns: filters() });
  if (error || !data?.wipWorkOrders?.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv(
    'work-orders',
    ['orderNo', 'productCode', 'status', 'plannedQty'],
    data.wipWorkOrders.map(row => [row.orderNo, row.productCode, row.status, row.plannedQty])
  );
}

async function batchDelete() {
  const ids = checked.value.map(item => Number(item));
  const { data, error } = await mesBatch({ resource: 'wipWorkOrder', action: 'delete', ids });
  if (error || !data) return;
  checked.value = [];
  window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
  load();
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, { orderNo: '', productID: null, routeVersionID: null, plannedQty: 25, priority: 2, dueDate: '', note: '' });
  modal.value = true;
}

function openEdit(row: WorkOrderRow) {
  editingId.value = row.id;
  Object.assign(form, {
    orderNo: row.orderNo,
    productID: row.productID,
    routeVersionID: row.routeVersionID,
    plannedQty: row.plannedQty,
    priority: row.priority,
    dueDate: row.dueDateText,
    note: row.note
  });
  modal.value = true;
}

function onVersion(id: number) {
  const found = versions.value.find(item => item.versionID === id);
  form.productID = found ? found.productID : null;
}

async function submit() {
  await formRef.value?.validate();
  const { error } = await saveWorkOrder(
    {
      orderNo: form.orderNo,
      productID: form.productID,
      routeVersionID: form.routeVersionID,
      plannedQty: form.plannedQty,
      priority: form.priority,
      dueDate: form.dueDate,
      note: form.note
    },
    editingId.value || undefined
  );
  if (error) return;
  window.$message?.success($t('page.mes.wip.saved'));
  modal.value = false;
  load();
}

async function release(row: WorkOrderRow) {
  const { error } = await releaseWorkOrder(row.id);
  if (error) return;
  window.$message?.success($t('page.mes.wip.released'));
  load();
}

async function close(row: WorkOrderRow) {
  const { error } = await closeWorkOrder(row.id);
  if (error) return;
  window.$message?.success($t('page.mes.wip.closed'));
  load();
}

async function remove(row: WorkOrderRow) {
  const { error } = await removeWorkOrder(row.id);
  if (error) return;
  load();
}

function openStart(row: WorkOrderRow) {
  startOrder.value = row;
  startQty.value = Math.max(row.plannedQty - row.releasedQty, 1);
  startType.value = 'production';
  startOpen.value = true;
}

async function submitStart() {
  if (!startOrder.value) return;
  const { data, error } = await startLot(startOrder.value.id, startQty.value, startType.value);
  if (error || !data) return;
  window.$message?.success(data.lotNo);
  startOpen.value = false;
  load();
}

function onPage(next: number) {
  page.value = next;
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('wipWorkOrder');
  keyword.value = String(pref.search?.keyword || '');
  statuses.value = Array.isArray(pref.search?.statuses) ? (pref.search.statuses as string[]) : [];
  productCode.value = String(pref.search?.productCode || '');
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...dataColumnKeys];
  collapsed.value = Boolean(pref.collapsed);
  sort.value = pref.sort || '-id';
  dueRange.value = pref.createdRange || null;
  prefsReady.value = true;
  const { data } = await fetchReleasedRoutes();
  versions.value = data?.versions || [];
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NSpace class="mb-12px" wrap>
      <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
      <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
      <ColumnPicker
        :items="orderedKeys(dataColumnKeys, columnOrder).map(key => ({ key, label: key }))"
        :hidden="hidden"
        @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
        @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : [...dataColumnKeys], key, dir); persist(); }"
      />
      <NButton v-if="canDelete" :disabled="!checked.length" type="error" ghost @click="batchDelete">{{ $t('page.mes.query.batchDelete') }}</NButton>
      <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.wip.orderNo')" clearable class="w-180px" @keyup.enter="search" />
      <NInput v-model:value="productCode" :placeholder="$t('page.mes.wip.product')" clearable class="w-160px" @keyup.enter="search" />
      <NSelect
        v-model:value="statuses"
        multiple
        clearable
        class="w-220px"
        :placeholder="$t('page.mes.wip.status')"
        :options="['created', 'released', 'in_progress', 'completed', 'closed'].map(value => ({ label: statusLabel(value), value }))"
      />
      <NDatePicker v-model:value="dueRange" type="daterange" clearable />
      <NSelect v-model:value="sort" class="w-150px" :options="[{ label: 'ID ↓', value: '-id' }, { label: 'ID ↑', value: 'id' }, { label: 'No ↑', value: 'order_no' }, { label: 'Due ↑', value: 'due_date' }]" />
      <NButton type="primary" @click="search">{{ $t('common.search') }}</NButton>
      <NButton @click="reset">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable
      v-model:checked-row-keys="checked"
      remote
      :row-key="(row: WorkOrderRow) => row.id"
      :loading="loading"
      :columns="columns"
      :data="rows"
      :pagination="{ page, pageSize, itemCount: total, onUpdatePage: onPage }"
    />
    <NModal v-model:show="modal" preset="card" :title="editingId ? $t('common.edit') : $t('common.add')" class="w-560px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="120">
        <NFormItem :label="$t('page.mes.wip.orderNo')" path="orderNo">
          <NInput v-model:value="form.orderNo" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.routeVersion')" path="routeVersionID">
          <NSelect
            v-model:value="form.routeVersionID"
            :options="versionOptions"
            :placeholder="$t('page.mes.wip.selectVersion')"
            @update:value="onVersion"
          />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.plannedQty')" path="plannedQty">
          <NInputNumber v-model:value="form.plannedQty" :min="1" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.priority')">
          <NSelect v-model:value="form.priority" :options="priorityOptions" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.dueDate')">
          <NInput v-model:value="form.dueDate" placeholder="YYYY-MM-DD" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.note')">
          <NInput v-model:value="form.note" type="textarea" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="startOpen" preset="card" :title="$t('page.mes.wip.startLot')" class="w-420px">
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="$t('page.mes.wip.quantity')">
          <NInputNumber v-model:value="startQty" :min="1" class="w-full" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.lotType')">
          <NSelect v-model:value="startType" :options="lotTypeOptions" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="startOpen = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitStart">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
