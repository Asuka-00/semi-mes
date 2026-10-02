<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  {
    key: 'operationCode',
    label: 'page.mes.field.operationCode',
    search: true,
    searchColumn: 'operation_code',
    required: true
  },
  {
    key: 'operationName',
    label: 'page.mes.field.operationName',
    search: true,
    searchColumn: 'operation_name',
    required: true
  },
  { key: 'operationType', label: 'page.mes.field.operationType', search: true, searchColumn: 'operation_type' },
  { key: 'standardTime', label: 'page.mes.field.standardTime', type: 'number' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  {
    key: 'status',
    label: 'page.mes.field.status',
    type: 'status',
    search: true,
    searchColumn: 'status',
    required: true
  }
];
</script>

<template>
  <CrudPage resource="baseOperation" list-key="baseOperations" perm="base:operation" :fields="fields" />
</template>
