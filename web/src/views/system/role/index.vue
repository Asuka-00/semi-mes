<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'roleCode', label: 'page.mes.field.roleCode', search: true, searchColumn: 'role_code', required: true },
  { key: 'roleName', label: 'page.mes.field.roleName', search: true, searchColumn: 'role_name', required: true },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="sysRole" list-key="sysRoles" perm="system:role" assignment="role-menus" :fields="fields" />
</template>
