<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">黑名单管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理IP和账号黑名单</p>
      </div>
      <button @click="openAddDialog"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加黑名单
      </button>
    </div>

    <!-- 筛选 -->
    <div class="bg-white rounded-xl p-4 border border-gray-100 shadow-sm">
      <div class="flex items-center gap-4">
        <select v-model="filterType"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="">全部类型</option>
          <option value="1">IP黑名单</option>
          <option value="2">账号黑名单</option>
        </select>
        <button @click="page = 1; fetchList()"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm">
          筛选
        </button>
      </div>
    </div>

    <!-- 黑名单列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">类型</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">内容</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">备注</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">添加时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="b in list" :key="b.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ b.id }}</td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  b.type === 1 ? 'bg-red-100 text-red-700' : 'bg-orange-100 text-orange-700']">
                  {{ b.type === 1 ? 'IP黑名单' : '账号黑名单' }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-gray-900">{{ b.content }}</td>
              <td class="px-4 py-3 text-gray-500 max-w-xs truncate">{{ b.remark || '-' }}</td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(b.addtime) }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="handleDelete(b.id)"
                  class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">
                  删除
                </button>
              </td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                  </svg>
                  <span>暂无黑名单记录</span>
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

    <!-- 添加弹窗 -->
    <div v-if="dialogVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="dialogVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">添加黑名单</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <select v-model="form.type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">IP黑名单</option>
                <option :value="2">账号黑名单</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">内容</label>
              <input v-model="form.content" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="IP或账号" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
              <input v-model="form.remark" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="可选" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="dialogVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleAdd"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">添加</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { blacklistList, blacklistOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filterType = ref('')
const dialogVisible = ref(false)

const form = reactive({
  type: 1,
  content: '',
  remark: ''
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const params: any = { page: page.value, limit: pageSize.value }
    if (filterType.value) {
      params.type = filterType.value
    }
    const res = await blacklistList(params)
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取黑名单失败:', error)
  }
}

function openAddDialog() {
  form.type = 1
  form.content = ''
  form.remark = ''
  dialogVisible.value = true
}

async function handleAdd() {
  if (!form.content.trim()) {
    ElMessage.warning('请输入内容')
    return
  }
  try {
    const res = await blacklistOp({ action: 'add', type: form.type, content: form.content, remark: form.remark })
    if (res.code === 0) {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.msg || '添加失败')
    }
  } catch (error) {
    console.error('添加失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该黑名单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await blacklistOp({ action: 'delete', id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  fetchList()
})
</script>
