<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useAuth } from '@/hooks/business/auth';
import { $t } from '@/locales';
import { recordMeasurement } from '@/service/api/quality';
import { fetchStation } from '@/service/api/track';
import type { InspectPlan } from '@/service/api/quality';
import type { StationView } from '@/service/api/track';

const { hasAuth } = useAuth();
const canAdd = computed(() => hasAuth('qc:measure:add'));
const lotNo = ref('DEMO-WIP-001');
const station = ref<StationView | null>(null);
const equipmentId = ref<number | null>(null);
const values = reactive<Record<string, number[]>>({});
const judged = ref('');

const plan = computed(() => station.value?.inspectPlan || null);

async function load() {
  judged.value = '';
  const { data, error } = await fetchStation(lotNo.value.trim());
  if (error || !data) {
    station.value = null;
    return;
  }
  station.value = data;
  equipmentId.value = data.equipment[0]?.id ?? data.openMove?.equipmentID ?? null;
  Object.keys(values).forEach(key => delete values[key]);
  (data.inspectPlan?.items || []).forEach(item => {
    values[item.paramCode] = Array.from({ length: item.sampleSize || 1 }, () => item.target ?? 0);
  });
}

async function submit() {
  if (!station.value || !plan.value) return;
  const samples = plan.value.items.map(item => ({
    paramCode: item.paramCode,
    values: values[item.paramCode] || []
  }));
  const { data, error } = await recordMeasurement({
    lotId: Number(station.value.lot.id),
    equipmentId: equipmentId.value || 0,
    operationId: plan.value.operationID,
    samples
  });
  if (error || !data) return;
  judged.value = data.result;
  window.$message?.success(data.result);
}
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.qc.measure')">
    <NSpace class="mb-16px">
      <NInput v-model:value="lotNo" class="w-220px" :placeholder="$t('page.mes.track.lotNo')" @keyup.enter="load" />
      <NButton type="primary" @click="load">{{ $t('page.mes.track.load') }}</NButton>
    </NSpace>
    <template v-if="station">
      <NSpace class="mb-12px">
        <NTag>{{ station.lot.lotNo }}</NTag>
        <NTag>{{ station.lot.status }}</NTag>
        <span>{{ station.nodeName }}</span>
        <span v-if="station.latestResult">{{ $t('page.mes.qc.latest') }} {{ station.latestResult }}</span>
      </NSpace>
      <NEmpty v-if="!plan" :description="$t('page.mes.qc.noPlan')" />
      <template v-else>
        <div class="mb-8px">{{ plan.planName }} / {{ plan.operationCode }}</div>
        <div v-for="item in plan.items" :key="item.paramCode" class="mb-12px">
          <div class="mb-4px">
            {{ item.paramName || item.paramCode }} ({{ item.unit }})
            {{ item.lsl ?? '-' }} ~ {{ item.usl ?? '-' }}
            <NTag v-if="item.required" size="small" class="ml-8px">{{ $t('page.mes.qc.required') }}</NTag>
          </div>
          <NSpace>
            <NInputNumber
              v-for="(_, index) in values[item.paramCode]"
              :key="index"
              v-model:value="values[item.paramCode][index]"
              class="w-140px"
            />
          </NSpace>
        </div>
        <NSpace>
          <NSelect
            v-model:value="equipmentId"
            class="w-240px"
            :options="(station.equipment || []).map(item => ({ label: item.equipmentCode, value: item.id }))"
            :placeholder="$t('page.mes.track.equipment')"
          />
          <NButton v-if="canAdd" type="primary" @click="submit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
        <div v-if="judged" class="mt-12px">{{ $t('page.mes.qc.judgement') }} {{ judged }}</div>
      </template>
    </template>
  </NCard>
</template>
