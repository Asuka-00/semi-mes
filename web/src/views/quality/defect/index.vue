<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, loadPagePref, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { mesBatch } from '@/service/api/mes';
import { fetchDefectCodes, fetchDefects, recordDefect, saveDefectCode } from '@/service/api/quality';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('qc:defect:add'));
const codes = ref<Array<Record<string, any>>>([]);
const defects = ref<Array<Record<string, any>>>([]);
const codeKeyword = ref('');
const category = ref('');
const lotNo = ref('');
const disposition = ref<string[]>([]);
const collapsed = ref(false);
const prefsReady = ref(false);
const checked = ref<DataTableRowKey[]>([]);
const codeModal = ref(false);
const codeForm = reactive({ defectCode: '', defectName: '', category: '', severity: 'major' });
const form = reactive({
  lotId: 0,
  defectCode: 'PARTICLE',
  quantity: 1,
  disposition: 'use_as_is',
  note: ''
});

const dispositions = computed(() => [
  { label: $t('page.mes.qc.useAsIs'), value: 'use_as_is' },
  { label: $t('page.mes.qc.rework'), value: 'rework' },
  { label: $t('page.mes.qc.scrap'), value: 'scrap' },
  { label: $t('page.mes.qc.hold'), value: 'hold' }
]);

function codeFilters() {
  const columns = [];
  if (codeKeyword.value) columns.push({ name: 'defect_code', exp: 'like', value: codeKeyword.value });
  if (category.value) columns.push({ name: 'category', exp: 'like', value: category.value });
  return columns;
}

function defectFilters() {
  const columns = [];
  if (lotNo.value) columns.push({ name: 'lot_no', exp: 'like', value: lotNo.value });
  if (codeKeyword.value) columns.push({ name: 'defect_code', exp: 'like', value: codeKeyword.value });
  if (disposition.value.length) columns.push({ name: 'disposition', exp: 'in', value: disposition.value.join(',') });
  return columns;
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('qcDefect', {
    search: { codeKeyword: codeKeyword.value, category: category.value, lotNo: lotNo.value, disposition: disposition.value },
    collapsed: collapsed.value
  });
}

async function load() {
  const [codeRes, defectRes] = await Promise.all([
    fetchDefectCodes({ page: 0, limit: 200, columns: codeFilters() }),
    fetchDefects(undefined, { page: 0, limit: 200, columns: defectFilters() })
  ]);
  if (!codeRes.error && codeRes.data) codes.value = codeRes.data.codes || [];
  if (!defectRes.error && defectRes.data) defects.value = defectRes.data.defects || [];
  if (!form.defectCode && codes.value[0]) form.defectCode = codes.value[0].defectCode;
}

async function addCode() {
  const { error } = await saveDefectCode({ ...codeForm, status: 1 });
  if (error) return;
  codeModal.value = false;
  load();
}

async function submit() {
  const { error } = await recordDefect({ ...form, lotId: Number(form.lotId), quantity: Number(form.quantity) });
  if (error) return;
  window.$message?.success($t('page.mes.qc.saved'));
  load();
}

async function exportRows() {
  if (!defects.value.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('defects', ['lotNo', 'defectCode', 'quantity', 'disposition'], defects.value.map(row => [row.lotNo, row.defectCode, row.quantity, row.disposition]));
}

async function batchDisable() {
  const ids = checked.value.map(item => Number(item));
  const { data, error } = await mesBatch({ resource: 'qcDefectCode', action: 'disable', ids });
  if (error || !data) return;
  checked.value = [];
  window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
  load();
}

const codeColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { type: 'selection' },
  { title: $t('page.mes.qc.code'), key: 'defectCode' },
  { title: $t('page.mes.qc.name'), key: 'defectName' },
  { title: $t('page.mes.qc.category'), key: 'category' },
  { title: $t('page.mes.qc.severity'), key: 'severity' }
]);

const defectColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.track.lotNo'), key: 'lotNo' },
  { title: $t('page.mes.qc.code'), key: 'defectCode' },
  { title: $t('page.mes.track.qty'), key: 'quantity' },
  { title: $t('page.mes.qc.disposition'), key: 'disposition' },
  { title: $t('page.mes.wip.reason'), key: 'note' }
]);

onMounted(async () => {
  const pref = await loadPagePref('qcDefect');
  codeKeyword.value = String(pref.search?.codeKeyword || '');
  category.value = String(pref.search?.category || '');
  lotNo.value = String(pref.search?.lotNo || '');
  disposition.value = Array.isArray(pref.search?.disposition) ? (pref.search.disposition as string[]) : [];
  collapsed.value = Boolean(pref.collapsed);
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.defect')">
    <NSpace class="mb-12px" wrap>
      <NButton @click="collapsed = !collapsed; persist()">{{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}</NButton>
      <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
      <NButton v-if="canAdd" :disabled="!checked.length" @click="batchDisable">{{ $t('page.mes.query.batchDisable') }}</NButton>
      <NButton v-if="canAdd" @click="codeModal = true">{{ $t('page.mes.qc.addCode') }}</NButton>
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="codeKeyword" class="w-160px" clearable :placeholder="$t('page.mes.qc.code')" @keyup.enter="persist(); load()" />
      <NInput v-model:value="category" class="w-140px" clearable :placeholder="$t('page.mes.qc.category')" @keyup.enter="persist(); load()" />
      <NInput v-model:value="lotNo" class="w-160px" clearable :placeholder="$t('page.mes.track.lotNo')" @keyup.enter="persist(); load()" />
      <NSelect v-model:value="disposition" multiple clearable class="w-220px" :options="dispositions" :placeholder="$t('page.mes.qc.disposition')" />
      <NButton type="primary" @click="persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="codeKeyword = ''; category = ''; lotNo = ''; disposition = []; persist(); load()">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable v-model:checked-row-keys="checked" :row-key="(row: Record<string, any>) => row.id" :columns="codeColumns" :data="codes" class="mb-16px" />
    <NForm v-if="canAdd" inline>
      <NFormItem :label="$t('page.mes.qc.lotId')"><NInputNumber v-model:value="form.lotId" /></NFormItem>
      <NFormItem :label="$t('page.mes.qc.code')">
        <NSelect v-model:value="form.defectCode" class="w-160px" :options="codes.map(item => ({ label: item.defectCode, value: item.defectCode }))" />
      </NFormItem>
      <NFormItem :label="$t('page.mes.track.qty')"><NInputNumber v-model:value="form.quantity" :min="1" /></NFormItem>
      <NFormItem :label="$t('page.mes.qc.disposition')">
        <NSelect v-model:value="form.disposition" class="w-160px" :options="dispositions" />
      </NFormItem>
      <NFormItem :label="$t('page.mes.wip.reason')"><NInput v-model:value="form.note" /></NFormItem>
      <NButton type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
    </NForm>
    <NDataTable :columns="defectColumns" :data="defects" class="mt-16px" />
    <NModal v-model:show="codeModal" preset="card" class="w-480px" :title="$t('page.mes.qc.addCode')">
      <NForm label-placement="left" label-width="90">
        <NFormItem :label="$t('page.mes.qc.code')"><NInput v-model:value="codeForm.defectCode" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.name')"><NInput v-model:value="codeForm.defectName" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.category')"><NInput v-model:value="codeForm.category" /></NFormItem>
        <NFormItem :label="$t('page.mes.qc.severity')"><NInput v-model:value="codeForm.severity" /></NFormItem>
      </NForm>
      <template #footer>
        <NButton type="primary" @click="addCode">{{ $t('common.confirm') }}</NButton>
      </template>
    </NModal>
  </NCard>
</template>
