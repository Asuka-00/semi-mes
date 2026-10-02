<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'factoryCode', label: 'page.mes.field.factoryCode', search: true, searchColumn: 'factory_code', required: true },
  { key: 'factoryName', label: 'page.mes.field.factoryName', search: true, searchColumn: 'factory_name', required: true },
  { key: 'address', label: 'page.mes.field.address' },
  { key: 'contact', label: 'page.mes.field.contact' },
  { key: 'phone', label: 'page.mes.field.phone' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseFactory" list-key="baseFactorys" perm="base:factory" :fields="fields" />
</template>
