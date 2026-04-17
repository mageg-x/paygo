<template>
  <div class="max-w-md mx-auto space-y-4">
    <!-- 页面标题 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h1 class="text-2xl font-bold text-gray-900 mb-2">转让用户组</h1>
      <p class="text-sm text-gray-500">将您的用户组转让给其他商户</p>
    </div>

    <!-- 转让表单 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">当前用户组</label>
          <div class="px-4 py-2.5 bg-gray-50 rounded-lg text-gray-900">
            {{ currentGroup || '默认组' }}
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">目标商户UID</label>
          <input v-model="form.targetUid" type="number" placeholder="输入目标商户UID"
            class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">目标用户组</label>
          <select v-model="form.targetGroup"
            class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">选择目标用户组</option>
            <option v-for="g in groups" :key="g.gid" :value="g.gid">{{ g.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">转让价格 (元)</label>
          <input v-model.number="form.price" type="number" min="0" step="0.01" placeholder="输入转让价格"
            class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>

        <div class="pt-2">
          <button @click="handleTransfer" :disabled="submitting"
            class="w-full px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">
            {{ submitting ? '提交中...' : '确认转让' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 转让记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100">
        <h3 class="text-lg font-semibold text-gray-900">转让记录</h3>
      </div>
      <div class="divide-y divide-gray-50">
        <div v-for="t in transfers" :key="t.id" class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-medium text-gray-900">转让给 UID: {{ t.to_uid }}</div>
              <div class="text-xs text-gray-500 mt-0.5">{{ t.created_at }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold text-green-600">￥{{ t.price }}</div>
              <span :class="['inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium mt-1',
                t.status === 1 ? 'bg-green-100 text-green-700' :
                t.status === 2 ? 'bg-yellow-100 text-yellow-700' :
                'bg-gray-100 text-gray-600']">
                {{ t.status === 1 ? '已完成' : t.status === 2 ? '处理中' : '已取消' }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="transfers.length === 0" class="p-8 text-center text-gray-400">
          暂无转让记录
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserGroupTransferList, createUserGroupTransfer, getUserGroupList } from '@/api/user'
import { ElMessage } from 'element-plus'

const currentGroup = ref('')
const groups = ref<any[]>([])
const transfers = ref<any[]>([])
const submitting = ref(false)
const form = ref({
  targetUid: 0,
  targetGroup: '',
  price: 0
})

async function fetchGroups() {
  try {
    const res = await getUserGroupList()
    if (res.code === 0) {
      groups.value = res.data || []
    }
  } catch (error) {
    console.error('获取用户组失败:', error)
  }
}

async function fetchTransfers() {
  try {
    const res = await getUserGroupTransferList()
    if (res.code === 0) {
      transfers.value = res.data || []
    }
  } catch (error) {
    console.error('获取转让记录失败:', error)
  }
}

async function handleTransfer() {
  if (!form.value.targetUid) {
    ElMessage.warning('请输入目标商户UID')
    return
  }
  if (!form.value.targetGroup) {
    ElMessage.warning('请选择目标用户组')
    return
  }
  if (form.value.price < 0) {
    ElMessage.warning('转让价格不能为负数')
    return
  }
  submitting.value = true
  try {
    const res = await createUserGroupTransfer({
      target_uid: form.value.targetUid,
      group_id: Number(form.value.targetGroup),
      price: form.value.price
    })
    if (res.code === 0) {
      ElMessage.success('转让请求已提交')
      form.value = { targetUid: 0, targetGroup: '', price: 0 }
      fetchTransfers()
    } else {
      ElMessage.error(res.msg || '转让失败')
    }
  } catch (error) {
    console.error('转让失败:', error)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchGroups()
  fetchTransfers()
})
</script>
