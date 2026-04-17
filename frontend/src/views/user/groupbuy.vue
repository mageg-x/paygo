<template>
  <div class="space-y-4">
    <h2 class="text-2xl font-bold text-gray-800">会员购买</h2>
    <p class="text-gray-500">升级用户组，享受更低费率</p>

    <!-- 当前用户组 -->
    <div class="bg-blue-50 rounded-xl p-4 border border-blue-100">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-blue-600">当前用户组</p>
          <p class="text-xl font-bold text-blue-800">{{ currentGroup?.name || '默认组' }}</p>
        </div>
        <div class="text-right">
          <p class="text-sm text-blue-600">结算费率</p>
          <p class="text-xl font-bold text-blue-800">{{ currentGroup?.settle_rate || '0' }}%</p>
        </div>
      </div>
    </div>

    <!-- 用户组列表 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div v-for="group in groups" :key="group.gid"
        :class="['bg-white rounded-xl border-2 p-6 transition-all cursor-pointer relative',
          group.gid === currentGid ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200 hover:border-blue-300']"
        @click="selectGroup(group)">
        <div v-if="group.gid === currentGid"
          class="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-500 text-white text-xs px-3 py-1 rounded-full">
          当前
        </div>
        <div class="text-center">
          <h3 class="text-lg font-bold text-gray-800">{{ group.name }}</h3>
          <p class="text-sm text-gray-500 mt-1">{{ group.info || '暂无描述' }}</p>
          <div class="mt-4">
            <span class="text-3xl font-bold text-blue-600">¥{{ group.price }}</span>
            <span class="text-gray-500 text-sm">/{{ group.expire === 0 ? '永久' : group.expire + '个月' }}</span>
          </div>
          <div class="mt-4 space-y-2 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-gray-500">结算费率</span>
              <span class="font-medium text-emerald-600">{{ group.settle_rate }}%</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-gray-500">结算周期</span>
              <span class="font-medium">{{ settleCycleName(group.settle_type) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-gray-500">自动结算</span>
              <span :class="group.settle_open ? 'text-emerald-600' : 'text-gray-400'">
                {{ group.settle_open ? '✓ 支持' : '✗ 不支持' }}
              </span>
            </div>
          </div>
          <button
            :class="['w-full mt-4 py-2 rounded-lg font-medium transition-colors',
              group.gid === currentGid
                ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                : 'bg-blue-600 text-white hover:bg-blue-700']"
            :disabled="group.gid === currentGid">
            {{ group.gid === currentGid ? '当前组' : '立即升级' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 升级说明 -->
    <div class="bg-gray-50 rounded-xl p-4 mt-6">
      <h4 class="font-medium text-gray-800 mb-2">升级说明</h4>
      <ul class="text-sm text-gray-600 space-y-1 list-disc list-inside">
        <li>升级后立即生效，有效期内可享受对应费率优惠</li>
        <li>费用按剩余月数比例计算</li>
        <li>如需取消升级，请联系客服</li>
      </ul>
    </div>

    <!-- 购买确认弹窗 -->
    <div v-if="selectedGroup" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="selectedGroup = null"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">确认升级</h3>
          <div class="space-y-4">
            <div class="bg-gray-50 rounded-lg p-4">
              <div class="flex justify-between mb-2">
                <span class="text-gray-500">当前用户组</span>
                <span>{{ currentGroup?.name }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">升级至</span>
                <span class="text-blue-600 font-medium">{{ selectedGroup.name }}</span>
              </div>
            </div>
            <div class="border-t pt-4">
              <div class="flex justify-between items-center">
                <span class="text-gray-600">应付金额</span>
                <span class="text-2xl font-bold text-blue-600">¥{{ selectedGroup.price }}</span>
              </div>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="selectedGroup = null"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
            <button @click="handleBuy"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700">确认支付</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUserInfo, getUserGroupList, buyUserGroup } from '@/api/user'
import { ElMessage } from 'element-plus'

const currentGid = ref(0)
const currentGroup = ref<any>(null)
const selectedGroup = ref<any>(null)
const groups = ref<any[]>([])

function settleCycleName(cycle: number) {
  const map: Record<number, string> = {
    0: '实时结算',
    1: '每日结算',
    2: '每周结算'
  }
  return map[cycle] || '未知'
}

function selectGroup(group: any) {
  if (group.gid === currentGid.value) return
  selectedGroup.value = group
}

async function handleBuy() {
  if (!selectedGroup.value) return
  try {
    const res = await buyUserGroup({ group_id: selectedGroup.value.gid })
    if (res.code === 0) {
      ElMessage.success(res.msg || '购买成功')
      selectedGroup.value = null
      await fetchData()
    }
  } catch (error) {
    console.error('购买失败:', error)
  }
}

async function fetchData() {
  try {
    const infoRes = await getUserInfo()
    if (infoRes.code === 0 && infoRes.data) {
      currentGid.value = infoRes.data.gid || 0
    }

    const groupRes = await getUserGroupList()
    if (groupRes.code === 0) {
      groups.value = (groupRes.data || []).filter((g: any) => g.isbuy === 1)
    }

    currentGroup.value = groups.value.find(g => g.gid === currentGid.value) || null
  } catch (error) {
    console.error('获取数据失败:', error)
  }
}

onMounted(() => {
  fetchData()
})
</script>
