<script setup lang="ts">
import { watch } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useEcharts } from '@/hooks/common/echarts';

defineOptions({ name: 'HomeBarChart' });

interface Row {
  label: string;
  count: number;
  qty?: number;
}

const props = defineProps<{
  rows: Row[];
  seriesName: string;
  horizontal?: boolean;
}>();

const appStore = useAppStore();
const horizontal = props.horizontal === true;

const { domRef, updateOptions } = useEcharts(() =>
  horizontal
    ? {
        tooltip: { trigger: 'axis' as const },
        grid: { left: 12, right: 16, top: 16, bottom: 8, containLabel: true },
        xAxis: { type: 'value' as const, minInterval: 1 },
        yAxis: { type: 'category' as const, data: [] as string[], inverse: true },
        series: [{ name: '', type: 'bar' as const, data: [] as number[], barMaxWidth: 28 }]
      }
    : {
        tooltip: { trigger: 'axis' as const },
        grid: { left: 12, right: 16, top: 16, bottom: 8, containLabel: true },
        xAxis: { type: 'category' as const, data: [] as string[], axisLabel: { interval: 0, hideOverlap: true } },
        yAxis: { type: 'value' as const, minInterval: 1 },
        series: [{ name: '', type: 'bar' as const, data: [] as number[], barMaxWidth: 36 }]
      }
);

function paint() {
  updateOptions(opts => {
    const chart = opts as {
      xAxis: { data?: string[] };
      yAxis: { data?: string[] };
      series: Array<{ name: string; data: number[] }>;
    };
    const labels = props.rows.map(row => row.label || '-');
    const values = props.rows.map(row => row.count);
    chart.series[0].name = props.seriesName;
    chart.series[0].data = values;
    if (props.horizontal) {
      chart.yAxis.data = labels;
    } else {
      chart.xAxis.data = labels;
    }
    return opts;
  });
}

watch(() => [props.rows, props.seriesName, props.horizontal, appStore.locale], paint, { deep: true, immediate: true });
</script>

<template>
  <div ref="domRef" class="h-280px overflow-hidden"></div>
</template>
