<template>
  <div class="login-container">
    <n-card class="login-card" :title="$t('auth.welcomeBack')">
      <n-form ref="formRef" :model="formData" :rules="rules">
        <n-form-item path="username" :label="$t('auth.username')">
          <n-input
            v-model:value="formData.username"
            :placeholder="$t('auth.username')"
            @keydown.enter="handleLogin"
          />
        </n-form-item>
        <n-form-item path="password" :label="$t('auth.password')">
          <n-input
            v-model:value="formData.password"
            type="password"
            show-password-on="click"
            :placeholder="$t('auth.password')"
            @keydown.enter="handleLogin"
          />
        </n-form-item>
        <n-form-item>
          <n-button
            type="primary"
            block
            :loading="loading"
            @click="handleLogin"
          >
            {{ $t('auth.login') }}
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const formRef = ref()
const loading = ref(false)

const formData = reactive({
  username: 'admin',
  password: 'admin123'
})

const rules = {
  username: [
    { required: true, message: 'Please input username', trigger: 'blur' }
  ],
  password: [
    { required: true, message: 'Please input password', trigger: 'blur' }
  ]
}

async function handleLogin() {
  try {
    await formRef.value?.validate()
    loading.value = true
    await userStore.login(formData)
    message.success('Login success')
    router.push('/')
  } catch (error: any) {
    message.error(error.message || 'Login failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  max-width: 90%;
}
</style>
