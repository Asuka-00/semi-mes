<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue';
import type { VNode } from 'vue';
import type { DataTableColumns, FormInst, FormRules, SelectOption, TreeOption } from 'naive-ui';
import { NButton, NPopconfirm, NSpace } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import {
  mesCreate,
  mesDelete,
  mesGetRoleMenus,
  mesGetUserRoles,
  mesList,
  mesSetRoleMenus,
  mesSetUserRoles,
  mesUpdate
} from '@/service/api/mes';
import type { MesChildren, MesField, MesOption } from './types';

defineOptions({ name: 'MesCrudPage' });

const props = defineProps<{
  resource: string;
  listKey: string;
  perm: string;
  fields: MesField[];
  children?: MesChildren;
  assignment?: 'user-roles' | 'role-menus';
}>();

const { hasAuth } = useAuth();

const loading = ref(false);
const rows = ref<Record<string, any>[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const searchModel = reactive<Record<string, any>>({});
const relationOptions = reactive<Record<string, MesOption[]>>({});
const selectedId = ref<number | null>(null);

const modalVisible = ref(false);
const editingId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);
const formModel = reactive<Record<string, any>>({});
const roleIds = ref<number[]>([]);
const menuIds = ref<Array<string | number>>([]);
const roleOptions = ref<SelectOption[]>([]);
const menuRows = ref<Record<string, any>[]>([]);

const childLoading = ref(false);
const childRows = ref<Record<string, any>[]>([]);
const childVisible = ref(false);
const childEditingId = ref<number | null>(null);
const childFormRef = ref<FormInst | null>(null);
const childModel = reactive<Record<string, any>>({});

const canAdd = computed(() => hasAuth(`${props.perm}:add`));
const canEdit = computed(() => hasAuth(`${props.perm}:edit`));
const canDelete = computed(() => hasAuth(`${props.perm}:delete`));
const canChildAdd = computed(() => (props.children ? hasAuth(`${props.children.perm}:add`) : false));
const canChildEdit = computed(() => (props.children ? hasAuth(`${props.children.perm}:edit`) : false));
const canChildDelete = computed(() => (props.children ? hasAuth(`${props.children.perm}:delete`) : false));

const statusOptions = computed<MesOption[]>(() => [
  { label: $t('page.mes.enabled'), value: 1 },
  { label: $t('page.mes.disabled'), value: 2 }
]);

const yesNoOptions = computed<MesOption[]>(() => [
  { label: $t('page.mes.yes'), value: 1 },
  { label: $t('page.mes.no'), value: 0 }
]);

const menuTypeOptions = computed<MesOption[]>(() => [
  { label: $t('page.mes.directory'), value: 1 },
  { label: $t('page.mes.menuPage'), value: 2 },
  { label: $t('page.mes.button'), value: 3 }
]);

function fieldOptions(field: MesField): SelectOption[] {
  if (field.type === 'status') return statusOptions.value as SelectOption[];
  if (field.type === 'yesno') return yesNoOptions.value as SelectOption[];
  if (field.key === 'menuType') return menuTypeOptions.value as SelectOption[];
  if (field.relation) return (relationOptions[field.key] || []) as SelectOption[];
  return (field.options || []) as SelectOption[];
}

function displayValue(field: MesField, row: Record<string, any>) {
  const value = row[field.key];
  if (field.type === 'status') return value === 1 ? $t('page.mes.enabled') : $t('page.mes.disabled');
  if (field.type === 'yesno') return value === 1 ? $t('page.mes.yes') : $t('page.mes.no');
  if (field.type === 'password') return '';
  const options = fieldOptions(field);
  if (options.length) {
    const found = options.find(item => item.value === value);
    if (found) return found.label;
  }
  return value ?? '';
}

function buildColumns(fields: MesField[], actions: 'main' | 'child'): DataTableColumns<Record<string, any>> {
  const columns: DataTableColumns<Record<string, any>> = fields
    .filter(field => field.table !== false && field.type !== 'password' && field.type !== 'textarea')
    .map(field => ({
      title: $t(field.label),
      key: field.key,
      minWidth: 120,
      render: row => displayValue(field, row)
    }));

  columns.push({
    title: $t('common.action'),
    key: 'actions',
    width: actions === 'child' ? 280 : 180,
    render: row => {
      const buttons: VNode[] = [];
      const allowEdit = actions === 'main' ? canEdit.value : canChildEdit.value;
      const allowDelete = actions === 'main' ? canDelete.value : canChildDelete.value;
      if (allowEdit) {
        buttons.push(
          h(
            NButton,
            { size: 'small', type: 'primary', ghost: true, onClick: () => (actions === 'main' ? openEdit(row) : openChildEdit(row)) },
            { default: () => $t('common.edit') }
          )
        );
      }
      if (actions === 'child' && allowEdit) {
        buttons.push(
          h(NButton, { size: 'small', onClick: () => moveChild(row, -1) }, { default: () => $t('page.mes.moveUp') }),
          h(NButton, { size: 'small', onClick: () => moveChild(row, 1) }, { default: () => $t('page.mes.moveDown') })
        );
      }
      if (allowDelete) {
        buttons.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => (actions === 'main' ? removeRow(row) : removeChild(row)) },
            {
              trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => $t('common.delete') }),
              default: () => $t('common.confirmDelete')
            }
          )
        );
      }
      return h(NSpace, { size: 8 }, { default: () => buttons });
    }
  });
  return columns;
}

const columns = computed(() => buildColumns(props.fields, 'main'));
const childColumns = computed(() => (props.children ? buildColumns(props.children.fields, 'child') : []));

const formRules = computed<FormRules>(() => {
  const rules: FormRules = {};
  props.fields.forEach(field => {
    if (!field.required || field.form === false) return;
    if (field.type === 'password' && editingId.value) return;
    const numeric = field.type === 'status' || field.type === 'number' || field.type === 'select' || field.type === 'yesno' || Boolean(field.relation);
    rules[field.key] = {
      required: true,
      type: numeric ? 'number' : 'string',
      message: $t('form.required'),
      trigger: ['blur', 'change']
    };
  });
  return rules;
});

const childRules = computed<FormRules>(() => {
  const rules: FormRules = {};
  props.children?.fields.forEach(field => {
    if (!field.required || field.form === false) return;
    const numeric = field.type === 'status' || field.type === 'number' || field.type === 'select' || field.type === 'yesno' || Boolean(field.relation);
    rules[field.key] = {
      required: true,
      type: numeric ? 'number' : 'string',
      message: $t('form.required'),
      trigger: ['blur', 'change']
    };
  });
  return rules;
});

const menuTree = computed<TreeOption[]>(() => {
  const nodes = new Map<number, TreeOption>();
  menuRows.value.forEach(row => {
    const id = Number(row.id);
    nodes.set(id, { key: id, label: menuLabel(row), children: [] });
  });
  const roots: TreeOption[] = [];
  menuRows.value.forEach(row => {
    const id = Number(row.id);
    const node = nodes.get(id);
    if (!node) return;
    const parent = nodes.get(Number(row.parentID));
    if (parent && parent.children) parent.children.push(node);
    else roots.push(node);
  });
  return roots;
});

const parentOptions = computed<SelectOption[]>(() => [
  { label: $t('page.mes.root'), value: 0 },
  ...menuRows.value.map(row => ({ label: menuLabel(row), value: Number(row.id) }))
] as SelectOption[]);

function menuLabel(row: Record<string, any>) {
  if (Number(row.menuType) === 3) return String(row.permissionCode || row.menuName || row.id);
  const name = String(row.menuName || '');
  if (name.startsWith('route.')) return $t(name as App.I18n.I18nKey);
  return name;
}

function resetModel(target: Record<string, any>, fields: MesField[], extra?: Record<string, any>) {
  fields.forEach(field => {
    if (field.type === 'number' || field.type === 'status' || field.type === 'yesno' || field.type === 'select' || field.relation) {
      target[field.key] = field.type === 'status' ? 1 : field.type === 'yesno' ? 0 : null;
    } else {
      target[field.key] = '';
    }
  });
  Object.assign(target, extra || {});
}

function payloadFrom(fields: MesField[], model: Record<string, any>, editing: boolean) {
  const data: Record<string, unknown> = {};
  fields.forEach(field => {
    if (field.form === false) return;
    const value = model[field.key];
    if (field.type === 'password') {
      if (!editing || value) data[field.key] = value || '';
      return;
    }
    if (field.type === 'number' || field.type === 'status' || field.type === 'yesno' || field.relation || field.type === 'select') {
      data[field.key] = value === null || value === '' || value === undefined ? 0 : Number(value);
      return;
    }
    data[field.key] = value ?? '';
  });
  return data;
}

function searchColumns() {
  const filters = props.fields
    .filter(field => field.search && field.searchColumn)
    .flatMap(field => {
      const value = searchModel[field.key];
      if (value === null || value === undefined || value === '') return [];
      const numeric = field.type === 'status' || field.type === 'yesno' || field.type === 'select' || field.relation;
      return [
        {
          name: field.searchColumn as string,
          exp: numeric ? '=' : 'like',
          value: numeric ? Number(value) : `%${value}%`,
          logic: 'and'
        }
      ];
    });
  return filters;
}

async function loadOptions() {
  const tasks = props.fields
    .filter(field => field.relation)
    .map(async field => {
      const relation = field.relation!;
      const { data, error } = await mesList(relation.resource, { page: 0, limit: 200, sort: '-id' });
      if (error || !data) return;
      const list = Array.isArray(data[relation.listKey]) ? data[relation.listKey] : [];
      relationOptions[field.key] = list.map((item: Record<string, any>) => ({
        value: Number(item.id),
        label: relation.extraLabelKey ? `${item[relation.labelKey]} ${item[relation.extraLabelKey] || ''}`.trim() : String(item[relation.labelKey] ?? item.id)
      }));
    });
  await Promise.all(tasks);
  if (props.assignment === 'user-roles' || props.assignment === 'role-menus') {
    const { data } = await mesList('sysRole', { page: 0, limit: 200, sort: '-id' });
    const roles = Array.isArray(data?.sysRoles) ? data.sysRoles : [];
    roleOptions.value = roles.map((item: Record<string, any>) => ({
      value: Number(item.id),
      label: `${item.roleName} (${item.roleCode})`
    })) as SelectOption[];
  }
  if (props.assignment === 'role-menus' || props.fields.some(field => field.key === 'parentID')) {
    const menus = await mesList('sysMenu', { page: 0, limit: 500, sort: 'sort_order' });
    menuRows.value = Array.isArray(menus.data?.sysMenus) ? menus.data.sysMenus : [];
  }
}

async function loadList() {
  loading.value = true;
  const filters = searchColumns();
  const { data, error } = await mesList(props.resource, {
    page: page.value - 1,
    limit: pageSize.value,
    sort: '-id',
    columns: filters.length ? filters : undefined
  });
  loading.value = false;
  if (error || !data) return;
  rows.value = Array.isArray(data[props.listKey]) ? data[props.listKey] : [];
  total.value = Number(data.total || 0);
}

async function loadChildren() {
  if (!props.children || !selectedId.value) {
    childRows.value = [];
    return;
  }
  childLoading.value = true;
  const { data, error } = await mesList(props.children.resource, {
    page: 0,
    limit: 200,
    sort: props.children.sequenceKey,
    columns: [{ name: props.children.parentColumn, exp: '=', value: selectedId.value, logic: 'and' }]
  });
  childLoading.value = false;
  if (error || !data) return;
  const list = Array.isArray(data[props.children.listKey]) ? data[props.children.listKey] : [];
  childRows.value = [...list].sort((a, b) => Number(a[props.children!.sequenceKey]) - Number(b[props.children!.sequenceKey]) || Number(a.id) - Number(b.id));
}

function handleSearch() {
  page.value = 1;
  loadList();
}

function handleReset() {
  props.fields.forEach(field => {
    if (field.search) searchModel[field.key] = null;
  });
  handleSearch();
}

function openCreate() {
  editingId.value = null;
  resetModel(formModel, props.fields);
  roleIds.value = [];
  menuIds.value = [];
  modalVisible.value = true;
}

async function openEdit(row: Record<string, any>) {
  editingId.value = Number(row.id);
  resetModel(formModel, props.fields, row);
  if (props.assignment === 'user-roles') {
    const { data } = await mesGetUserRoles(editingId.value);
    roleIds.value = data?.roleIds || [];
  }
  if (props.assignment === 'role-menus') {
    const { data } = await mesGetRoleMenus(editingId.value);
    menuIds.value = data?.menuIds || [];
  }
  modalVisible.value = true;
}

async function submitForm() {
  await formRef.value?.validate();
  const data = payloadFrom(props.fields, formModel, Boolean(editingId.value));
  const result = editingId.value
    ? await mesUpdate(props.resource, editingId.value, data)
    : await mesCreate(props.resource, data);
  if (result.error) return;
  const created = result.data as { id?: number } | null;
  const id = editingId.value || Number(created?.id || 0);
  if (props.assignment === 'user-roles' && id) {
    const saved = await mesSetUserRoles(id, roleIds.value.map(Number));
    if (saved.error) return;
  }
  if (props.assignment === 'role-menus' && id) {
    const saved = await mesSetRoleMenus(id, withAncestors(menuIds.value.map(Number)));
    if (saved.error) return;
  }
  window.$message?.success(editingId.value ? $t('common.updateSuccess') : $t('common.addSuccess'));
  modalVisible.value = false;
  await loadList();
  if (props.assignment === 'role-menus') await loadOptions();
}

function withAncestors(ids: number[]) {
  const byId = new Map(menuRows.value.map(row => [Number(row.id), Number(row.parentID)]));
  const set = new Set(ids);
  ids.forEach(id => {
    let parent = byId.get(id);
    while (parent) {
      set.add(parent);
      parent = byId.get(parent);
    }
  });
  return [...set];
}

async function removeRow(row: Record<string, any>) {
  const { error } = await mesDelete(props.resource, Number(row.id));
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  if (selectedId.value === Number(row.id)) selectedId.value = null;
  await loadList();
}

function rowProps(row: Record<string, any>) {
  return {
    style: selectedId.value === Number(row.id) ? 'cursor: pointer; background: rgba(32, 128, 240, 0.08);' : 'cursor: pointer;',
    onClick: () => {
      selectedId.value = Number(row.id);
    }
  };
}

function openChildCreate() {
  if (!props.children || !selectedId.value) return;
  childEditingId.value = null;
  const maxSeq = childRows.value.reduce((max, row) => Math.max(max, Number(row[props.children!.sequenceKey]) || 0), 0);
  resetModel(childModel, props.children.fields, {
    [props.children.parentKey]: selectedId.value,
    [props.children.sequenceKey]: maxSeq + 1,
    status: 1
  });
  childVisible.value = true;
}

function openChildEdit(row: Record<string, any>) {
  if (!props.children) return;
  childEditingId.value = Number(row.id);
  resetModel(childModel, props.children.fields, row);
  childVisible.value = true;
}

async function submitChild() {
  if (!props.children) return;
  await childFormRef.value?.validate();
  const data = payloadFrom(props.children.fields, childModel, Boolean(childEditingId.value));
  data[props.children.parentKey] = selectedId.value;
  const result = childEditingId.value
    ? await mesUpdate(props.children.resource, childEditingId.value, data)
    : await mesCreate(props.children.resource, data);
  if (result.error) return;
  window.$message?.success(childEditingId.value ? $t('common.updateSuccess') : $t('common.addSuccess'));
  childVisible.value = false;
  await loadChildren();
}

async function removeChild(row: Record<string, any>) {
  if (!props.children) return;
  const { error } = await mesDelete(props.children.resource, Number(row.id));
  if (error) return;
  window.$message?.success($t('common.deleteSuccess'));
  await loadChildren();
}

async function moveChild(row: Record<string, any>, direction: number) {
  if (!props.children) return;
  const list = [...childRows.value];
  const index = list.findIndex(item => Number(item.id) === Number(row.id));
  const target = index + direction;
  if (index < 0 || target < 0 || target >= list.length) return;
  const current = list[index];
  list[index] = list[target];
  list[target] = current;
  const seqKey = props.children.sequenceKey;
  await Promise.all(
    list.map((item, itemIndex) => {
      const sequence = itemIndex + 1;
      if (Number(item[seqKey]) === sequence) return Promise.resolve();
      return mesUpdate(props.children!.resource, Number(item.id), { ...item, [seqKey]: sequence });
    })
  );
  await loadChildren();
}

watch(page, loadList);
watch(pageSize, () => {
  if (page.value === 1) loadList();
  else page.value = 1;
});
watch(selectedId, loadChildren);

onMounted(async () => {
  props.fields.forEach(field => {
    if (field.search) searchModel[field.key] = null;
  });
  await loadOptions();
  await loadList();
});
</script>

<template>
  <NSpace vertical :size="16">
    <NCard :bordered="false" class="card-wrapper">
      <NSpace class="mb-16px" wrap>
        <template v-for="field in fields" :key="field.key">
          <NSelect
            v-if="field.search && (field.type === 'status' || field.type === 'yesno' || field.type === 'select' || field.relation)"
            v-model:value="searchModel[field.key]"
            class="w-180px"
            clearable
            :options="fieldOptions(field)"
            :placeholder="$t(field.label)"
          />
          <NInput
            v-else-if="field.search"
            v-model:value="searchModel[field.key]"
            class="w-180px"
            clearable
            :placeholder="$t(field.label)"
            @keyup.enter="handleSearch"
          />
        </template>
        <NButton type="primary" @click="handleSearch">{{ $t('common.search') }}</NButton>
        <NButton @click="handleReset">{{ $t('common.reset') }}</NButton>
        <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('common.add') }}</NButton>
      </NSpace>
      <NDataTable
        remote
        :columns="columns"
        :data="rows"
        :loading="loading"
        :pagination="{
          page,
          pageSize,
          itemCount: total,
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          onUpdatePage: (value: number) => (page = value),
          onUpdatePageSize: (value: number) => (pageSize = value)
        }"
        :row-key="(row: Record<string, any>) => row.id"
        :row-props="rowProps"
      />
    </NCard>

    <NCard v-if="children" :bordered="false" class="card-wrapper" :title="$t('page.mes.operations')">
      <template #header-extra>
        <NButton v-if="canChildAdd && selectedId" type="primary" size="small" @click="openChildCreate">
          {{ $t('common.add') }}
        </NButton>
      </template>
      <NEmpty v-if="!selectedId" :description="$t('page.mes.selectRecord')" />
      <NDataTable v-else :columns="childColumns" :data="childRows" :loading="childLoading" :row-key="(row: Record<string, any>) => row.id" />
    </NCard>

    <NModal v-model:show="modalVisible" preset="card" class="w-720px" :title="editingId ? $t('common.edit') : $t('common.add')">
      <NForm ref="formRef" :model="formModel" :rules="formRules" label-placement="left" label-width="120">
        <NFormItem v-for="field in fields.filter(item => item.form !== false)" :key="field.key" :label="$t(field.label)" :path="field.key">
          <NInputNumber v-if="field.type === 'number'" v-model:value="formModel[field.key]" class="w-full" />
          <NSelect
            v-else-if="field.type === 'status' || field.type === 'yesno' || field.type === 'select' || field.relation || field.key === 'parentID'"
            v-model:value="formModel[field.key]"
            :options="field.key === 'parentID' ? parentOptions : fieldOptions(field)"
            :placeholder="$t('page.mes.selectPlaceholder')"
          />
          <NInput v-else-if="field.type === 'textarea'" v-model:value="formModel[field.key]" type="textarea" />
          <NInput
            v-else
            v-model:value="formModel[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :placeholder="field.type === 'password' && editingId ? $t('page.mes.passwordOptional') : ''"
          />
        </NFormItem>
        <NFormItem v-if="assignment === 'user-roles'" :label="$t('page.mes.assignRoles')">
          <NSelect v-model:value="roleIds" multiple :options="roleOptions" />
        </NFormItem>
        <NFormItem v-if="assignment === 'role-menus'" :label="$t('page.mes.assignMenus')">
          <NTree v-model:checked-keys="menuIds" block-line cascade checkable :data="menuTree" class="max-h-320px w-full overflow-auto" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modalVisible = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitForm">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal v-if="children" v-model:show="childVisible" preset="card" class="w-640px" :title="childEditingId ? $t('common.edit') : $t('common.add')">
      <NForm ref="childFormRef" :model="childModel" :rules="childRules" label-placement="left" label-width="120">
        <NFormItem
          v-for="field in children?.fields.filter(item => item.form !== false && item.key !== children?.parentKey)"
          :key="field.key"
          :label="$t(field.label)"
          :path="field.key"
        >
          <NInputNumber v-if="field.type === 'number'" v-model:value="childModel[field.key]" class="w-full" />
          <NSelect
            v-else-if="field.type === 'status' || field.type === 'yesno' || field.relation"
            v-model:value="childModel[field.key]"
            :options="fieldOptions(field)"
          />
          <NInput v-else v-model:value="childModel[field.key]" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="childVisible = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="submitChild">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </NSpace>
</template>
