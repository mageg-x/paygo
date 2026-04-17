<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">风控管理</h1>
        <p class="text-sm text-gray-500 mt-1">查看和处理风控记录</p>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
      <div class="flex items-center gap-4 flex-wrap">
        <div class="flex items-center gap-2">
          <input v-model="searchUid" type="number" placeholder="商户ID"
            class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            @keyup.enter="page = 1; fetchList()" />
        </div>
        <button @click="page = 1; fetchList()"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm">
          搜索
        </button>
      </div>
    </div>

    <!-- 风控列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">类型</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">内容</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="r in list" :key="r.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ r.id }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="font-medium">{{ r.user_name || r.uid }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  typeClass(r.type)]">
                  {{ typeName(r.type) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-600 max-w-xs truncate">{{ r.content }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="toggleStatus(r)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  r.status === 1 ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ r.status === 1 ? '已处理' : '待处理' }}
                </button>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(r.date) }}</td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                  <span>暂无风控记录</span>
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
import { riskList, riskOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchUid = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function typeName(type: number) {
  const map: Record<number, string> = {
    1: '订单成功率低',
    2: '通道连续失败',
    3: 'IP限制',
    4: '账号限制',
    5: '商品屏蔽'
  }
  return map[type] || '未知'
}

function typeClass(type: number) {
  const map: Record<number, string> = {
    1: 'bg-orange-100 text-orange-700',
    2: 'bg-red-100 text-red-700',
    3: 'bg-purple-100 text-purple-700',
    4: 'bg-blue-100 text-blue-700',
    5: 'bg-yellow-100 text-yellow-700'
  }
  return map[type] || 'bg-gray-100 text-gray-700'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const res = await riskList({ page: page.value, limit: pageSize.value, uid: searchUid.value })
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取风控列表失败:', error)
  }
}

async function toggleStatus(r: any) {
  try {
    const newStatus = r.status === 1 ? 0 : 1
    await riskOp({ action: 'set_status', id: r.id, status: newStatus })
    ElMessage.success(newStatus === 1 ? '已标记为已处理' : '已标记为待处理')
    r.status = newStatus
  } catch (error) {
    console.error('操作失败:', error)
  }
}

onMounted(() => {
  fetchList()
})
</script>
