<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">资料管理</h2>

    <div class="flex gap-6">
      <div class="w-48">
        <div class="bg-white rounded-lg border p-2">
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'info' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'info'"
          >
            基本信息
          </button>
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'cert' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'cert'"
          >
            实名认证
          </button>
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'api' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'api'"
          >
            API信息
          </button>
        </div>
      </div>

      <div class="flex-1">
        <div v-if="activeTab === 'info'" class="card">
          <div class="card-body">
            <form @submit.prevent="handleSaveProfile" class="space-y-4">
              <div class="mb-4">
                <label class="form-label">用户名</label>
                <input v-model="user.username" type="text" class="form-input" />
              </div>
              <div class="mb-4">
                <label class="form-label">邮箱</label>
                <input v-model="user.email" type="email" class="form-input" />
              </div>
              <div class="mb-4">
                <label class="form-label">手机</label>
                <input v-model="user.phone" type="tel" class="form-input" />
              </div>
              <div class="mb-4">
                <label class="form-label">QQ</label>
                <input v-model="user.qq" type="text" class="form-input" />
              </div>
              <button type="submit" class="btn btn-primary">保存</button>
            </form>
          </div>
        </div>

        <div v-if="activeTab === 'cert'" class="card">
          <div class="card-body">
            <template v-if="user.cert === 0">
              <form @submit.prevent="handleCertSubmit" class="space-y-4">
                <div class="mb-4">
                  <label class="form-label">真实姓名</label>
                  <input v-model="certForm.certname" type="text" class="form-input" required />
                </div>
                <div class="mb-4">
                  <label class="form-label">身份证号</label>
                  <input v-model="certForm.certno" type="text" class="form-input" required />
                </div>
                <button type="submit" class="btn btn-primary">提交认证</button>
              </form>
            </template>
            <template v-else>
              <p class="text-success">已实名认证</p>
            </template>
          </div>
        </div>

        <div v-if="activeTab === 'api'" class="card">
          <div class="card-body">
            <div class="mb-4">
              <label class="form-label">商户ID</label>
              <input :value="user.uid" type="text" class="form-input" readonly />
            </div>
            <div class="mb-4">
              <label class="form-label">API密钥</label>
              <input :value="user.key" type="text" class="form-input" readonly />
            </div>
            <p class="text-sm text-gray-500">如需重置密钥请联系管理员</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { updateProfile, submitCertificate } from '@/api/user'

const activeTab = ref('info')

const user = reactive({
  uid: 1001,
  username: '',
  email: '',
  phone: '',
  qq: '',
  key: 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  cert: 0
})

const certForm = reactive({
  certname: '',
  certno: '',
  certtype: 0
})

const saving = ref(false)

async function handleSaveProfile() {
  saving.value = true
  try {
    const res = await updateProfile(user)
    if (res.code === 0) {
      alert('保存成功')
    }
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

async function handleCertSubmit() {
  try {
    const res = await submitCertificate(certForm)
    if (res.code === 0) {
      alert('提交成功')
    }
  } catch (error) {
    console.error('提交失败:', error)
  }
}
</script>
