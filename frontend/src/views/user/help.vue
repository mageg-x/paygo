<template>
  <div class="space-y-6">
    <div class="rounded-2xl bg-gradient-to-r from-blue-600 to-indigo-600 text-white p-6 shadow-lg">
      <h2 class="text-2xl font-bold">商户接入文档</h2>
      <p class="text-blue-100 mt-2">OpenAPI 接入流程、接口参数、签名规则与完整示例代码</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
      <div class="lg:col-span-1">
        <div class="card sticky top-4">
          <div class="card-body p-3">
            <div class="text-xs text-gray-500 mb-2 px-2">目录</div>
            <a
              v-for="s in sections"
              :key="s.id"
              :href="`#${s.id}`"
              class="block px-3 py-2 rounded-lg text-sm text-gray-700 hover:bg-blue-50 hover:text-blue-700 transition-colors"
            >
              {{ s.title }}
            </a>
          </div>
        </div>
      </div>

      <div class="lg:col-span-3 space-y-6">
        <section id="prepare" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">1. 接入准备</h3>
            <div class="text-sm text-gray-700 leading-7">
              <div>1. 登录商户后台，在「资料管理 / API信息」获取 `pid`（商户ID）和 `api key`（商户密钥）。</div>
              <div>2. 你的业务服务端负责签名并请求 OpenAPI，不要在前端页面暴露 `api key`。</div>
              <div>3. 配置公网可访问的 `notify_url`，用于接收支付成功回调。</div>
              <div>4. 回调处理完成后必须返回包含 `success` 的字符串，否则平台会判定失败并进入重试。</div>
            </div>
            <div class="text-xs text-gray-600 bg-blue-50 border border-blue-200 rounded-lg px-3 py-2 leading-6">
              <div>接口基地址示例：`http://localhost:3000`（本地）或 `https://你的平台域名`（生产）。</div>
              <div>`/api/pay/create` 仅支持 `POST JSON`；`query/refund` 使用 `GET` 或 `POST form`。</div>
            </div>
          </div>
        </section>

        <section id="apis" class="card">
          <div class="card-body space-y-3">
            <h3 class="text-lg font-semibold text-gray-900">2. 接口清单</h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">接口</th>
                    <th class="py-2 pr-3">方法</th>
                    <th class="py-2 pr-3">说明</th>
                    <th class="py-2 pr-3">返回重点</th>
                    <th class="py-2">签名</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in apiRows" :key="row.path" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ row.path }}</td>
                    <td class="py-2 pr-3">{{ row.method }}</td>
                    <td class="py-2 pr-3">{{ row.desc }}</td>
                    <td class="py-2 pr-3">{{ row.ret }}</td>
                    <td class="py-2">{{ row.sign }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section id="sign" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">3. 签名规则（MD5）</h3>
            <div class="text-sm text-gray-700 leading-7">
              <div>签名算法：`MD5`（小写）。</div>
              <div>签名步骤：</div>
              <div class="pl-4">1. 参数按 `key` 字典序排序。</div>
              <div class="pl-4">2. 跳过空值参数与 `sign/sign_type`。</div>
              <div class="pl-4">3. 拼接：`k1=v1&k2=v2...&key=API_KEY`。</div>
              <div class="pl-4">4. 对拼接串做 MD5 并转小写，得到 `sign`。</div>
            </div>

            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>function makeSign(params, apiKey) {
  const keys = Object.keys(params).sort()
  const pairs = []
  for (const k of keys) {
    if (k === 'sign' || k === 'sign_type') continue
    const v = params[k]
    if (v === undefined || v === null || v === '') continue
    pairs.push(`${k}=${v}`)
  }
  pairs.push(`key=${apiKey}`)
  return md5(pairs.join('&')).toLowerCase()
}</code></pre>
            </div>

            <div class="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
              注意：`money` 参与签名时请使用普通数值格式（如 `1`、`1.23`），不要强制固定为 `1.00`，否则会导致签名不一致。
            </div>
          </div>
        </section>

        <section id="flow" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">4. 最小接入流程</h3>
            <div class="text-sm text-gray-700 leading-7">
              <div>1. 服务端读取商户 `pid` 与 `api key`。</div>
              <div>2. 服务端组织 `/api/pay/create` 参数并生成 `sign`。</div>
              <div>3. 调用下单接口，获得 `trade_no` 与支付数据（HTML/URL/二维码）。</div>
              <div>4. 前端拉起支付页并等待支付完成。</div>
              <div>5. 以 `notify_url` 异步回调为支付成功的最终依据；可辅以 `/api/pay/query` 主动查单。</div>
            </div>
          </div>
        </section>

        <section id="create" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">5. 创建订单：/api/pay/create</h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">参数</th>
                    <th class="py-2 pr-3">必填</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in createParams" :key="p.name" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ p.name }}</td>
                    <td class="py-2 pr-3">{{ p.required }}</td>
                    <td class="py-2">{{ p.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>// 请求体示例（POST JSON）
{
  "pid": 1,
  "type": 2,
  "out_trade_no": "ORD_20260417190001",
  "name": "测试商品",
  "money": 1,
  "notify_url": "https://merchant.example.com/pay/notify",
  "return_url": "https://merchant.example.com/pay/return",
  "clientip": "127.0.0.1",
  "device": "mobile",
  "param": "biz_param",
  "sign_type": "MD5",
  "sign": "md5签名结果"
}</code></pre>
            </div>

            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>// Node.js 服务端示例（axios + crypto）
import axios from 'axios'
import crypto from 'crypto'

function sign(params, apiKey) {
  const data = Object.keys(params)
    .sort()
    .filter(k => !['sign', 'sign_type'].includes(k) && params[k] !== '' && params[k] !== undefined && params[k] !== null)
    .map(k => `${k}=${params[k]}`)
    .join('&') + `&key=${apiKey}`
  return crypto.createHash('md5').update(data, 'utf8').digest('hex')
}

const base = 'http://localhost:3000'
const pid = 1
const apiKey = 'YOUR_API_KEY'
const req = {
  pid,
  type: 2,
  out_trade_no: `ORD_${Date.now()}`,
  name: '测试商品',
  money: 1,
  notify_url: 'https://merchant.example.com/pay/notify',
  return_url: 'https://merchant.example.com/pay/return',
  clientip: '127.0.0.1',
  device: 'mobile',
  param: 'biz_param'
}

req.sign = sign(req, apiKey)
req.sign_type = 'MD5'

const res = await axios.post(`${base}/api/pay/create`, req)
console.log(res.data)

// 按支付类型处理
// pay_type=html: 直接渲染 pay_data(form html) 并自动提交
// pay_type=qrcode: 展示 pay_data/pay_info 为二维码
// pay_type=jump/page/url: 直接跳转 pay_info 或 result.URL</code></pre>
            </div>

            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>// 成功响应示例
{
  "code": 0,
  "trade_no": "20260418120000123456",
  "pay_type": "html",
  "pay_info": "",
  "pay_data": "&lt;form ...&gt;&lt;/form&gt;",
  "result": {
    "Type": "html",
    "URL": "",
    "Page": "",
    "Data": "&lt;form ...&gt;&lt;/form&gt;"
  },
  "timestamp": 1776500000
}</code></pre>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">字段</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="r in createResponseRows" :key="r.name" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ r.name }}</td>
                    <td class="py-2">{{ r.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section id="query" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">6. 查询订单：/api/pay/query</h3>
            <div class="text-sm text-gray-700">
              方式：`GET` 或 `POST`，至少传 `trade_no` 或 `out_trade_no` 之一，并带签名。
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">参数</th>
                    <th class="py-2 pr-3">必填</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in queryParams" :key="p.name" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ p.name }}</td>
                    <td class="py-2 pr-3">{{ p.required }}</td>
                    <td class="py-2">{{ p.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>GET /api/pay/query?pid=1&trade_no=20260418120000123456&sign=xxx&sign_type=MD5

// 或 POST form:
// pid=1&out_trade_no=ORD_20260417190001&sign=xxx&sign_type=MD5</code></pre>
            </div>
            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>// 查询响应示例
{
  "code": 0,
  "trade_no": "20260418120000123456",
  "out_trade_no": "ORD_20260417190001",
  "type": "支付宝",
  "pid": 1,
  "name": "测试商品",
  "money": 1,
  "status": 1,
  "buyer": "",
  "addtime": "2026-04-18 12:00:00",
  "endtime": "2026-04-18 12:01:05"
}</code></pre>
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">status</th>
                    <th class="py-2">含义</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="s in statusRows" :key="s.status" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ s.status }}</td>
                    <td class="py-2">{{ s.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section id="refund" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">7. 退款：/api/pay/refund</h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">参数</th>
                    <th class="py-2 pr-3">必填</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in refundParams" :key="p.name" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ p.name }}</td>
                    <td class="py-2 pr-3">{{ p.required }}</td>
                    <td class="py-2">{{ p.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>POST /api/pay/refund
Content-Type: application/x-www-form-urlencoded

pid=1
trade_no=20260418120000123456
money=0.5
sign_type=MD5
sign=md5签名结果</code></pre>
            </div>
            <div class="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
              说明：`money` 支持部分退款；请以你提交的退款金额参与签名。
            </div>
          </div>
        </section>

        <section id="notify" class="card">
          <div class="card-body space-y-4">
            <h3 class="text-lg font-semibold text-gray-900">8. 异步回调说明（notify_url）</h3>
            <div class="text-sm text-gray-700 leading-7">
              <div>支付成功后，平台会向你下单时传入的 `notify_url` 发起 `POST form`。</div>
              <div>你需要验签、判断业务是否已处理（幂等），最后返回 `success`。</div>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">参数</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in notifyParams" :key="p.name" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ p.name }}</td>
                    <td class="py-2">{{ p.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="bg-gray-900 text-gray-100 rounded-xl p-4 overflow-x-auto text-xs">
              <pre class="whitespace-pre-wrap"><code v-pre>// Express 回调示例（application/x-www-form-urlencoded）
app.post('/pay/notify', async (req, res) => {
  const body = { ...req.body }
  const remoteSign = String(body.sign || '').toLowerCase()
  delete body.sign
  delete body.sign_type

  const localSign = makeSign(body, process.env.API_KEY)
  if (localSign !== remoteSign) return res.send('fail')

  // 幂等处理：trade_no 已处理则直接 success
  // 更新你的业务订单状态...

  return res.send('success')
})</code></pre>
            </div>
          </div>
        </section>

        <section id="codes" class="card">
          <div class="card-body space-y-3">
            <h3 class="text-lg font-semibold text-gray-900">9. 常见错误码</h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="text-left text-gray-500 border-b">
                    <th class="py-2 pr-3">返回</th>
                    <th class="py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="e in errorRows" :key="e.code" class="border-b last:border-0">
                    <td class="py-2 pr-3 font-mono text-xs">{{ e.code }}</td>
                    <td class="py-2">{{ e.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const sections = [
  { id: 'prepare', title: '1. 接入准备' },
  { id: 'apis', title: '2. 接口清单' },
  { id: 'sign', title: '3. 签名规则' },
  { id: 'flow', title: '4. 接入流程' },
  { id: 'create', title: '5. 创建订单' },
  { id: 'query', title: '6. 查询订单' },
  { id: 'refund', title: '7. 发起退款' },
  { id: 'notify', title: '8. 异步回调' },
  { id: 'codes', title: '9. 错误码' }
]

const apiRows = [
  { path: '/api/pay/create', method: 'POST JSON', desc: '创建支付订单', ret: 'trade_no + pay_data', sign: '需要' },
  { path: '/api/pay/query', method: 'GET/POST', desc: '查询订单状态', ret: 'status/money/type', sign: '需要' },
  { path: '/api/pay/refund', method: 'POST form', desc: '订单退款', ret: '退款结果', sign: '需要' },
  { path: '/api/pay/types', method: 'GET', desc: '获取商户可用支付类型', ret: '类型列表', sign: '不需要' },
  { path: '/api/pay/channels', method: 'GET', desc: '按支付类型获取可用通道（pid,type）', ret: '通道列表', sign: '不需要' }
]

const createParams = [
  { name: 'pid', required: '是', desc: '商户ID' },
  { name: 'type', required: '是', desc: '支付类型ID' },
  { name: 'out_trade_no', required: '是', desc: '商户订单号（唯一）' },
  { name: 'name', required: '是', desc: '商品名称' },
  { name: 'money', required: '是', desc: '订单金额（数字）' },
  { name: 'notify_url', required: '建议', desc: '支付异步回调地址' },
  { name: 'return_url', required: '否', desc: '同步返回地址' },
  { name: 'clientip', required: '否', desc: '用户IP，不传则平台自动获取' },
  { name: 'device', required: '否', desc: '设备类型：pc/mobile，不传自动识别UA' },
  { name: 'param', required: '否', desc: '透传参数' },
  { name: 'sign_type', required: '否', desc: '固定 MD5（默认 MD5）' },
  { name: 'sign', required: '是', desc: '签名值' }
]

const createResponseRows = [
  { name: 'trade_no', desc: '平台订单号，用于后续查单和退款' },
  { name: 'pay_type', desc: '支付拉起类型，如 html/qrcode/jump/page' },
  { name: 'pay_info', desc: '当为跳转型支付时可直接跳转的URL' },
  { name: 'pay_data', desc: '当为 html 或 qrcode 时的表单HTML或二维码内容' },
  { name: 'result', desc: '插件原始结构：Type/URL/Page/Data/Msg' }
]

const queryParams = [
  { name: 'pid', required: '是', desc: '商户ID' },
  { name: 'trade_no', required: '二选一', desc: '平台订单号' },
  { name: 'out_trade_no', required: '二选一', desc: '商户订单号' },
  { name: 'sign_type', required: '否', desc: 'MD5（默认 MD5）' },
  { name: 'sign', required: '是', desc: '签名值' }
]

const statusRows = [
  { status: '0', desc: '待支付' },
  { status: '1', desc: '已支付' },
  { status: '2', desc: '已退款' },
  { status: '3', desc: '已冻结' },
  { status: '4', desc: '预授权' }
]

const refundParams = [
  { name: 'pid', required: '是', desc: '商户ID' },
  { name: 'trade_no', required: '是', desc: '平台订单号' },
  { name: 'money', required: '是', desc: '退款金额' },
  { name: 'sign_type', required: '否', desc: 'MD5（默认 MD5）' },
  { name: 'sign', required: '是', desc: '签名值' }
]

const notifyParams = [
  { name: 'trade_no', desc: '平台订单号' },
  { name: 'out_trade_no', desc: '商户订单号' },
  { name: 'type', desc: '支付类型ID' },
  { name: 'status', desc: '订单状态，成功回调固定为 1' },
  { name: 'money', desc: '订单金额' },
  { name: 'realmoney', desc: '平台实收金额' },
  { name: 'sign', desc: '回调签名（使用商户 API Key 验签）' }
]

const errorRows = [
  { code: 'code=1,msg=签名错误', desc: '签名算法/参数顺序/金额格式不一致' },
  { code: 'code=1,msg=sign_type不支持', desc: '仅支持 MD5' },
  { code: 'code=1,msg=签名不能为空', desc: '缺少 sign 参数' },
  { code: 'code=1,msg=商户不存在', desc: 'pid 无效或未找到商户' },
  { code: 'code=1,msg=参数错误', desc: '必填参数缺失或格式错误' },
  { code: 'code=1,msg=订单不存在', desc: '订单号不存在或不属于当前商户' },
  { code: 'code=1,msg=商户已被禁用', desc: '商户状态异常' },
  { code: 'code=1,msg=域名未授权', desc: '触发域名授权规则（参数中携带域名时）' }
]
</script>
