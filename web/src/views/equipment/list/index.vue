<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { DataTableColumns, DataTableRowKey, FormInst, FormRules } from 'naive-ui';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, rangeColumns, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { changeEquipmentState, fetchEquipmentDetail, fetchEquipmentPage, removeEquipment, saveEquipment } from '@/service/api/equipment';
import type { EquipmentRow } from '@/service/api/equipment';
import { mesBatch, mesList } from '@/service/api/mes';

const router = useRouter();
const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('eqp:equipment:add'));
const canEdit = computed(() => hasAuth('eqp:equipment:edit'));
const canDelete = computed(() => hasAuth('eqp:equipment:delete'));

const rows = ref<EquipmentRow[]>([]);
const total = ref(0);
const page = ref(1);
const keyword = ref('');
const name = ref('');
const group = ref('');
const status = ref<string[]>([]);
const createdRange = ref<[number, number] | null>(null);
const sort = ref('equipment_code');
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const checked = ref<DataTableRowKey[]>([]);
const dataColumnKeys = ['equipmentCode', 'equipmentName', 'equipmentGroup', 'equipmentType', 'modelName', 'manufacturer', 'location', 'capacity', 'status'];
const loading = ref(false);
const modal = ref(false);
const stateOpen = ref(false);
const editingId = ref<number | null>(null);
const stateTarget = ref<EquipmentRow | null>(null);
const formRef = ref<FormInst | null>(null);
const form = reactive({
  equipmentCode: '',
  equipmentName: '',
  equipmentGroup: '',
  equipmentType: '',
  status: 'standby',
  modelName: '',
  manufacturer: '',
  serialNo: '',
  location: '',
  chamberCount: 1,
  capacity: 1,
  lineID: 0,
  capabilities: [] as Array<{ operationID: number | null; recipeID: number | null }>
});
const stateForm = reactive({ toState: 'standby', reasonCode: 'OTHER', reason: '' });

const states = ['standby', 'engineering', 'scheduled_down', 'unscheduled_down', 'non_scheduled', 'productive'];
const reasons = ['ENG_SETUP', 'ENG_DONE', 'PM_START', 'PM_DONE', 'BREAKDOWN', 'REPAIR_DONE', 'NO_WIP', 'SHIFT_END', 'SHIFT_START', 'OTHER'];

function stateLabel(value: string) {
  if (!value) return '';
  return $t(`page.mes.eqp.state_${value}` as App.I18n.I18nKey);
}

const operations = ref<Array<{ id: number; operationCode: string; operationName: string }>>([]);
const recipes = ref<Array<{ id: number; operationID: number; recipeCode: string; recipeName: string }>>([]);
const operationOptions = computed(() =>
  operations.value.map(item => ({ label: `${item.operationCode} ${item.operationName}`, value: item.id }))
);

function recipeOptions(operationID: number | null) {
  const matched = recipes.value
    .filter(item => item.operationID === operationID)
    .map(item => ({ label: `${item.recipeCode} ${item.recipeName}`, value: item.id }));
  return [{ label: $t('page.mes.eqp.anyRecipe'), value: 0 }, ...matched];
}

async function loadRefs() {
  if (operations.value.length) return;
  const [opRes, recipeRes] = await Promise.all([
    mesList('baseOperation', { page: 0, limit: 200 }),
    mesList('baseRecipe', { page: 0, limit: 200 })
  ]);
  operations.value = opRes.data?.baseOperations || [];
  recipes.value = recipeRes.data?.baseRecipes || [];
}

const stateOptions = computed(() => states.map(value => ({ label: stateLabel(value), value })));
const reasonOptions = computed(() => reasons.map(value => ({ label: value, value })));
const rules = computed<FormRules>(() => ({
  equipmentCode: { required: true, message: $t('form.required'), trigger: 'blur' },
  equipmentName: { required: true, message: $t('form.required'), trigger: 'blur' }
}));

function filters() {
  const columns = [];
  if (keyword.value) columns.push({ name: 'equipment_code', exp: 'like', value: keyword.value });
  if (name.value) columns.push({ name: 'equipment_name', exp: 'like', value: name.value });
  if (group.value) columns.push({ name: 'equipment_group', exp: 'like', value: group.value });
  if (status.value.length) columns.push({ name: 'status', exp: 'in', value: status.value.join(',') });
  return [...columns, ...rangeColumns('created_at', createdRange.value)];
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('eqpEquipment', {
    search: { keyword: keyword.value, name: name.value, group: group.value, status: status.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value,
    sort: sort.value,
    createdRange: createdRange.value
  });
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchEquipmentPage({ page: page.value - 1, limit: 10, sort: sort.value, columns: filters() });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.eqpEquipments || [];
  total.value = Number(data.total || 0);
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, {
    equipmentCode: '',
    equipmentName: '',
    equipmentGroup: '',
    equipmentType: '',
    status: 'standby',
    modelName: '',
    manufacturer: '',
    serialNo: '',
    location: '',
    chamberCount: 1,
    capacity: 1,
    lineID: 0,
    capabilities: []
  });
  loadRefs();
  modal.value = true;
}

async function openEdit(row: EquipmentRow) {
  editingId.value = row.id;
  Object.assign(form, {
    equipmentCode: row.equipmentCode,
    equipmentName: row.equipmentName,
    equipmentGroup: row.equipmentGroup,
    equipmentType: row.equipmentType,
    status: row.status,
    modelName: row.modelName,
    manufacturer: row.manufacturer,
    serialNo: row.serialNo,
    location: row.location,
    chamberCount: row.chamberCount || 1,
    capacity: row.capacity || 1,
    lineID: row.lineID || 0,
    capabilities: []
  });
  loadRefs();
  const { data } = await fetchEquipmentDetail(row.id);
  form.capabilities = (data?.capabilities || []).map(item => ({
    operationID: item.operationID,
    recipeID: item.recipeID || 0
  }));
  modal.value = true;
}

async function submit() {
  await formRef.value?.validate();
  const capabilities = form.capabilities
    .filter(item => item.operationID)
    .map(item => ({ operationID: item.operationID, recipeID: item.recipeID || 0 }));
  const { error } = await saveEquipment({ ...form, capabilities }, editingId.value || undefined);
  if (error) return;
  modal.value = false;
  load();
}

async function submitState() {
  if (!stateTarget.value) return;
  const { error } = await changeEquipmentState(stateTarget.value.id, { ...stateForm });
  if (error) return;
  stateOpen.value = false;
  load();
}

const columns = computed<DataTableColumns<EquipmentRow>>(() => {
  const visible = new Set(orderedKeys(dataColumnKeys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  const all: DataTableColumns<EquipmentRow> = [
  { type: 'selection' },
  { title: $t('page.mes.eqp.code'), key: 'equipmentCode', minWidth: 120 },
  { title: $t('page.mes.eqp.name'), key: 'equipmentName', minWidth: 160 },
  { title: $t('page.mes.eqp.group'), key: 'equipmentGroup', width: 100 },
  { title: $t('page.mes.eqp.type'), key: 'equipmentType', width: 100 },
  { title: $t('page.mes.eqp.model'), key: 'modelName', width: 110 },
  { title: $t('page.mes.eqp.vendor'), key: 'manufacturer', width: 110 },
  { title: $t('page.mes.eqp.location'), key: 'location', minWidth: 100 },
  { title: $t('page.mes.eqp.capacity'), key: 'capacity', width: 80 },
  {
    title: $t('page.mes.eqp.state'),
    key: 'status',
    width: 140,
    render: row => h(NTag, { size: 'small' }, { default: () => stateLabel(row.status) })
  },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 240,
    render: row =>
      h('div', { class: 'flex gap-8px' }, [
        h(NButton, { size: 'small', onClick: () => router.push({ name: 'equipment_detail', params: { id: String(row.id) } }) }, { default: () => $t('page.mes.eqp.detail') }),
        canEdit.value ? h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => $t('common.edit') }) : null,
        canEdit.value
          ? h(
              NButton,
              {
                size: 'small',
                onClick: () => {
                  stateTarget.value = row;
                  const next: Record<string, string> = {
                    standby: 'engineering',
                    engineering: 'standby',
                    scheduled_down: 'standby',
                    unscheduled_down: 'standby',
                    non_scheduled: 'standby',
                    productive: 'unscheduled_down'
                  };
                  stateForm.toState = next[row.status] || 'standby';
                  stateForm.reasonCode = row.status === 'productive' ? 'BREAKDOWN' : row.status === 'engineering' ? 'ENG_DONE' : 'ENG_SETUP';
                  stateForm.reason = '';
                  stateOpen.value = true;
                }
              },
              { default: () => $t('page.mes.eqp.changeState') }
            )
          : null,
        canDelete.value
          ? h(
              NPopconfirm,
              { onPositiveClick: async () => { await removeEquipment(row.id); load(); } },
              { trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => $t('common.delete') }), default: () => $t('common.confirmDelete') }
            )
          : null
      ])
  }
  ];
  return all.filter(column => keepColumn(column, visible));
});

async function exportRows() {
  const { data, error } = await fetchEquipmentPage({ page: 0, limit: 2000, sort: sort.value, columns: filters() });
  if (error || !data?.eqpEquipments?.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv(
    'equipment',
    ['code', 'name', 'group', 'status'],
    data.eqpEquipments.map(row => [row.equipmentCode, row.equipmentName, row.equipmentGroup, row.status])
  );
}

async function batchDelete() {
  const ids = checked.value.map(item => Number(item));
  const { data, error } = await mesBatch({ resource: 'eqpEquipment', action: 'delete', ids });
  if (error || !data) return;
  checked.value = [];
  window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
  load();
}

function reset() {
  keyword.value = '';
  name.value = '';
  group.value = '';
  status.value = [];
  createdRange.value = null;
  sort.value = 'equipment_code';
  page.value = 1;
  persist();
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('eqpEquipment');
  keyword.value = String(pref.search?.keyword || '');
  name.value = String(pref.search?.name || '');
  group.value = String(pref.search?.group || '');
  status.value = Array.isArray(pref.search?.status) ? (pref.search.status as string[]) : [];
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...dataColumnKeys];
  collapsed.value = Boolean(pref.collapsed);
  sort.value = pref.sort || 'equipment_code';
  createdRange.value = pref.createdRange || null;
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('route.equipment_list')">
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
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.eqp.code')" class="w-160px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="name" :placeholder="$t('page.mes.eqp.name')" class="w-160px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="group" :placeholder="$t('page.mes.eqp.group')" class="w-140px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NSelect v-model:value="status" :options="stateOptions" multiple clearable class="w-220px" :placeholder="$t('page.mes.eqp.state')" />
      <NDatePicker v-model:value="createdRange" type="daterange" clearable />
      <NSelect v-model:value="sort" class="w-160px" :options="[{ label: 'Code ↑', value: 'equipment_code' }, { label: 'Code ↓', value: '-equipment_code' }, { label: 'Name ↑', value: 'equipment_name' }]" />
      <NButton type="primary" @click="page = 1; persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="reset">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable v-model:checked-row-keys="checked" remote :row-key="(row: EquipmentRow) => row.id" :loading="loading" :columns="columns" :data="rows" :scroll-x="1400" :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }" />
    <NModal v-model:show="modal" preset="card" :title="editingId ? $t('common.edit') : $t('common.add')" class="w-640px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="100">
        <NFormItem :label="$t('page.mes.eqp.code')" path="equipmentCode"><NInput v-model:value="form.equipmentCode" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.name')" path="equipmentName"><NInput v-model:value="form.equipmentName" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.group')"><NInput v-model:value="form.equipmentGroup" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.type')"><NInput v-model:value="form.equipmentType" /></NFormItem>
        <NFormItem v-if="!editingId" :label="$t('page.mes.eqp.state')"><NSelect v-model:value="form.status" :options="stateOptions.filter(item => item.value !== 'productive')" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.model')"><NInput v-model:value="form.modelName" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.vendor')"><NInput v-model:value="form.manufacturer" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.serial')"><NInput v-model:value="form.serialNo" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.location')"><NInput v-model:value="form.location" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.line')"><NInputNumber v-model:value="form.lineID" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.chambers')"><NInputNumber v-model:value="form.chamberCount" :min="1" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.capacity')"><NInputNumber v-model:value="form.capacity" :min="1" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.capability')">
          <div class="w-full flex flex-col gap-8px">
            <div v-for="(cap, index) in form.capabilities" :key="index" class="flex gap-8px">
              <NSelect v-model:value="cap.operationID" :options="operationOptions" filterable class="flex-1" />
              <NSelect v-model:value="cap.recipeID" :options="recipeOptions(cap.operationID)" class="w-180px" />
              <NButton @click="form.capabilities.splice(index, 1)">{{ $t('common.delete') }}</NButton>
            </div>
            <NButton @click="form.capabilities.push({ operationID: null, recipeID: 0 })">{{ $t('common.add') }}</NButton>
            <span class="text-12px opacity-70">{{ $t('page.mes.eqp.capabilityHint') }}</span>
          </div>
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="stateOpen" preset="card" :title="$t('page.mes.eqp.changeState')" class="w-480px">
      <NForm label-placement="left" label-width="90">
        <NFormItem :label="$t('page.mes.eqp.state')"><NSelect v-model:value="stateForm.toState" :options="stateOptions.filter(item => item.value !== 'productive')" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.reason')"><NSelect v-model:value="stateForm.reasonCode" :options="reasonOptions" /></NFormItem>
        <NFormItem :label="$t('page.mes.wip.reason')"><NInput v-model:value="stateForm.reason" /></NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="stateOpen = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitState">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
