<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">计划任务</h1>
        <p class="text-sm text-gray-500 mt-1">管理系统定时任务</p>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">任务名称</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">执行周期</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">下次执行</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="task in taskList" :key="task.name" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full" :class="task.running ? 'bg-green-500' : 'bg-gray-300'"></span>
                  <span class="font-medium text-gray-900">{{ taskName(task.name) }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-center">
                <span class="font-mono text-xs bg-gray-100 px-2 py-1 rounded">{{ task.spec || '默认组' }}</span>
              </td>
              <td class="px-4 py-3 text-center text-gray-500 text-xs">{{ task.next || '-' }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="toggleTask(task)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  task.enabled ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ task.enabled ? '启用' : '禁用' }}
                </button>
              </td>
              <td class="px-4 py-3 text-center">
                <button @click="runTask(task)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors mr-1">
                  立即执行
                </button>
                <button @click="openEditDialog(task)"
                  class="px-3 py-1 text-xs text-gray-600 hover:bg-gray-100 rounded transition-colors">
                  编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <div v-if="dialogVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="dialogVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">编辑执行周期</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">任务名称</label>
              <input :value="currentTask.name" type="text" disabled
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-gray-50" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Cron表达式</label>
              <input v-model="form.spec" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="0 * * * * ?" />
              <p class="text-xs text-gray-500 mt-1">
                格式: 秒 分 时 日 月 周(可选)
                <br />常用: <span class="cursor-pointer text-blue-500" @click="form.spec = '0 */5 * * * ?'">每5分钟</span> |
                <span class="cursor-pointer text-blue-500" @click="form.spec = '0 0 * * * ?'">每小时</span> |
                <span class="cursor-pointer text-blue-500" @click="form.spec = '0 0 0 * * ?'">每天</span>
              </p>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="dialogVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="saveTask"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { cronList, cronOp } from '@/api/admin'
import { ElMessage } from 'element-plus'

const taskList = ref<any[]>([])
const dialogVisible = ref(false)
const currentTask = ref<any>({})
const form = ref({
  spec: ''
})

const defaultTasks = [
  { name: 'auto_settle', desc: '自动结算', defaultSpec: '0 0 * * * ?' },
  { name: 'retry_notify', desc: '回调重试', defaultSpec: '0 */5 * * * ?' },
  { name: 'risk_check', desc: '风控检查', defaultSpec: '0 */30 * * * ?' },
  { name: 'cleanup', desc: '清理过期数据', defaultSpec: '0 0 0 * * ?' }
]

function taskName(name: string) {
  const task = defaultTasks.find(t => t.name === name)
  return task ? task.desc : name
}

async function fetchTasks() {
  try {
    const res = await cronList()
    if (res.code === 0) {
      // 合并任务列表
      const serverTasks = res.data || []
      taskList.value = defaultTasks.map(t => {
        const serverTask = serverTasks.find((s: any) => s.name === t.name)
        return {
          name: t.name,
          desc: t.desc,
          spec: serverTask?.spec || t.defaultSpec,
          next: serverTask?.next || '-',
          running: serverTask?.running || false,
          enabled: serverTask?.next !== undefined
        }
      })
    }
  } catch (error) {
    console.error('获取任务列表失败:', error)
  }
}

async function toggleTask(task: any) {
  try {
    await cronOp({
      action: 'set',
      name: task.name,
      enable: !task.enabled
    })
    ElMessage.success(task.enabled ? '已禁用' : '已启用')
    task.enabled = !task.enabled
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function runTask(task: any) {
  try {
    await cronOp({ action: 'run', name: task.name })
    ElMessage.success('任务已触发')
  } catch (error) {
    console.error('操作失败:', error)
  }
}

function openEditDialog(task: any) {
  currentTask.value = task
  form.value.spec = task.spec
  dialogVisible.value = true
}

async function saveTask() {
  if (!form.value.spec.trim()) {
    ElMessage.warning('请输入执行周期')
    return
  }
  try {
    await cronOp({
      action: 'set',
      name: currentTask.value.name,
      spec: form.value.spec
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchTasks()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

onMounted(() => {
  fetchTasks()
})
</script>
