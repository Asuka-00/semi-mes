<template>
  <div>
    <n-card :title="$t('baseData.factory.title')">
      <template #header-extra>
        <n-space>
          <n-button v-if="userStore.hasPermission('base:factory:add')" type="primary" @click="handleAdd">
            {{ $t('common.add') }}
          </n-button>
        </n-space>
      </template>

      <n-space vertical>
        <n-form inline :model="searchForm" label-placement="left">
          <n-form-item :label="$t('common.search')">
            <n-input v-model:value="searchForm.keyword" :placeholder="$t('baseData.factory.code') + ' / ' + $t('baseData.factory.name')" clearable />
          </n-form-item>
          <n-form-item>
            <n-button type="primary" @click="handleSearch">{{ $t('common.query') }}</n-button>
            <n-button @click="handleReset" style="margin-left: 12px;">{{ $t('common.reset') }}</n-button>
          </n-form-item>
        </n-form>

        <n-data-table
          :columns="columns"
          :data="tableData"
          :loading="loading"
          :pagination="pagination"
          @update:page="handlePageChange"
        />
      </n-space>
    </n-card>

    <n-modal v-model:show="showModal" :title="modalTitle">
      <n-card style="width: 600px;" :title="modalTitle" :bordered="false" size="huge">
        <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="auto">
          <n-form-item :label="$t('baseData.factory.code')" path="factory_code">
            <n-input v-model:value="formData.factory_code" :disabled="!!formData.id" />
          </n-form-item>
          <n-form-item :label="$t('baseData.factory.name')" path="factory_name">
            <n-input v-model:value="formData.factory_name" />
          </n-form-item>
          <n-form-item :label="$t('baseData.factory.address')">
            <n-input v-model:value="formData.address" />
          </n-form-item>
          <n-form-item :label="$t('baseData.factory.contact')">
            <n-input v-model:value="formData.contact" />
          </n-form-item>
          <n-form-item :label="$t('baseData.factory.phone')">
            <n-input v-model:value="formData.phone" />
          </n-form-item>
          <n-form-item :label="$t('common.status')">
            <n-switch v-model:value="formData.status" :checked-value="1" :unchecked-value="0" />
          </n-form-item>
        </n-form>
        <template #footer>
          <n-space justify="end">
            <n-button @click="showModal = false">{{ $t('common.cancel') }}</n-button>
            <n-button type="primary" @click="handleSubmit">{{ $t('common.submit') }}</n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h, onMounted } from 'vue'
import { NButton, NSpace, NTag, useMessage, useDialog } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { getFactories, createFactory, updateFactory, deleteFactory, type Factory } from '@/api/basedata'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const loading = ref(false)
const showModal = ref(false)
const formRef = ref()
const tableData = ref<Factory[]>([])
const searchForm = reactive({
  keyword: ''
})
const formData = reactive<Factory>({
  factory_code: '',
  factory_name: '',
  address: '',
  contact: '',
  phone: '',
  status: 1
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100]
})

const modalTitle = computed(() => {
  return formData.id ? t('common.edit') : t('common.add')
})

const rules = {
  factory_code: [
    { required: true, message: 'Please input factory code', trigger: 'blur' }
  ],
  factory_name: [
    { required: true, message: 'Please input factory name', trigger: 'blur' }
  ]
}

const columns = computed<DataTableColumns<Factory>>(() => [
  {
    title: t('baseData.factory.code'),
    key: 'factory_code'
  },
  {
    title: t('baseData.factory.name'),
    key: 'factory_name'
  },
  {
    title: t('baseData.factory.address'),
    key: 'address'
  },
  {
    title: t('baseData.factory.contact'),
    key: 'contact'
  },
  {
    title: t('baseData.factory.phone'),
    key: 'phone'
  },
  {
    title: t('common.status'),
    key: 'status',
    render: (row) => {
      return h(NTag, { type: row.status === 1 ? 'success' : 'default' }, 
        { default: () => row.status === 1 ? t('common.enable') : t('common.disable') }
      )
    }
  },
  {
    title: t('common.action'),
    key: 'action',
    render: (row) => {
      return h(NSpace, {}, {
        default: () => [
          userStore.hasPermission('base:factory:edit') && h(
            NButton,
            { size: 'small', onClick: () => handleEdit(row) },
            { default: () => t('common.edit') }
          ),
          userStore.hasPermission('base:factory:delete') && h(
            NButton,
            { size: 'small', type: 'error', onClick: () => handleDelete(row) },
            { default: () => t('common.delete') }
          )
        ]
      })
    }
  }
])

async function loadData() {
  loading.value = true
  try {
    const result = await getFactories({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword
    })
    tableData.value = result.list
    pagination.itemCount = result.total
  } catch (error: any) {
    message.error(error.message || 'Load data failed')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  searchForm.keyword = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadData()
}

function handleAdd() {
  Object.assign(formData, {
    id: undefined,
    factory_code: '',
    factory_name: '',
    address: '',
    contact: '',
    phone: '',
    status: 1
  })
  showModal.value = true
}

function handleEdit(row: Factory) {
  Object.assign(formData, { ...row })
  showModal.value = true
}

async function handleDelete(row: Factory) {
  dialog.warning({
    title: t('common.confirm'),
    content: t('common.confirmDelete'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        await deleteFactory(row.id!)
        message.success(t('common.success'))
        loadData()
      } catch (error: any) {
        message.error(error.message || 'Delete failed')
      }
    }
  })
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    if (formData.id) {
      await updateFactory(formData.id, formData)
    } else {
      await createFactory(formData)
    }
    message.success(t('common.success'))
    showModal.value = false
    loadData()
  } catch (error: any) {
    if (error?.message) {
      message.error(error.message)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>
