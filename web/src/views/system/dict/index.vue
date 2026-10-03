<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesChildren, MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'typeCode', label: 'page.mes.dict.typeCode', search: true, searchColumn: 'type_code', required: true },
  { key: 'typeName', label: 'page.mes.dict.typeName', search: true, searchColumn: 'type_name', required: true },
  { key: 'remark', label: 'page.mes.dict.remark' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];

const childFields: MesField[] = [
  { key: 'itemCode', label: 'page.mes.dict.itemCode', required: true },
  { key: 'labelZh', label: 'page.mes.dict.labelZh', required: true },
  { key: 'labelEn', label: 'page.mes.dict.labelEn', required: true },
  { key: 'color', label: 'page.mes.dict.color' },
  { key: 'tag', label: 'page.mes.dict.tag' },
  { key: 'sortOrder', label: 'page.mes.dict.sort', type: 'number' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', required: true }
];

const children: MesChildren = {
  resource: 'sysDictItem',
  listKey: 'sysDictItems',
  perm: 'system:dict',
  parentKey: 'typeID',
  parentColumn: 'type_id',
  sequenceKey: 'sortOrder',
  fields: childFields
};
</script>

<template>
  <CrudPage resource="sysDictType" list-key="sysDictTypes" perm="system:dict" :fields="fields" :children="children" />
</template>
