<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { fetchPmPlans, removePmPlan, savePmPlan } from '@/service/api/equipment';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('eqp:pm:add'));
const canEdit = computed(() => hasAuth('eqp:pm:edit'));
const canDelete = computed(() => hasAuth('eqp:pm:delete'));
const rows = ref<Array<Record<string, any>>>([]);
const modal = ref(false);
const editingId = ref<number | null>(null);
const form = reactive({
  planName: '',
  equipmentID: 0,
  equipmentGroup: '',
  triggerType: 'time',
  intervalDays: 30,
  intervalCount: 0,
  checklistText: '',
  blockTrackIn: false,
  nextDueAt: '',
  enabled: true
});

async function load() {
  const { data, error } = await fetchPmPlans();
  if (error || !data) return;
  rows.value = data.plans || [];
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, { planName: '', equipmentID: 0, equipmentGroup: '', triggerType: 'time', intervalDays: 30, intervalCount: 0, checklistText: '', blockTrackIn: false, nextDueAt: '', enabled: true });
  modal.value = true;
}

function openEdit(row: Record<string, any>) {
  editingId.value = row.id;
  let names: string[] = [];
  try {
    const parsed = JSON.parse(row.checklistJson || '[]');
    names = Array.isArray(parsed) ? parsed.map((item: string | { name: string }) => (typeof item === 'string' ? item : item.name)) : [];
  } catch {
    names = [];
  }
  Object.assign(form, {
    planName: row.planName,
    equipmentID: row.equipmentID || 0,
    equipmentGroup: row.equipmentGroup || '',
    triggerType: row.triggerType,
    intervalDays: row.intervalDays || 0,
    intervalCount: row.intervalCount || 0,
    checklistText: names.join('\n'),
    blockTrackIn: !!row.blockTrackIn,
    nextDueAt: row.nextDueAt ? String(row.nextDueAt).slice(0, 10) : '',
    enabled: row.enabled !== false
  });
  modal.value = true;
}

async function submit() {
  const checklist = form.checklistText.split('\n').map(item => item.trim()).filter(Boolean);
  const { error } = await savePmPlan(
    {
      planName: form.planName,
      equipmentID: Number(form.equipmentID || 0),
      equipmentGroup: form.equipmentGroup,
      triggerType: form.triggerType,
      intervalDays: form.intervalDays,
      intervalCount: form.intervalCount,
      checklist,
      blockTrackIn: form.blockTrackIn,
      nextDueAt: form.nextDueAt,
      enabled: form.enabled
    },
    editingId.value || undefined
  );
  if (error) return;
  modal.value = false;
  load();
}

const columns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.eqp.plan'), key: 'planName', minWidth: 160 },
  { title: $t('page.mes.eqp.group'), key: 'equipmentGroup', width: 100 },
  { title: $t('page.mes.eqp.trigger'), key: 'triggerType', width: 90 },
  { title: $t('page.mes.eqp.intervalDays'), key: 'intervalDays', width: 90 },
  { title: $t('page.mes.eqp.intervalCount'), key: 'intervalCount', width: 90 },
  { title: $t('page.mes.eqp.lotsSince'), key: 'lotsSince', width: 90 },
  { title: $t('page.mes.eqp.block'), key: 'blockTrackIn', width: 90, render: row => (row.blockTrackIn ? $t('common.yesOrNo.yes') : $t('common.yesOrNo.no')) },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 160,
    render: row =>
      h('div', { class: 'flex gap-8px' }, [
        canEdit.value ? h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => $t('common.edit') }) : null,
        canDelete.value
          ? h(NPopconfirm, { onPositiveClick: async () => { await removePmPlan(row.id); load(); } }, {
              trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => $t('common.delete') }),
              default: () => $t('common.confirmDelete')
            })
          : null
      ])
  }
]);

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('route.equipment_pm-plan')">
    <NButton v-if="canAdd" type="primary" class="mb-12px" @click="openCreate">{{ $t('common.add') }}</NButton>
    <NDataTable :columns="columns" :data="rows" size="small" />
    <NModal v-model:show="modal" preset="card" :title="$t('page.mes.eqp.plan')" class="w-560px">
      <NForm label-placement="left" label-width="110">
        <NFormItem :label="$t('page.mes.eqp.plan')"><NInput v-model:value="form.planName" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.equipmentId')"><NInputNumber v-model:value="form.equipmentID" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.group')"><NInput v-model:value="form.equipmentGroup" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.trigger')">
          <NSelect v-model:value="form.triggerType" :options="[{ label: $t('page.mes.eqp.byTime'), value: 'time' }, { label: $t('page.mes.eqp.byCount'), value: 'count' }, { label: $t('page.mes.eqp.byBoth'), value: 'both' }]" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.eqp.intervalDays')"><NInputNumber v-model:value="form.intervalDays" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.intervalCount')"><NInputNumber v-model:value="form.intervalCount" :min="0" class="w-full" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.nextDue')"><NInput v-model:value="form.nextDueAt" placeholder="YYYY-MM-DD" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.checklist')"><NInput v-model:value="form.checklistText" type="textarea" :placeholder="$t('page.mes.eqp.checklistHint')" /></NFormItem>
        <NFormItem :label="$t('page.mes.eqp.block')"><NSwitch v-model:value="form.blockTrackIn" /></NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
