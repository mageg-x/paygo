<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">SSO单点登录</h1>
        <p class="text-sm text-gray-500 mt-1">管理员快速登录商户账号</p>
      </div>
    </div>

    <!-- 搜索商户 -->
    <div class="bg-white rounded-xl p-6 border border-gray-100 shadow-sm">
      <div class="max-w-md">
        <label class="block text-sm font-medium text-gray-700 mb-2">请输入要登录的商户ID</label>
        <div class="flex gap-3">
          <input v-model="uid" type="number" placeholder="输入商户ID"
            class="flex-1 px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            @keyup.enter="handleSSOLogin" />
          <button @click="handleSSOLogin"
            class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
            登录
          </button>
        </div>
        <p class="text-xs text-gray-500 mt-2">说明：管理员可通过此功能直接登录商户后台，无需知道商户密码</p>
      </div>
    </div>

    <!-- 最近登录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-100">
        <h3 class="font-semibold text-gray-700">最近登录的商户</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户ID</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="(item, index) in recentList" :key="index" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-500">{{ index + 1 }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                    <span class="text-blue-600 font-medium">{{ (item.username || 'U')[0].toUpperCase() }}</span>
                  </div>
                  <span class="font-medium">{{ item.username || '商户' + item.uid }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-500">{{ item.uid }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="quickLogin(item.uid)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">
                  登录
                </button>
              </td>
            </tr>
            <tr v-if="recentList.length === 0">
              <td colspan="4" class="px-4 py-8 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-10 h-10 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  <span>暂无最近登录记录</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ssoLogin } from '@/api/admin'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const uid = ref('')
const recentList = ref<any[]>([])

async function handleSSOLogin() {
  if (!uid.value) {
    ElMessage.warning('请输入商户ID')
    return
  }
  await quickLogin(parseInt(uid.value))
}

async function quickLogin(uidVal: number) {
  try {
    const res = await ssoLogin({ uid: uidVal })
    if (res.code === 0) {
      // 保存到localStorage
      localStorage.setItem('user_token', res.token)
      localStorage.setItem('user_uid', res.uid)
      // 跳转到商户后台
      window.open('/user/index', '_blank')
      // 更新最近登录
      updateRecentList(uidVal)
      ElMessage.success('登录成功')
    } else {
      ElMessage.error(res.msg || '登录失败')
    }
  } catch (error) {
    console.error('SSO登录失败:', error)
  }
}

function updateRecentList(uidVal: number) {
  const recent = JSON.parse(localStorage.getItem('sso_recent') || '[]')
  const filtered = recent.filter((item: any) => item.uid !== uidVal)
  filtered.unshift({ uid: uidVal, username: '商户' + uidVal })
  localStorage.setItem('sso_recent', JSON.stringify(filtered.slice(0, 10)))
  recentList.value = filtered.slice(0, 10)
}

onMounted(() => {
  const recent = JSON.parse(localStorage.getItem('sso_recent') || '[]')
  recentList.value = recent
})
</script>
