<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">批量转账</h1>
        <p class="text-sm text-gray-500 mt-1">上传文件批量执行转账操作</p>
      </div>
    </div>

    <!-- 文件上传 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">上传转账文件</h3>
      <div class="border-2 border-dashed border-gray-200 rounded-lg p-8 text-center hover:border-blue-400 transition-colors">
        <input type="file" ref="fileInput" @change="handleFileSelect" accept=".csv,.xlsx,.xls" class="hidden" />
        <button @click="triggerFileInput"
          class="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2 mx-auto">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          选择文件
        </button>
        <p class="text-gray-500 text-sm mt-3">支持 CSV、Excel 格式</p>
      </div>

      <!-- 文件预览 -->
      <div v-if="fileData.length > 0" class="mt-6">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium text-gray-800">文件预览 ({{ fileName }})</h4>
          <span class="text-sm text-gray-500">共 {{ fileData.length }} 条记录</span>
        </div>
        <div class="overflow-x-auto border border-gray-200 rounded-lg">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50 border-b border-gray-200">
                <th class="px-4 py-2 text-left font-semibold text-gray-600">序号</th>
                <th class="px-4 py-2 text-left font-semibold text-gray-600">商户UID</th>
                <th class="px-4 py-2 text-left font-semibold text-gray-600">真实姓名</th>
                <th class="px-4 py-2 text-left font-semibold text-gray-600">账号</th>
                <th class="px-4 py-2 text-right font-semibold text-gray-600">金额</th>
                <th class="px-4 py-2 text-left font-semibold text-gray-600">备注</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <tr v-for="(row, idx) in fileData.slice(0, 10)" :key="idx" class="hover:bg-gray-50">
                <td class="px-4 py-2 text-gray-500">{{ idx + 1 }}</td>
                <td class="px-4 py-2 text-gray-900">{{ row.uid }}</td>
                <td class="px-4 py-2 text-gray-900">{{ row.name }}</td>
                <td class="px-4 py-2 text-gray-900 font-mono text-xs">{{ row.account }}</td>
                <td class="px-4 py-2 text-right font-semibold text-green-600">￥{{ row.amount }}</td>
                <td class="px-4 py-2 text-gray-500">{{ row.remark || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-if="fileData.length > 10" class="text-sm text-gray-500 mt-2">... 还有 {{ fileData.length - 10 }} 条记录未显示</p>

        <div class="flex items-center gap-4 mt-6">
          <button @click="executeBatchTransfer"
            :disabled="executing"
            class="px-6 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">
            {{ executing ? '执行中...' : '确认执行批量转账' }}
          </button>
          <button @click="clearFile"
            class="px-6 py-2.5 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors text-sm font-medium">
            清除文件
          </button>
        </div>
      </div>
    </div>

    <!-- 转账记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100">
        <h3 class="text-lg font-semibold text-gray-900">批量转账记录</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">批次号</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">文件名</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">总数量</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">成功</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">失败</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">总金额</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="batch in batchList" :key="batch.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-mono text-xs">{{ batch.batch_no }}</td>
              <td class="px-4 py-3 text-gray-600">{{ batch.filename }}</td>
              <td class="px-4 py-3 text-center text-gray-600">{{ batch.total }}</td>
              <td class="px-4 py-3 text-center text-green-600 font-medium">{{ batch.success }}</td>
              <td class="px-4 py-3 text-center text-red-600 font-medium">{{ batch.failed }}</td>
              <td class="px-4 py-3 text-right font-semibold text-green-600">￥{{ batch.amount }}</td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  batch.status === 1 ? 'bg-green-100 text-green-700' :
                  batch.status === 2 ? 'bg-yellow-100 text-yellow-700' :
                  'bg-gray-100 text-gray-600']">
                  {{ batch.status === 1 ? '已完成' : batch.status === 2 ? '处理中' : '待处理' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center text-gray-500 text-xs">{{ batch.created_at }}</td>
            </tr>
            <tr v-if="batchList.length === 0">
              <td colspan="8" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <span>暂无批量转账记录</span>
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
import { getTransferBatchList, transferBatchCreate, transferBatchOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = ref('')
const fileData = ref<any[]>([])
const executing = ref(false)
const batchList = ref<any[]>([])

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  fileName.value = file.name
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const text = e.target?.result as string
      parseCSV(text)
    } catch (error) {
      ElMessage.error('文件解析失败')
    }
  }
  reader.readAsText(file)
}

function parseCSV(text: string) {
  const lines = text.trim().split('\n')
  const data: any[] = []
  for (let i = 1; i < lines.length; i++) {
    const cols = lines[i].split(',').map(c => c.trim())
    if (cols.length >= 4) {
      data.push({
        uid: cols[0],
        name: cols[1],
        account: cols[2],
        amount: parseFloat(cols[3]) || 0,
        remark: cols[4] || ''
      })
    }
  }
  fileData.value = data
}

function clearFile() {
  fileName.value = ''
  fileData.value = []
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

async function executeBatchTransfer() {
  if (fileData.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  executing.value = true
  try {
    const res = await transferBatchCreate({
      filename: fileName.value,
      data: JSON.stringify(fileData.value)
    })
    ElMessage.success(res.msg || '批量转账任务已创建')
    clearFile()
    fetchBatchList()
  } catch (error) {
    console.error('创建批量转账失败:', error)
  } finally {
    executing.value = false
  }
}

async function fetchBatchList() {
  try {
    const res = await getTransferBatchList()
    if (res.code === 0) {
      batchList.value = res.data || []
    }
  } catch (error) {
    console.error('获取批量转账记录失败:', error)
  }
}

onMounted(() => {
  fetchBatchList()
})
</script>
