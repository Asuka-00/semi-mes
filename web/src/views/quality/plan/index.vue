<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { fetchPlans, removePlan, savePlan } from '@/service/api/quality';
import type { InspectPlan } from '@/service/api/quality';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('qc:plan:add'));
const canEdit = computed(() => hasAuth('qc:plan:edit'));
const canDelete = computed(() => hasAuth('qc:plan:delete'));
const rows = ref<InspectPlan[]>([]);
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

async function load() {
  const { data, error } = await fetchPlans();
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

const columns = computed<DataTableColumns<InspectPlan>>(() => [
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
]);

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

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.plan')">
    <template #header-extra>
      <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
    </template>
    <NDataTable :columns="columns" :data="rows" :row-key="(row: InspectPlan) => row.id" />
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
