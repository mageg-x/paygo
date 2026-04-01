<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">邀请推广</h1>
        <p class="text-gray-500 mt-1">邀请新商户注册，获得返现奖励</p>
      </div>
    </div>

    <!-- 推广链接卡片 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">我的推广链接</h3>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">推广链接</label>
          <div class="flex gap-2">
            <input :value="inviteUrl" readonly
              class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm bg-gray-50 font-mono" />
            <button @click="copyUrl"
              class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm">
              复制链接
            </button>
          </div>
          <p class="text-xs text-gray-400 mt-1">复制链接发送给好友，好友通过此链接注册后自动成为您的下线</p>
        </div>

        <div class="grid grid-cols-2 gap-4 pt-4 border-t">
          <div class="text-center p-4 bg-gray-50 rounded-lg">
            <div class="text-2xl font-bold text-blue-600">{{ stats.inviteCount }}</div>
            <div class="text-sm text-gray-500">邀请人数</div>
          </div>
          <div class="text-center p-4 bg-gray-50 rounded-lg">
            <div class="text-2xl font-bold text-green-600">￥{{ stats.inviteMoney }}</div>
            <div class="text-sm text-gray-500">返现总额</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 邀请返现规则 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">返现规则</h3>
      <div class="space-y-2 text-sm text-gray-600">
        <div class="flex items-start gap-2">
          <span class="text-green-500">1.</span>
          <span>每邀请一位商户成功注册并完成首笔交易，您可获得相应的返现奖励</span>
        </div>
        <div class="flex items-start gap-2">
          <span class="text-green-500">2.</span>
          <span>返现金额从被邀请商户的交易手续费中扣除，不影响商户收益</span>
        </div>
        <div class="flex items-start gap-2">
          <span class="text-green-500">3.</span>
          <span>返现比例由管理员在后台设置，详情请联系管理员</span>
        </div>
      </div>
    </div>

    <!-- 邀请记录 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100">
        <h3 class="text-lg font-semibold text-gray-900">邀请记录</h3>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-100">
              <th class="px-4 py-3 text-left font-semibold text-gray-600">商户ID</th>
              <th class="px-4 py-3 text-left font-semibold text-gray-600">注册时间</th>
              <th class="px-4 py-3 text-right font-semibold text-gray-600">返现金额</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            <tr v-for="record in records" :key="record.id" class="hover:bg-gray-50/50">
              <td class="px-4 py-3">{{ record.uid }}</td>
              <td class="px-4 py-3 text-gray-500">{{ formatTime(record.addtime) }}</td>
              <td class="px-4 py-3 text-right text-green-600 font-medium">+￥{{ record.money }}</td>
            </tr>
            <tr v-if="records.length === 0">
              <td colspan="3" class="px-4 py-8 text-center text-gray-400">
                暂无邀请记录
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const records = ref<any[]>([])

const stats = computed(() => ({
  inviteCount: records.value.length,
  inviteMoney: records.value.reduce((sum, r) => sum + (r.money || 0), 0).toFixed(2)
}))

const inviteUrl = computed(() => {
  if (!appStore.userInfo?.uid) return ''
  const code = btoa(String(appStore.userInfo.uid))
  return `${window.location.origin}/user/reg?invite=${code}`
})

function copyUrl() {
  navigator.clipboard.writeText(inviteUrl.value)
  ElMessage.success('推广链接已复制')
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(async () => {
  // 暂时使用模拟数据
  records.value = []
})
</script>
