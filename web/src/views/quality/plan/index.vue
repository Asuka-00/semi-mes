<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { mesBatch } from '@/service/api/mes';
import { fetchPlans, removePlan, savePlan } from '@/service/api/quality';
import type { InspectPlan } from '@/service/api/quality';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('qc:plan:add'));
const canEdit = computed(() => hasAuth('qc:plan:edit'));
const canDelete = computed(() => hasAuth('qc:plan:delete'));
const rows = ref<InspectPlan[]>([]);
const keyword = ref('');
const enabled = ref<string[]>([]);
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const checked = ref<DataTableRowKey[]>([]);
const keys = ['planName', 'operationCode', 'param', 'enabled'];
const modal = ref(false);
const editingId = ref<number | null>(null);
const form = reactive({
  planName: '',
  operationID: 0,
  productID: 0,
  enabled: true,
  paramCode: 'CD',
  paramName: '',
  unit: 'nm',
  target: 500,
  lsl: 470,
  usl: 530,
  sampleSize: 1,
  required: true
});

function filters() {
  const columns = [];
  if (keyword.value) columns.push({ name: 'plan_name', exp: 'like', value: keyword.value });
  if (enabled.value.length === 1) columns.push({ name: 'enabled', exp: '=', value: enabled.value[0] });
  return columns;
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('qcInspectPlan', {
    search: { keyword: keyword.value, enabled: enabled.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value
  });
}

async function load() {
  const { data, error } = await fetchPlans({ page: 0, limit: 200, columns: filters() });
  if (error || !data) return;
  rows.value = data.plans || [];
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, {
    planName: '',
    operationID: 0,
    productID: 0,
    enabled: true,
    paramCode: 'CD',
    paramName: '',
    unit: 'nm',
    target: 500,
    lsl: 470,
    usl: 530,
    sampleSize: 1,
    required: true
  });
  modal.value = true;
}

function openEdit(row: InspectPlan) {
  const item = row.items?.[0];
  editingId.value = row.id;
  Object.assign(form, {
    planName: row.planName,
    operationID: row.operationID,
    productID: row.productID || 0,
    enabled: row.enabled,
    paramCode: item?.paramCode || 'CD',
    paramName: item?.paramName || '',
    unit: item?.unit || '',
    target: item?.target ?? 0,
    lsl: item?.lsl ?? 0,
    usl: item?.usl ?? 0,
    sampleSize: item?.sampleSize || 1,
    required: item?.required !== false
  });
  modal.value = true;
}

async function submit() {
  const { error } = await savePlan(
    {
      planName: form.planName,
      operationID: Number(form.operationID),
      productID: Number(form.productID || 0),
      enabled: form.enabled,
      items: [
        {
          paramCode: form.paramCode,
          paramName: form.paramName,
          unit: form.unit,
          target: form.target,
          lsl: form.lsl,
          usl: form.usl,
          sampleSize: form.sampleSize,
          required: form.required
        }
      ]
    },
    editingId.value || undefined
  );
  if (error) return;
  modal.value = false;
  load();
}

async function remove(id: number) {
  const { error } = await removePlan(id);
  if (error) return;
  load();
}

const columns = computed<DataTableColumns<InspectPlan>>(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  const all: DataTableColumns<InspectPlan> = [
  { type: 'selection' },
  { title: $t('page.mes.qc.plan'), key: 'planName' },
  { title: $t('page.mes.qc.operation'), key: 'operationCode' },
  { title: $t('page.mes.qc.param'), key: 'param', render: row => row.items?.map(item => item.paramCode).join(', ') },
  {
    title: $t('page.mes.qc.enabled'),
    key: 'enabled',
    render: row => (row.enabled ? $t('page.mes.qc.yes') : $t('page.mes.qc.no'))
  },
  {
    title: $t('common.action'),
    key: 'action',
    render: row =>
      hActions(row)
  }
  ];
  return all.filter(column => keepColumn(column, visible));
});

function hActions(row: InspectPlan) {
  return [
    canEdit.value ? hBtn(() => openEdit(row), $t('common.edit')) : null,
    canDelete.value
      ? hPop(() => remove(row.id))
      : null
  ];
}

function hBtn(onClick: () => void, label: string) {
  return h(NButton, { size: 'small', onClick }, { default: () => label });
}

function hPop(onPositiveClick: () => void) {
  return h(
    NPopconfirm,
    { onPositiveClick },
    {
      trigger: () => h(NButton, { size: 'small', type: 'error', class: 'ml-8px' }, { default: () => $t('common.delete') }),
      default: () => $t('common.confirmDelete')
    }
  );
}

async function exportRows() {
  if (!rows.value.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('inspect-plans', ['planName', 'operationCode'], rows.value.map(row => [row.planName, row.operationCode || '']));
}

async function runBatch(action: 'delete' | 'enable' | 'disable') {
  const ids = checked.value.map(item => Number(item));
  const { data, error } = await mesBatch({ resource: 'qcInspectPlan', action, ids });
  if (error || !data) return;
  checked.value = [];
  window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('qcInspectPlan');
  keyword.value = String(pref.search?.keyword || '');
  enabled.value = Array.isArray(pref.search?.enabled) ? (pref.search.enabled as string[]) : [];
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.plan')">
    <NSpace class="mb-12px" wrap>
      <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
      <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
      <ColumnPicker
        :items="orderedKeys(keys, columnOrder).map(key => ({ key, label: key }))"
        :hidden="hidden"
        @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
        @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : [...keys], key, dir); persist(); }"
      />
      <NButton v-if="canEdit" :disabled="!checked.length" @click="runBatch('enable')">{{ $t('page.mes.query.batchEnable') }}</NButton>
      <NButton v-if="canEdit" :disabled="!checked.length" @click="runBatch('disable')">{{ $t('page.mes.query.batchDisable') }}</NButton>
      <NButton v-if="canDelete" :disabled="!checked.length" type="error" ghost @click="runBatch('delete')">{{ $t('page.mes.query.batchDelete') }}</NButton>
      <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="keyword" class="w-180px" clearable :placeholder="$t('page.mes.qc.plan')" @keyup.enter="persist(); load()" />
      <NSelect v-model:value="enabled" multiple clearable class="w-160px" :options="[{ label: $t('page.mes.enabled'), value: '1' }, { label: $t('page.mes.disabled'), value: '0' }]" />
      <NButton type="primary" @click="persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="keyword = ''; enabled = []; persist(); load()">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable v-model:checked-row-keys="checked" :columns="columns" :data="rows" :row-key="(row: InspectPlan) => row.id" />
    <NModal v-model:show="modal" preset="card" class="w-560px" :title="$t('page.mes.qc.plan')">
      <NForm label-placement="left" label-width="110">
        <NFormItem :label="$t('page.mes.qc.plan')"><NInput v-model:value="form.planName" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.operation')"><NInputNumber v-model:value="form.operationID" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.product')"><NInputNumber v-model:value="form.productID" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.param')"><NInput v-model:value="form.paramCode" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.paramName')"><NInput v-model:value="form.paramName" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.unit')"><NInput v-model:value="form.unit" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.target')"><NInputNumber v-model:value="form.target" class="w-full" /></NFormItem>
        <NFormItem label="LSL"><NInputNumber v-model:value="form.lsl" class="w-full" /></NFormItem>
        <NFormItem label="USL"><NInputNumber v-model:value="form.usl" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.sample')"><NInputNumber v-model:value="form.sampleSize" :min="1" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.required')"><NSwitch v-model:value="form.required" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.enabled')"><NSwitch v-model:value="form.enabled" /></NFormItem>
      </NForm>
      <template #footer>
        <NButton type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
      </template>
    </NModal>
  </NCard>
</template>
