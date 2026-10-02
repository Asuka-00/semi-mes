<script setup lang="ts">
import { watch } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useEcharts } from '@/hooks/common/echarts';

defineOptions({ name: 'HomeTrendChart' });

const props = defineProps<{
  days: Array<{ day: string; moves: number; scrap: number }>;
  moveName: string;
  scrapName: string;
}>();

const appStore = useAppStore();

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: [] as string[], top: 0 },
  grid: { left: 12, right: 16, top: 36, bottom: 8, containLabel: true },
  xAxis: { type: 'category', data: [] as string[] },
  yAxis: { type: 'value', minInterval: 1 },
  series: [
    { name: '', type: 'line', smooth: true, data: [] as number[] },
    { name: '', type: 'bar', data: [] as number[], barMaxWidth: 18 }
  ]
}));

function paint() {
  updateOptions(opts => {
    opts.legend.data = [props.moveName, props.scrapName];
    opts.xAxis.data = props.days.map(day => day.day.slice(5));
    opts.series[0].name = props.moveName;
    opts.series[0].data = props.days.map(day => day.moves);
    opts.series[1].name = props.scrapName;
    opts.series[1].data = props.days.map(day => day.scrap);
    return opts;
  });
}

watch(() => [props.days, props.moveName, props.scrapName, appStore.locale], paint, { deep: true, immediate: true });
</script>

<template>
  <div ref="domRef" class="h-280px overflow-hidden"></div>
</template>
