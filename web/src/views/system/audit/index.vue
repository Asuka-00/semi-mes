<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { NButton } from 'naive-ui';
import ColumnPicker from '@/components/mes/column-picker.vue';
import { downloadCsv, keepColumn, loadPagePref, moveKey, orderedKeys, rangeColumns, savePagePref } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { fetchAuditList } from '@/service/api/audit';

const username = ref('');
const moduleName = ref('');
const action = ref('');
const entityCode = ref('');
const createdRange = ref<[number, number] | null>(null);
const sort = ref('-id');
const collapsed = ref(false);
const hidden = ref<string[]>([]);
const columnOrder = ref<string[]>([]);
const prefsReady = ref(false);
const rows = ref<Array<Record<string, any>>>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const drawer = ref(false);
const current = ref<Record<string, any> | null>(null);
const keys = ['createdAt', 'username', 'module', 'action', 'entityType', 'entityCode', 'summary', 'ip'];

const labels = computed<Record<string, string>>(() => ({
  createdAt: $t('page.mes.audit.time'),
  username: $t('page.mes.audit.user'),
  module: $t('page.mes.audit.module'),
  action: $t('page.mes.audit.action'),
  entityType: $t('page.mes.audit.entityType'),
  entityCode: $t('page.mes.audit.target'),
  summary: $t('page.mes.audit.summary'),
  ip: $t('page.mes.audit.ip')
}));

function pretty(value: unknown) {
  if (!value) return '';
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}

const allColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: labels.value.createdAt, key: 'createdAt', minWidth: 170 },
  { title: labels.value.username, key: 'username', width: 120 },
  { title: labels.value.module, key: 'module', width: 100 },
  { title: labels.value.action, key: 'action', width: 120 },
  { title: labels.value.entityType, key: 'entityType', width: 120 },
  { title: labels.value.entityCode, key: 'entityCode', minWidth: 140 },
  { title: labels.value.summary, key: 'summary', minWidth: 160 },
  { title: labels.value.ip, key: 'ip', width: 130 },
  {
    title: $t('page.mes.audit.diff'),
    key: 'diff',
    width: 90,
    render: row =>
      h(
        NButton,
        {
          text: true,
          type: 'primary',
          onClick: () => {
            current.value = row;
            drawer.value = true;
          }
        },
        { default: () => $t('page.mes.audit.diff') }
      )
  }
]);

const columns = computed(() => {
  const visible = new Set(orderedKeys(keys, columnOrder.value).filter(key => !hidden.value.includes(key)));
  return allColumns.value.filter(column => keepColumn(column, visible, ['diff']));
});

function filters() {
  const query: Array<{ name: string; exp: string; value: string; logic: string }> = [];
  if (username.value) query.push({ name: 'username', exp: 'like', value: username.value, logic: 'and' });
  if (moduleName.value) query.push({ name: 'module', exp: 'eq', value: moduleName.value, logic: 'and' });
  if (action.value) query.push({ name: 'action', exp: 'eq', value: action.value, logic: 'and' });
  if (entityCode.value) query.push({ name: 'entity_code', exp: 'like', value: entityCode.value, logic: 'and' });
  return [...query, ...rangeColumns('created_at', createdRange.value)];
}

async function persist() {
  if (!prefsReady.value) return;
  await savePagePref('sysAudit', {
    search: { username: username.value, module: moduleName.value, action: action.value, entityCode: entityCode.value },
    hidden: hidden.value,
    order: columnOrder.value,
    collapsed: collapsed.value,
    sort: sort.value,
    createdRange: createdRange.value
  });
}

async function load() {
  loading.value = true;
  const { data, error } = await fetchAuditList({ page: page.value - 1, limit: 10, sort: sort.value, columns: filters() });
  loading.value = false;
  if (error || !data) return;
  rows.value = data.sysAudits || [];
  total.value = Number(data.total || 0);
}

async function exportRows() {
  const { data, error } = await fetchAuditList({ page: 0, limit: 2000, sort: sort.value, columns: filters() });
  const list = data?.sysAudits || [];
  if (error || !list.length) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  downloadCsv(
    'audit-log',
    keys,
    list.map(row => keys.map(key => row[key]))
  );
  window.$message?.success($t('page.mes.query.exported', { count: list.length }));
}

function reset() {
  username.value = '';
  moduleName.value = '';
  action.value = '';
  entityCode.value = '';
  createdRange.value = null;
  sort.value = '-id';
  page.value = 1;
  persist();
  load();
}

onMounted(async () => {
  const pref = await loadPagePref('sysAudit');
  const search = pref.search || {};
  username.value = String(search.username || '');
  moduleName.value = String(search.module || '');
  action.value = String(search.action || '');
  entityCode.value = String(search.entityCode || '');
  hidden.value = pref.hidden || [];
  columnOrder.value = pref.order || [...keys];
  collapsed.value = Boolean(pref.collapsed);
  sort.value = pref.sort || '-id';
  createdRange.value = pref.createdRange || null;
  prefsReady.value = true;
  load();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.audit.title')">
    <NSpace class="mb-12px" wrap>
      <NButton @click="collapsed = !collapsed; persist()">
        {{ collapsed ? $t('page.mes.query.expand') : $t('page.mes.query.collapse') }}
      </NButton>
      <NButton @click="exportRows">{{ $t('page.mes.query.export') }}</NButton>
      <ColumnPicker
        :items="orderedKeys(keys, columnOrder).map(key => ({ key, label: labels[key] || key }))"
        :hidden="hidden"
        @toggle="(key: string, shown: boolean) => { hidden = shown ? hidden.filter(item => item !== key) : [...hidden, key]; persist(); }"
        @reorder="(key: string, dir: number) => { columnOrder = moveKey(columnOrder.length ? columnOrder : [...keys], key, dir); persist(); }"
      />
    </NSpace>
    <NSpace v-show="!collapsed" class="mb-12px" wrap>
      <NInput v-model:value="username" :placeholder="$t('page.mes.audit.user')" class="w-140px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="moduleName" :placeholder="$t('page.mes.audit.module')" class="w-140px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="action" :placeholder="$t('page.mes.audit.action')" class="w-140px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NInput v-model:value="entityCode" :placeholder="$t('page.mes.audit.target')" class="w-160px" clearable @keyup.enter="page = 1; persist(); load()" />
      <NDatePicker v-model:value="createdRange" type="daterange" clearable />
      <NSelect v-model:value="sort" class="w-140px" :options="[{ label: 'ID ↓', value: '-id' }, { label: 'ID ↑', value: 'id' }, { label: $t('page.mes.audit.time'), value: '-created_at' }]" />
      <NButton type="primary" @click="page = 1; persist(); load()">{{ $t('common.search') }}</NButton>
      <NButton @click="reset">{{ $t('common.reset') }}</NButton>
    </NSpace>
    <NDataTable
      remote
      :loading="loading"
      :columns="columns"
      :data="rows"
      :scroll-x="1200"
      :pagination="{ page, pageSize: 10, itemCount: total, onUpdatePage: (next: number) => { page = next; load(); } }"
    />
    <NDrawer v-model:show="drawer" :width="520">
      <NDrawerContent :title="$t('page.mes.audit.diff')" closable>
        <div class="mb-8px text-13px text-#666">{{ current?.summary }} · {{ current?.entityCode }}</div>
        <div class="mb-4px font-600">{{ $t('page.mes.audit.diff') }}</div>
        <pre class="mb-12px whitespace-pre-wrap break-all text-12px">{{ pretty(current?.diffJson) || $t('page.mes.audit.emptyDiff') }}</pre>
        <div class="mb-4px font-600">{{ $t('page.mes.audit.before') }}</div>
        <pre class="mb-12px whitespace-pre-wrap break-all text-12px">{{ pretty(current?.beforeJson) || '-' }}</pre>
        <div class="mb-4px font-600">{{ $t('page.mes.audit.after') }}</div>
        <pre class="whitespace-pre-wrap break-all text-12px">{{ pretty(current?.afterJson) || '-' }}</pre>
      </NDrawerContent>
    </NDrawer>
  </NCard>
</template>
