<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import type { DataTableColumns, SelectOption } from 'naive-ui';
import { Handle, MarkerType, Position, VueFlow, useVueFlow } from '@vue-flow/core';
import type { Connection } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Controls } from '@vue-flow/controls';
import { MiniMap } from '@vue-flow/minimap';
import { graphlib, layout as dagreLayout } from '@dagrejs/dagre';
import { useI18n } from 'vue-i18n';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { mesCreate, mesDelete, mesList, mesUpdate } from '@/service/api/mes';
import {
  createRouteVersion,
  fetchRouteGraph,
  fetchRouteVersions,
  releaseRouteGraph,
  resolveRouteStep,
  saveRouteGraph,
  validateRouteGraph
} from '@/service/api/route-graph';
import type {
  GraphCondition,
  GraphEdge,
  GraphIssue,
  GraphNode,
  GraphPredicate,
  RouteVersion
} from '@/service/api/route-graph';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import '@vue-flow/controls/dist/style.css';
import '@vue-flow/minimap/dist/style.css';

defineOptions({ name: 'BaseDataRoute' });

interface FlowNodeData {
  name: string;
  nodeType: string;
  operationID: number;
  recipeID: number;
  equipmentGroup: string;
}

interface FlowPredicate {
  field: string;
  op: string;
  value: string;
}

interface FlowEdgeData {
  edgeKind: 'normal' | 'rework';
  isDefault: boolean;
  priority: number;
  maxRework: number;
  onExceed: string;
  label: string;
  match: 'all' | 'any';
  predicates: FlowPredicate[];
}

interface FlowNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: FlowNodeData;
}

interface FlowEdge {
  id: string;
  source: string;
  target: string;
  type?: string;
  label?: string;
  animated?: boolean;
  markerEnd?: string;
  style?: { stroke: string; strokeDasharray?: string };
  data: FlowEdgeData;
}

const { hasAuth } = useAuth();
const { locale } = useI18n();
const { screenToFlowCoordinate, fitView } = useVueFlow({ id: 'route-editor' });

const canAdd = computed(() => hasAuth('base:route:add'));
const canEdit = computed(() => hasAuth('base:route:edit'));
const canDelete = computed(() => hasAuth('base:route:delete'));

const routes = ref<Record<string, any>[]>([]);
const products = ref<Record<string, any>[]>([]);
const operations = ref<Record<string, any>[]>([]);
const recipes = ref<Record<string, any>[]>([]);
const selectedRouteId = ref<number | null>(null);
const versions = ref<RouteVersion[]>([]);
const versionId = ref<number | null>(null);
const nodes = ref<FlowNode[]>([]);
const edges = ref<FlowEdge[]>([]);
const issues = ref<GraphIssue[]>([]);
const selectedNodeId = ref<string | null>(null);
const selectedEdgeId = ref<string | null>(null);
const busy = ref(false);
const modalVisible = ref(false);
const editingRouteId = ref<number | null>(null);
const form = reactive({
  productID: null as number | null,
  routeCode: '',
  routeName: '',
  isDefault: 0,
  description: '',
  status: 1
});
const sim = reactive({ node: '', result: 'pass', lotType: 'production', rework: 0, text: '' });

const currentVersion = computed(() => versions.value.find(item => item.id === versionId.value) || null);
const editable = computed(() => currentVersion.value?.state === 'draft' && canEdit.value);
const selectedNode = computed(() => nodes.value.find(item => item.id === selectedNodeId.value) || null);
const selectedEdge = computed(() => edges.value.find(item => item.id === selectedEdgeId.value) || null);
const selectedRoute = computed(() => routes.value.find(item => item.id === selectedRouteId.value) || null);

const numberFields = new Set(['lot.priority', 'rework.count']);
const stringOps = ['eq', 'ne', 'in', 'contains'];
const numberOps = ['eq', 'ne', 'gt', 'gte', 'lt', 'lte', 'in'];

const fieldOptions = computed<SelectOption[]>(() => [
  { label: $t('page.mes.routeGraph.fieldInspectionResult'), value: 'inspection.result' },
  { label: $t('page.mes.routeGraph.fieldInspectionGrade'), value: 'inspection.grade' },
  { label: $t('page.mes.routeGraph.fieldDefectCode'), value: 'defect.code' },
  { label: $t('page.mes.routeGraph.fieldProductCode'), value: 'lot.productCode' },
  { label: $t('page.mes.routeGraph.fieldPriority'), value: 'lot.priority' },
  { label: $t('page.mes.routeGraph.fieldLotType'), value: 'lot.type' },
  { label: $t('page.mes.routeGraph.fieldReworkCount'), value: 'rework.count' }
]);

const opLabel: Record<string, string> = {
  eq: 'page.mes.routeGraph.opEq',
  ne: 'page.mes.routeGraph.opNe',
  gt: 'page.mes.routeGraph.opGt',
  gte: 'page.mes.routeGraph.opGte',
  lt: 'page.mes.routeGraph.opLt',
  lte: 'page.mes.routeGraph.opLte',
  in: 'page.mes.routeGraph.opIn',
  contains: 'page.mes.routeGraph.opContains'
};

const productOptions = computed<SelectOption[]>(() =>
  products.value.map(item => ({ label: `${item.productCode} ${item.productName}`, value: item.id }))
);
const operationOptions = computed<SelectOption[]>(() => [
  { label: $t('page.mes.routeGraph.none'), value: 0 },
  ...operations.value.map(item => ({ label: `${item.operationCode} ${item.operationName}`, value: item.id }))
]);
const recipeOptions = computed<SelectOption[]>(() => {
  const opId = selectedNode.value?.data.operationID;
  const rows = recipes.value.filter(item => !opId || item.operationID === opId);
  return [
    { label: $t('page.mes.routeGraph.none'), value: 0 },
    ...rows.map(item => ({ label: item.recipeName, value: item.id }))
  ];
});
const nodeOptions = computed<SelectOption[]>(() =>
  nodes.value.map(item => ({ label: item.data.name || item.id, value: item.id }))
);
const versionOptions = computed<SelectOption[]>(() =>
  versions.value.map(item => ({ label: `v${item.versionNo} ${stateLabel(item.state)}`, value: item.id }))
);

const columns = computed<DataTableColumns<Record<string, any>>>(() => [
  { title: $t('page.mes.field.routeCode'), key: 'routeCode' },
  { title: $t('page.mes.field.routeName'), key: 'routeName' },
  { title: $t('page.mes.field.version'), key: 'version' },
  {
    title: $t('page.mes.field.status'),
    key: 'status',
    render: row => (row.status === 1 ? $t('page.mes.enabled') : $t('page.mes.disabled'))
  }
]);

function stateLabel(state: string) {
  if (state === 'released') return $t('page.mes.routeGraph.stateReleased');
  if (state === 'obsolete') return $t('page.mes.routeGraph.stateObsolete');
  return $t('page.mes.routeGraph.stateDraft');
}

function defaultName(type: string) {
  if (type === 'start') return $t('page.mes.routeGraph.nodeStart');
  if (type === 'end') return $t('page.mes.routeGraph.nodeEnd');
  if (type === 'decision') return $t('page.mes.routeGraph.nodeDecision');
  return $t('page.mes.routeGraph.nodeOperation');
}

function opsFor(field: string) {
  const ops = numberFields.has(field) ? numberOps : stringOps;
  return ops.map(op => ({ label: $t(opLabel[op] as App.I18n.I18nKey), value: op }));
}

function decorate(edge: FlowEdge): FlowEdge {
  const data = edge.data!;
  const rework = data.edgeKind === 'rework';
  let label = data.label;
  if (!label && rework) label = $t('page.mes.routeGraph.rework');
  if (!label && data.isDefault) label = $t('page.mes.routeGraph.defaultEdge');
  return {
    ...edge,
    type: 'smoothstep',
    label,
    animated: rework,
    markerEnd: MarkerType.ArrowClosed,
    style: rework ? { stroke: '#d97706', strokeDasharray: '6 4' } : { stroke: '#64748b' }
  };
}

function encodeValue(field: string, op: string, raw: string): unknown {
  if (op === 'in') {
    const parts = raw
      .split(',')
      .map(item => item.trim())
      .filter(Boolean);
    return numberFields.has(field) ? parts.map(item => Number(item)) : parts;
  }
  return numberFields.has(field) ? Number(raw) : raw;
}

function predicatesFrom(condition?: GraphCondition | null): { match: 'all' | 'any'; predicates: FlowPredicate[] } {
  const toRow = (item: GraphPredicate): FlowPredicate => ({
    field: item.field,
    op: item.op,
    value: Array.isArray(item.value) ? item.value.join(',') : String(item.value ?? '')
  });
  if (condition?.any?.length) return { match: 'any', predicates: condition.any.map(toRow) };
  return { match: 'all', predicates: (condition?.all || []).map(toRow) };
}

function conditionFrom(data: FlowEdgeData): GraphCondition | null {
  const rows = data.predicates.filter(item => item.field && item.op && item.value !== '');
  if (!rows.length || data.isDefault) return null;
  const predicates = rows.map(item => ({
    field: item.field,
    op: item.op,
    value: encodeValue(item.field, item.op, item.value)
  }));
  return data.match === 'any' ? { any: predicates } : { all: predicates };
}

function toPayload(): { nodes: GraphNode[]; edges: GraphEdge[] } {
  return {
    nodes: nodes.value.map(node => ({
      nodeKey: node.id,
      nodeType: node.data.nodeType,
      name: node.data.name,
      operationID: Number(node.data.operationID) || 0,
      recipeID: Number(node.data.recipeID) || 0,
      equipmentGroup: node.data.equipmentGroup || '',
      posX: node.position.x,
      posY: node.position.y
    })),
    edges: edges.value.map(edge => ({
      edgeKey: edge.id,
      fromKey: edge.source,
      toKey: edge.target,
      edgeKind: edge.data!.edgeKind,
      isDefault: edge.data!.isDefault,
      priority: Number(edge.data!.priority) || 0,
      condition: conditionFrom(edge.data!),
      maxRework: Number(edge.data!.maxRework) || 0,
      onExceed: edge.data!.edgeKind === 'rework' ? 'hold' : '',
      label: edge.data!.label || ''
    }))
  };
}

async function loadMasters() {
  const [routeRes, productRes, operationRes, recipeRes] = await Promise.all([
    mesList('baseProcessRoute', { page: 0, limit: 100, sort: '-id' }),
    mesList('baseProduct', { page: 0, limit: 200, sort: '-id' }),
    mesList('baseOperation', { page: 0, limit: 200, sort: '-id' }),
    mesList('baseRecipe', { page: 0, limit: 200, sort: '-id' })
  ]);
  routes.value = routeRes.data?.baseProcessRoutes || [];
  products.value = productRes.data?.baseProducts || [];
  operations.value = operationRes.data?.baseOperations || [];
  recipes.value = recipeRes.data?.baseRecipes || [];
}

async function selectRoute(id: number) {
  selectedRouteId.value = id;
  issues.value = [];
  sim.text = '';
  const { data, error } = await fetchRouteVersions(id);
  if (error || !data) return;
  versions.value = data.versions || [];
  if (!versions.value.length) {
    versionId.value = null;
    nodes.value = [];
    edges.value = [];
    return;
  }
  const preferred = versions.value.find(item => item.state === 'released') || versions.value[0];
  await selectVersion(preferred.id);
}

async function selectVersion(id: number) {
  if (!selectedRouteId.value) return;
  versionId.value = id;
  selectedNodeId.value = null;
  selectedEdgeId.value = null;
  issues.value = [];
  sim.text = '';
  const { data, error } = await fetchRouteGraph(selectedRouteId.value, id);
  if (error || !data) return;
  nodes.value = (data.nodes || []).map(node => ({
    id: node.nodeKey,
    type: node.nodeType,
    position: { x: node.posX, y: node.posY },
    data: {
      name: node.name,
      nodeType: node.nodeType,
      operationID: node.operationID || 0,
      recipeID: node.recipeID || 0,
      equipmentGroup: node.equipmentGroup || ''
    }
  }));
  edges.value = (data.edges || []).map(edge => {
    const parsed = predicatesFrom(edge.condition);
    return decorate({
      id: edge.edgeKey,
      source: edge.fromKey,
      target: edge.toKey,
      data: {
        edgeKind: edge.edgeKind === 'rework' ? 'rework' : 'normal',
        isDefault: edge.isDefault,
        priority: edge.priority,
        maxRework: edge.maxRework,
        onExceed: edge.onExceed || 'hold',
        label: edge.label || '',
        match: parsed.match,
        predicates: parsed.predicates
      }
    });
  });
  sim.node = nodes.value.find(node => node.data.nodeType === 'decision')?.id || nodes.value[0]?.id || '';
  await nextTick();
  fitView({ padding: 0.16 });
}

const flowWrap = ref<HTMLElement | null>(null);
let flowObserver: ResizeObserver | null = null;

function fitGraph() {
  nextTick(() => {
    fitView({ padding: 0.2, duration: 180 });
    window.setTimeout(() => fitView({ padding: 0.2, duration: 180 }), 60);
  });
}

function addNode(type: string, position?: { x: number; y: number }) {
  if (!editable.value) return;
  const id = `n_${Date.now().toString(36)}`;
  const spot = position || { x: 80 + nodes.value.length * 30, y: 80 + (nodes.value.length % 4) * 20 };
  nodes.value = [
    ...nodes.value,
    {
      id,
      type,
      position: spot,
      data: { name: defaultName(type), nodeType: type, operationID: 0, recipeID: 0, equipmentGroup: '' }
    }
  ];
  selectedNodeId.value = id;
  selectedEdgeId.value = null;
}

function onDragStart(event: DragEvent, type: string) {
  event.dataTransfer?.setData('application/vueflow', type);
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
}

function onDrop(event: DragEvent) {
  const type = event.dataTransfer?.getData('application/vueflow');
  if (!type) return;
  addNode(type, screenToFlowCoordinate({ x: event.clientX, y: event.clientY }));
}

function onConnect(connection: Connection) {
  if (!editable.value || !connection.source || !connection.target) return;
  const outgoing = edges.value.filter(item => item.source === connection.source);
  const id = `e_${Date.now().toString(36)}`;
  edges.value = [
    ...edges.value,
    decorate({
      id,
      source: connection.source,
      target: connection.target,
      data: {
        edgeKind: 'normal',
        isDefault: outgoing.length === 0,
        priority: (outgoing.length + 1) * 10,
        maxRework: 0,
        onExceed: 'hold',
        label: '',
        match: 'all',
        predicates: []
      }
    })
  ];
  selectedEdgeId.value = id;
  selectedNodeId.value = null;
}

function patchNode(partial: Partial<FlowNodeData>) {
  if (!selectedNodeId.value) return;
  nodes.value = nodes.value.map(node =>
    node.id === selectedNodeId.value ? { ...node, data: { ...node.data, ...partial } } : node
  );
}

function patchEdge(partial: Partial<FlowEdgeData>) {
  if (!selectedEdgeId.value) return;
  edges.value = edges.value.map(edge => {
    if (edge.id !== selectedEdgeId.value) return edge;
    const data = { ...edge.data!, ...partial };
    if (partial.isDefault) {
      return decorate({ ...edge, data });
    }
    return decorate({ ...edge, data });
  });
  if (partial.isDefault) {
    const source = edges.value.find(item => item.id === selectedEdgeId.value)?.source;
    edges.value = edges.value.map(edge => {
      if (edge.source !== source || edge.id === selectedEdgeId.value) return edge;
      return decorate({ ...edge, data: { ...edge.data!, isDefault: false } });
    });
  }
}

function connectSelected(target: string) {
  if (!selectedNodeId.value || !target || target === selectedNodeId.value) return;
  onConnect({ source: selectedNodeId.value, target, sourceHandle: null, targetHandle: null });
}

function removeNode() {
  if (!selectedNodeId.value || !editable.value) return;
  const id = selectedNodeId.value;
  nodes.value = nodes.value.filter(node => node.id !== id);
  edges.value = edges.value.filter(edge => edge.source !== id && edge.target !== id);
  selectedNodeId.value = null;
}

function removeEdge() {
  if (!selectedEdgeId.value || !editable.value) return;
  edges.value = edges.value.filter(edge => edge.id !== selectedEdgeId.value);
  selectedEdgeId.value = null;
}

function addPredicate() {
  const edge = selectedEdge.value;
  if (!edge) return;
  patchEdge({ predicates: [...edge.data!.predicates, { field: 'inspection.result', op: 'eq', value: '' }] });
}

function updatePredicate(index: number, partial: Partial<FlowPredicate>) {
  const edge = selectedEdge.value;
  if (!edge) return;
  const predicates = edge.data!.predicates.map((item, i) => (i === index ? { ...item, ...partial } : item));
  patchEdge({ predicates });
}

function removePredicate(index: number) {
  const edge = selectedEdge.value;
  if (!edge) return;
  patchEdge({ predicates: edge.data!.predicates.filter((_, i) => i !== index) });
}

function autoLayout() {
  const graph = new graphlib.Graph();
  graph.setDefaultEdgeLabel(() => ({}));
  graph.setGraph({ rankdir: 'LR', nodesep: 48, ranksep: 90 });
  nodes.value.forEach(node => graph.setNode(node.id, { width: 168, height: 64 }));
  edges.value.forEach(edge => graph.setEdge(edge.source, edge.target));
  dagreLayout(graph);
  nodes.value = nodes.value.map(node => {
    const pos = graph.node(node.id);
    return { ...node, position: { x: pos.x - 84, y: pos.y - 32 } };
  });
  nextTick(() => fitView({ padding: 0.2 }));
}

async function createDraft(copyFrom?: number) {
  if (!selectedRouteId.value || !canAdd.value) return;
  busy.value = true;
  try {
    const { data, error } = await createRouteVersion(selectedRouteId.value, { copyFrom: copyFrom || 0, note: '' });
    if (error || !data) return;
    await selectRoute(selectedRouteId.value);
    await selectVersion(data.version.id);
  } finally {
    busy.value = false;
  }
}

async function save(silent = false) {
  if (!selectedRouteId.value || !versionId.value || !editable.value) return true;
  busy.value = true;
  try {
    const { error } = await saveRouteGraph(selectedRouteId.value, versionId.value, toPayload());
    if (error) return false;
    if (!silent) window.$message?.success($t('page.mes.routeGraph.savedOk'));
    return true;
  } catch {
    return false;
  } finally {
    busy.value = false;
  }
}

async function validate() {
  if (!selectedRouteId.value || !versionId.value) return;
  if (editable.value && !(await save(true))) return;
  busy.value = true;
  try {
    const { data, error } = await validateRouteGraph(selectedRouteId.value, versionId.value);
    if (error || !data) return;
    issues.value = data.issues || [];
    if (data.valid) window.$message?.success($t('page.mes.routeGraph.validOk'));
  } catch {
    /* shown by request */
  } finally {
    busy.value = false;
  }
}

async function release() {
  if (!selectedRouteId.value || !versionId.value || !editable.value) return;
  if (!(await save(true))) return;
  busy.value = true;
  try {
    const { data, error } = await releaseRouteGraph(selectedRouteId.value, versionId.value);
    if (error || !data) return;
    issues.value = data.issues || [];
    if (data.released) {
      window.$message?.success($t('page.mes.routeGraph.releasedOk'));
      await loadMasters();
      await selectRoute(selectedRouteId.value);
      await selectVersion(versionId.value);
    }
  } catch {
    /* shown by request */
  } finally {
    busy.value = false;
  }
}

async function simulate() {
  if (!selectedRouteId.value || !versionId.value || !sim.node) return;
  if (editable.value && !(await save(true))) return;
  const reworkCounts: Record<string, number> = {};
  edges.value.forEach(edge => {
    if (edge.source === sim.node && edge.data?.edgeKind === 'rework') reworkCounts[edge.id] = Number(sim.rework) || 0;
  });
  try {
    const { data, error } = await resolveRouteStep(selectedRouteId.value, versionId.value, {
      currentNodeKey: sim.node,
      context: { inspection: { result: sim.result }, lot: { type: sim.lotType }, reworkCounts }
    });
    if (error || !data) {
      sim.text = '';
      return;
    }
    const action =
      data.action === 'hold'
        ? $t('page.mes.routeGraph.resultHold')
        : data.action === 'end'
          ? $t('page.mes.routeGraph.resultEnd')
          : $t('page.mes.routeGraph.resultMove');
    const reasonMap: Record<string, string> = {
      matched: $t('page.mes.routeGraph.reasonMatched'),
      default: $t('page.mes.routeGraph.reasonDefault'),
      rework_exceeded: $t('page.mes.routeGraph.reasonRework'),
      already_end: $t('page.mes.routeGraph.reasonEnd')
    };
    sim.text = `${action} ${data.nextNodeName || data.nextNodeKey || ''} · ${reasonMap[data.reason] || data.reason}`;
  } catch {
    sim.text = '';
  }
}

function openCreate() {
  editingRouteId.value = null;
  form.productID = products.value[0]?.id ?? null;
  form.routeCode = '';
  form.routeName = '';
  form.isDefault = 0;
  form.description = '';
  form.status = 1;
  modalVisible.value = true;
}

function openEdit() {
  const row = selectedRoute.value;
  if (!row) return;
  editingRouteId.value = row.id;
  form.productID = row.productID;
  form.routeCode = row.routeCode;
  form.routeName = row.routeName;
  form.isDefault = row.isDefault;
  form.description = row.description || '';
  form.status = row.status;
  modalVisible.value = true;
}

async function submitRoute() {
  if (!form.productID || !form.routeCode || !form.routeName) return;
  const body = { ...form, productID: form.productID };
  busy.value = true;
  try {
    if (editingRouteId.value) {
      const updated = await mesUpdate('baseProcessRoute', editingRouteId.value, body);
      if (updated.error) return;
    } else {
      const created = await mesCreate('baseProcessRoute', body);
      if (created.error || !created.data) return;
      await loadMasters();
      await selectRoute(created.data.id);
      if (canAdd.value) await createDraft();
    }
    modalVisible.value = false;
    await loadMasters();
  } catch {
    /* shown by request */
  } finally {
    busy.value = false;
  }
}

async function removeRoute() {
  if (!selectedRouteId.value) return;
  await mesDelete('baseProcessRoute', selectedRouteId.value);
  selectedRouteId.value = null;
  versions.value = [];
  nodes.value = [];
  edges.value = [];
  await loadMasters();
}

watch(locale, () => {
  edges.value = edges.value.map(edge => decorate(edge));
});

onMounted(async () => {
  await loadMasters();
  const seeded = routes.value.find(item => item.routeCode === 'ROUTE-CMOS');
  if (seeded) await selectRoute(seeded.id);
  if (flowWrap.value) {
    flowObserver = new ResizeObserver(() => fitGraph());
    flowObserver.observe(flowWrap.value);
  }
});

onUnmounted(() => flowObserver?.disconnect());
</script>

<template>
  <div class="flex flex-col gap-12px">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <NSpace class="mb-12px">
        <NButton v-if="canAdd" type="primary" @click="openCreate">{{ $t('page.mes.routeGraph.createRoute') }}</NButton>
        <NButton v-if="canEdit && selectedRoute" @click="openEdit">{{ $t('page.mes.routeGraph.editRoute') }}</NButton>
        <NPopconfirm v-if="canDelete && selectedRoute" @positive-click="removeRoute">
          <template #trigger>
            <NButton type="error" ghost>{{ $t('page.mes.routeGraph.deleteRoute') }}</NButton>
          </template>
          {{ $t('common.confirmDelete') }}
        </NPopconfirm>
      </NSpace>
      <NDataTable
        :columns="columns"
        :data="routes"
        :row-key="row => row.id"
        :row-props="row => ({ style: 'cursor: pointer', onClick: () => selectRoute(row.id) })"
        size="small"
        max-height="140"
      />
    </NCard>

    <NCard v-if="!selectedRoute" :bordered="false" size="small" class="card-wrapper">
      <NEmpty :description="$t('page.mes.routeGraph.selectRoute')" />
    </NCard>

    <NCard v-else :bordered="false" size="small" class="card-wrapper" :title="selectedRoute.routeName">
      <NSpace class="mb-12px" align="center">
        <span>{{ $t('page.mes.routeGraph.versions') }}</span>
        <NSelect
          v-if="versions.length"
          :value="versionId"
          :options="versionOptions"
          class="w-220px"
          @update:value="selectVersion"
        />
        <NTag
          v-if="currentVersion"
          :type="
            currentVersion.state === 'released' ? 'success' : currentVersion.state === 'draft' ? 'warning' : 'default'
          "
        >
          {{ stateLabel(currentVersion.state) }}
        </NTag>
        <NButton v-if="canAdd" :loading="busy" @click="createDraft()">{{ $t('page.mes.routeGraph.newDraft') }}</NButton>
        <NButton v-if="canAdd && versionId" :loading="busy" @click="createDraft(versionId)">
          {{ $t('page.mes.routeGraph.copyDraft') }}
        </NButton>
        <NButton v-if="editable" type="primary" :loading="busy" @click="() => save()">
          {{ $t('page.mes.routeGraph.save') }}
        </NButton>
        <NButton :loading="busy" :disabled="!versionId" @click="validate">
          {{ $t('page.mes.routeGraph.validate') }}
        </NButton>
        <NPopconfirm v-if="editable" @positive-click="release">
          <template #trigger>
            <NButton type="primary" ghost :loading="busy">{{ $t('page.mes.routeGraph.release') }}</NButton>
          </template>
          {{ $t('page.mes.routeGraph.release') }}
        </NPopconfirm>
        <NButton :disabled="!nodes.length" @click="autoLayout">{{ $t('page.mes.routeGraph.autoLayout') }}</NButton>
      </NSpace>
      <NAlert v-if="currentVersion && currentVersion.state !== 'draft'" type="info" class="mb-12px">
        {{ $t('page.mes.routeGraph.readonlyHint') }}
      </NAlert>
      <NAlert v-if="!versions.length" type="warning" class="mb-12px">{{ $t('page.mes.routeGraph.noVersion') }}</NAlert>
      <NAlert v-if="issues.length" type="error" class="mb-12px" :title="$t('page.mes.routeGraph.issueTitle')">
        <div v-for="issue in issues" :key="`${issue.code}-${issue.ref || ''}`">{{ issue.message }} {{ issue.ref }}</div>
      </NAlert>

      <div v-if="versionId" class="flex gap-12px">
        <div class="w-140px flex flex-col gap-8px">
          <div class="text-13px font-600">{{ $t('page.mes.routeGraph.palette') }}</div>
          <div
            v-for="item in [
              ['start', $t('page.mes.routeGraph.addStart')],
              ['operation', $t('page.mes.routeGraph.addOperation')],
              ['decision', $t('page.mes.routeGraph.addDecision')],
              ['end', $t('page.mes.routeGraph.addEnd')]
            ]"
            :key="item[0]"
            class="palette-item"
            :class="{ disabled: !editable }"
            :draggable="editable"
            @dragstart="onDragStart($event, item[0])"
            @click="addNode(item[0])"
          >
            {{ item[1] }}
          </div>
          <div class="text-12px opacity-70">{{ $t('page.mes.routeGraph.dragHint') }}</div>
        </div>

        <div ref="flowWrap" class="flow-wrap" @drop="onDrop" @dragover.prevent>
          <VueFlow
            id="route-editor"
            v-model:nodes="nodes"
            v-model:edges="edges"
            :nodes-draggable="editable"
            :nodes-connectable="editable"
            :elements-selectable="true"
            fit-view-on-init
            @nodes-initialized="fitGraph"
            @connect="onConnect"
            @node-click="
              ({ node }) => {
                selectedNodeId = node.id;
                selectedEdgeId = null;
              }
            "
            @edge-click="
              ({ edge }) => {
                selectedEdgeId = edge.id;
                selectedNodeId = null;
              }
            "
            @pane-click="
              selectedNodeId = null;
              selectedEdgeId = null;
            "
          >
            <template #node-start="{ data }">
              <div class="mes-node mes-start">
                {{ data.name }}
                <Handle type="source" :position="Position.Right" />
              </div>
            </template>
            <template #node-end="{ data }">
              <div class="mes-node mes-end">
                <Handle type="target" :position="Position.Left" />
                {{ data.name }}
              </div>
            </template>
            <template #node-operation="{ data }">
              <div class="mes-node mes-op">
                <Handle type="target" :position="Position.Left" />
                <div class="mes-kind">{{ $t('page.mes.routeGraph.nodeOperation') }}</div>
                {{ data.name }}
                <Handle type="source" :position="Position.Right" />
              </div>
            </template>
            <template #node-decision="{ data }">
              <div class="mes-node mes-decision">
                <Handle type="target" :position="Position.Left" />
                <div class="mes-kind">{{ $t('page.mes.routeGraph.nodeDecision') }}</div>
                {{ data.name }}
                <Handle type="source" :position="Position.Right" />
              </div>
            </template>
            <Background />
            <Controls />
            <MiniMap />
          </VueFlow>
        </div>

        <div class="w-300px overflow-auto">
          <template v-if="selectedNode">
            <div class="mb-8px font-600">{{ $t('page.mes.routeGraph.nodePanel') }}</div>
            <NForm label-placement="top" size="small">
              <NFormItem :label="$t('page.mes.routeGraph.nodeName')">
                <NInput
                  :value="selectedNode.data.name"
                  :disabled="!editable"
                  @update:value="name => patchNode({ name })"
                />
              </NFormItem>
              <NFormItem v-if="selectedNode.data.nodeType === 'operation'" :label="$t('page.mes.field.operation')">
                <NSelect
                  :value="selectedNode.data.operationID"
                  :options="operationOptions"
                  :disabled="!editable"
                  filterable
                  @update:value="operationID => patchNode({ operationID, recipeID: 0 })"
                />
              </NFormItem>
              <NFormItem v-if="selectedNode.data.nodeType === 'operation'" :label="$t('page.mes.routeGraph.recipe')">
                <NSelect
                  :value="selectedNode.data.recipeID"
                  :options="recipeOptions"
                  :disabled="!editable"
                  @update:value="recipeID => patchNode({ recipeID })"
                />
              </NFormItem>
              <NFormItem
                v-if="selectedNode.data.nodeType === 'operation'"
                :label="$t('page.mes.routeGraph.equipmentGroup')"
              >
                <NInput
                  :value="selectedNode.data.equipmentGroup"
                  :disabled="!editable"
                  @update:value="equipmentGroup => patchNode({ equipmentGroup })"
                />
              </NFormItem>
              <NFormItem :label="$t('page.mes.routeGraph.connectTo')">
                <NSelect
                  :options="nodeOptions.filter(item => item.value !== selectedNode?.id)"
                  :disabled="!editable"
                  :placeholder="$t('page.mes.selectPlaceholder')"
                  @update:value="connectSelected"
                />
              </NFormItem>
              <NButton v-if="editable" type="error" ghost size="small" @click="removeNode">
                {{ $t('page.mes.routeGraph.deleteNode') }}
              </NButton>
            </NForm>
          </template>
          <template v-else-if="selectedEdge && selectedEdge.data">
            <div class="mb-8px font-600">{{ $t('page.mes.routeGraph.edgePanel') }}</div>
            <NForm label-placement="top" size="small">
              <NFormItem :label="$t('page.mes.routeGraph.edgeKind')">
                <NSelect
                  :value="selectedEdge.data.edgeKind"
                  :disabled="!editable"
                  :options="[
                    { label: $t('page.mes.routeGraph.kindNormal'), value: 'normal' },
                    { label: $t('page.mes.routeGraph.kindRework'), value: 'rework' }
                  ]"
                  @update:value="
                    edgeKind =>
                      patchEdge({ edgeKind, maxRework: edgeKind === 'rework' ? selectedEdge?.data?.maxRework || 2 : 0 })
                  "
                />
              </NFormItem>
              <NFormItem :label="$t('page.mes.routeGraph.defaultEdge')">
                <NSwitch
                  :value="selectedEdge.data.isDefault"
                  :disabled="!editable"
                  @update:value="isDefault => patchEdge({ isDefault })"
                />
              </NFormItem>
              <NFormItem :label="$t('page.mes.routeGraph.priority')">
                <NInputNumber
                  :value="selectedEdge.data.priority"
                  :disabled="!editable"
                  class="w-full"
                  @update:value="priority => patchEdge({ priority: priority || 0 })"
                />
              </NFormItem>
              <NFormItem :label="$t('page.mes.field.description')">
                <NInput
                  :value="selectedEdge.data.label"
                  :disabled="!editable"
                  @update:value="label => patchEdge({ label })"
                />
              </NFormItem>
              <template v-if="selectedEdge.data.edgeKind === 'rework'">
                <NFormItem :label="$t('page.mes.routeGraph.maxRework')">
                  <NInputNumber
                    :value="selectedEdge.data.maxRework"
                    :min="1"
                    :disabled="!editable"
                    class="w-full"
                    @update:value="maxRework => patchEdge({ maxRework: maxRework || 1 })"
                  />
                </NFormItem>
                <div class="mb-8px text-12px opacity-70">{{ $t('page.mes.routeGraph.onExceedHold') }}</div>
              </template>
              <template v-if="!selectedEdge.data.isDefault">
                <NFormItem :label="$t('page.mes.routeGraph.conditionMode')">
                  <NSelect
                    :value="selectedEdge.data.match"
                    :disabled="!editable"
                    :options="[
                      { label: $t('page.mes.routeGraph.matchAll'), value: 'all' },
                      { label: $t('page.mes.routeGraph.matchAny'), value: 'any' }
                    ]"
                    @update:value="match => patchEdge({ match })"
                  />
                </NFormItem>
                <div
                  v-for="(pred, index) in selectedEdge.data.predicates"
                  :key="index"
                  class="mb-8px flex flex-col gap-4px"
                >
                  <NSelect
                    :value="pred.field"
                    :options="fieldOptions"
                    :disabled="!editable"
                    @update:value="field => updatePredicate(index, { field, op: 'eq' })"
                  />
                  <NSelect
                    :value="pred.op"
                    :options="opsFor(pred.field)"
                    :disabled="!editable"
                    @update:value="op => updatePredicate(index, { op })"
                  />
                  <NInput
                    :value="pred.value"
                    :disabled="!editable"
                    :placeholder="pred.op === 'in' ? $t('page.mes.routeGraph.inHint') : ''"
                    @update:value="value => updatePredicate(index, { value })"
                  />
                  <NButton v-if="editable" text type="error" @click="removePredicate(index)">
                    {{ $t('common.delete') }}
                  </NButton>
                </div>
                <NButton v-if="editable" size="small" class="mb-8px" @click="addPredicate">
                  {{ $t('page.mes.routeGraph.addPredicate') }}
                </NButton>
              </template>
              <NButton v-if="editable" type="error" ghost size="small" @click="removeEdge">
                {{ $t('page.mes.routeGraph.deleteEdge') }}
              </NButton>
            </NForm>
          </template>
          <NEmpty v-else :description="$t('page.mes.routeGraph.emptySelection')" />

          <NDivider />
          <div class="mb-8px font-600">{{ $t('page.mes.routeGraph.simulate') }}</div>
          <NForm label-placement="top" size="small">
            <NFormItem :label="$t('page.mes.routeGraph.currentNode')">
              <NSelect v-model:value="sim.node" :options="nodeOptions" />
            </NFormItem>
            <NFormItem :label="$t('page.mes.routeGraph.inspectionResult')">
              <NSelect
                v-model:value="sim.result"
                :options="[
                  { label: $t('page.mes.routeGraph.pass'), value: 'pass' },
                  { label: $t('page.mes.routeGraph.fail'), value: 'fail' }
                ]"
              />
            </NFormItem>
            <NFormItem :label="$t('page.mes.routeGraph.lotType')">
              <NSelect
                v-model:value="sim.lotType"
                :options="[
                  { label: $t('page.mes.routeGraph.production'), value: 'production' },
                  { label: $t('page.mes.routeGraph.engineering'), value: 'engineering' }
                ]"
              />
            </NFormItem>
            <NFormItem :label="$t('page.mes.routeGraph.reworkCount')">
              <NInputNumber v-model:value="sim.rework" :min="0" class="w-full" />
            </NFormItem>
            <NButton size="small" @click="simulate">{{ $t('page.mes.routeGraph.runResolve') }}</NButton>
            <div v-if="sim.text" class="mt-8px text-13px">{{ sim.text }}</div>
          </NForm>
        </div>
      </div>
    </NCard>

    <NModal
      v-model:show="modalVisible"
      preset="card"
      :title="editingRouteId ? $t('page.mes.routeGraph.editRoute') : $t('page.mes.routeGraph.createRoute')"
      class="w-520px"
    >
      <NForm label-placement="left" label-width="90">
        <NFormItem :label="$t('page.mes.field.product')">
          <NSelect v-model:value="form.productID" :options="productOptions" filterable />
        </NFormItem>
        <NFormItem :label="$t('page.mes.field.routeCode')">
          <NInput v-model:value="form.routeCode" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.field.routeName')">
          <NInput v-model:value="form.routeName" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.field.isDefault')">
          <NSelect
            v-model:value="form.isDefault"
            :options="[
              { label: $t('page.mes.yes'), value: 1 },
              { label: $t('page.mes.no'), value: 0 }
            ]"
          />
        </NFormItem>
        <NFormItem :label="$t('page.mes.field.description')">
          <NInput v-model:value="form.description" type="textarea" />
        </NFormItem>
        <NFormItem :label="$t('page.mes.field.status')">
          <NSelect
            v-model:value="form.status"
            :options="[
              { label: $t('page.mes.enabled'), value: 1 },
              { label: $t('page.mes.disabled'), value: 2 }
            ]"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="modalVisible = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="busy" @click="submitRoute">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.flow-wrap {
  flex: 1;
  min-width: 0;
  height: min(560px, calc(100vh - 180px));
  min-height: 420px;
  border: 1px solid var(--n-border-color, #e5e7eb);
  border-radius: 8px;
  overflow: hidden;
}

.flow-wrap :deep(.vue-flow) {
  width: 100%;
  height: 100%;
}

.palette-item {
  padding: 8px 10px;
  border: 1px dashed #94a3b8;
  border-radius: 8px;
  cursor: grab;
  text-align: center;
}

.palette-item.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.mes-node {
  min-width: 128px;
  padding: 8px 12px;
  border: 1px solid #94a3b8;
  border-radius: 8px;
  background: #fff;
  color: #0f172a;
  text-align: center;
}

.mes-kind {
  font-size: 11px;
  opacity: 0.65;
}

.mes-start,
.mes-end {
  border-radius: 999px;
}

.mes-start {
  border-color: #059669;
}

.mes-op {
  border-color: #2563eb;
}

.mes-decision {
  border-color: #d97706;
  background: #fffbeb;
}

.mes-end {
  border-color: #475569;
}

:deep(.vue-flow__node.selected) .mes-node {
  box-shadow: 0 0 0 2px #2563eb;
}
</style>
