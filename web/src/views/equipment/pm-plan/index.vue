<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { NButton, NPopconfirm } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { useAuth } from '@/hooks/business/auth';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { fetchPmPlans, removePmPlan, savePmPlan } from '@/service/api/equipment';
import { mesBatch } from '@/service/api/mes';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('eqp:pm:add'));
const canEdit = computed(() => hasAuth('eqp:pm:edit'));
const canDelete = computed(() => hasAuth('eqp:pm:delete'));
const rows = ref<Array<Record<string, any>>>([]);
const keyword = ref('');
const trigger = ref<string[]>([]);
const enabled = ref<string[]>([]);
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const checked = ref<DataTableRowKey[]>([]);
const keys = ['planName', 'equipmentGroup', 'triggerType', 'intervalDays', 'intervalCount', 'lotsSince', 'blockTrackIn'];
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

function filters() {
  const columns = [];
  if (keyword.value) columns.push({ name: 'plan_name', exp: 'like', value: keyword.value });
  if (trigger.value.length) columns.push({ name: 'trigger_type', exp: 'in', value: trigger.value.join(',') });
  if (enabled.value.length === 1) columns.push({ name: 'enabled', exp: '=', value: enabled.value[0] });
  return columns;
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('eqpPmPlan', {
    search: { keyword: keyword.value, trigger: trigger.value, enabled: enabled.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value
  });
}

async function load() {
  const { data, error } = await fetchPmPlans({ page: 0, limit: 200, columns: filters() });
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

const columns = computed<DataTableColumns<Record<string, any>>>(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  const all: DataTableColumns<Record<string, any>> = [
  { type: 'selection' },
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
  ];
  return all.filter(column => keepColumn(column, visible));
});

async function exportRows() {
  const list = rows.value;
  if (!list.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv('pm-plans', ['planName', 'triggerType'], list.map(row => [row.planName, row.triggerType]));
}

async function runBatch(action: 'delete' | 'enable' | 'disable') {
  const ids = checked.value.map(item => Number(item));
  const { data, error } = await mesBatch({ resource: 'eqpPmPlan', action, ids });
  if (error || !data) return;
  checked.value = [];
  window.$message?.success($t('page.mes.query.partial', { ok: data.ok?.length || 0, failed: data.failed?.length || 0 }));
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('eqpPmPlan');
  keyword.value = String(pref.search?.keyword || '');
  trigger.value = Array.isArray(pref.search?.trigger) ? (pref.search.trigger as string[]) : [];
  enabled.value = Array.isArray(pref.search?.enabled) ? (pref.search.enabled as string[]) : [];
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('route.equipment_pm-plan')">
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
      <NInput v-model:value="keyword" class="w-180px" clearable :placeholder="$t('page.mes.eqp.plan')" @keyup.enter="persist(); load()" />
      <NSelect v-model:value="trigger" multiple clearable class="w-200px" :options="[{ label: $t('page.mes.eqp.byTime'), value: 'time' }, { label: $t('page.mes.eqp.byCount'), value: 'count' }, { label: $t('page.mes.eqp.byBoth'), value: 'both' }]" />
      <NSelect v-model:value="enabled" multiple clearable class="w-160px" :options="[{ label: $t('page.mes.enabled'), value: '1' }, { label: $t('page.mes.disabled'), value: '0' }]" />
      <NButton type="primary" @click="persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="keyword = ''; trigger = []; enabled = []; persist(); load()">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable v-model:checked-row-keys="checked" :row-key="(row: Record<string, any>) => row.id" :columns="columns" :data="rows" size="small" />
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
