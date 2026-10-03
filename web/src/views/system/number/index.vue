<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { NButton, NPopconfirm, NSpace } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { previewNumber } from '@/hooks/business/dict';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { mesCreate, mesDelete, mesList, mesUpdate } from '@/service/api/mes';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('system:number:add'));
const canEdit = computed(() => hasAuth('system:number:edit'));
const canDelete = computed(() => hasAuth('system:number:delete'));
const keyword = ref('');
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
const preview = ref('');
const keys = ['ruleCode', 'ruleName', 'prefix', 'datePart', 'seqLength', 'resetPeriod', 'separator', 'status'];
const form = reactive({
  ruleCode: '',
  ruleName: '',
  prefix: '',
  datePart: 'yyyyMMdd',
  seqLength: 3,
  resetPeriod: 'daily',
  separator: '-',
  status: 1
});

const dateOptions = [
  { label: 'none', value: 'none' },
  { label: 'yyyy', value: 'yyyy' },
  { label: 'yyyyMM', value: 'yyyyMM' },
  { label: 'yyyyMMdd', value: 'yyyyMMdd' }
];
const resetOptions = computed(() => [
  { label: $t('page.mes.number.daily'), value: 'daily' },
  { label: $t('page.mes.number.monthly'), value: 'monthly' },
  { label: $t('page.mes.number.yearly'), value: 'yearly' },
  { label: $t('page.mes.number.never'), value: 'never' }
]);
const rules: FormRules = {
  ruleCode: { required: true, message: $t('form.required'), trigger: 'blur' },
  ruleName: { required: true, message: $t('form.required'), trigger: 'blur' }
};

const allColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.number.ruleCode'), key: 'ruleCode', minWidth: 140 },
  { title: $t('page.mes.number.ruleName'), key: 'ruleName', minWidth: 140 },
  { title: $t('page.mes.number.prefix'), key: 'prefix', width: 100 },
  { title: $t('page.mes.number.datePart'), key: 'datePart', width: 120 },
  { title: $t('page.mes.number.seqLength'), key: 'seqLength', width: 100 },
  { title: $t('page.mes.number.reset'), key: 'resetPeriod', width: 120, render: row => resetOptions.value.find(item => item.value === row.resetPeriod)?.label || row.resetPeriod },
  { title: $t('page.mes.number.separator'), key: 'separator', width: 80 },
  { title: $t('page.mes.field.status'), key: 'status', width: 90, render: row => (Number(row.status) === 1 ? $t('page.mes.enabled') : $t('page.mes.disabled')) },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 220,
    render: row =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => showPreview(String(row.ruleCode)) }, { default: () => $t('page.mes.number.preview') }),
          canEdit.value ? h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => $t('common.edit') }) : null,
          canDelete.value
            ? h(NPopconfirm, { onPositiveClick: () => remove(Number(row.id)) }, {
                trigger: () => h(NButton, { size: 'tiny', type: 'error' }, { default: () => $t('common.delete') }),
                default: () => $t('common.confirmDelete')
              })
            : null
        ]
      })
  }
]);
const columns = computed(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  return allColumns.value.filter(column => keepColumn(column, visible, ['actions']));
});

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('sysNumberRule', {
    search: { keyword: keyword.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value
  });
}

async function load() {
  loading.value = true;
  const columnsQuery = keyword.value ? [{ name: 'rule_code', exp: 'like', value: keyword.value, logic: 'and' }] : undefined;
  const { data, error } = await mesList('sysNumberRule', { page: page.value - 1, limit: 10, sort: 'id', columns: columnsQuery });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.sysNumberRules || [];
  total.value = Number(data.total || 0);
}

function openCreate() {
  editingId.value = null;
  preview.value = '';
  Object.assign(form, { ruleCode: '', ruleName: '', prefix: '', datePart: 'yyyyMMdd', seqLength: 3, resetPeriod: 'daily', separator: '-', status: 1 });
  modal.value = true;
}

function openEdit(row: Record<string, any>) {
  editingId.value = Number(row.id);
  preview.value = '';
  Object.assign(form, row);
  modal.value = true;
}

async function save() {
  await formRef.value?.validate();
  const payload = { ...form };
  const { error } = editingId.value ? await mesUpdate('sysNumberRule', editingId.value, payload) : await mesCreate('sysNumberRule', payload);
  if (error) return;
  modal.value = false;
  load();
}

async function remove(id: number) {
  const { error } = await mesDelete('sysNumberRule', id);
  if (error) return;
  load();
}

async function showPreview(ruleCode: string) {
  const { data, error } = await previewNumber(ruleCode);
  if (error || !data) return;
  preview.value = data.preview;
  window.$message?.success(data.preview);
}

async function exportRows() {
  const { data, error } = await mesList('sysNumberRule', { page: 0, limit: 2000, sort: 'id' });
  const list = data?.sysNumberRules || [];
  if (error || !list.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('number-rules', keys, list.map((row: Record<string, any>) => keys.map(key => row[key])));
}

onMounted(async () => {
  const pref = await loadPagePref('sysNumberRule');
  keyword.value = String(pref.search?.keyword || '');
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.number.title')">
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
    <NSpace v-show="!collapsed" class="mb-12px">
      <NInput v-model:value="keyword" :placeholder="$t('page.mes.number.ruleCode')" class="w-180px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NButton type="primary" @click="page = 1; persist(); load()">{{ $t('common.search') }}</NButton>
    </NSpace>
    <NDataTable remote :loading="loading" :columns="columns" :data="rows" :scroll-x="980" :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }" />
    <div v-if="preview" class="mt-8px">{{ $t('page.mes.number.preview') }}: {{ preview }}</div>
    <NModal v-model:show="modal" preset="card" :title="editingId ? $t('common.edit') : $t('common.add')" class="w-520px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="110">
        <NFormItem :label="$t('page.mes.number.ruleCode')" path="ruleCode"><NInput v-model:value="form.ruleCode" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.ruleName')" path="ruleName"><NInput v-model:value="form.ruleName" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.prefix')"><NInput v-model:value="form.prefix" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.datePart')"><NSelect v-model:value="form.datePart" :options="dateOptions" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.seqLength')"><NInputNumber v-model:value="form.seqLength" :min="1" :max="12" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.reset')"><NSelect v-model:value="form.resetPeriod" :options="resetOptions" /></NFormItem>
        <NFormItem :label="$t('page.mes.number.separator')"><NInput v-model:value="form.separator" /></NFormItem>
        <NFormItem :label="$t('page.mes.field.status')">
          <NSelect v-model:value="form.status" :options="[{ label: $t('page.mes.enabled'), value: 1 }, { label: $t('page.mes.disabled'), value: 2 }]" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="save">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
