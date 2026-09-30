<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'productCode', label: 'page.mes.field.productCode', search: true, searchColumn: 'product_code', required: true },
  { key: 'productName', label: 'page.mes.field.productName', search: true, searchColumn: 'product_name', required: true },
  { key: 'productType', label: 'page.mes.field.productType', search: true, searchColumn: 'product_type' },
  { key: 'version', label: 'page.mes.field.version' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseProduct" list-key="baseProducts" perm="base:product" :fields="fields" />
</template>
