<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesChildren, MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'productID', label: 'page.mes.field.product', type: 'select', search: true, searchColumn: 'product_id', required: true, relation: { resource: 'baseProduct', listKey: 'baseProducts', labelKey: 'productName', extraLabelKey: 'productCode' } },
  { key: 'routeCode', label: 'page.mes.field.routeCode', search: true, searchColumn: 'route_code', required: true },
  { key: 'routeName', label: 'page.mes.field.routeName', search: true, searchColumn: 'route_name', required: true },
  { key: 'version', label: 'page.mes.field.version' },
  { key: 'isDefault', label: 'page.mes.field.isDefault', type: 'yesno' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];

const children: MesChildren = {
  resource: 'baseOperation',
  listKey: 'baseOperations',
  perm: 'base:operation',
  parentKey: 'routeID',
  parentColumn: 'route_id',
  sequenceKey: 'sequence',
  fields: [
    { key: 'routeID', label: 'page.mes.field.route', form: false, table: false },
    { key: 'operationCode', label: 'page.mes.field.operationCode', required: true },
    { key: 'operationName', label: 'page.mes.field.operationName', required: true },
    { key: 'operationType', label: 'page.mes.field.operationType' },
    { key: 'sequence', label: 'page.mes.field.sequence', type: 'number' },
    { key: 'standardTime', label: 'page.mes.field.standardTime', type: 'number' },
    { key: 'status', label: 'page.mes.field.status', type: 'status', required: true }
  ]
};
</script>

<template>
  <CrudPage resource="baseProcessRoute" list-key="baseProcessRoutes" perm="base:route" :fields="fields" :children="children" />
</template>
