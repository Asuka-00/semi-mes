<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { VNode } from 'vue';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { NButton, NPopconfirm, NSpace, NTag } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
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

const columns = computed<DataTableColumns<WorkOrderRow>>(() => [
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
]);

async function load() {
  loading.value = true;
  const { data, error } = await fetchWorkOrders({
    page: page.value - 1,
    limit: pageSize.value,
    sort: '-id',
    columns: keyword.value ? [{ name: 'order_no', exp: 'like', value: keyword.value, logic: 'and' }] : undefined
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.wipWorkOrders || [];
  total.value = Number(data.total || 0);
}

function search() {
  page.value = 1;
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
  const { data } = await fetchReleasedRoutes();
  versions.value = data?.versions || [];
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NSpace class="mb-12px" justify="space-between">
      <NSpace>
        <NInput v-model:value="keyword" :placeholder="$t('page.mes.wip.orderNo')" clearable class="w-220px" @keyup.enter="search" />
        <NButton @click="search">{{ $t('common.search') }}</NButton>
      </NSpace>
      <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
    </NSpace>
    <NDataTable
      remote
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
