<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">数据清理</h1>
        <p class="text-sm text-gray-500 mt-1">清理过期订单和临时数据</p>
      </div>
    </div>

    <!-- 清理配置 -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- 超时订单清理 -->
      <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center">
            <Clock class="w-5 h-5 text-amber-600" />
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">超时订单清理</h3>
            <p class="text-sm text-gray-500">清理长时间未支付的订单</p>
          </div>
        </div>
        <div class="space-y-3">
          <div>
            <label class="block text-sm text-gray-600 mb-1">超时时间</label>
            <select v-model="cleanForm.order_timeout"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="1">1小时</option>
              <option value="3">3小时</option>
              <option value="6">6小时</option>
              <option value="12">12小时</option>
              <option value="24">24小时</option>
              <option value="48">48小时</option>
            </select>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-500">预计清理数量</span>
            <span class="font-medium text-amber-600">{{ orderCount }} 条</span>
          </div>
          <button @click="cleanOrders"
            class="w-full py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-700 transition-colors text-sm font-medium">
            立即清理
          </button>
        </div>
      </div>

      <!-- 回调失败清理 -->
      <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center">
            <AlertCircle class="w-5 h-5 text-red-600" />
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">失败回调清理</h3>
            <p class="text-sm text-gray-500">清理回调一直失败的订单</p>
          </div>
        </div>
        <div class="space-y-3">
          <div>
            <label class="block text-sm text-gray-600 mb-1">最大重试次数</label>
            <select v-model="cleanForm.max_retry"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="3">3次</option>
              <option value="5">5次</option>
              <option value="10">10次</option>
              <option value="20">20次</option>
            </select>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-500">预计清理数量</span>
            <span class="font-medium text-red-600">{{ failedNotifyCount }} 条</span>
          </div>
          <button @click="cleanFailedNotifies"
            class="w-full py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors text-sm font-medium">
            立即清理
          </button>
        </div>
      </div>

      <!-- 日志清理 -->
      <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
            <FileText class="w-5 h-5 text-blue-600" />
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">操作日志清理</h3>
            <p class="text-sm text-gray-500">清理过期的操作日志</p>
          </div>
        </div>
        <div class="space-y-3">
          <div>
            <label class="block text-sm text-gray-600 mb-1">保留天数</label>
            <select v-model="cleanForm.log_days"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="7">7天</option>
              <option value="14">14天</option>
              <option value="30">30天</option>
              <option value="60">60天</option>
              <option value="90">90天</option>
            </select>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-500">预计清理数量</span>
            <span class="font-medium text-blue-600">{{ logCount }} 条</span>
          </div>
          <button @click="cleanLogs"
            class="w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
            立即清理
          </button>
        </div>
      </div>

      <!-- 缓存清理 -->
      <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
            <Database class="w-5 h-5 text-green-600" />
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">缓存数据清理</h3>
            <p class="text-sm text-gray-500">清理系统缓存和临时数据</p>
          </div>
        </div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-500">缓存大小</span>
            <span class="font-medium text-green-600">{{ cacheSize }}</span>
          </div>
          <button @click="cleanCache"
            class="w-full py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium">
            立即清理
          </button>
        </div>
      </div>
    </div>

    <!-- 清理记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-100">
        <h3 class="font-semibold text-gray-700">清理记录</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">类型</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">清理数量</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">操作人</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="record in records" :key="record.id" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 text-gray-900">{{ record.type }}</td>
              <td class="px-4 py-3 text-right text-gray-600">{{ record.count }} 条</td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ record.time }}</td>
              <td class="px-4 py-3 text-gray-500">{{ record.operator }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="4" class="px-4 py-8 text-center text-gray-400">
                暂无清理记录
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
import { getCleanStats, runClean } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Clock, AlertCircle, FileText, Database } from 'lucide-vue-next'

const cleanForm = ref({
  order_timeout: '24',
  max_retry: '10',
  log_days: '30'
})

const orderCount = ref(0)
const failedNotifyCount = ref(0)
const logCount = ref(0)
const cacheSize = ref('0 MB')

const records = ref<any[]>([])

async function fetchStats() {
  try {
    const res = await getCleanStats({
      order_timeout: cleanForm.value.order_timeout,
      max_retry: cleanForm.value.max_retry,
      log_days: cleanForm.value.log_days
    })
    if (res.code === 0 && res.data) {
      orderCount.value = res.data.order_count || 0
      failedNotifyCount.value = res.data.failed_notify_count || 0
      logCount.value = res.data.log_count || 0
      cacheSize.value = res.data.cache_size || '0 B'
    }
  } catch (error) {
    console.error('获取清理统计失败:', error)
  }
}

async function cleanOrders() {
  try {
    await ElMessageBox.confirm(
      `确定要清理超时的订单吗？预计清理 ${orderCount.value} 条`,
      '清理确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await runClean({
      action: 'orders',
      order_timeout: Number(cleanForm.value.order_timeout)
    })
    const count = res.data?.count ?? orderCount.value
    ElMessage.success(`清理完成，共处理 ${count} 条`)
    addRecord('超时订单清理', count)
    fetchStats()
  } catch {
    return
  }
}

async function cleanFailedNotifies() {
  try {
    await ElMessageBox.confirm(
      `确定要清理回调失败的订单吗？预计清理 ${failedNotifyCount.value} 条`,
      '清理确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await runClean({
      action: 'failed_notifies',
      max_retry: Number(cleanForm.value.max_retry)
    })
    const count = res.data?.count ?? failedNotifyCount.value
    ElMessage.success(`清理完成，共处理 ${count} 条`)
    addRecord('失败回调清理', count)
    fetchStats()
  } catch {
    return
  }
}

async function cleanLogs() {
  try {
    await ElMessageBox.confirm(
      `确定要清理过期的操作日志吗？预计清理 ${logCount.value} 条`,
      '清理确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await runClean({
      action: 'logs',
      log_days: Number(cleanForm.value.log_days)
    })
    const count = res.data?.count ?? logCount.value
    ElMessage.success(`清理完成，共处理 ${count} 条`)
    addRecord('操作日志清理', count)
    fetchStats()
  } catch {
    return
  }
}

async function cleanCache() {
  try {
    await ElMessageBox.confirm(
      '确定要清理系统缓存吗？',
      '清理确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await runClean({ action: 'cache' })
    const count = res.data?.count ?? 0
    ElMessage.success('缓存已清理')
    addRecord('缓存清理', count)
    await fetchStats()
  } catch {
    return
  }
}

function addRecord(typeName: string, count: number) {
  records.value.unshift({
    id: Date.now(),
    type: typeName,
    count: count,
    time: new Date().toLocaleString('zh-CN'),
    operator: 'admin'
  })
}

onMounted(() => {
  fetchStats()
})
</script>
