<script setup lang="ts">
import { ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { downloadCsv, formatDay } from '@/hooks/business/list-kit';
import { $t } from '@/locales';
import { fetchLotTrace, fetchTraceReverse, type TraceReport } from '@/service/api/trace';
import TraceTree from './tree.vue';

const lotNo = ref('DEMO-WIP-001');
const loading = ref(false);
const report = ref<TraceReport | null>(null);
const equipmentCode = ref('');
const nodeKey = ref('');
const from = ref<number | null>(null);
const to = ref<number | null>(null);
const reverseRows = ref<Array<Record<string, any>>>([]);
const reverseLoading = ref(false);

const historyColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.wip.event'), key: 'eventType', width: 120 },
  { title: $t('page.mes.wip.fromNode'), key: 'fromNodeKey', width: 120 },
  { title: $t('page.mes.wip.toNode'), key: 'toNodeKey', width: 120 },
  { title: $t('page.mes.wip.reasonCode'), key: 'reasonCode', width: 120 },
  { title: $t('page.mes.trace.reason'), key: 'reason', minWidth: 160 },
  { title: $t('page.mes.trace.related'), key: 'relatedLotNo', width: 140 },
  { title: $t('page.mes.audit.time'), key: 'createdAt', minWidth: 170 }
];

const moveColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.track.step'), key: 'nodeKey', width: 110 },
  { title: $t('page.mes.trace.equipment'), key: 'equipmentCode', width: 130 },
  { title: $t('page.mes.trace.operator'), key: 'operatorName', width: 110 },
  { title: $t('page.mes.trace.recipe'), key: 'recipeCode', width: 120 },
  { title: $t('page.mes.trace.params'), key: 'parameters', minWidth: 140 },
  { title: $t('page.mes.track.qty'), key: 'qtyIn', width: 80 },
  { title: $t('page.mes.track.qtyOut'), key: 'qtyOut', width: 80 },
  { title: $t('page.mes.track.state'), key: 'state', width: 100 },
  { title: $t('page.mes.wip.result'), key: 'inspectionResult', width: 100 },
  { title: $t('page.mes.audit.time'), key: 'trackInAt', minWidth: 170 }
];

const defectColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.wip.defectCode'), key: 'defectCode', width: 140 },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 90 },
  { title: $t('page.mes.trace.disposition'), key: 'disposition', width: 120 },
  { title: $t('page.mes.track.step'), key: 'nodeKey', width: 110 },
  { title: $t('page.mes.wip.note'), key: 'note', minWidth: 140 }
];

const measureColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.trace.params'), key: 'paramCode', width: 140 },
  { title: $t('page.mes.trace.value'), key: 'value', width: 100 },
  { title: $t('page.mes.wip.result'), key: 'specResult', width: 100 },
  { title: $t('page.mes.audit.time'), key: 'measuredAt', minWidth: 170 }
];

const reverseColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.trace.lotNo'), key: 'lotNo', minWidth: 140 },
  { title: $t('page.mes.wip.status'), key: 'status', width: 120 },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 90 },
  { title: $t('page.mes.wip.currentNode'), key: 'currentNodeKey', minWidth: 120 }
];

async function load() {
  if (!lotNo.value.trim()) return;
  loading.value = true;
  const { data, error } = await fetchLotTrace(lotNo.value.trim());
  loading.value = false;
  if (error || !data) {
    report.value = null;
    window.$message?.warning($t('page.mes.trace.notFound'));
    return;
  }
  report.value = data;
}

async function reverse() {
  reverseLoading.value = true;
  const { data, error } = await fetchTraceReverse({
    equipmentCode: equipmentCode.value.trim(),
    nodeKey: nodeKey.value.trim(),
    from: from.value ? formatDay(from.value) : '',
    to: to.value ? formatDay(to.value) : ''
  });
  reverseLoading.value = false;
  if (error || !data) {
    reverseRows.value = [];
    return;
  }
  reverseRows.value = data.lots || [];
}

function exportReport() {
  const current = report.value;
  if (!current) {
    window.$message?.warning($t('page.mes.query.exportEmpty'));
    return;
  }
  const rows: Array<Array<string | number>> = [['section', 'lot', 'detail']];
  rows.push(['lot', current.lot.lotNo, `${current.lot.status} ${current.lot.currentNodeKey}`]);
  for (const item of current.history || []) {
    rows.push(['history', current.lot.lotNo, `${item.eventType} ${item.reasonCode || ''} ${item.relatedLotNo || ''}`]);
  }
  for (const item of current.moves || []) {
    rows.push(['move', current.lot.lotNo, `${item.nodeKey} ${item.equipmentCode || ''} ${item.operatorName || ''} ${item.recipeCode || ''}`]);
  }
  for (const item of current.defects || []) {
    rows.push(['defect', current.lot.lotNo, `${item.defectCode} ${item.quantity} ${item.disposition}`]);
  }
  for (const item of current.measurements || []) {
    rows.push(['measurement', current.lot.lotNo, `${item.paramCode} ${item.value} ${item.specResult || ''}`]);
  }
  downloadCsv(`trace-${current.lot.lotNo}`, ['section', 'lot', 'detail'], rows.slice(1).length ? rows.slice(1) : rows);
}

function printReport() {
  const node = document.getElementById('trace-report');
  if (!node) return;
  const win = window.open('', '_blank', 'noopener,noreferrer');
  if (!win) return;
  win.document.write(`<html><head><title>${lotNo.value}</title></head><body>${node.innerHTML}</body></html>`);
  win.document.close();
  win.focus();
  win.print();
}

load();
</script>

<template>
  <div class="flex-col gap-12px">
    <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.trace.title')">
      <NSpace class="mb-12px" wrap>
        <NInput v-model:value="lotNo" :placeholder="$t('page.mes.trace.lotNo')" class="w-220px" clearable @keyup.enter="load" />
        <NButton type="primary" :loading="loading" @click="load">{{ $t('page.mes.trace.load') }}</NButton>
        <NButton :disabled="!report" @click="exportReport">{{ $t('page.mes.trace.export') }}</NButton>
        <NButton :disabled="!report" @click="printReport">{{ $t('page.mes.trace.print') }}</NButton>
      </NSpace>
      <div class="mb-12px text-13px text-#888">{{ $t('page.mes.trace.wafersLater') }}</div>
      <div id="trace-report">
        <template v-if="report">
          <div class="mb-12px">
            <span class="mr-16px font-600">{{ report.lot.lotNo }}</span>
            <span class="mr-16px">{{ report.lot.status }}</span>
            <span class="mr-16px">{{ report.lot.currentNodeKey }}</span>
            <span>{{ report.lot.quantity }}</span>
            <span v-if="report.lot.holdReasonCode" class="ml-16px">{{ report.lot.holdReasonCode }} {{ report.lot.holdReason }}</span>
          </div>
          <div class="grid gap-12px md:grid-cols-2">
            <NCard size="small" :title="$t('page.mes.trace.backward')">
              <TraceTree v-if="report.backward?.length" :nodes="report.backward" />
              <div v-else class="text-#888">{{ $t('page.mes.trace.emptyTree') }}</div>
            </NCard>
            <NCard size="small" :title="$t('page.mes.trace.forward')">
              <TraceTree v-if="report.forward?.length" :nodes="report.forward" />
              <div v-else class="text-#888">{{ $t('page.mes.trace.emptyTree') }}</div>
            </NCard>
          </div>
          <NDivider>{{ $t('page.mes.trace.history') }}</NDivider>
          <NDataTable size="small" :columns="historyColumns" :data="report.history || []" :scroll-x="980" />
          <NDivider>{{ $t('page.mes.trace.moves') }}</NDivider>
          <NDataTable size="small" :columns="moveColumns" :data="report.moves || []" :scroll-x="1200" />
          <NDivider>{{ $t('page.mes.trace.defects') }}</NDivider>
          <NDataTable size="small" :columns="defectColumns" :data="report.defects || []" />
          <NDivider>{{ $t('page.mes.trace.measurements') }}</NDivider>
          <NDataTable size="small" :columns="measureColumns" :data="report.measurements || []" />
        </template>
      </div>
    </NCard>
    <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.trace.reverse')">
      <NSpace class="mb-12px" wrap>
        <NInput v-model:value="equipmentCode" :placeholder="$t('page.mes.trace.equipment')" class="w-160px" clearable />
        <NInput v-model:value="nodeKey" :placeholder="$t('page.mes.trace.node')" class="w-140px" clearable />
        <NDatePicker v-model:value="from" type="date" clearable />
        <NDatePicker v-model:value="to" type="date" clearable />
        <NButton type="primary" :loading="reverseLoading" @click="reverse">{{ $t('page.mes.trace.searchReverse') }}</NButton>
      </NSpace>
      <NDataTable size="small" :loading="reverseLoading" :columns="reverseColumns" :data="reverseRows" />
    </NCard>
  </div>
</template>
