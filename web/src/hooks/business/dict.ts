import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { request } from '@/service/request';

export interface CodeOption {
  label: string;
  value: string;
  color?: string;
}

function pickLabel(row: Record<string, any>, english: boolean, zhKey: string, enKey: string, codeKey: string) {
  const zh = String(row[zhKey] || '');
  const en = String(row[enKey] || '');
  const code = String(row[codeKey] || '');
  if (english) return en || zh || code;
  return zh || en || code;
}

export function useDictOptions(typeCode: string, fallback: () => CodeOption[]) {
  const options = ref<CodeOption[]>([]);
  const { locale } = useI18n();

  async function reload() {
    const { data, error } = await request<{ items: Array<Record<string, any>> }>({
      url: '/dict/items',
      params: { typeCode }
    });
    const items = !error && data?.items?.length ? data.items : [];
    if (!items.length) {
      options.value = fallback();
      return;
    }
    const english = String(locale.value).toLowerCase().startsWith('en');
    options.value = items.map(item => ({
      value: String(item.itemCode),
      label: pickLabel(item, english, 'labelZh', 'labelEn', 'itemCode'),
      color: String(item.color || '')
    }));
  }

  watch(locale, () => {
    reload();
  });
  reload();

  function labelOf(value: string | number | null | undefined) {
    const text = String(value ?? '');
    return options.value.find(item => item.value === text)?.label || fallback().find(item => item.value === text)?.label || text;
  }

  return { options, labelOf, reload };
}

export function useReasonOptions(category: string, fallback: () => CodeOption[]) {
  const options = ref<CodeOption[]>([]);
  const { locale } = useI18n();

  async function reload() {
    const { data, error } = await request<{ reasons: Array<Record<string, any>> }>({
      url: '/dict/reasons',
      params: { category }
    });
    const rows = !error && data?.reasons?.length ? data.reasons : [];
    if (!rows.length) {
      options.value = fallback();
      return;
    }
    const english = String(locale.value).toLowerCase().startsWith('en');
    options.value = rows.map(item => ({
      value: String(item.reasonCode),
      label: `${item.reasonCode} ${pickLabel(item, english, 'nameZh', 'nameEn', 'reasonCode')}`.trim()
    }));
  }

  watch(locale, () => {
    reload();
  });
  reload();
  return { options, reload };
}

export function previewNumber(ruleCode: string) {
  return request<{ preview: string }>({ url: '/sysNumberRule/preview', method: 'post', data: { ruleCode } });
}
