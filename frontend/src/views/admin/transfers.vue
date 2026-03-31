<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">转账管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理所有转账记录</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-4 gap-4">
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">总转账笔数</div>
        <div class="text-2xl font-bold text-gray-900 mt-1">{{ total }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">待处理</div>
        <div class="text-2xl font-bold text-yellow-600 mt-1">{{ statusCount(0) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">成功</div>
        <div class="text-2xl font-bold text-green-600 mt-1">{{ statusCount(1) }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
        <div class="text-sm text-gray-500">失败</div>
        <div class="text-2xl font-bold text-red-600 mt-1">{{ statusCount(2) }}</div>
      </div>
    </div>

    <!-- 搜索筛选 -->
    <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
      <div class="flex items-center gap-4 flex-wrap">
        <div class="flex-1 min-w-[200px]">
          <input v-model="searchForm.search" type="text" placeholder="搜索交易号/账号/姓名..."
            class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
            @keyup.enter="page = 1; fetchTransfers()" />
        </div>
        <select v-model="searchForm.status"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="-1">全部状态</option>
          <option value="0">待处理</option>
          <option value="1">成功</option>
          <option value="2">失败</option>
        </select>
        <button
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
          @click="page = 1; fetchTransfers()">
          搜索
        </button>
      </div>
    </div>

    <!-- 转账列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">交易号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">类型</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">账号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">姓名</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">金额</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="t in transfers" :key="t.biz_no" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ t.id }}</td>
              <td class="px-4 py-3 text-gray-600">{{ t.user_name || t.uid }}</td>
              <td class="px-4 py-3 text-gray-500 font-mono text-xs">{{ t.biz_no }}</td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700">
                  {{ typeName(t.type) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-600">{{ t.account }}</td>
              <td class="px-4 py-3 text-gray-600">{{ t.username }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900">￥{{ t.money }}</td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium', statusClass(t.status)]">
                  {{ statusName(t.status) }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(t.paytime) }}</td>
              <td class="px-4 py-3 text-center">
                <div class="inline-flex items-center gap-1">
                  <button @click="queryStatus(t.biz_no)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">查询</button>
                  <button @click="showSetStatus(t)"
                    class="px-3 py-1 text-xs text-gray-600 hover:bg-gray-100 rounded transition-colors">改状态</button>
                  <button @click="handleRefund(t.biz_no)"
                    class="px-3 py-1 text-xs text-yellow-600 hover:bg-yellow-50 rounded transition-colors">退回</button>
                  <button @click="handleDelete(t.biz_no)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="transfers.length === 0">
              <td colspan="10" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>暂无转账记录</span>
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
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page <= 1" @click="page--; fetchTransfers()">上一页</button>
          <span class="px-3 py-1 text-sm">{{ page }} / {{ totalPages }}</span>
          <button
            class="px-3 py-1 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="page >= totalPages" @click="page++; fetchTransfers()">下一页</button>
        </div>
      </div>
    </div>

    <!-- 修改状态弹窗 -->
    <div v-if="showStatusModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showStatusModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">修改转账状态</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select v-model="newStatus"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="0">待处理</option>
                <option value="1">成功</option>
                <option value="2">失败</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">失败原因（可选）</label>
              <input v-model="newResult" type="text" placeholder="请输入失败原因"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showStatusModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="confirmSetStatus"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">确定</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getTransferList, transferOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const transfers = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const searchForm = ref({
  search: '',
  status: '-1'
})
const showStatusModal = ref(false)
const currentBizNo = ref('')
const newStatus = ref('1')
const newResult = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function statusCount(status: number) {
  return transfers.value.filter(t => t.status === status).length
}

async function fetchTransfers() {
  try {
    const res = await getTransferList({
      page: page.value,
      limit: pageSize.value,
      status: searchForm.value.status,
      search: searchForm.value.search
    })
    if (res.code === 0) {
      transfers.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取转账列表失败:', error)
  }
}

async function queryStatus(bizNo: string) {
  try {
    const res = await transferOp({ action: 'query', biz_no: bizNo })
    ElMessage.success(res.msg || '查询成功')
    fetchTransfers()
  } catch (error) {
    console.error('查询失败:', error)
  }
}

function showSetStatus(t: any) {
  currentBizNo.value = t.biz_no
  newStatus.value = String(t.status)
  newResult.value = ''
  showStatusModal.value = true
}

async function confirmSetStatus() {
  try {
    const res = await transferOp({
      action: 'set_status',
      biz_no: currentBizNo.value,
      status: parseInt(newStatus.value),
      result: newResult.value
    })
    ElMessage.success(res.msg || '状态已更新')
    showStatusModal.value = false
    fetchTransfers()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

async function handleRefund(bizNo: string) {
  if (!confirm('确定要退回这笔转账吗？')) return
  try {
    const res = await transferOp({ action: 'refund', biz_no: bizNo })
    ElMessage.success(res.msg || '退回成功')
    fetchTransfers()
  } catch (error) {
    console.error('退回失败:', error)
  }
}

async function handleDelete(bizNo: string) {
  if (!confirm('确定要删除这条记录吗？')) return
  try {
    const res = await transferOp({ action: 'delete', biz_no: bizNo })
    ElMessage.success(res.msg || '删除成功')
    fetchTransfers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

function typeName(type: string) {
  const map: Record<string, string> = {
    alipay: '支付宝',
    wxpay: '微信',
  }
  return map[type] || type
}

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待处理',
    1: '成功',
    2: '失败'
  }
  return map[status] || '未知'
}

function statusClass(status: number) {
  const map: Record<number, string> = {
    0: 'bg-yellow-100 text-yellow-700',
    1: 'bg-green-100 text-green-700',
    2: 'bg-red-100 text-red-700'
  }
  return map[status] || 'bg-gray-100 text-gray-700'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchTransfers()
})
</script>
