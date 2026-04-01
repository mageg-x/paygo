<template>
  <div class="space-y-4">
    <h2 class="text-2xl font-bold text-gray-800">帮助中心</h2>

    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <!-- 左侧目录 -->
      <div class="md:col-span-1">
        <div class="card">
          <div class="card-body p-2">
            <button v-for="(item, index) in categories" :key="index" @click="activeCategory = index"
              :class="[
                'w-full text-left px-4 py-3 rounded-lg text-sm transition-colors',
                activeCategory === index ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-50'
              ]">
              {{ item.title }}
            </button>
          </div>
        </div>
      </div>

      <!-- 右侧内容 -->
      <div class="md:col-span-3">
        <div class="card">
          <div class="card-body">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">{{ categories[activeCategory].title }}</h3>
            <div class="space-y-4">
              <div v-for="(qa, qaIndex) in categories[activeCategory].items" :key="qaIndex"
                class="border-b border-gray-100 pb-4 last:border-0">
                <button @click="qa.open = !qa.open"
                  class="w-full flex items-center justify-between text-left">
                  <span class="font-medium text-gray-800">{{ qa.q }}</span>
                  <span :class="['transition-transform', qa.open ? 'rotate-180' : '']">▼</span>
                </button>
                <div v-show="qa.open" class="mt-3 text-gray-600 text-sm leading-relaxed">
                  {{ qa.a }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 联系客服 -->
        <div class="card mt-4">
          <div class="card-body">
            <h4 class="font-medium text-gray-800 mb-3">联系客服</h4>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="bg-gray-50 rounded-lg p-4">
                <p class="text-sm text-gray-500 mb-1">商务合作</p>
                <p class="text-gray-800">contact@example.com</p>
              </div>
              <div class="bg-gray-50 rounded-lg p-4">
                <p class="text-sm text-gray-500 mb-1">技术支持</p>
                <p class="text-gray-800">support@example.com</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const activeCategory = ref(0)

const categories = ref([
  {
    title: '快速开始',
    items: [
      {
        q: '如何注册商户账号？',
        a: '点击首页的"商户注册"按钮，填写邮箱、密码等信息即可完成注册。如果开启了邀请码功能，请填写有效的邀请码以获得返现。',
        open: true
      },
      {
        q: '如何获取API密钥？',
        a: '登录商户后台后，进入"资料管理"页面即可查看您的商户ID和API密钥。请妥善保管密钥，不要泄露给他人。',
        open: false
      },
      {
        q: '如何配置支付回调？',
        a: '在商户后台的"资料管理"中填写您的notify_url，这是接收支付结果通知的地址。当有订单支付成功时，系统会自动POST订单信息到该地址。',
        open: false
      }
    ]
  },
  {
    title: '支付集成',
    items: [
      {
        q: '支付接口地址是什么？',
        a: '支付接口地址为：/api/pay/create（POST请求）。请参考API文档获取详细的参数说明和签名方式。',
        open: false
      },
      {
        q: '如何生成支付订单？',
        a: '调用支付接口时需要传入：pid（商户ID）、out_trade_no（商户订单号）、type（支付方式）、money（金额）等参数。具体请参考接口文档。',
        open: false
      },
      {
        q: '订单状态如何查询？',
        a: '使用您的商户ID和密钥，通过/api/pay/query接口查询订单状态。返回的status字段：0=待支付，1=已支付，2=已关闭。',
        open: false
      },
      {
        q: '支付成功后如何通知我的系统？',
        a: '当订单状态变化时，系统会向您的notify_url发送POST请求，包含trade_no、out_trade_no、status等参数。请验证签名后处理您的业务逻辑。',
        open: false
      }
    ]
  },
  {
    title: '结算相关',
    items: [
      {
        q: '结算周期是多久？',
        a: '默认结算周期为每日结算，即当日的收入会在次日进行结算。具体结算周期可能因用户组不同而有所差异。',
        open: false
      },
      {
        q: '如何申请结算？',
        a: '当账户余额达到最低结算金额（默认30元）后，您可以在"结算管理"页面申请结算。选择结算方式和账户信息后提交即可。',
        open: false
      },
      {
        q: '结算手续费是多少？',
        a: '结算手续费根据用户组不同而有所差异，一般为1%-3%。具体费率请查看您所在用户组的费率配置。',
        open: false
      }
    ]
  },
  {
    title: '常见问题',
    items: [
      {
        q: '支付失败怎么办？',
        a: '支付失败可能原因：1) 余额不足；2) 超过限额；3) 通道维护；4) 风控拦截。请查看错误提示，或联系客服协助解决。',
        open: false
      },
      {
        q: '如何联系客服？',
        a: '您可以通过以下方式联系客服：1) 商户后台右上角客服入口；2) 发送邮件至support@example.com；3) 技术支持QQ群。',
        open: false
      },
      {
        q: '账户密码忘记了怎么办？',
        a: '在登录页面点击"忘记密码"，输入注册的邮箱地址，系统会发送验证码给您，通过验证后即可重置密码。',
        open: false
      }
    ]
  }
])
</script>
