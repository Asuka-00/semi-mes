<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'username', label: 'page.mes.field.username', search: true, searchColumn: 'username', required: true },
  { key: 'password', label: 'page.mes.field.password', type: 'password', table: false, required: true },
  { key: 'realName', label: 'page.mes.field.realName', search: true, searchColumn: 'real_name' },
  { key: 'email', label: 'page.mes.field.email' },
  { key: 'phone', label: 'page.mes.field.phone' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="sysUser" list-key="sysUsers" perm="system:user" assignment="user-roles" :fields="fields" />
</template>
