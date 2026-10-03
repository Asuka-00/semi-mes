<script setup lang="ts">
defineOptions({ name: 'TraceTree' });
defineProps<{ nodes: Array<Record<string, any>> }>();
</script>

<template>
  <ul v-if="nodes?.length" class="m-0 list-disc pl-18px">
    <li v-for="node in nodes" :key="`${node.linkType}-${node.lotId}`" class="py-2px">
      <span class="font-600">{{ node.lotNo }}</span>
      <span class="ml-8px text-13px text-#888">{{ node.linkType }} · {{ node.status }} · {{ node.quantity }}</span>
      <TraceTree :nodes="node.nodes || []" />
    </li>
  </ul>
</template>
