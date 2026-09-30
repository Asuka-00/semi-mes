<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'workshopID', label: 'page.mes.field.workshop', type: 'select', search: true, searchColumn: 'workshop_id', required: true, relation: { resource: 'baseWorkshop', listKey: 'baseWorkshops', labelKey: 'workshopName', extraLabelKey: 'workshopCode' } },
  { key: 'lineCode', label: 'page.mes.field.lineCode', search: true, searchColumn: 'line_code', required: true },
  { key: 'lineName', label: 'page.mes.field.lineName', search: true, searchColumn: 'line_name', required: true },
  { key: 'capacity', label: 'page.mes.field.capacity', type: 'number' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseProductionLine" list-key="baseProductionLines" perm="base:line" :fields="fields" />
</template>
