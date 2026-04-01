<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">操作日志</h1>
        <p class="text-sm text-gray-500 mt-1">查看商户操作记录</p>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
      <div class="flex items-center gap-4">
        <input v-model="searchUid" type="number" placeholder="商户ID"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          @keyup.enter="page = 1; fetchList()" />
        <button @click="page = 1; fetchList()"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm">
          搜索
        </button>
      </div>
    </div>

    <!-- 日志列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">商户</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">操作</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">详情</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">IP</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="l in list" :key="l.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ l.id }}</td>
              <td class="px-4 py-3 text-center">
                <span class="font-medium">{{ l.user_name || l.uid }}</span>
              </td>
              <td class="px-4 py-3">
                <span :class="typeClass(l.type)">
                  {{ typeName(l.type) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-600 max-w-xs truncate">{{ l.content }}</td>
              <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ l.ip }}</td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(l.time) }}</td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无操作日志</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div class="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex items-center gap-2">
          <button @click="page--; fetchList()" :disabled="page <= 1"
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50">
            上一页
          </button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button @click="page++; fetchList()" :disabled="page >= totalPages"
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { logList } from '@/api/admin'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchUid = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(type: string) {
  const map: Record<string, string> = {
    'login': '登录',
    'logout': '登出',
    'order': '订单操作',
    'settle': '结算操作',
    'edit': '资料修改',
    'recharge': '余额操作',
    'other': '其他'
  }
  return map[type] || type
}

function typeClass(type: string) {
  const map: Record<string, string> = {
    'login': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700',
    'logout': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700',
    'order': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700',
    'settle': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700',
    'edit': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-orange-100 text-orange-700',
    'recharge': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-700',
    'other': 'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700'
  }
  return map[type] || map['other']
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (searchUid.value) {
      params.uid = searchUid.value
    }
    const res = await logList(params)
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取日志列表失败:', error)
  }
}

onMounted(() => {
  fetchList()
})
</script>
