<template>
  <div class="space-y-4">
    <h2 class="text-2xl font-bold text-gray-800">支付账号绑定</h2>
    <p class="text-gray-500">绑定您的支付宝/微信账号用于收款</p>

    <!-- 支付宝绑定 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-blue-500 rounded-xl flex items-center justify-center">
            <span class="text-2xl">💙</span>
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">支付宝</h3>
            <p class="text-sm text-gray-500">{{ userInfo?.alipay_uid || '未绑定' }}</p>
          </div>
        </div>
        <button @click="showAlipayDialog = true"
          :class="['px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            userInfo?.alipay_uid ? 'bg-gray-100 text-gray-700 hover:bg-gray-200' : 'bg-blue-600 text-white hover:bg-blue-700']">
          {{ userInfo?.alipay_uid ? '更换' : '立即绑定' }}
        </button>
      </div>
      <div class="mt-4 bg-blue-50 rounded-lg p-4">
        <p class="text-sm text-blue-800">
          绑定支付宝账号后，客户可以使用支付宝扫码支付，款项将直接进入您的支付宝账户。
        </p>
      </div>
    </div>

    <!-- 微信支付绑定 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-green-500 rounded-xl flex items-center justify-center">
            <span class="text-2xl">🟢</span>
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">微信支付</h3>
            <p class="text-sm text-gray-500">{{ userInfo?.wx_uid || '未绑定' }}</p>
          </div>
        </div>
        <button @click="showWxDialog = true"
          :class="['px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            userInfo?.wx_uid ? 'bg-gray-100 text-gray-700 hover:bg-gray-200' : 'bg-green-600 text-white hover:bg-green-700']">
          {{ userInfo?.wx_uid ? '更换' : '立即绑定' }}
        </button>
      </div>
      <div class="mt-4 bg-green-50 rounded-lg p-4">
        <p class="text-sm text-green-800">
          绑定微信支付账号后，客户可以使用微信扫码支付。
        </p>
      </div>
    </div>

    <!-- QQ钱包绑定 -->
    <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-purple-500 rounded-xl flex items-center justify-center">
            <span class="text-2xl">💜</span>
          </div>
          <div>
            <h3 class="font-semibold text-gray-800">QQ钱包</h3>
            <p class="text-sm text-gray-500">{{ userInfo?.qq_uid || '未绑定' }}</p>
          </div>
        </div>
        <button @click="showQQDialog = true"
          :class="['px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            userInfo?.qq_uid ? 'bg-gray-100 text-gray-700 hover:bg-gray-200' : 'bg-purple-600 text-white hover:bg-purple-700']">
          {{ userInfo?.qq_uid ? '更换' : '立即绑定' }}
        </button>
      </div>
      <div class="mt-4 bg-purple-50 rounded-lg p-4">
        <p class="text-sm text-purple-800">
          绑定QQ钱包账号后，客户可以使用QQ扫码支付。
        </p>
      </div>
    </div>

    <!-- 支付宝绑定弹窗 -->
    <div v-if="showAlipayDialog" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showAlipayDialog = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">绑定支付宝</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">支付宝账号</label>
              <input v-model="alipayForm.account" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="手机号/邮箱" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">真实姓名</label>
              <input v-model="alipayForm.name" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入真实姓名" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showAlipayDialog = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
            <button @click="bindAlipay"
              class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700">确认绑定</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 微信绑定弹窗 -->
    <div v-if="showWxDialog" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showWxDialog = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">绑定微信支付</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">OpenID</label>
              <input v-model="wxForm.openid" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                placeholder="请输入微信OpenID" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showWxDialog = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
            <button @click="bindWx"
              class="px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700">确认绑定</button>
          </div>
        </div>
      </div>
    </div>

    <!-- QQ绑定弹窗 -->
    <div v-if="showQQDialog" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="showQQDialog = false"></div>
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">绑定QQ钱包</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">QQ号</label>
              <input v-model="qqForm.qq" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="请输入QQ号" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showQQDialog = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
            <button @click="bindQQ"
              class="px-4 py-2 text-sm bg-purple-600 text-white rounded-lg hover:bg-purple-700">确认绑定</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { getUserInfo, updateProfile } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { ElMessage } from 'element-plus'

const appStore = useAppStore()
const userInfo = computed(() => appStore.userInfo)

const showAlipayDialog = ref(false)
const showWxDialog = ref(false)
const showQQDialog = ref(false)

const alipayForm = ref({ account: '', name: '' })
const wxForm = ref({ openid: '' })
const qqForm = ref({ qq: '' })

async function bindAlipay() {
  if (!alipayForm.value.account) {
    ElMessage.warning('请输入支付宝账号')
    return
  }
  try {
    await updateProfile({ alipay_uid: alipayForm.value.account })
    ElMessage.success('支付宝绑定成功')
    showAlipayDialog.value = false
    refreshUserInfo()
  } catch (error) {
    console.error('绑定失败:', error)
    ElMessage.error('绑定失败')
  }
}

async function bindWx() {
  if (!wxForm.value.openid) {
    ElMessage.warning('请输入OpenID')
    return
  }
  try {
    await updateProfile({ wx_uid: wxForm.value.openid })
    ElMessage.success('微信绑定成功')
    showWxDialog.value = false
    refreshUserInfo()
  } catch (error) {
    console.error('绑定失败:', error)
    ElMessage.error('绑定失败')
  }
}

async function bindQQ() {
  if (!qqForm.value.qq) {
    ElMessage.warning('请输入QQ号')
    return
  }
  try {
    await updateProfile({ qq_uid: qqForm.value.qq })
    ElMessage.success('QQ钱包绑定成功')
    showQQDialog.value = false
    refreshUserInfo()
  } catch (error) {
    console.error('绑定失败:', error)
    ElMessage.error('绑定失败')
  }
}

async function refreshUserInfo() {
  try {
    const res = await getUserInfo()
    if (res.code === 0 && res.data) {
      appStore.userLogin(appStore.userToken!, {
        uid: res.data.uid,
        username: res.data.username || '',
        email: res.data.email || '',
        phone: res.data.phone || '',
        money: res.data.money || 0,
        status: res.data.status || 1
      })
    }
  } catch (error) {
    console.error('刷新用户信息失败:', error)
  }
}
</script>
