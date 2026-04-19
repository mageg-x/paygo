package main

import (
	// 导入所有内置插件，触发 init() 注册
	_ "gopay/src/plugin/channels/alipay"
	_ "gopay/src/plugin/channels/wxpay"
)
