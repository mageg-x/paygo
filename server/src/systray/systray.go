//go:build gui

package systray

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/getlantern/systray"
)

var (
	appName = "GoPay"
	appTip  = "GoPay 支付服务运行中"
	openURL string
	quitFn  func()
)

func Init(name, tooltip, homepage string, onQuit func()) {
	if strings.TrimSpace(name) != "" {
		appName = strings.TrimSpace(name)
	}
	if strings.TrimSpace(tooltip) != "" {
		appTip = strings.TrimSpace(tooltip)
	}
	if strings.TrimSpace(homepage) != "" {
		openURL = strings.TrimSpace(homepage)
	}
	quitFn = onQuit
}

func Stop() {
	systray.Quit()
}

func Run(onReady func()) {
	systray.Run(func() {
		setup()
		if onReady != nil {
			onReady()
		}
	}, func() {})
}

func setup() {
	systray.SetIcon(getIcon())
	systray.SetTitle(appName)
	systray.SetTooltip(appTip)

	systray.AddMenuItem("● 服务运行中", "当前状态")
	systray.AddSeparator()
	mOpen := systray.AddMenuItem("打开首页", "在浏览器中打开首页")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出程序")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				if openURL != "" {
					_ = openBrowser(openURL)
				}
			case <-mQuit.ClickedCh:
				if quitFn != nil {
					quitFn()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
