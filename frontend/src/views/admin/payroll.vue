<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">支付轮询配置</h1>
        <p class="text-sm text-gray-500 mt-1">配置通道轮询规则和权重</p>
      </div>
      <button @click="showAddModal"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加规则
      </button>
    </div>

    <!-- 轮询规则列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">规则名称</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">支付类型</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">轮询方式</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">通道配置</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">优先级</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">状态</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="r in rolls" :key="r.id" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900 font-medium">{{ r.id }}</td>
              <td class="px-4 py-3 text-gray-900">{{ r.name }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <SvgIcon :name="r.type === 1 ? 'alipay' : 'wechatpay'" :size="16" />
                  <span>{{ typeName(r.type) }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-500">{{ kindName(r.kind) }}</td>
              <td class="px-4 py-3 text-gray-500 text-xs">
                <div v-if="r.info" class="max-w-xs truncate">{{ r.info }}</div>
                <div v-else class="text-gray-400">未配置</div>
              </td>
              <td class="px-4 py-3 text-center text-gray-500">{{ r.index }}</td>
              <td class="px-4 py-3 text-center">
                <span
                  :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', r.status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                  {{ r.status ? '开启' : '关闭' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <div class="inline-flex items-center gap-1">
                  <button @click="showEditModal(r)"
                    class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">编辑</button>
                  <button @click="toggleStatus(r)"
                    :class="['px-3 py-1 text-xs rounded transition-colors', r.status ? 'text-yellow-600 hover:bg-yellow-50' : 'text-green-600 hover:bg-green-50']">
                    {{ r.status ? '关闭' : '开启' }}
                  </button>
                  <button @click="handleDelete(r.id)"
                    class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="rolls.length === 0">
              <td colspan="8" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  <span>暂无轮询规则</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 添加/编辑弹窗 -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showModal = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">{{ isEdit ? '编辑轮询规则' : '添加轮询规则' }}</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">规则名称</label>
              <input v-model="form.name" type="text" placeholder="如：默认轮询规则"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">支付类型</label>
              <select v-model="form.type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">支付宝</option>
                <option :value="2">微信支付</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">轮询方式</label>
              <select v-model="form.kind"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="0">权重轮询</option>
                <option :value="1">顺序轮询</option>
                <option :value="2">随机分配</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">通道配置 (JSON)</label>
              <textarea v-model="form.info" rows="3" placeholder='[{"channel_id":1,"weight":60},{"channel_id":2,"weight":40}]'
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">优先级</label>
              <input v-model.number="form.index" type="number"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select v-model="form.status"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-8">
            <button @click="showModal = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleSave"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRollList, rollOp } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'
import SvgIcon from '@/components/svgicon.vue'

const rolls = ref<any[]>([])
const showModal = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  type: 1,
  kind: 0,
  info: '',
  index: 0,
  status: 1
})

async function fetchRolls() {
  try {
    const res = await getRollList()
    if (res.code === 0) {
      rolls.value = res.data || []
    }
  } catch (error) {
    console.error('获取轮询配置失败:', error)
  }
}

function showAddModal() {
  isEdit.value = false
  form.value = {
    id: 0,
    name: '',
    type: 1,
    kind: 0,
    info: '',
    index: 0,
    status: 1
  }
  showModal.value = true
}

function showEditModal(r: any) {
  isEdit.value = true
  form.value = {
    id: r.id,
    name: r.name,
    type: r.type,
    kind: r.kind,
    info: r.info || '',
    index: r.index || 0,
    status: r.status
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入规则名称')
    return
  }
  try {
    const res = await rollOp({
      action: isEdit.value ? 'edit' : 'add',
      ...form.value
    })
    ElMessage.success(res.msg || '保存成功')
    showModal.value = false
    fetchRolls()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确定要删除这个规则吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await rollOp({ action: 'delete', id })
    ElMessage.success(res.msg || '删除成功')
    fetchRolls()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function toggleStatus(r: any) {
  try {
    const res = await rollOp({
      action: 'set_status',
      id: r.id,
      status: r.status ? 0 : 1
    })
    ElMessage.success(res.msg || '状态已更新')
    fetchRolls()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

function typeName(type: number): string {
  const map: Record<number, string> = {
    1: '支付宝',
    2: '微信支付'
  }
  return map[type] || '未知'
}

function kindName(kind: number): string {
  const map: Record<number, string> = {
    0: '权重轮询',
    1: '顺序轮询',
    2: '随机分配'
  }
  return map[kind] || '未知'
}

onMounted(() => {
  fetchRolls()
})
</script>
