<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
      <template v-if="loading">
        <div class="flex justify-center py-12">
          <div class="loading-spinner w-8 h-8"></div>
        </div>
      </template>

      <template v-else-if="order">
        <div class="text-center mb-6">
          <h2 class="text-2xl font-bold text-gray-800 mb-2">{{ order.name }}</h2>
          <div class="text-3xl font-bold text-primary-600">
            ¥{{ order.money }}
          </div>
        </div>

        <div class="mb-6">
          <div class="text-sm text-gray-600 mb-3">选择支付方式</div>
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="pt in payTypes"
              :key="pt.id"
              :class="[
                'p-4 rounded-xl border-2 text-center transition-all',
                selectedType === pt.id
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-gray-200 hover:border-gray-300'
              ]"
              @click="selectPayType(pt.id)"
            >
              <div class="text-lg font-medium">{{ pt.name }}</div>
            </button>
          </div>
        </div>

        <div v-if="qrCodeUrl" class="text-center mb-6">
          <img :src="qrCodeUrl" alt="支付二维码" class="mx-auto mb-4" />
          <p class="text-sm text-gray-500">请使用{{ selectedType === 'alipay' ? '支付宝' : selectedType === 'wxpay' ? '微信' : 'QQ' }}扫码支付</p>
        </div>

        <div class="text-center text-sm text-gray-500">
          订单号: {{ tradeNo }}
        </div>
      </template>

      <template v-else>
        <div class="text-center py-12">
          <p class="text-gray-500">订单不存在或已过期</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import QRCode from 'qrcode'

const route = useRoute()
const tradeNo = route.params.trade_no as string

const loading = ref(true)
const order = ref<any>(null)
const qrCodeUrl = ref('')
const selectedType = ref('alipay')

const payTypes = [
  { id: 'alipay', name: '支付宝', color: '#1677FF' },
  { id: 'wxpay', name: '微信支付', color: '#07C160' },
]

onMounted(async () => {
  try {
    const res = await fetch(`/api/pay/query?trade_no=${tradeNo}`)
    const data = await res.json()

    if (data.code === 0) {
      order.value = data

      if (data.pay_info) {
        qrCodeUrl.value = await QRCode.toDataURL(data.pay_info, {
          width: 200,
          margin: 2
        })
      }
    }
  } catch (error) {
    console.error('获取订单失败:', error)
  } finally {
    loading.value = false
  }
})

function selectPayType(typeId: string) {
  selectedType.value = typeId
}
</script>
