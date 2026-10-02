<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'parentID', label: 'page.mes.field.parent', type: 'select' },
  { key: 'menuType', label: 'page.mes.field.menuType', type: 'select', required: true },
  { key: 'menuName', label: 'page.mes.field.menuName', search: true, searchColumn: 'menu_name', required: true },
  { key: 'permissionCode', label: 'page.mes.field.permissionCode', search: true, searchColumn: 'permission_code' },
  { key: 'routeName', label: 'page.mes.field.menuRouteName' },
  { key: 'routePath', label: 'page.mes.field.menuRoutePath', table: false },
  { key: 'componentPath', label: 'page.mes.field.menuComponentPath', table: false },
  { key: 'icon', label: 'page.mes.field.icon', table: false },
  { key: 'sortOrder', label: 'page.mes.field.sortOrder', type: 'number' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="sysMenu" list-key="sysMenus" perm="system:menu" :fields="fields" />
</template>
