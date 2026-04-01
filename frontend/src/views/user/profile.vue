<template>
  <div>
    <h2 class="text-2xl font-bold text-gray-800 mb-6">资料管理</h2>

    <div class="flex gap-6">
      <div class="w-48">
        <div class="bg-white rounded-lg border p-2">
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'info' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'info'">
            基本信息
          </button>
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'cert' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'cert'">
            实名认证
          </button>
          <button
            :class="['w-full text-left px-4 py-2 rounded-lg text-sm', activeTab === 'api' ? 'bg-primary-50 text-primary-700' : 'text-gray-600']"
            @click="activeTab = 'api'">
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
                <input v-model="user.username" type="text" class="form-input px-3" />
              </div>
              <div class="mb-4">
                <label class="form-label">邮箱</label>
                <input v-model="user.email" type="email" class="form-input px-3" />
              </div>
              <div class="mb-4">
                <label class="form-label">手机</label>
                <input v-model="user.phone" type="tel" class="form-input px-3" />
              </div>
              <div class="mb-4">
                <label class="form-label">QQ</label>
                <input v-model="user.qq" type="text" class="form-input px-3" />
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
                  <input v-model="certForm.certname" type="text" class="form-input px-3" required />
                </div>
                <div class="mb-4">
                  <label class="form-label">身份证号</label>
                  <input v-model="certForm.certno" type="text" class="form-input px-3" required />
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
              <input :value="user.uid" type="text" class="form-input px-3" readonly />
            </div>
            <div class="mb-4">
              <label class="form-label">API密钥</label>
              <input :value="user.key" type="text" class="form-input px-3" readonly />
            </div>
            <p class="text-sm text-gray-500">如需重置密钥请联系管理员</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { updateProfile, submitCertificate, getUserInfo } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { ElMessage } from 'element-plus'

const appStore = useAppStore()
const activeTab = ref('info')

const user = reactive({
  uid: 0,
  username: '',
  email: '',
  phone: '',
  qq: '',
  key: '',
  cert: 0
})

const certForm = reactive({
  certname: '',
  certno: '',
  certtype: 0
})

const saving = ref(false)
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getUserInfo()
    if (res.code === 0 && res.data) {
      const u = res.data
      user.uid = u.uid
      user.username = u.username || ''
      user.email = u.email || ''
      user.phone = u.phone || ''
      user.qq = u.qq || ''
      user.key = u.key || ''
      user.cert = u.cert || 0
      // 更新全局状态
      if (u.uid) {
        appStore.userLogin(appStore.userToken, {
          uid: u.uid,
          username: u.username || '',
          email: u.email || '',
          phone: u.phone || '',
          money: u.money || 0,
          status: u.status || 1
        })
      }
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  } finally {
    loading.value = false
  }
})

async function handleSaveProfile() {
  saving.value = true
  try {
    const res = await updateProfile({ username: user.username, phone: user.phone, qq: user.qq })
    if (res.code === 0) {
      ElMessage.success('保存成功')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleCertSubmit() {
  try {
    const res = await submitCertificate(certForm)
    if (res.code === 0) {
      ElMessage.success('提交成功')
      user.cert = 1 // 待审核
    }
  } catch (error: any) {
    ElMessage.error(error.message || '提交失败')
  }
}
</script>
