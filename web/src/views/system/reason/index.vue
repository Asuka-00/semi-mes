<script setup lang="ts">
import { computed } from 'vue';
import CrudPage from '@/components/mes/crud-page.vue';
import type { MesField } from '@/components/mes/types';
import { $t } from '@/locales';

const fields = computed<MesField[]>(() => [
  {
    key: 'category',
    label: 'page.mes.reason.category',
    type: 'select',
    search: true,
    searchColumn: 'category',
    required: true,
    options: ['hold', 'release', 'scrap', 'rework', 'eqp_down', 'eqp'].map(value => ({
      label: $t(`page.mes.reason.cat_${value}` as App.I18n.I18nKey),
      value
    }))
  },
  { key: 'reasonCode', label: 'page.mes.reason.code', search: true, searchColumn: 'reason_code', required: true },
  { key: 'nameZh', label: 'page.mes.dict.labelZh', required: true },
  { key: 'nameEn', label: 'page.mes.dict.labelEn', required: true },
  { key: 'sortOrder', label: 'page.mes.dict.sort', type: 'number' },
  { key: 'status', label: 'page.mes.field.status', type: 'status', search: true, searchColumn: 'status', required: true }
]);
</script>

<template>
  <CrudPage resource="sysReasonCode" list-key="sysReasonCodes" perm="system:reason" :fields="fields" />
</template>
