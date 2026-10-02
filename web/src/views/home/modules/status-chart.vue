<script setup lang="ts">
import { watch } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useEcharts } from '@/hooks/common/echarts';

defineOptions({ name: 'HomeStatusChart' });

const props = defineProps<{
  rows: Array<{ label: string; count: number }>;
}>();

const appStore = useAppStore();

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0, type: 'scroll' },
  series: [
    {
      type: 'pie',
      radius: ['42%', '68%'],
      center: ['50%', '44%'],
      data: [] as Array<{ name: string; value: number }>,
      label: { formatter: '{b}\n{c}' }
    }
  ]
}));

function paint() {
  updateOptions(opts => {
    opts.series[0].data = props.rows.map(row => ({ name: row.label, value: row.count }));
    return opts;
  });
}

watch(() => [props.rows, appStore.locale], paint, { deep: true, immediate: true });
</script>

<template>
  <div ref="domRef" class="h-280px overflow-hidden"></div>
</template>
