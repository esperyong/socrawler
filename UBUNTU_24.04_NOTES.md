# Ubuntu 24.04 (Noble Numbat) 兼容性说明

## 重要变化

Ubuntu 24.04 引入了一些包命名变化，主要是为了支持 64 位时间戳（time64）。这影响了 Chrome 浏览器的依赖包。

## 包名变化对照表

| Ubuntu 22.04 及更早 | Ubuntu 24.04+ | 说明 |
|-------------------|--------------|------|
| `libasound2` | `libasound2t64` | ALSA 音频库 |
| `libatk-bridge2.0-0` | `libatk-bridge2.0-0t64` | ATK 桥接库 |
| `libatk1.0-0` | `libatk1.0-0t64` | ATK 工具包 |
| `libcups2` | `libcups2t64` | CUPS 打印系统 |

**注意：** `t64` 后缀表示支持 64 位时间戳（time64），这是为了解决 2038 年问题。

## 自动检测和安装

我们的 `setup_browser_ubuntu.sh` 脚本会自动检测 Ubuntu 版本并使用正确的包名：

```bash
# 脚本会自动检测版本
./setup_browser_ubuntu.sh
```

### 手动安装（Ubuntu 24.04）

如果你需要手动安装，使用以下命令：

```bash
# 安装依赖（Ubuntu 24.04+）
sudo apt-get install -y \
    wget gnupg ca-certificates \
    fonts-liberation libappindicator3-1 \
    libasound2t64 libatk-bridge2.0-0t64 \
    libatk1.0-0t64 libcups2t64 \
    libdbus-1-3 libgdk-pixbuf2.0-0 \
    libnspr4 libnss3 libx11-xcb1 \
    libxcomposite1 libxdamage1 libxrandr2 \
    xdg-utils libgbm1 libxkbcommon0 \
    libpango-1.0-0 libcairo2

# 添加 Chrome 仓库（新方法，apt-key 已废弃）
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | \
    sudo gpg --dearmor -o /usr/share/keyrings/google-chrome-keyring.gpg

echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome-keyring.gpg] http://dl.google.com/linux/chrome/deb/ stable main" | \
    sudo tee /etc/apt/sources.list.d/google-chrome.list

# 安装 Chrome
sudo apt-get update
sudo apt-get install -y google-chrome-stable
```

### 手动安装（Ubuntu 22.04）

```bash
# 安装依赖（Ubuntu 22.04）
sudo apt-get install -y \
    wget gnupg ca-certificates \
    fonts-liberation libappindicator3-1 \
    libasound2 libatk-bridge2.0-0 \
    libatk1.0-0 libcups2 \
    libdbus-1-3 libgdk-pixbuf2.0-0 \
    libnspr4 libnss3 libx11-xcb1 \
    libxcomposite1 libxdamage1 libxrandr2 \
    xdg-utils libgbm1 libxkbcommon0 \
    libpango-1.0-0 libcairo2

# 添加 Chrome 仓库（旧方法仍然可用）
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | \
    sudo tee /etc/apt/sources.list.d/google-chrome.list

# 安装 Chrome
sudo apt-get update
sudo apt-get install -y google-chrome-stable
```

## apt-key 废弃说明

Ubuntu 24.04 中 `apt-key` 命令已被废弃。新的推荐方法是：

**旧方法（已废弃）：**
```bash
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list
```

**新方法（推荐）：**
```bash
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | \
    sudo gpg --dearmor -o /usr/share/keyrings/google-chrome-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome-keyring.gpg] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list
```

## Docker 镜像

我们的 Dockerfile 已更新为使用 Ubuntu 24.04 作为基础镜像，并使用正确的包名。

## 常见问题

### Q: 为什么会出现 "Package 'libasound2' has no installation candidate" 错误？

**A:** 这是因为在 Ubuntu 24.04 中，包名已更改为 `libasound2t64`。使用我们更新后的脚本会自动处理这个问题。

### Q: 我的系统是 Ubuntu 22.04，脚本还能用吗？

**A:** 可以！脚本会自动检测 Ubuntu 版本并使用正确的包名。Ubuntu 22.04 及更早版本会使用不带 `t64` 后缀的包名。

### Q: 如何检查我的 Ubuntu 版本？

**A:** 运行以下命令：
```bash
lsb_release -a
# 或
cat /etc/os-release
```

### Q: 我可以在 Ubuntu 20.04 上使用吗？

**A:** 可以，但建议升级到 Ubuntu 22.04 LTS 或 24.04 LTS 以获得更好的支持和安全更新。

## 验证安装

安装完成后，验证 Chrome 是否正常工作：

```bash
# 检查 Chrome 版本
google-chrome --version

# 测试无头模式
google-chrome --headless --disable-gpu --dump-dom https://www.google.com

# 检查依赖
ldd /usr/bin/google-chrome | grep "not found"
```

如果最后一个命令没有输出，说明所有依赖都已正确安装。

## 参考资源

- [Ubuntu 24.04 Release Notes](https://discourse.ubuntu.com/t/noble-numbat-release-notes/39890)
- [Time64 Migration](https://wiki.ubuntu.com/Time64Migration)
- [Chrome on Linux](https://www.google.com/chrome/browser/desktop/index.html?platform=linux)

---

**最后更新：** 2025年10月

