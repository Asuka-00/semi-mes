<script setup lang="ts">
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';

const fields: MesField[] = [
  { key: 'operationID', label: 'page.mes.field.operation', type: 'select', search: true, searchColumn: 'operation_id', required: true, relation: { resource: 'baseOperation', listKey: 'baseOperations', labelKey: 'operationName', extraLabelKey: 'operationCode' } },
  { key: 'recipeCode', label: 'page.mes.field.recipeCode', search: true, searchColumn: 'recipe_code', required: true },
  { key: 'recipeName', label: 'page.mes.field.recipeName', search: true, searchColumn: 'recipe_name', required: true },
  { key: 'version', label: 'page.mes.field.version' },
  { key: 'parameters', label: 'page.mes.field.parameters', type: 'textarea' },
  { key: 'isDefault', label: 'page.mes.field.isDefault', type: 'yesno' },
  { key: 'description', label: 'page.mes.field.description', type: 'textarea', table: false },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
];
</script>

<template>
  <CrudPage resource="baseRecipe" list-key="baseRecipes" perm="base:recipe" :fields="fields" />
</template>
