<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">邀请码管理</h1>
        <p class="text-sm text-gray-500 mt-1">生成和管理商户注册邀请码</p>
      </div>
      <button @click="showGenerateModal = true"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        生成邀请码
      </button>
    </div>

    <!-- 搜索框 -->
    <div class="flex gap-4">
      <div class="relative flex-1 max-w-md">
        <input v-model="search" type="text" placeholder="搜索邀请码..."
          class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          @keyup.enter="fetchList" />
        <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <button @click="fetchList"
        class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm">
        搜索
      </button>
    </div>

    <!-- 邀请码列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">邀请码</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">生成时间</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">使用人</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">使用时间</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="code in list" :key="code.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-medium">{{ code.id }}</td>
              <td class="px-4 py-3">
                <code class="bg-gray-100 px-2 py-1 rounded text-sm font-mono">{{ code.code }}</code>
              </td>
              <td class="px-4 py-3 text-gray-500">{{ formatTime(code.addtime) }}</td>
              <td class="px-4 py-3 text-gray-500">{{ code.uid || '-' }}</td>
              <td class="px-4 py-3 text-gray-500">{{ code.usetime ? formatTime(code.usetime) : '-' }}</td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  code.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                  {{ code.status ? '已使用' : '未使用' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <button v-if="!code.status" @click="handleDelete(code.id)"
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
                      d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                  </svg>
                  <span>暂无邀请码</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="total > limit" class="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <div class="flex gap-2">
          <button @click="page--; fetchList()" :disabled="page <= 1"
            class="px-3 py-1 text-sm border rounded hover:bg-gray-50 disabled:opacity-50">
            上一页
          </button>
          <button @click="page++; fetchList()" :disabled="page * limit >= total"
            class="px-3 py-1 text-sm border rounded hover:bg-gray-50 disabled:opacity-50">
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 生成邀请码弹窗 -->
    <div v-if="showGenerateModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showGenerateModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">生成邀请码</h3>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">生成数量</label>
            <input v-model.number="generateNum" type="number" min="1" max="100"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="1-100" />
            <p class="text-xs text-gray-400 mt-1">每次最多生成 100 个</p>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showGenerateModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleGenerate" :disabled="generating"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50">
              {{ generating ? '生成中...' : '生成' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 生成的邀请码弹窗 -->
    <div v-if="showCodesModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showCodesModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">生成的邀请码</h3>
          <div class="max-h-60 overflow-y-auto space-y-2">
            <div v-for="code in generatedCodes" :key="code"
              class="flex items-center justify-between bg-gray-50 rounded-lg px-3 py-2">
              <code class="font-mono">{{ code }}</code>
              <button @click="copyCode(code)"
                class="text-xs text-blue-600 hover:text-blue-700">复制</button>
            </div>
          </div>
          <div class="flex justify-end mt-4">
            <button @click="showCodesModal = false"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getInviteCodeList, generateInviteCode, deleteInviteCode } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const search = ref('')
const showGenerateModal = ref(false)
const showCodesModal = ref(false)
const generateNum = ref(1)
const generating = ref(false)
const generatedCodes = ref<string[]>([])

async function fetchList() {
  try {
    const res = await getInviteCodeList({ page: page.value, limit, search: search.value })
    if (res.code === 0) {
      list.value = res.data || []
      total.value = res.count || 0
    }
  } catch (error) {
    console.error('获取邀请码列表失败:', error)
  }
}

async function handleGenerate() {
  if (generateNum.value < 1 || generateNum.value > 100) {
    ElMessage.warning('数量必须在 1-100 之间')
    return
  }
  generating.value = true
  try {
    const res = await generateInviteCode(generateNum.value)
    if (res.code === 0) {
      generatedCodes.value = res.codes || []
      showGenerateModal.value = false
      showCodesModal.value = true
      fetchList()
    } else {
      ElMessage.error(res.msg || '生成失败')
    }
  } catch (error) {
    console.error('生成邀请码失败:', error)
  } finally {
    generating.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个邀请码吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await deleteInviteCode(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    }
  } catch (error) {
    console.error('删除失败:', error)
  }
}

function copyCode(code: string) {
  navigator.clipboard.writeText(code)
  ElMessage.success('已复制')
}

function formatTime(time: string) {
  if (!time) return ''
  const d = new Date(time)
  return d.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchList()
})
</script>
