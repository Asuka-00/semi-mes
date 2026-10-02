<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { $t } from '@/locales';
import { fetchChart, fetchPlans, fetchPolicies, savePolicy } from '@/service/api/quality';
import { fetchFeatures } from '@/service/api/notice';
import { fetchEquipment } from '@/service/api/track';
import type { ChartView, InspectPlan } from '@/service/api/quality';

const spcOn = ref(true);
const plans = ref<InspectPlan[]>([]);
const tools = ref<Array<{ id: number; equipmentCode: string }>>([]);
const chart = ref<ChartView | null>(null);
const policies = ref<Array<Record<string, any>>>([]);
const query = reactive({
  param: 'CD',
  operationId: null as number | null,
  equipmentId: null as number | null,
  from: '',
  to: ''
});
const policy = reactive({ onOOC: 'hold_lot', onOOS: 'hold_lot' });

const width = 720;
const height = 260;
const pad = 28;

const plot = computed(() => {
  const points = chart.value?.points || [];
  if (!points.length || !chart.value) {
    return { line: '', dots: [] as Array<{ x: number; y: number; bad: boolean; label: string }>, yUcl: 0, yCenter: 0, yLcl: 0 };
  }
  const values = points.map(point => point.value);
  const low = Math.min(chart.value.lcl, ...values);
  const high = Math.max(chart.value.ucl, ...values);
  const span = high - low || 1;
  const xOf = (index: number) => pad + (index * (width - pad * 2)) / Math.max(points.length - 1, 1);
  const yOf = (value: number) => pad + ((high - value) * (height - pad * 2)) / span;
  const line = points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${xOf(index)} ${yOf(point.value)}`).join(' ');
  const dots = points.map((point, index) => ({
    x: xOf(index),
    y: yOf(point.value),
    bad: Boolean(point.oos || (point.violations && point.violations.length)),
    label: (point.violations || []).join(',')
  }));
  return {
    line,
    dots,
    yUcl: yOf(chart.value.ucl),
    yCenter: yOf(chart.value.center),
    yLcl: yOf(chart.value.lcl)
  };
});

async function load() {
  const params: Record<string, unknown> = { param: query.param };
  if (query.operationId) params.operationId = query.operationId;
  if (query.equipmentId) params.equipmentId = query.equipmentId;
  if (query.from) params.from = query.from;
  if (query.to) params.to = query.to;
  const { data, error } = await fetchChart(params);
  if (error || !data) return;
  chart.value = data;
  const matched = policies.value.find(item => item.paramCode === query.param);
  if (matched) {
    policy.onOOC = matched.onOOC;
    policy.onOOS = matched.onOOS;
  }
}

async function saveReaction() {
  const { error } = await savePolicy({
    paramCode: query.param,
    operationID: query.operationId || 0,
    onOOC: policy.onOOC,
    onOOS: policy.onOOS
  });
  if (error) return;
  window.$message?.success($t('page.mes.qc.saved'));
  const { data } = await fetchPolicies();
  policies.value = data?.policies || [];
}

onMounted(async () => {
  const feature = await fetchFeatures();
  spcOn.value = Boolean(feature.data?.spc);
  if (!spcOn.value) return;
  const [planRes, eqpRes, policyRes] = await Promise.all([fetchPlans(), fetchEquipment(), fetchPolicies()]);
  plans.value = planRes.data?.plans || [];
  tools.value = eqpRes.data?.equipment || [];
  policies.value = policyRes.data?.policies || [];
  if (plans.value[0]) {
    query.operationId = plans.value[0].operationID;
    query.param = plans.value[0].items?.[0]?.paramCode || 'CD';
  }
  load();
});

const reactions = ['none', 'hold_lot', 'eqp_engineering', 'eqp_down', 'hold_and_engineering', 'hold_and_down'];
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.spc')">
    <NEmpty v-if="!spcOn" :description="$t('page.mes.qc.spcOff')" />
    <template v-else>
      <NSpace class="mb-12px">
        <NSelect
          v-model:value="query.param"
          class="w-140px"
          :options="[{ label: 'CD', value: 'CD' }]"
        />
        <NSelect
          v-model:value="query.operationId"
          class="w-180px"
          clearable
          :placeholder="$t('page.mes.qc.operation')"
          :options="plans.map(item => ({ label: item.operationCode || item.planName, value: item.operationID }))"
        />
        <NSelect
          v-model:value="query.equipmentId"
          class="w-180px"
          clearable
          :placeholder="$t('page.mes.track.equipment')"
          :options="tools.map(item => ({ label: item.equipmentCode, value: item.id }))"
        />
        <NInput v-model:value="query.from" class="w-140px" placeholder="YYYY-MM-DD" />
        <NInput v-model:value="query.to" class="w-140px" placeholder="YYYY-MM-DD" />
        <NButton type="primary" @click="load">{{ $t('page.mes.track.load') }}</NButton>
      </NSpace>
      <div v-if="chart" class="mb-8px">
        {{ chart.chartType }} · LCL {{ chart.lcl.toFixed(2) }} · CL {{ chart.center.toFixed(2) }} · UCL {{ chart.ucl.toFixed(2) }}
      </div>
      <svg v-if="chart && chart.points.length" :viewBox="`0 0 ${width} ${height}`" class="w-full max-w-900px border border-#efeff5">
        <line :x1="pad" :x2="width - pad" :y1="plot.yUcl" :y2="plot.yUcl" stroke="#d03050" stroke-dasharray="4" />
        <line :x1="pad" :x2="width - pad" :y1="plot.yCenter" :y2="plot.yCenter" stroke="#18a058" />
        <line :x1="pad" :x2="width - pad" :y1="plot.yLcl" :y2="plot.yLcl" stroke="#d03050" stroke-dasharray="4" />
        <path :d="plot.line" fill="none" stroke="#2080f0" stroke-width="2" />
        <circle
          v-for="(dot, index) in plot.dots"
          :key="index"
          :cx="dot.x"
          :cy="dot.y"
          r="5"
          :fill="dot.bad ? '#d03050' : '#2080f0'"
        >
          <title>{{ dot.label }}</title>
        </circle>
      </svg>
      <NEmpty v-else-if="chart" :description="$t('common.noData')" />
      <NDivider>{{ $t('page.mes.qc.reaction') }}</NDivider>
      <NSpace>
        <NSelect v-model:value="policy.onOOC" class="w-220px" :options="reactions.map(item => ({ label: `OOC ${item}`, value: item }))" />
        <NSelect v-model:value="policy.onOOS" class="w-220px" :options="reactions.map(item => ({ label: `OOS ${item}`, value: item }))" />
        <NButton @click="saveReaction">{{ $t('common.confirm') }}</NButton>
      </NSpace>
      <NDataTable
        class="mt-16px"
        :columns="[
          { title: $t('page.mes.qc.kind'), key: 'kind' },
          { title: $t('page.mes.qc.rule'), key: 'rule' },
          { title: $t('page.mes.qc.value'), key: 'value' },
          { title: $t('page.mes.qc.reaction'), key: 'reaction' }
        ]"
        :data="chart?.events || []"
      />
    </template>
  </NCard>
</template>
