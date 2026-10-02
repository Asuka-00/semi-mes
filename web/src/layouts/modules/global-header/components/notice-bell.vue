<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import { fetchNotices, markAllNoticesRead, markNoticeRead } from '@/service/api/notice';
import type { NoticeItem } from '@/service/api/notice';

defineOptions({
  name: 'NoticeBell'
});

const appStore = useAppStore();
const notices = ref<NoticeItem[]>([]);
const unread = ref(0);
let timer = 0;

const zh = computed(() => appStore.locale === 'zh-CN');

function titleOf(item: NoticeItem) {
  return zh.value ? item.titleZh : item.titleEn;
}

function bodyOf(item: NoticeItem) {
  return zh.value ? item.bodyZh : item.bodyEn;
}

async function load() {
  const { data, error } = await fetchNotices();
  if (error || !data) return;
  notices.value = data.notices || [];
  unread.value = data.unread || 0;
}

async function readOne(item: NoticeItem) {
  if (item.read) return;
  const { error } = await markNoticeRead(item.id);
  if (error) return;
  item.read = true;
  unread.value = Math.max(0, unread.value - 1);
}

async function readAll() {
  const { error } = await markAllNoticesRead();
  if (error) return;
  notices.value.forEach(item => {
    item.read = true;
  });
  unread.value = 0;
}

onMounted(() => {
  load();
  timer = window.setInterval(load, 30000);
});

onUnmounted(() => {
  window.clearInterval(timer);
});
</script>

<template>
  <NPopover trigger="click" placement="bottom-end" :width="360" @update:show="show => show && load()">
    <template #trigger>
      <NBadge :value="unread" :max="99" :show="unread > 0">
        <ButtonIcon icon="mdi:bell-outline" :tooltip-content="$t('notice.title')" />
      </NBadge>
    </template>
    <div class="max-h-420px overflow-auto">
      <div class="mb-8px flex-y-center justify-between">
        <span class="font-600">{{ $t('notice.title') }}</span>
        <NButton text type="primary" size="small" :disabled="unread === 0" @click="readAll">
          {{ $t('notice.markAll') }}
        </NButton>
      </div>
      <NEmpty v-if="notices.length === 0" :description="$t('notice.empty')" />
      <div
        v-for="item in notices"
        :key="item.id"
        class="mb-8px cursor-pointer border-b border-#efeff5 pb-8px"
        @click="readOne(item)"
      >
        <div class="flex-y-center justify-between gap-8px">
          <span :class="item.read ? 'text-#999' : 'font-600'">{{ titleOf(item) }}</span>
          <span v-if="!item.read" class="h-8px w-8px rounded-full bg-#d03050"></span>
        </div>
        <div class="mt-4px text-13px text-#666">{{ bodyOf(item) }}</div>
      </div>
    </div>
  </NPopover>
</template>
