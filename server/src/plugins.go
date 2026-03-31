package main

import (
	// 导入所有内置插件，触发 init() 注册
	_ "paygo/src/plugin/channels/alipay"
	_ "paygo/src/plugin/channels/wxpay"
)
