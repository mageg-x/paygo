<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">用户组管理</h1>
        <p class="text-sm text-gray-500 mt-1">管理商户用户组及费率配置</p>
      </div>
      <button @click="openAddDialog"
        class="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加用户组
      </button>
    </div>

    <!-- 用户组列表 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm whitespace-nowrap">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">组名称</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">备注</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">费率</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">结算</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">排序</th>
              <th class="px-4 py-3 text-center font-semibold text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="g in groups" :key="g.gid" class="hover:bg-gray-50/50 transition-colors">
              <td class="px-4 py-3 text-gray-900">{{ g.gid }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="text-gray-900 font-medium">{{ g.name }}</span>
                  <span v-if="g.gid === defaultGid"
                    class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700">
                    默认组
                  </span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-500">{{ g.info || '-' }}</td>
              <td class="px-4 py-3 text-center">
                <span class="text-emerald-600 font-medium">{{ (g.settle_rate || '0') }}%</span>
              </td>
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  g.settle_open === 1 ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600']">
                  {{ g.settle_open === 1 ? '开启' : '关闭' }}
                </span>
              </td>
              <td class="px-4 py-3 text-center text-gray-500">{{ g.sort }}</td>
              <td class="px-4 py-3 text-center">
                <button @click="openEditDialog(g)"
                  class="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded transition-colors">编辑</button>
                <button v-if="g.gid !== defaultGid" @click="setDefault(g.gid)"
                  class="px-3 py-1 text-xs text-purple-600 hover:bg-purple-50 rounded transition-colors ml-1">设默认</button>
                <button v-if="g.gid !== 1 && g.gid !== defaultGid" @click="handleDelete(g.gid)"
                  class="px-3 py-1 text-xs text-red-600 hover:bg-red-50 rounded transition-colors ml-1">删除</button>
              </td>
            </tr>
            <tr v-if="groups.length === 0">
              <td colspan="7" class="px-4 py-12 text-center text-gray-400">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                      d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  <span>暂无用户组</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 添加/编辑弹窗 -->
    <div v-if="dialogVisible" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="dialogVisible = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ isEdit ? '编辑用户组' : '添加用户组' }}</h3>

          <div class="space-y-4 max-h-[60vh] overflow-y-auto">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">组名称 *</label>
              <input v-model="form.name" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
              <input v-model="form.info" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">排序</label>
                <input v-model.number="form.sort" type="number"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">结算手续费率 (%)</label>
                <input v-model.number="form.settle_rate" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">结算开关</label>
              <select v-model="form.settle_open"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="1">开启</option>
                <option :value="0">关闭</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">允许购买</label>
                <select v-model="form.isbuy"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option :value="1">是</option>
                  <option :value="0">否</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">购买价格</label>
                <input v-model.number="form.price" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">结算周期</label>
              <select v-model="form.settle_type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option :value="0">实时结算</option>
                <option :value="1">每日结算</option>
                <option :value="2">每周结算</option>
                <option :value="3">每月结算</option>
              </select>
            </div>
          </div>

          <div class="flex justify-end gap-3 mt-6">
            <button @click="dialogVisible = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">取消</button>
            <button @click="handleSubmit" :disabled="submitting"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50">
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getGroupList, groupOp } from '@/api/admin'

interface Group {
  gid: number
  name: string
  info: string
  sort: number
  isbuy: number
  price: number
  expire: number
  settle_open: number
  settle_type: number
  settle_rate: string
  settings: string
  config: string
}

const groups = ref<Group[]>([])
const defaultGid = ref(1)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)

const form = reactive({
  gid: 0,
  name: '',
  info: '',
  sort: 0,
  isbuy: 0,
  price: 0,
  expire: 0,
  settle_open: 1,
  settle_type: 1,
  settle_rate: 0,
  settings: '',
  config: ''
})

async function fetchGroups() {
  try {
    const res = await getGroupList()
    if (res.code === 0) {
      groups.value = res.data || []
      if (res.default_gid) {
        defaultGid.value = res.default_gid
      }
    }
  } catch (error) {
    console.error('获取用户组列表失败:', error)
  }
}

async function setDefault(gid: number) {
  try {
    await ElMessageBox.confirm('确定要将此用户组设为默认分组吗？新注册商户将自动使用此分组。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await groupOp({ action: 'set_default', gid })
    if (res.code === 0) {
      ElMessage.success('设置成功')
      defaultGid.value = gid
    } else {
      ElMessage.error(res.msg || '设置失败')
    }
  } catch (error) {
    console.error('设置默认分组失败:', error)
  }
}

function openAddDialog() {
  isEdit.value = false
  form.gid = 0
  form.name = ''
  form.info = ''
  form.sort = 0
  form.isbuy = 0
  form.price = 0
  form.expire = 0
  form.settle_open = 1
  form.settle_type = 1
  form.settle_rate = 0
  form.settings = ''
  form.config = ''
  dialogVisible.value = true
}

function openEditDialog(g: Group) {
  isEdit.value = true
  form.gid = g.gid
  form.name = g.name
  form.info = g.info || ''
  form.sort = g.sort || 0
  form.isbuy = g.isbuy || 0
  form.price = g.price || 0
  form.expire = g.expire || 0
  form.settle_open = g.settle_open || 0
  form.settle_type = g.settle_type || 1
  form.settle_rate = Number(g.settle_rate || 0)
  form.settings = g.settings || ''
  form.config = g.config || ''
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入组名称')
    return
  }

  submitting.value = true
  try {
    const res = await groupOp({
      action: isEdit.value ? 'edit' : 'add',
      gid: form.gid,
      name: form.name,
      info: form.info,
      sort: form.sort,
      isbuy: form.isbuy,
      price: form.price,
      expire: form.expire,
      settle_open: form.settle_open,
      settle_type: form.settle_type,
      settle_rate: form.settle_rate,
      settings: form.settings,
      config: form.config
    })
    if (res.code === 0) {
      ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
      dialogVisible.value = false
      fetchGroups()
    } else {
      ElMessage.error(res.msg || '操作失败')
    }
  } catch (error) {
    console.error('操作失败:', error)
  } finally {
    submitting.value = false
  }
}

async function handleDelete(gid: number) {
  try {
    await ElMessageBox.confirm('确定要删除该用户组吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await groupOp({ action: 'delete', gid })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchGroups()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  fetchGroups()
})
</script>
