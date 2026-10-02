<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { DataTableColumns } from 'naive-ui';
import { NButton } from 'naive-ui';
import { $t } from '@/locales';
import { fetchOverview } from '@/service/api/track';

const router = useRouter();
const byStatus = ref<Array<{ key: string; count: number; qty: number }>>([]);
const byNode = ref<Array<{ key: string; count: number; qty: number }>>([]);
const byProduct = ref<Array<{ key: string; count: number; qty: number }>>([]);
const holds = ref<Array<Record<string, any>>>([]);

function countColumns(title: string): DataTableColumns<{ key: string; count: number; qty: number }> {
  return [
    { title, key: 'key', ellipsis: { tooltip: true } },
    { title: $t('page.mes.track.count'), key: 'count', width: 72 },
    { title: $t('page.mes.track.qty'), key: 'qty', width: 72 }
  ];
}

const holdColumns: DataTableColumns<Record<string, any>> = [
  { title: $t('page.mes.wip.lotNo'), key: 'lotNo' },
  { title: $t('page.mes.wip.currentNode'), key: 'nodeName', render: row => row.nodeName || row.currentNodeKey },
  { title: $t('page.mes.wip.quantity'), key: 'quantity', width: 80 },
  { title: $t('page.mes.wip.reasonCode'), key: 'holdReasonCode' },
  { title: $t('page.mes.wip.reason'), key: 'holdReason' },
  {
    title: $t('common.action'),
    key: 'actions',
    width: 100,
    render: row =>
      h(NButton, { size: 'small', onClick: () => router.push({ name: 'lot_detail', params: { id: String(row.id) } }) }, { default: () => $t('page.mes.wip.detail') })
  }
];

onMounted(async () => {
  const { data, error } = await fetchOverview();
  if (error || !data) return;
  byStatus.value = data.byStatus || [];
  byNode.value = data.byNode || [];
  byProduct.value = data.byProduct || [];
  holds.value = data.holds || [];
});
</script>

<template>
  <div class="flex flex-col gap-12px">
    <NGrid cols="1 s:1 m:3" responsive="screen" :x-gap="12" :y-gap="12">
      <NGi>
        <NCard :title="$t('page.mes.track.byStatus')" size="small">
          <NDataTable :columns="countColumns($t('page.mes.wip.status'))" :data="byStatus" size="small" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :title="$t('page.mes.track.byNode')" size="small">
          <NDataTable :columns="countColumns($t('page.mes.track.step'))" :data="byNode" size="small" />
        </NCard>
      </NGi>
      <NGi>
        <NCard :title="$t('page.mes.track.byProduct')" size="small">
          <NDataTable :columns="countColumns($t('page.mes.wip.product'))" :data="byProduct" size="small" />
        </NCard>
      </NGi>
    </NGrid>
    <NCard :title="$t('page.mes.track.holdList')" size="small">
      <NDataTable :columns="holdColumns" :data="holds" size="small" />
    </NCard>
  </div>
</template>
