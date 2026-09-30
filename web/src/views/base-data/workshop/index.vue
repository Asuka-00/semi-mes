<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'factoryID', label: 'page.mes.field.factory', type: 'select', search: true, searchColumn: 'factory_id', required: true, relation: { resource: 'baseFactory', listKey: 'baseFactorys', labelKey: 'factoryName', extraLabelKey: 'factoryCode' } },
  { key: 'workshopCode', label: 'page.mes.field.workshopCode', search: true, searchColumn: 'workshop_code', required: true },
  { key: 'workshopName', label: 'page.mes.field.workshopName', search: true, searchColumn: 'workshop_name', required: true },
  { key: 'workshopType', label: 'page.mes.field.workshopType', search: true, searchColumn: 'workshop_type' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseWorkshop" list-key="baseWorkshops" perm="base:workshop" :fields="fields" />
</template>
