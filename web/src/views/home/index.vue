<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import dayjs from 'dayjs';
import type { DataTableColumns } from 'naive-ui';
import { NButton } from 'naive-ui';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import { fetchDashboard } from '@/service/api/dashboard';
import type { DashBucket, DashEvent, DashHold, ShopDashboard } from '@/service/api/dashboard';
import BarChart from './modules/bar-chart.vue';
import StatusChart from './modules/status-chart.vue';
import TrendChart from './modules/trend-chart.vue';

const router = useRouter();
const appStore = useAppStore();
const { hasAuth } = useAuth();
const loading = ref(false);
const board = ref<ShopDashboard | null>(null);
let timer = 0;

const zh = computed(() => appStore.locale === 'zh-CN');
const gap = computed(() => (appStore.isMobile ? 0 : 16));

const kpis = computed(() => {
  const data = board.value;
  if (!data?.wip) return [];
  return [
    { key: 'wipLots', label: $t('page.home.dash.wipLots'), value: data.wipLots, tone: '#646cff' },
    { key: 'wipQty', label: $t('page.home.dash.wipQty'), value: data.wipQty, tone: '#18a058' },
    { key: 'holdLots', label: $t('page.home.dash.holdLots'), value: data.holdLots, tone: '#d03050' },
    { key: 'runningLots', label: $t('page.home.dash.runningLots'), value: data.runningLots, tone: '#f0a020' },
    { key: 'todayMoves', label: $t('page.home.dash.todayMoves'), value: data.todayMoves, tone: '#2080f0' },
    { key: 'todayCompleted', label: $t('page.home.dash.todayCompleted'), value: data.todayCompleted, tone: '#36ad6a' },
    { key: 'todayScrap', label: $t('page.home.dash.todayScrap'), value: data.todayScrap, tone: '#c97c10' }
  ];
});

const lotStatusKey: Record<string, string> = {
  waiting: 'page.mes.wip.waiting',
  running: 'page.mes.wip.running',
  hold: 'page.mes.wip.hold',
  completed: 'page.mes.wip.completed',
  scrapped: 'page.mes.wip.scrapped',
  merged: 'page.mes.wip.merged'
};

const eventKey: Record<string, string> = {
  start: 'page.mes.wip.eventStart',
  advance: 'page.mes.wip.eventAdvance',
  hold: 'page.mes.wip.eventHold',
  release: 'page.mes.wip.eventRelease',
  split: 'page.mes.wip.eventSplit',
  merge: 'page.mes.wip.eventMerge',
  complete: 'page.mes.wip.eventComplete',
  track_in: 'page.mes.wip.eventTrackIn',
  track_out: 'page.mes.wip.eventTrackOut',
  abort: 'page.mes.wip.eventAbort',
  pass: 'page.mes.wip.eventPass',
  scrap: 'page.mes.wip.eventScrap'
};

function statusRows(rows: DashBucket[]) {
  return rows.map(row => ({
    label: lotStatusKey[row.key] ? $t(lotStatusKey[row.key] as App.I18n.I18nKey) : row.label || row.key,
    count: row.count
  }));
}

function equipmentRows(rows: DashBucket[]) {
  return rows.map(row => ({
    label: $t(`page.mes.eqp.state_${row.key}` as App.I18n.I18nKey),
    count: row.count,
    qty: row.qty
  }));
}

function duration(seconds: number) {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return $t('page.home.dash.durationValue', { hours, minutes });
}

function eventLabel(item: DashEvent) {
  const key = eventKey[item.eventType];
  return key ? $t(key as App.I18n.I18nKey) : item.eventType;
}

function noticeTitle(item: ShopDashboard['notices'][number]) {
  return zh.value ? item.titleZh : item.titleEn;
}

function noticeBody(item: ShopDashboard['notices'][number]) {
  return zh.value ? item.bodyZh : item.bodyEn;
}

const updatedText = computed(() => {
  if (!board.value?.updatedAt) return '';
  return $t('page.home.dash.updated', { time: dayjs(board.value.updatedAt).format('YYYY-MM-DD HH:mm:ss') });
});

const holdColumns = computed<DataTableColumns<DashHold>>(() => [
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo', width: 140 },
  { title: $t('page.mes.wip.product'), key: 'productCode', width: 120 },
  {
    title: $t('page.mes.wip.currentNode'),
    key: 'nodeName',
    render: row => row.nodeName || row.nodeKey
  },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 72 },
  { title: $t('page.mes.wip.reason'), key: 'holdReason', ellipsis: { tooltip: true } },
  {
    title: $t('page.home.dash.duration'),
    key: 'holdSeconds',
    width: 130,
    render: row => duration(row.holdSeconds || 0)
  },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 88,
    render: row =>
      hasAuth('lot:lot:query')
        ? h(
            NButton,
            { size: 'small', text: true, type: 'primary', onClick: () => router.push({ name: 'lot_detail', params: { id: String(row.id) } }) },
            { default: () => $t('page.mes.wip.detail') }
          )
        : null
  }
]);

async function load() {
  loading.value = true;
  const { data, error } = await fetchDashboard();
  loading.value = false;
  if (error || !data) return;
  board.value = data;
}

onMounted(() => {
  load();
  timer = window.setInterval(load, 45000);
});

onUnmounted(() => {
  window.clearInterval(timer);
});
</script>

<template>
  <NSpace vertical :size="16">
    <NCard :bordered="false" size="small" class="card-wrapper">
      <div class="flex flex-wrap items-center justify-between gap-12px">
        <div>
          <div class="text-18px font-600">{{ $t('page.home.dash.title') }}</div>
          <div class="mt-4px text-13px text-#666">{{ $t('page.home.dash.subtitle') }}</div>
        </div>
        <div class="flex items-center gap-12px">
          <span class="text-13px text-#888">{{ updatedText }}</span>
          <NButton size="small" :loading="loading" @click="load">{{ $t('page.home.dash.refresh') }}</NButton>
        </div>
      </div>
    </NCard>

    <template v-if="board?.wip">
      <NGrid :x-gap="gap" :y-gap="16" cols="2 s:3 m:4 l:7" responsive="screen">
        <NGi v-for="item in kpis" :key="item.key">
          <NCard :bordered="false" size="small" class="card-wrapper">
            <div class="text-13px text-#666">{{ item.label }}</div>
            <div class="mt-8px text-28px font-600" :style="{ color: item.tone }">{{ item.value }}</div>
          </NCard>
        </NGi>
      </NGrid>

      <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
        <NGi span="24 s:24 m:14">
          <NCard :title="$t('page.home.dash.byStep')" :bordered="false" size="small" class="card-wrapper">
            <BarChart :rows="board.byStep" :series-name="$t('page.home.dash.lots')" />
          </NCard>
        </NGi>
        <NGi span="24 s:24 m:10">
          <NCard :title="$t('page.home.dash.byStatus')" :bordered="false" size="small" class="card-wrapper">
            <StatusChart :rows="statusRows(board.byStatus)" />
          </NCard>
        </NGi>
      </NGrid>

      <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
        <NGi span="24 s:24 m:10">
          <NCard :title="$t('page.home.dash.byProduct')" :bordered="false" size="small" class="card-wrapper">
            <BarChart :rows="board.byProduct" :series-name="$t('page.home.dash.lots')" horizontal />
          </NCard>
        </NGi>
        <NGi span="24 s:24 m:14">
          <NCard :title="$t('page.home.dash.trend')" :bordered="false" size="small" class="card-wrapper">
            <TrendChart :days="board.trend" :move-name="$t('page.home.dash.moves')" :scrap-name="$t('page.home.dash.scrap')" />
          </NCard>
        </NGi>
      </NGrid>

      <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
        <NGi span="24 s:24 m:14">
          <NCard :title="$t('page.home.dash.holds')" :bordered="false" size="small" class="card-wrapper">
            <template #header-extra>
              <NButton text type="primary" @click="router.push({ name: 'wip_overview' })">
                {{ $t('page.home.dash.openOverview') }}
              </NButton>
            </template>
            <NDataTable :columns="holdColumns" :data="board.holds" size="small" :pagination="false" />
          </NCard>
        </NGi>
        <NGi span="24 s:24 m:10">
          <NCard :title="$t('page.home.dash.events')" :bordered="false" size="small" class="card-wrapper">
            <NEmpty v-if="board.events.length === 0" :description="$t('page.home.dash.noEvents')" />
            <div v-for="item in board.events" :key="item.id" class="mb-8px border-b border-#efeff5 pb-8px">
              <div class="flex items-center justify-between gap-8px">
                <NButton
                  v-if="hasAuth('lot:lot:query')"
                  text
                  type="primary"
                  @click="router.push({ name: 'lot_detail', params: { id: String(item.lotId) } })"
                >
                  {{ item.lotNo }}
                </NButton>
                <span v-else>{{ item.lotNo }}</span>
                <span class="text-12px text-#888">{{ eventLabel(item) }}</span>
              </div>
              <div v-if="item.reason" class="mt-4px text-13px text-#666">{{ item.reason }}</div>
            </div>
          </NCard>
        </NGi>
      </NGrid>
    </template>
    <NCard v-else-if="board" :bordered="false" size="small" class="card-wrapper">
      <NEmpty :description="$t('page.home.dash.noWip')" />
    </NCard>

    <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
      <NGi span="24 s:24 m:12">
        <NCard :title="$t('page.home.dash.equipment')" :bordered="false" size="small" class="card-wrapper">
          <template v-if="board?.equipment" #header-extra>
            <NButton text type="primary" @click="router.push({ name: 'equipment_board' })">
              {{ $t('page.home.dash.openBoard') }}
            </NButton>
          </template>
          <NEmpty v-if="board && !board.equipment" :description="$t('page.home.dash.noEquipment')" />
          <template v-else-if="board?.equipment">
            <div class="mb-12px text-14px">
              {{ $t('page.home.dash.overduePm') }}
              <span class="ml-8px text-20px font-600 text-#d03050">{{ board.overduePm }}</span>
            </div>
            <BarChart :rows="equipmentRows(board.byEquipment)" :series-name="$t('page.home.dash.tools')" horizontal />
          </template>
        </NCard>
      </NGi>
      <NGi span="24 s:24 m:12">
        <NCard :title="$t('page.home.dash.notices')" :bordered="false" size="small" class="card-wrapper">
          <template #header-extra>
            <span class="text-13px text-#888">{{ $t('page.home.dash.unread', { count: board?.unread || 0 }) }}</span>
          </template>
          <NEmpty v-if="!board || board.notices.length === 0" :description="$t('page.home.dash.noNotices')" />
          <div v-for="item in board?.notices || []" :key="item.id" class="mb-8px border-b border-#efeff5 pb-8px">
            <div class="flex items-center justify-between gap-8px">
              <span :class="item.read ? 'text-#999' : 'font-600'">{{ noticeTitle(item) }}</span>
              <span v-if="!item.read" class="h-8px w-8px rounded-full bg-#d03050"></span>
            </div>
            <div class="mt-4px text-13px text-#666">{{ noticeBody(item) }}</div>
          </div>
        </NCard>
      </NGi>
    </NGrid>
  </NSpace>
</template>
