<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { DataTableColumns } from 'naive-ui';
import { $t } from '@/locales';
import { fetchEquipmentDetail } from '@/service/api/equipment';
import type { CapabilityRow, EquipmentRow } from '@/service/api/equipment';

const route = useRoute();
const router = useRouter();
const equipment = ref<EquipmentRow | null>(null);
const capabilities = ref<CapabilityRow[]>([]);
const logs = ref<Array<Record<string, any>>>([]);
const tasks = ref<Array<Record<string, any>>>([]);
const lots = ref<Array<Record<string, any>>>([]);
const openLots = ref(0);
const id = computed(() => Number(route.params.id));

function stateLabel(value: string) {
  if (!value) return '';
  return $t(`page.mes.eqp.state_${value}` as App.I18n.I18nKey);
}

const logColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.eqp.from'), key: 'fromState', render: row => stateLabel(row.fromState || '') },
  { title: $t('page.mes.eqp.to'), key: 'toState', render: row => stateLabel(row.toState) },
  { title: $t('page.mes.eqp.reason'), key: 'reasonCode' },
  { title: $t('page.mes.wip.reason'), key: 'reason' },
  { title: $t('page.mes.eqp.started'), key: 'startedAt', minWidth: 160 },
  { title: $t('page.mes.eqp.duration'), key: 'durationSeconds', width: 100 }
];
const lotColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo' },
  { title: $t('page.mes.track.step'), key: 'nodeKey' },
  { title: $t('page.mes.track.qty'), key: 'qtyIn', width: 80 },
  { title: $t('page.mes.track.qtyOut'), key: 'qtyOut', width: 90 },
  { title: $t('page.mes.eqp.state'), key: 'state', width: 110 }
];

async function load() {
  if (!id.value) return;
  const { data, error } = await fetchEquipmentDetail(id.value);
  if (error || !data) return;
  equipment.value = data.equipment;
  capabilities.value = data.capabilities || [];
  logs.value = data.stateLogs || [];
  tasks.value = data.pmTasks || [];
  lots.value = data.recentLots || [];
  openLots.value = data.openLots || 0;
}

onMounted(load);
watch(id, load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="equipment?.equipmentCode || $t('route.equipment_detail')">
    <NButton class="mb-12px" @click="router.push({ name: 'equipment_list' })">{{ $t('page.mes.wip.back') }}</NButton>
    <NSpace v-if="equipment" class="mb-16px">
      <NTag type="info">{{ equipment.equipmentName }}</NTag>
      <NTag>{{ stateLabel(equipment.status) }}</NTag>
      <span>{{ $t('page.mes.eqp.group') }} {{ equipment.equipmentGroup }}</span>
      <span>{{ $t('page.mes.eqp.vendor') }} {{ equipment.manufacturer }}</span>
      <span>{{ $t('page.mes.eqp.model') }} {{ equipment.modelName }}</span>
      <span>{{ $t('page.mes.eqp.serial') }} {{ equipment.serialNo }}</span>
      <span>{{ $t('page.mes.eqp.location') }} {{ equipment.location }}</span>
      <span>{{ $t('page.mes.eqp.chambers') }} {{ equipment.chamberCount }}</span>
      <span>{{ $t('page.mes.eqp.capacity') }} {{ equipment.capacity }}</span>
      <span>{{ $t('page.mes.eqp.openLots') }} {{ openLots }}</span>
    </NSpace>
    <div class="mb-8px font-600">{{ $t('page.mes.eqp.capability') }}</div>
    <div class="mb-16px">
      <NTag v-for="item in capabilities" :key="item.id" class="mr-8px">{{ item.operationCode || item.operationID }}<template v-if="item.recipeID"> / {{ item.recipeCode || item.recipeID }}</template></NTag>
      <span v-if="!capabilities.length">{{ $t('page.mes.eqp.anyCapability') }}</span>
    </div>
    <div class="mb-8px font-600">{{ $t('page.mes.eqp.stateHistory') }}</div>
    <NDataTable :columns="logColumns" :data="logs" size="small" class="mb-16px" />
    <div class="mb-8px font-600">{{ $t('page.mes.eqp.recentLots') }}</div>
    <NDataTable :columns="lotColumns" :data="lots" size="small" class="mb-16px" />
    <div class="mb-8px font-600">{{ $t('page.mes.eqp.pmHistory') }}</div>
    <NDataTable
      size="small"
      :data="tasks"
      :columns="[
        { title: $t('page.mes.eqp.plan'), key: 'planName' },
        { title: $t('page.mes.eqp.taskStatus'), key: 'status' },
        { title: $t('page.mes.eqp.result'), key: 'result' },
        { title: $t('page.mes.wip.note'), key: 'note' }
      ]"
    />
  </NCard>
</template>
