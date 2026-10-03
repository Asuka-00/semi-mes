<script setup lang="ts">
import { computed, ref } from 'vue';
import { useAuth } from '@/hooks/business/auth';
import { useReasonOptions } from '@/hooks/business/dict';
import { $t } from '@/locales';
import { abortTrack, fetchStation, passNode, trackIn, trackOut } from '@/service/api/track';
import type { StationView } from '@/service/api/track';

const { hasAuth } = useAuth();
const canTrack = computed(() => hasAuth('wip:move:track'));
const lotNo = ref('');
const station = ref<StationView | null>(null);
const equipmentId = ref<number | null>(null);
const qtyOut = ref(0);
const qtyScrap = ref(0);
const scrapReason = ref('PARTICLE');
const inspection = ref('pass');
const readings = ref<Record<string, number[]>>({});
const defectCode = ref('');
const abortReason = ref('');
const carrierNo = ref('');
const resultText = ref('');

const { options: scrapOptions } = useReasonOptions('scrap', () => [
  { label: $t('page.mes.track.reasonParticle'), value: 'PARTICLE' },
  { label: $t('page.mes.track.reasonBroken'), value: 'BROKEN' },
  { label: $t('page.mes.track.reasonScratch'), value: 'SCRATCH' },
  { label: $t('page.mes.track.reasonOther'), value: 'OTHER' }
]);

const equipmentOptions = computed(() =>
  (station.value?.equipment || []).map(item => ({
    label: `${item.equipmentCode} ${item.equipmentName}`,
    value: item.id
  }))
);

async function load() {
  resultText.value = '';
  if (!lotNo.value.trim()) return;
  const { data, error } = await fetchStation(lotNo.value.trim());
  if (error || !data) {
    station.value = null;
    return;
  }
  station.value = data;
  equipmentId.value = data.equipment[0]?.id ?? null;
  carrierNo.value = data.carrierNo || '';
  readings.value = {};
  (data.inspectPlan?.items || []).forEach(item => {
    readings.value[item.paramCode] = Array.from({ length: item.sampleSize || 1 }, () => item.target ?? 0);
  });
  if (data.latestResult) inspection.value = '';
  const qty = Number(data.lot?.quantity || 0);
  qtyOut.value = qty;
  qtyScrap.value = 0;
}

async function doTrackIn() {
  if (!station.value || !equipmentId.value) return;
  const { error } = await trackIn(Number(station.value.lot.id), equipmentId.value, carrierNo.value);
  if (error) return;
  window.$message?.success($t('page.mes.track.trackIn'));
  lotNo.value = station.value.lot.lotNo;
  load();
}

async function doTrackOut() {
  if (!station.value) return;
  const { data, error } = await trackOut({
    lotId: Number(station.value.lot.id),
    qtyOut: qtyOut.value,
    qtyScrap: qtyScrap.value,
    scrapReasonCode: qtyScrap.value > 0 ? scrapReason.value : '',
    inspectionResult: station.value.inspectPlan ? '' : station.value.inspectionRequired ? inspection.value : '',
    measurements: station.value.inspectPlan
      ? Object.entries(readings.value).map(([paramCode, values]) => ({ paramCode, values }))
      : [],
    defectCode: defectCode.value
  });
  if (error || !data) return;
  resultText.value = `${data.result.action} ${data.result.nextNodeKey || ''} ${data.result.reason}`.trim();
  load();
}

async function doAbort() {
  if (!station.value) return;
  const { error } = await abortTrack(Number(station.value.lot.id), abortReason.value);
  if (error) return;
  load();
}

async function doPass() {
  if (!station.value) return;
  const { data, error } = await passNode({
    lotId: Number(station.value.lot.id),
    inspectionResult: station.value.inspectionRequired ? inspection.value : '',
    defectCode: defectCode.value
  });
  if (error || !data) return;
  resultText.value = `${data.result.action} ${data.result.nextNodeKey || ''} ${data.result.reason}`.trim();
  load();
}
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" :title="$t('page.mes.track.station')">
    <NSpace class="mb-16px">
      <NInput v-model:value="lotNo" :placeholder="$t('page.mes.track.lotNo')" class="w-260px" @keyup.enter="load" />
      <NButton type="primary" @click="load">{{ $t('page.mes.track.load') }}</NButton>
    </NSpace>
    <template v-if="station">
      <NSpace class="mb-12px">
        <NTag type="info">{{ station.lot.lotNo }}</NTag>
        <NTag>{{ station.lot.status }}</NTag>
        <span>{{ $t('page.mes.track.step') }}：{{ station.nodeName || station.lot.currentNodeKey }}</span>
        <span>{{ $t('page.mes.wip.quantity') }} {{ station.lot.quantity }}</span>
        <span v-if="station.equipmentGroup">{{ $t('page.mes.track.group') }} {{ station.equipmentGroup }}</span>
      </NSpace>
      <NAlert v-if="station.lot.status === 'hold'" type="warning" class="mb-12px">{{ $t('page.mes.track.holdHint') }}</NAlert>
      <div v-if="canTrack && station.lot.status === 'waiting' && station.nodeType === 'operation'" class="mb-12px">
        <NSpace>
          <NSelect v-model:value="equipmentId" :options="equipmentOptions" class="w-280px" :placeholder="$t('page.mes.track.equipment')" />
          <NInput v-model:value="carrierNo" class="w-180px" :placeholder="$t('page.mes.carrier.no')" />
          <NButton type="primary" :disabled="!equipmentId" @click="doTrackIn">{{ $t('page.mes.track.trackIn') }}</NButton>
        </NSpace>
        <div v-if="!equipmentOptions.length" class="mt-8px">{{ $t('page.mes.track.notAllowed') }}</div>
      </div>
      <div v-if="canTrack && station.lot.status === 'waiting' && station.nodeType !== 'operation' && station.nodeType !== 'end'">
        <NSpace>
          <NSelect
            v-if="station.inspectionRequired"
            v-model:value="inspection"
            class="w-180px"
            :options="[
              { label: $t('page.mes.qc.useJudgement'), value: '' },
              { label: $t('page.mes.routeGraph.pass'), value: 'pass' },
              { label: $t('page.mes.routeGraph.fail'), value: 'fail' }
            ]"
          />
          <span v-if="station.latestResult">{{ $t('page.mes.qc.latest') }} {{ station.latestResult }}</span>
          <NInput v-model:value="defectCode" class="w-160px" :placeholder="$t('page.mes.wip.defectCode')" />
          <NButton type="primary" @click="doPass">{{ $t('page.mes.track.pass') }}</NButton>
        </NSpace>
      </div>
      <div v-if="canTrack && station.lot.status === 'running'">
        <NAlert type="info" class="mb-12px">{{ $t('page.mes.track.runningHint') }}</NAlert>
        <NSpace class="mb-12px">
          <NInputNumber v-model:value="qtyOut" :min="0" class="w-140px" :placeholder="$t('page.mes.track.qtyOut')" />
          <NInputNumber v-model:value="qtyScrap" :min="0" class="w-140px" :placeholder="$t('page.mes.track.qtyScrap')" />
          <NSelect v-if="qtyScrap > 0" v-model:value="scrapReason" :options="scrapOptions" class="w-160px" />
          <template v-if="station.inspectPlan">
            <div v-for="item in station.inspectPlan.items" :key="item.paramCode" class="flex items-center gap-8px">
              <span>{{ item.paramCode }}</span>
              <NInputNumber
                v-for="(_, index) in readings[item.paramCode]"
                :key="index"
                v-model:value="readings[item.paramCode][index]"
                class="w-120px"
              />
            </div>
          </template>
          <NSelect
            v-else-if="station.inspectionRequired"
            v-model:value="inspection"
            class="w-140px"
            :options="[
              { label: $t('page.mes.routeGraph.pass'), value: 'pass' },
              { label: $t('page.mes.routeGraph.fail'), value: 'fail' }
            ]"
          />
          <NButton type="primary" @click="doTrackOut">{{ $t('page.mes.track.trackOut') }}</NButton>
        </NSpace>
        <NSpace>
          <NInput v-model:value="abortReason" class="w-220px" :placeholder="$t('page.mes.wip.reason')" />
          <NButton @click="doAbort">{{ $t('page.mes.track.abort') }}</NButton>
        </NSpace>
      </div>
      <div v-if="resultText" class="mt-12px">{{ $t('page.mes.wip.result') }} {{ resultText }}</div>
    </template>
  </NCard>
</template>
