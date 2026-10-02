<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { fetchDefectCodes, fetchDefects, recordDefect, saveDefectCode } from '@/service/api/quality';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('qc:defect:add'));
const codes = ref<Array<Record<string, any>>>([]);
const defects = ref<Array<Record<string, any>>>([]);
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

async function load() {
  const [codeRes, defectRes] = await Promise.all([fetchDefectCodes(), fetchDefects()]);
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

const codeColumns = computed<DataTableColumns<Record<string, any>>>(() => [
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

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.defect')">
    <template #header-extra>
      <NButton v-if="canAdd" @click="codeModal = true">{{ $t('page.mes.qc.addCode') }}</NButton>
    </template>
    <NDataTable :columns="codeColumns" :data="codes" class="mb-16px" />
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
