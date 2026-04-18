<template>
  <div class="space-y-4">
    <!-- 页面标题 -->
    <div>
      <h1 class="text-2xl font-bold text-gray-900">系统设置</h1>
      <p class="text-sm text-gray-500 mt-1">配置网站各项参数</p>
    </div>

    <div class="flex flex-col lg:flex-row gap-4 lg:gap-6">
      <!-- 标签导航 -->
      <div class="w-full lg:w-56 flex-shrink-0">
        <div class="bg-white rounded-xl border border-gray-100 shadow-sm p-2 overflow-x-auto">
          <button v-for="tab in tabs" :key="tab.id" :class="[
            'w-full text-left px-4 py-2.5 rounded-lg text-sm transition-colors no-wrap',
            activeTab === tab.id
              ? 'bg-blue-50 text-blue-700 font-medium'
              : 'text-gray-600 hover:bg-gray-50'
          ]" @click="activeTab = tab.id">
            {{ tab.name }}
          </button>
        </div>
      </div>

      <!-- 设置表单 -->
      <div class="flex-1 min-w-0">
        <!-- 成功/错误提示 -->
        <div v-if="successMsg" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
          <p class="text-sm text-green-600">{{ successMsg }}</p>
        </div>
        <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
          <p class="text-sm text-red-600">{{ errorMsg }}</p>
        </div>

        <!-- 网站设置 -->
        <div v-show="activeTab === 'site'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">网站信息</h3>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">网站名称</label>
                <input v-model="form.sitename" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">网站标题</label>
                <input v-model="form.title" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">回调地址</label>
                <input v-model="form.localurl" type="text" placeholder="以 http:// 或 https:// 开头，以 / 结尾"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">API地址</label>
                <input v-model="form.apiurl" type="text" placeholder="用户对接地址"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">联系邮箱</label>
                <input v-model="form.email" type="email"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">客服QQ</label>
                <input v-model="form.kfqq" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">开放注册</label>
              <select v-model="form.reg_open"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开放注册</option>
                <option value="0">关闭注册</option>
                <option value="2">仅邀请注册</option>
              </select>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 支付设置 -->
        <div v-show="activeTab === 'pay'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">支付相关配置</h3>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">最小支付金额</label>
                <input v-model="form.pay_min_money" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">最大支付金额</label>
                <input v-model="form.pay_max_money" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">交易手续费率 (%)</label>
                <input v-model="form.pay_fee_rate" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">邀请返现比例 (%)</label>
                <input v-model="form.invite_cashback" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">屏蔽的商品ID</label>
              <input v-model="form.pay_block_goods" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="多个用逗号分隔" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">聚合收款码</label>
              <select v-model="form.qrcode_enabled"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">测试支付</label>
              <select v-model="form.test_open"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
              <p class="text-xs text-gray-400 mt-1">开启后可以使用测试金额进行支付测试</p>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">支付成功页面</label>
                <input v-model="form.pay_success_page" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">支付失败页面</label>
                <input v-model="form.pay_error_page" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 结算设置 -->
        <div v-show="activeTab === 'settle'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">结算规则配置</h3>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">最低结算金额</label>
                <input v-model="form.settle_money" type="number"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">结算周期</label>
                <select v-model="form.settle_cycle"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="0">实时结算</option>
                  <option value="1">每日结算</option>
                  <option value="2">每周结算</option>
                  <option value="3">每月结算</option>
                </select>
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">支付宝结算</label>
                <select v-model="form.settle_alipay"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="1">开启</option>
                  <option value="0">关闭</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">微信结算</label>
                <select v-model="form.settle_wxpay"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="1">开启</option>
                  <option value="0">关闭</option>
                </select>
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">自动转账</label>
                <select v-model="form.settle_auto_transfer"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="1">开启</option>
                  <option value="0">关闭</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">结算手续费率 (%)</label>
                <input v-model="form.settle_fee_rate" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 转账设置 -->
        <div v-show="activeTab === 'transfer'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">企业付款配置</h3>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">转账最低金额</label>
                <input v-model="form.transfer_min" type="number"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">转账最高金额</label>
                <input v-model="form.transfer_max" type="number"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">转账手续费率 (%)</label>
                <input v-model="form.transfer_fee" type="number" step="0.01"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">代付显示名称</label>
                <input v-model="form.transfer_show_name" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">支付宝转账通道</label>
                <select v-model="form.transfer_alipay"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="">请选择</option>
                  <option v-for="ch in transferAlipayChannels" :key="`ali_${ch.id}`" :value="String(ch.id)">
                    #{{ ch.id }} {{ ch.name || ch.plugin_showname || ch.plugin }}
                  </option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">微信转账通道</label>
                <select v-model="form.transfer_wxpay"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="">请选择</option>
                  <option v-for="ch in transferWxpayChannels" :key="`wx_${ch.id}`" :value="String(ch.id)">
                    #{{ ch.id }} {{ ch.name || ch.plugin_showname || ch.plugin }}
                  </option>
                </select>
              </div>
            </div>
            <div class="text-xs text-gray-500 space-y-1">
              <p>提示：这里只能选择支持对应转账能力的已启用通道。</p>
              <p v-if="transferChannelWarn" class="text-rose-600">{{ transferChannelWarn }}</p>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 快捷登录 -->
        <div v-show="activeTab === 'oauth'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">快捷登录配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">支付宝登录</label>
              <select v-model="form.login_alipay"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">QQ登录</label>
              <select v-model="form.login_qq"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">微信登录</label>
              <select v-model="form.login_wx"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 通知设置 -->
        <div v-show="activeTab === 'notice'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">消息提醒配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">订单通知邮箱</label>
              <input v-model="form.notify_email" type="email"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮件通知</label>
              <select v-model="form.email_notify"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">商户下单通知</label>
              <select v-model="form.order_notify"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 实名认证 -->
        <div v-show="activeTab === 'certificate'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">实名认证配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">实名认证必填</label>
              <select v-model="form.certificate_required"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
              <p class="text-xs text-gray-400 mt-1">开启后商户必须完成实名认证才能进行支付</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">认证方式</label>
              <select v-model="form.certificate_types"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">支付宝实人认证</option>
                <option value="1,2">支付宝+腾讯云实人认证</option>
                <option value="1,2,3">全部认证方式</option>
              </select>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- IP类型配置 -->
        <div v-show="activeTab === 'iptype'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">IP获取方式配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">IP获取方式</label>
              <select v-model="form.ip_type"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="0">REMOTE_ADDR</option>
                <option value="1">X_FORWARDED_FOR</option>
                <option value="2">X_REAL_IP</option>
              </select>
              <p class="text-xs text-gray-400 mt-1">根据服务器配置选择正确的IP获取方式</p>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 代理设置 -->
        <div v-show="activeTab === 'proxy'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">代理配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">代理开关</label>
              <select v-model="form.proxy_enabled"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">代理地址</label>
                <input v-model="form.proxy_host" type="text" placeholder="如 127.0.0.1"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">代理端口</label>
                <input v-model="form.proxy_port" type="text" placeholder="如 8080"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">代理账号</label>
                <input v-model="form.proxy_user" type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">代理密码</label>
                <input v-model="form.proxy_pass" type="password"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 邮件设置 -->
        <div v-show="activeTab === 'mail'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">邮件服务器配置</h3>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">SMTP服务器</label>
                <input v-model="form.mail_smtp_host" type="text" placeholder="如 smtp.qq.com"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">SMTP端口</label>
                <input v-model="form.mail_smtp_port" type="text" placeholder="如 587"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮箱账号</label>
              <input v-model="form.mail_username" type="text" placeholder="完整邮箱地址"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮箱密码</label>
              <input v-model="form.mail_password" type="password" placeholder="授权码而非登录密码"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">发件人昵称</label>
              <input v-model="form.mail_from" type="text" placeholder="如 GoPay支付"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 短信设置 -->
        <div v-show="activeTab === 'sms'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">阿里云短信配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">短信开关</label>
              <select v-model="form.sms_enabled"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="1">开启</option>
                <option value="0">关闭</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">AccessKey ID</label>
              <input v-model="form.sms_access_key_id" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="阿里云AccessKey ID" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">AccessKey Secret</label>
              <input v-model="form.sms_access_key_secret" type="password"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="阿里云AccessKey Secret" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">短信签名</label>
              <input v-model="form.sms_sign_name" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如: GoPay支付" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">验证码模板ID</label>
              <input v-model="form.sms_template_code" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如: SMS_xxx" />
              <p class="text-xs text-gray-400 mt-1">模板中变量名为code</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">订单通知模板ID</label>
              <input v-model="form.sms_order_template_code" type="text"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如: SMS_xxx" />
              <p class="text-xs text-gray-400 mt-1">模板中变量名为trade_no和amount</p>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 公告设置 -->
        <div v-show="activeTab === 'gonggao'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">公告内容配置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">公告内容（支持HTML）</label>
              <textarea v-model="form.gonggao_content" rows="6"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                placeholder="输入公告内容，支持HTML格式"></textarea>
              <p class="text-xs text-gray-400 mt-1">将显示在用户中心首页</p>
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handleSave" :disabled="saving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <!-- 账户设置 -->
        <div v-show="activeTab === 'account'" class="bg-white rounded-xl border border-gray-100 shadow-sm p-4 md:p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-6">修改管理员密码</h3>
          <div class="space-y-4 max-w-md">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">原密码</label>
              <input v-model="passwordForm.old_pwd" type="password"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">新密码</label>
              <input v-model="passwordForm.new_pwd" type="password" placeholder="至少8位"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">确认新密码</label>
              <input v-model="passwordForm.confirm_pwd" type="password"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
          <div class="mt-6 pt-4 border-t">
            <button @click="handlePasswordChange" :disabled="passwordSaving"
              class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50">
              {{ passwordSaving ? '修改中...' : '修改密码' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { saveConfig, getConfig, getChannelList } from '@/api/admin'
import { ElMessage } from 'element-plus'

const activeTab = ref('site')

const tabs = [
  { id: 'site', name: '网站设置' },
  { id: 'pay', name: '支付设置' },
  { id: 'settle', name: '结算设置' },
  { id: 'transfer', name: '转账设置' },
  { id: 'oauth', name: '快捷登录' },
  { id: 'notice', name: '通知设置' },
  { id: 'certificate', name: '实名认证' },
  { id: 'iptype', name: 'IP配置' },
  { id: 'proxy', name: '代理设置' },
  { id: 'mail', name: '邮件设置' },
  { id: 'sms', name: '短信设置' },
  { id: 'gonggao', name: '公告设置' },
  { id: 'account', name: '账户设置' }
]

const form = reactive({
  // 网站设置
  sitename: '',
  title: '',
  localurl: '',
  apiurl: '',
  email: '',
  kfqq: '',
  reg_open: '1',
  site_keywords: '',
  site_description: '',
  cdn_url: '',
  user_verification: '0',
  // 支付设置
  test_open: '0',
  pay_success_page: '',
  pay_error_page: '',
  pay_min_money: '1',
  pay_max_money: '100000',
  pay_block_goods: '',
  pay_fee_rate: '0',
  invite_cashback: '0',
  qrcode_enabled: '0',
  // 结算设置
  settle_money: '30',
  settle_cycle: '1',
  settle_alipay: '1',
  settle_wxpay: '1',
  settle_auto_transfer: '0',
  settle_fee_rate: '0',
  // 转账设置
  transfer_min: '1',
  transfer_max: '50000',
  transfer_fee: '0',
  transfer_alipay: '',
  transfer_wxpay: '',
  transfer_show_name: 'GoPay支付',
  // 快捷登录
  login_alipay: '0',
  login_qq: '0',
  login_wx: '0',
  // 通知设置
  notify_email: '',
  email_notify: '0',
  order_notify: '1',
  // 实名认证
  certificate_required: '0',
  certificate_types: '1,2,3',
  // IP类型
  ip_type: '0',
  // 代理设置
  proxy_enabled: '0',
  proxy_host: '',
  proxy_port: '',
  proxy_user: '',
  proxy_pass: '',
  // 邮件设置
  mail_smtp_host: '',
  mail_smtp_port: '587',
  mail_username: '',
  mail_password: '',
  mail_from: '',
  // 短信设置
  sms_enabled: '0',
  sms_access_key_id: '',
  sms_access_key_secret: '',
  sms_sign_name: '',
  sms_template_code: '',
  sms_order_template_code: '',
  // 公告
  gonggao_content: ''
})

const passwordForm = reactive({
  old_pwd: '',
  new_pwd: '',
  confirm_pwd: ''
})

const saving = ref(false)
const passwordSaving = ref(false)
const successMsg = ref('')
const errorMsg = ref('')
type ChannelItem = {
  id: number
  plugin: string
  name?: string
  status?: number
  plugin_showname?: string
  plugin_select?: Record<string, string>
}

const transferAlipayChannels = ref<ChannelItem[]>([])
const transferWxpayChannels = ref<ChannelItem[]>([])
const transferChannelWarn = ref('')

async function loadConfig() {
  try {
    const res = await getConfig()
    if (res.code === 0 && res.data) {
      const data = res.data as Record<string, string>
      // 网站设置
      if (data.sitename) form.sitename = data.sitename
      if (data.title) form.title = data.title
      if (data.localurl) form.localurl = data.localurl
      if (data.apiurl) form.apiurl = data.apiurl
      if (data.email) form.email = data.email
      if (data.kfqq) form.kfqq = data.kfqq
      if (data.reg_open) form.reg_open = data.reg_open
      if (data.site_keywords) form.site_keywords = data.site_keywords
      if (data.site_description) form.site_description = data.site_description
      if (data.cdn_url) form.cdn_url = data.cdn_url
      if (data.user_verification) form.user_verification = data.user_verification
      // 支付设置
      if (data.test_open) form.test_open = data.test_open
      if (data.pay_success_page) form.pay_success_page = data.pay_success_page
      if (data.pay_error_page) form.pay_error_page = data.pay_error_page
      if (data.pay_min_money) form.pay_min_money = data.pay_min_money
      if (data.pay_max_money) form.pay_max_money = data.pay_max_money
      if (data.pay_block_goods) form.pay_block_goods = data.pay_block_goods
      if (data.pay_fee_rate) form.pay_fee_rate = data.pay_fee_rate
      if (data.invite_cashback) form.invite_cashback = data.invite_cashback
      if (data.qrcode_enabled) form.qrcode_enabled = data.qrcode_enabled
      // 结算设置
      if (data.settle_money) form.settle_money = data.settle_money
      if (data.settle_cycle) form.settle_cycle = data.settle_cycle
      if (data.settle_alipay) form.settle_alipay = data.settle_alipay
      if (data.settle_wxpay) form.settle_wxpay = data.settle_wxpay
      if (data.settle_auto_transfer) form.settle_auto_transfer = data.settle_auto_transfer
      if (data.settle_fee_rate) form.settle_fee_rate = data.settle_fee_rate
      // 转账设置
      if (data.transfer_min) form.transfer_min = data.transfer_min
      if (data.transfer_max) form.transfer_max = data.transfer_max
      if (data.transfer_fee) form.transfer_fee = data.transfer_fee
      if (data.transfer_alipay) form.transfer_alipay = data.transfer_alipay
      if (data.transfer_wxpay) form.transfer_wxpay = data.transfer_wxpay
      if (data.transfer_show_name) form.transfer_show_name = data.transfer_show_name
      // 快捷登录
      if (data.login_alipay) form.login_alipay = data.login_alipay
      if (data.login_qq) form.login_qq = data.login_qq
      if (data.login_wx) form.login_wx = data.login_wx
      // 通知设置
      if (data.notify_email) form.notify_email = data.notify_email
      if (data.email_notify) form.email_notify = data.email_notify
      if (data.order_notify) form.order_notify = data.order_notify
      // 实名认证
      if (data.certificate_required) form.certificate_required = data.certificate_required
      if (data.certificate_types) form.certificate_types = data.certificate_types
      // IP类型
      if (data.ip_type) form.ip_type = data.ip_type
      // 代理设置
      if (data.proxy_enabled) form.proxy_enabled = data.proxy_enabled
      if (data.proxy_host) form.proxy_host = data.proxy_host
      if (data.proxy_port) form.proxy_port = data.proxy_port
      if (data.proxy_user) form.proxy_user = data.proxy_user
      if (data.proxy_pass) form.proxy_pass = data.proxy_pass
      // 邮件设置
      if (data.mail_smtp_host) form.mail_smtp_host = data.mail_smtp_host
      if (data.mail_smtp_port) form.mail_smtp_port = data.mail_smtp_port
      if (data.mail_username) form.mail_username = data.mail_username
      if (data.mail_password) form.mail_password = data.mail_password
      if (data.mail_from) form.mail_from = data.mail_from
      // 短信设置
      if (data.sms_enabled) form.sms_enabled = data.sms_enabled
      if (data.sms_access_key_id) form.sms_access_key_id = data.sms_access_key_id
      if (data.sms_access_key_secret) form.sms_access_key_secret = data.sms_access_key_secret
      if (data.sms_sign_name) form.sms_sign_name = data.sms_sign_name
      if (data.sms_template_code) form.sms_template_code = data.sms_template_code
      if (data.sms_order_template_code) form.sms_order_template_code = data.sms_order_template_code
      // 公告
      if (data.gonggao_content) form.gonggao_content = data.gonggao_content
    }
  } catch (error) {
    console.error('加载配置失败:', error)
  }
}

function pluginSupportsTransferType(item: ChannelItem, transferType: 'alipay' | 'wxpay') {
  const plugin = String(item.plugin || '').toLowerCase()
  if (transferType === 'alipay') return plugin === 'alipay'
  if (transferType === 'wxpay') return plugin === 'wxpay'
  return false
}

async function loadTransferChannels() {
  try {
    const res = await getChannelList()
    const list = Array.isArray(res?.data) ? (res.data as ChannelItem[]) : []
    const enabled = list.filter(ch => Number(ch.status || 0) === 1)
    transferAlipayChannels.value = enabled.filter(ch => pluginSupportsTransferType(ch, 'alipay'))
    transferWxpayChannels.value = enabled.filter(ch => pluginSupportsTransferType(ch, 'wxpay'))
  } catch (e) {
    transferAlipayChannels.value = []
    transferWxpayChannels.value = []
  }
}

function validateTransferSelectionLocal() {
  transferChannelWarn.value = ''
  if (activeTab.value !== 'transfer') return true

  if (form.transfer_alipay) {
    const ok = transferAlipayChannels.value.some(ch => String(ch.id) === String(form.transfer_alipay))
    if (!ok) {
      transferChannelWarn.value = '支付宝转账通道配置无效：请选择支持支付宝转账的已启用通道。'
      return false
    }
  }

  if (form.transfer_wxpay) {
    const ok = transferWxpayChannels.value.some(ch => String(ch.id) === String(form.transfer_wxpay))
    if (!ok) {
      transferChannelWarn.value = '微信转账通道配置无效：请选择支持微信转账的已启用通道。'
      return false
    }
  }

  return true
}

async function handleSave() {
  saving.value = true
  successMsg.value = ''
  errorMsg.value = ''

  try {
    if (!validateTransferSelectionLocal()) {
      errorMsg.value = transferChannelWarn.value || '转账通道配置不正确'
      return
    }
    const res = await saveConfig({ mod: activeTab.value, ...form })
    if (res.code === 0) {
      successMsg.value = '保存成功'
      setTimeout(() => { successMsg.value = '' }, 3000)
    } else {
      errorMsg.value = res.msg || '保存失败'
    }
  } catch (error: any) {
    console.error('保存失败:', error)
    errorMsg.value = error.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function handlePasswordChange() {
  if (!passwordForm.old_pwd || !passwordForm.new_pwd || !passwordForm.confirm_pwd) {
    ElMessage.warning('请填写所有密码字段')
    return
  }

  if (passwordForm.new_pwd !== passwordForm.confirm_pwd) {
    ElMessage.error('两次输入的新密码不一致')
    return
  }

  if (passwordForm.new_pwd.length < 8) {
    ElMessage.error('新密码长度至少8位')
    return
  }

  passwordSaving.value = true
  successMsg.value = ''
  errorMsg.value = ''

  try {
    const res = await saveConfig({ mod: 'account', ...passwordForm })
    if (res.code === 0) {
      ElMessage.success('密码修改成功')
      passwordForm.old_pwd = ''
      passwordForm.new_pwd = ''
      passwordForm.confirm_pwd = ''
      if (res.token) {
        sessionStorage.setItem('admin_token', res.token)
      }
    } else {
      ElMessage.error(res.msg || '密码修改失败')
    }
  } catch (error: any) {
    console.error('密码修改失败:', error)
    ElMessage.error(error.message || '密码修改失败')
  } finally {
    passwordSaving.value = false
  }
}

onMounted(() => {
  loadConfig()
  loadTransferChannels()
})
</script>
