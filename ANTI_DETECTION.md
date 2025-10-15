# Anti-Detection Implementation

## 问题描述

当执行爬虫时，sora.chatgpt.com 会显示 Cloudflare 验证页面（"Verifying you are human. This may take a few seconds."），阻止自动化访问。

## 解决方案

实现了多层反检测机制来绕过 Cloudflare 和其他反爬虫保护。

## 技术实现

### 1. Stealth.js 注入

使用 `go-rod/stealth` 库的 JavaScript 脚本来隐藏浏览器自动化特征：

```go
// browser/browser.go
_, err := page.Eval(stealth.JS)
```

这个脚本会：
- 隐藏 WebDriver 标识
- 伪造浏览器指纹
- 模拟真实浏览器的各种属性

### 2. 自定义反检测脚本

额外注入 JavaScript 来覆盖常见的检测点：

```javascript
// 覆盖 navigator.webdriver
Object.defineProperty(navigator, 'webdriver', {
    get: () => undefined
});

// 覆盖 Chrome 对象
window.chrome = {
    runtime: {}
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
    get: () => [1, 2, 3, 4, 5]
});

// 覆盖 languages
Object.defineProperty(navigator, 'languages', {
    get: () => ['en-US', 'en']
});
```

### 3. 真实的 User-Agent

设置真实的 Chrome User-Agent：

```go
userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
```

### 4. 应用时机

在导航到目标页面之前应用反检测脚本：

```go
// sora/crawler.go
// 应用反检测脚本
logrus.Info("Applying stealth mode to page...")
if err := browser.ApplyStealthToPage(page); err != nil {
    logrus.Warnf("Failed to apply stealth mode: %v", err)
}

// 导航到 Sora 页面
logrus.Info("Navigating to Sora page...")
if err := page.Navigate(SoraURL); err != nil {
    return nil, errors.Wrap(err, "failed to navigate to Sora page")
}
```

## 修改的文件

### 1. `browser/browser.go`

**改动：**
- 添加了 `ApplyStealthToPage()` 函数
- 注入 stealth.js 和自定义反检测脚本
- 设置真实的 User-Agent

**新增导入：**
```go
"github.com/go-rod/stealth"
```

### 2. `sora/crawler.go`

**改动：**
- 在导航前调用 `browser.ApplyStealthToPage(page)`
- 添加日志记录

**新增导入：**
```go
"github.com/esperyong/socrawler/browser"
```

### 3. `README.md`

**改动：**
- 添加了 "Anti-Detection & Cloudflare Bypass" 章节
- 更新了功能列表
- 添加了故障排除指南

## 使用方法

### 基本使用

1. 启动服务器（推荐先用非 headless 模式测试）：
```bash
go run . runserver --headless=false --debug
```

2. 发送爬取请求：
```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 10,
    "save_path": "./downloads/sora"
  }'
```

3. 观察浏览器窗口 - 应该能自动通过 Cloudflare 验证

### 验证反检测是否生效

检查日志输出：
```
level=info msg="Applying stealth mode to page..."
level=info msg="Navigating to Sora page..."
level=info msg="Page loaded, waiting for initial content..."
level=info msg="Page title: Sora, URL: https://sora.chatgpt.com/"
```

如果看到这些日志且页面标题是 "Sora"，说明成功绕过了 Cloudflare。

## 高级技巧

### 1. 非 Headless 模式

某些网站能检测 headless 模式，建议先用非 headless 模式测试：

```bash
go run . runserver --headless=false
```

### 2. 增加延迟

让行为更像真实用户：

```json
{
  "total_duration_seconds": 120,
  "scroll_interval_seconds": 15,
  "save_path": "./downloads/sora"
}
```

### 3. 使用代理

如果 IP 被封禁，可以考虑使用代理服务（需要额外配置）。

## 技术原理

### Cloudflare 检测机制

Cloudflare 通过多种方式检测自动化：

1. **JavaScript 挑战** - 执行复杂的 JS 代码验证浏览器环境
2. **浏览器指纹** - 检查 navigator、screen、plugins 等属性
3. **WebDriver 检测** - 检查 `navigator.webdriver` 属性
4. **行为分析** - 分析鼠标移动、键盘输入等行为模式
5. **TLS 指纹** - 分析 TLS 握手特征

### 我们的对策

1. **Stealth.js** - 修改浏览器指纹，使其看起来像真实浏览器
2. **JavaScript 覆盖** - 隐藏 WebDriver 和其他自动化标识
3. **真实 User-Agent** - 使用最新的 Chrome User-Agent
4. **自然行为** - 平滑滚动、合理的时间间隔

## 限制和注意事项

1. **不是 100% 保证** - Cloudflare 不断更新检测机制
2. **IP 限制** - 如果 IP 被封禁，反检测也无法绕过
3. **账号登录** - 如果网站需要登录，需要额外处理 cookies
4. **合法性** - 请确保你的爬虫行为符合网站的服务条款

## 故障排除

### 仍然看到 Cloudflare 验证页面

1. 检查截图：`./downloads/sora/debug_initial_page.png`
2. 查看日志中的页面标题
3. 尝试非 headless 模式
4. 增加等待时间
5. 检查 IP 是否被封禁

### 日志显示 "Failed to apply stealth mode"

这通常不是致命错误，但说明部分反检测脚本可能没有成功注入。检查：
1. go-rod/stealth 库是否正确安装
2. 浏览器版本是否兼容

## 未来改进

可能的增强方向：

1. **代理池** - 轮换 IP 地址
2. **Cookie 管理** - 保存和重用 Cloudflare cookies
3. **更多反检测技术** - Canvas 指纹、WebGL 指纹等
4. **人工智能行为模拟** - 更自然的鼠标移动和滚动模式
5. **验证码识别** - 自动处理 CAPTCHA（如果出现）

## 参考资料

- [go-rod/stealth](https://github.com/go-rod/stealth) - Rod 浏览器的隐身模式库
- [Puppeteer Extra Stealth Plugin](https://github.com/berstend/puppeteer-extra/tree/master/packages/puppeteer-extra-plugin-stealth) - 类似的 Node.js 实现
- [Cloudflare Bot Management](https://www.cloudflare.com/products/bot-management/) - Cloudflare 的反爬虫技术

