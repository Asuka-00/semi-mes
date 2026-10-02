<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { NButton } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { completePmTask, fetchPmTasks, startPmTask } from '@/service/api/equipment';

const { hasAuth } = useAuth();
const canEdit = computed(() => hasAuth('eqp:pm:edit'));
const rows = ref<Array<Record<string, any>>>([]);
const status = ref<string | null>(null);
const loading = ref(false);

const statuses = ['due', 'overdue', 'in_progress', 'done'];
const completeOpen = ref(false);
const completeRow = ref<Record<string, any> | null>(null);
const completeResult = ref('pass');
const completeNote = ref('');
const completeItems = ref<Array<{ name: string; result: string }>>([]);

function taskLabel(value: string) {
  return $t(`page.mes.eqp.task_${value}` as App.I18n.I18nKey);
}

function checklistNames(raw: string) {
  try {
    const parsed = JSON.parse(raw || '[]');
    if (!Array.isArray(parsed)) return [];
    return parsed.map((item: string | { name: string }) => (typeof item === 'string' ? item : item.name)).filter(Boolean);
  } catch {
    return [];
  }
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchPmTasks({ status: status.value || '', limit: 50 });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.tasks || [];
}

async function start(row: Record<string, any>) {
  const { error } = await startPmTask(row.id);
  if (!error) load();
}

function openComplete(row: Record<string, any>) {
  completeRow.value = row;
  completeResult.value = 'pass';
  completeNote.value = '';
  completeItems.value = checklistNames(row.checklistJson).map(name => ({ name, result: 'pass' }));
  completeOpen.value = true;
}

async function finish() {
  if (!completeRow.value) return;
  const { error } = await completePmTask(completeRow.value.id, {
    result: completeResult.value,
    note: completeNote.value,
    items: completeItems.value
  });
  if (error) return;
  completeOpen.value = false;
  load();
}

const columns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.eqp.taskNo'), key: 'taskNo', minWidth: 120 },
  { title: $t('page.mes.eqp.plan'), key: 'planName', minWidth: 140 },
  { title: $t('page.mes.eqp.code'), key: 'equipmentCode', width: 120 },
  { title: $t('page.mes.eqp.taskStatus'), key: 'status', width: 120, render: row => taskLabel(row.status) },
  { title: $t('page.mes.eqp.due'), key: 'dueAt', minWidth: 160 },
  { title: $t('page.mes.eqp.result'), key: 'result', width: 80 },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 180,
    render: row =>
      h('div', { class: 'flex gap-8px' }, [
        canEdit.value && (row.status === 'due' || row.status === 'overdue')
          ? h(NButton, { size: 'small', type: 'warning', onClick: () => start(row) }, { default: () => $t('page.mes.eqp.startPm') })
          : null,
        canEdit.value && row.status === 'in_progress'
          ? h(NButton, { size: 'small', type: 'primary', onClick: () => openComplete(row) }, { default: () => $t('page.mes.eqp.completePm') })
          : null
      ])
  }
]);

onMounted(load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('route.equipment_pm-task')">
    <NSpace class="mb-12px">
      <NSelect v-model:value="status" clearable class="w-180px" :placeholder="$t('page.mes.eqp.taskStatus')" :options="statuses.map(value => ({ label: taskLabel(value), value }))" />
      <NButton @click="load">{{ $t('common.search') }}</NButton>
    </NSpace>
    <NDataTable :loading="loading" :columns="columns" :data="rows" size="small" />
    <NModal v-model:show="completeOpen" preset="card" :title="$t('page.mes.eqp.completePm')" class="w-560px">
      <NForm label-placement="left" label-width="90">
        <NFormItem v-for="item in completeItems" :key="item.name" :label="item.name">
          <NSelect v-model:value="item.result" :options="[{ label: $t('page.mes.eqp.pass'), value: 'pass' }, { label: $t('page.mes.eqp.fail'), value: 'fail' }]" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.eqp.result')">
          <NSelect v-model:value="completeResult" :options="[{ label: $t('page.mes.eqp.pass'), value: 'pass' }, { label: $t('page.mes.eqp.fail'), value: 'fail' }]" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.wip.note')"><NInput v-model:value="completeNote" /></NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="completeOpen = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="finish">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
