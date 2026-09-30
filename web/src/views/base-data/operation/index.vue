<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'routeID', label: 'page.mes.field.route', type: 'select', search: true, searchColumn: 'route_id', required: true, relation: { resource: 'baseProcessRoute', listKey: 'baseProcessRoutes', labelKey: 'routeName', extraLabelKey: 'routeCode' } },
  { key: 'operationCode', label: 'page.mes.field.operationCode', search: true, searchColumn: 'operation_code', required: true },
  { key: 'operationName', label: 'page.mes.field.operationName', search: true, searchColumn: 'operation_name', required: true },
  { key: 'operationType', label: 'page.mes.field.operationType' },
  { key: 'sequence', label: 'page.mes.field.sequence', type: 'number' },
  { key: 'standardTime', label: 'page.mes.field.standardTime', type: 'number' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseOperation" list-key="baseOperations" perm="base:operation" :fields="fields" />
</template>
