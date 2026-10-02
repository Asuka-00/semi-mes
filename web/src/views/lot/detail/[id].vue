<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MarkerType, VueFlow, useVueFlow } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Controls } from '@vue-flow/controls';
import { MiniMap } from '@vue-flow/minimap';
import type { DataTableColumns } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { advanceLot, fetchLot, holdLot, releaseHold } from '@/service/api/wip';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import '@vue-flow/controls/dist/style.css';
import '@vue-flow/minimap/dist/style.css';

const props = defineProps<{ id?: string }>();
const route = useRoute();
const router = useRouter();
const { hasAuth } = useAuth();
const canEdit = computed(() => hasAuth('lot:lot:edit'));
const { fitView } = useVueFlow({ id: 'lot-detail' });

const lot = ref<Record<string, any> | null>(null);
const history = ref<Array<Record<string, any>>>([]);
const links = ref<Array<Record<string, any>>>([]);
const nodes = ref<Array<Record<string, any>>>([]);
const edges = ref<Array<Record<string, any>>>([]);
const inspection = ref('pass');
const defectCode = ref('');
const resultText = ref('');
const reasonCode = ref('HOLD');
const reason = ref('');

const lotId = computed(() => Number(props.id || route.params.id));

const flowNodes = computed(() =>
  nodes.value.map(node => ({
    id: String(node.nodeKey),
    position: { x: Number(node.posX || 0), y: Number(node.posY || 0) },
    data: { label: node.name || node.nodeKey },
    style:
      node.nodeKey === lot.value?.currentNodeKey
        ? { border: '2px solid #18a058', background: '#e8f7ee', borderRadius: '8px', padding: '8px 12px' }
        : { borderRadius: '8px', padding: '8px 12px' }
  }))
);

const flowEdges = computed(() =>
  edges.value.map(edge => ({
    id: String(edge.edgeKey),
    source: String(edge.fromKey),
    target: String(edge.toKey),
    label: edge.edgeKind === 'rework' ? $t('page.mes.routeGraph.rework') : edge.isDefault ? $t('page.mes.routeGraph.defaultEdge') : edge.label || '',
    animated: edge.edgeKind === 'rework',
    style: edge.edgeKind === 'rework' ? { stroke: '#f0a020' } : {},
    markerEnd: MarkerType.ArrowClosed
  }))
);

const historyColumns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.wip.event'), key: 'eventType', width: 120, render: row => eventLabel(row.eventType) },
  { title: $t('page.mes.wip.fromNode'), key: 'fromNodeKey' },
  { title: $t('page.mes.wip.toNode'), key: 'toNodeKey' },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 80 },
  { title: $t('page.mes.wip.reasonCode'), key: 'reasonCode' },
  { title: $t('page.mes.wip.reason'), key: 'reason' }
]);

function eventLabel(event: string) {
  const map: Record<string, string> = {
    start: 'page.mes.wip.eventStart',
    advance: 'page.mes.wip.eventAdvance',
    hold: 'page.mes.wip.eventHold',
    release: 'page.mes.wip.eventRelease',
    split: 'page.mes.wip.eventSplit',
    merge: 'page.mes.wip.eventMerge',
    complete: 'page.mes.wip.eventComplete'
  };
  return $t((map[event] || 'page.mes.wip.event') as App.I18n.I18nKey);
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    waiting: 'page.mes.wip.waiting',
    hold: 'page.mes.wip.hold',
    completed: 'page.mes.wip.completed',
    merged: 'page.mes.wip.merged'
  };
  return $t((map[status] || 'page.mes.wip.status') as App.I18n.I18nKey);
}

async function load() {
  if (!lotId.value) return;
  const { data, error } = await fetchLot(lotId.value);
  if (error || !data) return;
  lot.value = data.lot;
  history.value = data.history || [];
  links.value = data.links || [];
  nodes.value = data.nodes || [];
  edges.value = data.edges || [];
  await nextTick();
  fitView({ padding: 0.2 });
}

async function advance() {
  const { data, error } = await advanceLot(lotId.value, inspection.value, defectCode.value);
  if (error || !data) return;
  const why = data.result?.reason || '';
  resultText.value = `${data.result?.action || ''} ${data.result?.nextNodeKey || ''} ${why}`.trim();
  load();
}

async function hold() {
  const { error } = await holdLot(lotId.value, reasonCode.value || 'HOLD', reason.value);
  if (error) return;
  load();
}

async function release() {
  const { error } = await releaseHold(lotId.value, reasonCode.value || 'RELEASE', reason.value);
  if (error) return;
  load();
}

onMounted(load);
watch(lotId, load);
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NSpace class="mb-12px" justify="space-between">
      <NButton @click="router.push({ name: 'lot_list' })">{{ $t('page.mes.wip.back') }}</NButton>
      <NSpace v-if="lot">
        <NTag>{{ lot.lotNo }}</NTag>
        <NTag type="info">{{ statusLabel(lot.status) }}</NTag>
        <span>{{ $t('page.mes.wip.quantity') }} {{ lot.quantity }}</span>
        <span>{{ $t('page.mes.wip.currentNode') }} {{ lot.currentNodeKey }}</span>
      </NSpace>
    </NSpace>
    <NGrid v-if="canEdit && lot && (lot.status === 'waiting' || lot.status === 'hold')" :cols="2" :x-gap="16" class="mb-12px">
      <NGi v-if="lot.status === 'waiting'">
        <NSpace>
          <NSelect
            v-model:value="inspection"
            class="w-140px"
            :options="[
              { label: $t('page.mes.routeGraph.pass'), value: 'pass' },
              { label: $t('page.mes.routeGraph.fail'), value: 'fail' }
            ]"
          />
          <NInput v-model:value="defectCode" class="w-160px" :placeholder="$t('page.mes.wip.defectCode')" />
          <NButton type="primary" @click="advance">{{ $t('page.mes.wip.advance') }}</NButton>
          <span v-if="resultText">{{ $t('page.mes.wip.result') }} {{ resultText }}</span>
        </NSpace>
      </NGi>
      <NGi>
        <NSpace>
          <NInput v-model:value="reasonCode" class="w-140px" :placeholder="$t('page.mes.wip.reasonCode')" />
          <NInput v-model:value="reason" class="w-200px" :placeholder="$t('page.mes.wip.reason')" />
          <NButton v-if="lot.status === 'waiting'" @click="hold">{{ $t('page.mes.wip.holdAction') }}</NButton>
          <NButton v-else type="primary" @click="release">{{ $t('page.mes.wip.releaseHold') }}</NButton>
        </NSpace>
      </NGi>
    </NGrid>
    <div class="mb-8px font-600">{{ $t('page.mes.wip.position') }}</div>
    <div class="flow-wrap mb-16px">
      <VueFlow id="lot-detail" :nodes="flowNodes" :edges="flowEdges" :nodes-draggable="false" :nodes-connectable="false" :edges-updatable="false" fit-view-on-init @nodes-initialized="fitView({ padding: 0.2 })">
        <Background />
        <Controls />
        <MiniMap />
      </VueFlow>
    </div>
    <div class="mb-8px font-600">{{ $t('page.mes.wip.history') }}</div>
    <NDataTable :columns="historyColumns" :data="history" size="small" class="mb-16px" />
    <div class="mb-8px font-600">{{ $t('page.mes.wip.genealogy') }}</div>
    <NSpace vertical>
      <div v-for="link in links" :key="link.id">
        {{ link.linkType === 'split' ? $t('page.mes.wip.linkSplit') : $t('page.mes.wip.linkMerge') }}
        {{ link.parentLotNo }} → {{ link.childLotNo }} ({{ link.quantity }})
      </div>
      <NEmpty v-if="!links.length" />
    </NSpace>
  </NCard>
</template>

<style scoped>
.flow-wrap {
  height: 420px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
}
</style>
