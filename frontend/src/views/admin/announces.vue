<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">公告管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理网站公告</p>
      </div>
      <button @click="openAddDialog"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加公告
      </button>
    </div>

    <!-- 公告列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">内容</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">颜色</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">排序</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">添加时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="a in list" :key="a.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ a.id }}</td>
              <td class="px-4 py-3">
                <div :style="{ color: a.color || '#333' }" class="max-w-xs truncate">{{ a.content }}</div>
              </td>
              <td class="px-4 py-3 text-center">
                <div class="flex items-center justify-center gap-2">
                  <div class="w-4 h-4 rounded" :style="{ backgroundColor: a.color || '#333' }"></div>
                  <span class="text-xs text-gray-500">{{ a.color || '#333' }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-center text-gray-500">{{ a.sort }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="toggleStatus(a)" :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium cursor-pointer transition-colors',
                  a.status === 1 ? 'bg-green-100 text-green-700 hover:bg-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]">
                  {{ a.status === 1 ? '显示' : '隐藏' }}
                </button>
              </td>
              <td class="px-4 py-3 text-gray-500 text-xs">{{ formatTime(a.addtime) }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="openEditDialog(a)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors mr-1">
                  编辑
                </button>
                <button @click="handleDelete(a.id)"
                  class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">
                  删除
                </button>
              </td>
            </tr>
            <tr v-if="list.length === 0">
              <td colspan="7" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
                  </svg>
                  <span>暂无公告</span>
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

    <!-- 添加/编辑弹窗 -->
    <div v-if="dialogVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="dialogVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ isEdit ? '编辑公告' : '添加公告' }}</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">内容</label>
              <textarea v-model="form.content" rows="3"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="公告内容"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">颜色</label>
              <div class="flex items-center gap-3">
                <input v-model="form.color" type="color"
                  class="w-10 h-10 border border-gray-200 rounded cursor-pointer" />
                <input v-model="form.color" type="text"
                  class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="#333333" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">排序</label>
              <input v-model="form.sort" type="number"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="数值越大越靠前" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select v-model="form.status"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">显示</option>
                <option :value="0">隐藏</option>
              </select>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="dialogVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleSave"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">{{ isEdit ? '保存' : '添加' }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { anounceList, anounceOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)

const form = ref({
  id: 0,
  content: '',
  color: '#333333',
  sort: 0,
  status: 1
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

async function fetchList() {
  try {
    const res = await anounceList({ page: page.value, limit: pageSize.value })
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取公告列表失败:', error)
  }
}

function openAddDialog() {
  isEdit.value = false
  form.value = { id: 0, content: '', color: '#333333', sort: 0, status: 1 }
  dialogVisible.value = true
}

function openEditDialog(a: any) {
  isEdit.value = true
  form.value = {
    id: a.id,
    content: a.content,
    color: a.color || '#333333',
    sort: a.sort,
    status: a.status
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.content.trim()) {
    ElMessage.warning('请输入公告内容')
    return
  }
  try {
    const action = isEdit.value ? 'edit' : 'add'
    const res = await anounceOp({
      action,
      id: form.value.id,
      content: form.value.content,
      color: form.value.color,
      sort: form.value.sort,
      status: form.value.status
    })
    if (res.code === 0) {
      ElMessage.success(isEdit.value ? '保存成功' : '添加成功')
      dialogVisible.value = false
      fetchList()
    } else {
      ElMessage.error(res.msg || '操作失败')
    }
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function toggleStatus(a: any) {
  try {
    const newStatus = a.status === 1 ? 0 : 1
    await anounceOp({ action: 'edit', id: a.id, content: a.content, color: a.color, sort: a.sort, status: newStatus })
    ElMessage.success(newStatus === 1 ? '已显示' : '已隐藏')
    a.status = newStatus
  } catch (error) {
    console.error('操作失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除该公告吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await anounceOp({ action: 'delete', id })
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
