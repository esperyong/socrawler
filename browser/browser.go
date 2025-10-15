package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"
)

// StealthBrowser 包装了 headless_browser.Browser 并添加了反检测功能
type StealthBrowser struct {
	*headless_browser.Browser
	rodBrowser *rod.Browser
}

// NewCleanBrowser 创建一个带有强反检测功能的浏览器实例
func NewCleanBrowser(headless bool) *StealthBrowser {
	logrus.Infof("Creating stealth browser (headless: %v)", headless)

	// 创建自定义 launcher，添加更多反检测参数
	l := launcher.New().
		Headless(headless).
		// 关键：禁用自动化控制特征
		Set("disable-blink-features", "AutomationControlled").
		// 设置真实的 User-Agent
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").
		// 禁用各种检测特征
		Set("disable-features", "IsolateOrigins,site-per-process,SitePerProcess").
		// 设置窗口大小
		Set("window-size", "1920,1080").
		// 设置语言
		Set("lang", "en-US,en").
		// 禁用开发者工具
		Set("disable-dev-shm-usage", "").
		// 禁用 GPU 加速（某些环境下有帮助）
		Set("disable-gpu", "").
		// 不使用沙箱（某些环境需要）
		// NoSandbox(true).
		// 禁用扩展
		Set("disable-extensions", "")

	// 启动浏览器
	url := l.MustLaunch()
	logrus.Infof("Browser launched at: %s", url)

	// 创建 rod 浏览器实例
	rodBrowser := rod.New().ControlURL(url).MustConnect()

	// 使用 headless_browser 包装
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	hb := headless_browser.New(
		headless_browser.WithHeadless(headless),
		headless_browser.WithUserAgent(userAgent),
	)

	return &StealthBrowser{
		Browser:    hb,
		rodBrowser: rodBrowser,
	}
}

// NewPage 创建一个新页面并自动应用反检测脚本
func (sb *StealthBrowser) NewPage() *rod.Page {
	// 使用底层 rod browser 创建页面
	page := sb.rodBrowser.MustPage()

	// 在页面加载任何内容之前注入 stealth 脚本
	// 使用 EvalOnNewDocument 确保脚本在页面加载前执行
	page.MustEvalOnNewDocument(stealth.JS)

	// 注入额外的反检测脚本
	page.MustEvalOnNewDocument(`() => {
		// 覆盖 navigator.webdriver
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// 覆盖 Chrome 对象
		window.chrome = {
			runtime: {},
			loadTimes: function() {},
			csi: function() {},
			app: {}
		};

		// 覆盖 permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);

		// 覆盖 plugins - 添加一些常见插件
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				},
				{
					0: {type: "application/pdf", suffixes: "pdf", description: ""},
					description: "",
					filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
					length: 1,
					name: "Chrome PDF Viewer"
				}
			]
		});

		// 覆盖 languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});

		// 覆盖 platform
		Object.defineProperty(navigator, 'platform', {
			get: () => 'MacIntel'
		});

		// 添加 connection 属性
		Object.defineProperty(navigator, 'connection', {
			get: () => ({
				effectiveType: '4g',
				rtt: 50,
				downlink: 10,
				saveData: false
			})
		});

		// 覆盖 hardwareConcurrency
		Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => 8
		});

		// 覆盖 deviceMemory
		Object.defineProperty(navigator, 'deviceMemory', {
			get: () => 8
		});
	}`)

	logrus.Debug("Stealth scripts injected into new page")
	return page
}

// Close 关闭浏览器
func (sb *StealthBrowser) Close() {
	if sb.rodBrowser != nil {
		sb.rodBrowser.MustClose()
	}
	if sb.Browser != nil {
		sb.Browser.Close()
	}
}

// ApplyStealthToPage 对已存在的页面应用隐身模式脚本（向后兼容）
func ApplyStealthToPage(page *rod.Page) error {
	// 注入 stealth.js 脚本
	_, err := page.Eval(stealth.JS)
	if err != nil {
		return err
	}

	// 额外的反检测脚本
	_, err = page.Eval(`() => {
		// 覆盖 navigator.webdriver
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// 覆盖 Chrome 对象
		window.chrome = {
			runtime: {},
			loadTimes: function() {},
			csi: function() {},
			app: {}
		};

		// 覆盖 permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);

		// 覆盖 plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				}
			]
		});

		// 覆盖 languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});
	}`)

	return err
}
