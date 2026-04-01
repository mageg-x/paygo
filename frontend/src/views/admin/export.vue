<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">数据导出</h1>
        <p class="text-sm text-gray-500 mt-1">导出订单数据为Excel文件</p>
      </div>
    </div>

    <!-- 导出配置 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">导出配置</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">日期范围</label>
          <div class="flex items-center gap-2">
            <input v-model="form.start_date" type="date"
              class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            <span class="text-gray-400">至</span>
            <input v-model="form.end_date" type="date"
              class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">商户筛选</label>
          <input v-model="form.uid" type="number" placeholder="输入商户ID，不填则导出全部"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">订单状态</label>
          <select v-model="form.status"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">全部状态</option>
            <option value="0">待支付</option>
            <option value="1">已支付</option>
            <option value="2">已关闭</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">支付方式</label>
          <select v-model="form.type"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">全部方式</option>
            <option value="1">支付宝</option>
            <option value="2">微信支付</option>
            <option value="3">QQ钱包</option>
            <option value="4">银行卡</option>
          </select>
        </div>
      </div>

      <div class="mt-6 pt-4 border-t flex items-center gap-4">
        <button @click="handleExport"
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
          <Download class="w-4 h-4" />
          导出Excel
        </button>
        <span class="text-sm text-gray-500">导出文件格式为 .xlsx，每次最多导出10万条记录</span>
      </div>
    </div>

    <!-- 导出记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-100">
        <h3 class="font-semibold text-gray-700">导出记录</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">文件名</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">记录数</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="record in records" :key="record.id" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 text-gray-900">{{ record.filename }}</td>
              <td class="px-4 py-3 text-gray-600">{{ record.count }} 条</td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ record.time }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="downloadFile(record)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">
                  下载
                </button>
              </td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="4" class="px-4 py-8 text-center text-gray-400">
                暂无导出记录
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getOrderList } from '@/api/admin'
import { ElMessage } from 'element-plus'
import { Download } from 'lucide-vue-next'

const form = ref({
  start_date: '',
  end_date: '',
  uid: '',
  status: '',
  type: ''
})

const records = ref<any[]>([])

async function handleExport() {
  if (!form.value.start_date || !form.value.end_date) {
    ElMessage.warning('请选择日期范围')
    return
  }

  try {
    // TODO: 调用后端导出API
    // 后端应返回文件流或下载链接
    ElMessage.info('正在生成导出文件，请稍候...')

    // 模拟：获取数据并生成CSV
    const res = await getOrderList({ page: 1, limit: 1000 })
    if (res.code === 0) {
      const data = res.data || []
      const csv = generateCSV(data)
      downloadCSV(csv, `订单导出_${form.value.start_date}_${form.value.end_date}.csv`)
      ElMessage.success('导出成功')

      // 添加到记录
      records.value.unshift({
        id: Date.now(),
        filename: `订单导出_${form.value.start_date}_${form.value.end_date}.csv`,
        count: data.length,
        time: new Date().toLocaleString('zh-CN')
      })
    }
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

function generateCSV(data: any[]) {
  const headers = ['订单号', '商户订单号', '商户ID', '商品名称', '支付方式', '金额', '实付金额', '商户所得', '状态', '创建时间', '支付时间', '回调状态', 'IP']
  const rows = data.map((o: any) => [
    o.trade_no,
    o.out_trade_no || '',
    o.uid,
    o.name,
    typeName(o.type),
    o.money,
    o.realmoney || o.money,
    o.getmoney || '0',
    statusName(o.status),
    o.addtime,
    o.endtime || '',
    o.notify === 1 ? '已回调' : '未回调',
    o.ip || ''
  ])

  return [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
}

function downloadCSV(content: string, filename: string) {
  const BOM = '\uFEFF'
  const blob = new Blob([BOM + content], { type: 'text/csv;charset=utf-8' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
  URL.revokeObjectURL(link.href)
}

function downloadFile(record: any) {
  ElMessage.info('下载功能开发中')
}

function typeName(type: number) {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付',
    3: 'QQ钱包',
    4: '银行卡'
  }
  return map[type] || '其他'
}

function statusName(status: number) {
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '已关闭'
  }
  return map[status] || '未知'
}
</script>
