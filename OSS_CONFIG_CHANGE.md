# OSS Configuration Security Update

## 更新日期
2025年10月17日

## 变更概述

为了提高安全性，阿里云 OSS 的访问密钥不再硬编码在代码中，改为通过环境变量配置。

## 主要变更

### 1. 代码变更

#### `sora/oss_uploader.go`
- **移除**: 硬编码的 OSS 密钥常量
- **新增**: `NewOSSConfigFromEnv()` 函数，从环境变量读取配置
- **修改**: `NewOSSUploader()` 函数签名，需要传入 `OSSConfig` 参数
- **新增**: 配置验证，确保所有必需字段都已设置

#### `sora/goldcast_uploader.go`
- **修改**: `NewGoldcastUploader()` 函数签名，需要传入 `OSSConfig` 参数

#### `main.go`
- **修改**: `feed uploadgoldcast` 命令，在创建 uploader 之前从环境变量加载 OSS 配置
- **改进**: 如果 OSS 配置缺失，提供清晰的错误信息

### 2. 环境变量配置

| 环境变量 | 必需性 | 默认值 | 说明 |
|---------|--------|--------|------|
| `OSS_ACCESS_KEY_ID` | **必需** | 无 | 阿里云 Access Key ID |
| `OSS_ACCESS_KEY_SECRET` | **必需** | 无 | 阿里云 Access Key Secret |
| `OSS_BUCKET_NAME` | 可选 | `dreammedias` | OSS Bucket 名称 |
| `OSS_ENDPOINT` | 可选 | `oss-cn-beijing.aliyuncs.com` | OSS 端点地址 |
| `OSS_REGION` | 可选 | `cn-beijing` | OSS 区域 |

## 使用方法

### 方式一：通过命令行参数

```bash
# 最简单的方式：只提供必需的 Key
./socrawler feed uploadgoldcast \
  --oss-access-key-id="your-access-key-id" \
  --oss-access-key-secret="your-access-key-secret"

# 自定义 bucket 和 endpoint
./socrawler feed uploadgoldcast \
  --oss-access-key-id="your-access-key-id" \
  --oss-access-key-secret="your-access-key-secret" \
  --oss-bucket-name="my-bucket" \
  --oss-endpoint="oss-cn-shanghai.aliyuncs.com"
```

### 方式二：通过环境变量

```bash
# 设置必需的环境变量（其他使用默认值）
export OSS_ACCESS_KEY_ID="your-access-key-id"
export OSS_ACCESS_KEY_SECRET="your-access-key-secret"

# 运行上传命令
./socrawler feed uploadgoldcast

# 可选：覆盖默认值
export OSS_BUCKET_NAME="your-bucket-name"
export OSS_ENDPOINT="oss-cn-shanghai.aliyuncs.com"
```

### 方式三：使用 .env 文件

创建 `.env` 文件（参考 `env.example`）：

```bash
# Required
OSS_ACCESS_KEY_ID=your-access-key-id
OSS_ACCESS_KEY_SECRET=your-access-key-secret

# Optional (uncomment to override defaults)
# OSS_BUCKET_NAME=dreammedias
# OSS_ENDPOINT=oss-cn-beijing.aliyuncs.com
# OSS_REGION=cn-beijing
```

加载并运行：

```bash
# 加载环境变量
source .env

# 或使用 env 命令
env $(cat .env | xargs) ./socrawler feed uploadgoldcast
```

### 方式四：在脚本中设置

```bash
#!/bin/bash

# 设置 OSS 配置
export OSS_ACCESS_KEY_ID="your-access-key-id"
export OSS_ACCESS_KEY_SECRET="your-access-key-secret"
export OSS_BUCKET_NAME="your-bucket-name"
export OSS_ENDPOINT="oss-cn-beijing.aliyuncs.com"

# 运行上传
cd /path/to/socrawler
./socrawler feed uploadgoldcast --limit 100
```

## 错误处理

### 缺少 OSS 配置

如果未设置必需的参数，程序会报错并提示：

```
FATAL Failed to load OSS configuration: OSS_ACCESS_KEY_ID environment variable is required

Please provide OSS credentials via:
  1. Command-line flags: --oss-access-key-id and --oss-access-key-secret
  2. Environment variables: OSS_ACCESS_KEY_ID and OSS_ACCESS_KEY_SECRET

Optional (have defaults):
  - OSS_BUCKET_NAME (default: dreammedias)
  - OSS_ENDPOINT (default: oss-cn-beijing.aliyuncs.com)
  - OSS_REGION (default: cn-beijing)
```

### 调试配置问题

启用调试模式查看详细信息：

```bash
./socrawler feed uploadgoldcast --debug
```

## 迁移指南

### 对于现有部署

1. **找到现有的 OSS 配置**
   - 查看之前代码中的 OSS_ACCESS_KEY_ID 和 OSS_ACCESS_KEY_SECRET
   - 或从阿里云控制台获取新的密钥

2. **设置环境变量**
   - 在部署环境中设置相应的环境变量
   - 更新 systemd 服务文件或 Docker 配置

3. **测试配置**
   ```bash
   # 测试上传功能
   ./socrawler feed uploadgoldcast --limit 1 --debug
   ```

### 对于 Docker 部署

更新 `docker-compose.yml`：

```yaml
version: '3'
services:
  socrawler:
    image: socrawler:latest
    environment:
      - OSS_ACCESS_KEY_ID=${OSS_ACCESS_KEY_ID}
      - OSS_ACCESS_KEY_SECRET=${OSS_ACCESS_KEY_SECRET}
      - OSS_BUCKET_NAME=${OSS_BUCKET_NAME}
      - OSS_ENDPOINT=${OSS_ENDPOINT}
      - OSS_REGION=cn-beijing
    env_file:
      - .env
```

### 对于 Systemd 服务

更新服务文件 `/etc/systemd/system/socrawler.service`：

```ini
[Unit]
Description=Socrawler Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/socrawler
Environment="OSS_ACCESS_KEY_ID=your-access-key-id"
Environment="OSS_ACCESS_KEY_SECRET=your-access-key-secret"
Environment="OSS_BUCKET_NAME=your-bucket-name"
Environment="OSS_ENDPOINT=oss-cn-beijing.aliyuncs.com"
ExecStart=/path/to/socrawler/socrawler feed uploadgoldcast
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

或使用 EnvironmentFile：

```ini
[Service]
EnvironmentFile=/path/to/socrawler/.env
ExecStart=/path/to/socrawler/socrawler feed uploadgoldcast
```

### 对于 Cron 任务

更新 crontab：

```bash
# 编辑 crontab
crontab -e

# 添加环境变量和任务
OSS_ACCESS_KEY_ID=your-access-key-id
OSS_ACCESS_KEY_SECRET=your-access-key-secret
OSS_BUCKET_NAME=your-bucket-name
OSS_ENDPOINT=oss-cn-beijing.aliyuncs.com

30 3 * * * cd /path/to/socrawler && ./socrawler feed uploadgoldcast
```

或创建包装脚本：

```bash
#!/bin/bash
# /path/to/socrawler/upload_cron.sh

# 加载环境变量
source /path/to/socrawler/.env

# 执行上传
cd /path/to/socrawler
./socrawler feed uploadgoldcast --limit 100
```

然后在 crontab 中调用脚本：

```cron
30 3 * * * /path/to/socrawler/upload_cron.sh >> /var/log/socrawler-upload.log 2>&1
```

## 安全最佳实践

1. **不要提交 .env 文件到 Git**
   ```bash
   # 添加到 .gitignore
   echo ".env" >> .gitignore
   ```

2. **限制文件权限**
   ```bash
   chmod 600 .env
   ```

3. **使用 RAM 角色（推荐）**
   - 在阿里云 ECS 上，可以使用 RAM 角色代替 AccessKey
   - 这样无需在代码或配置文件中存储密钥

4. **定期轮换密钥**
   - 定期更新 AccessKey
   - 删除不再使用的旧密钥

5. **最小权限原则**
   - 确保 AccessKey 只有上传 OSS 所需的最小权限
   - 不要使用主账号的 AccessKey

## 影响范围

### 受影响的功能
- ✅ `feed uploadgoldcast` 命令 - 上传视频到 Goldcast 时需要先上传到 OSS

### 不受影响的功能
- ✅ `feed fetch` - 获取 feed 数据
- ✅ `feed download` - 下载视频到本地
- ✅ `feed sync` - 同步 feed 和下载
- ✅ `feed export` - 导出数据库
- ✅ `runserver` - HTTP 服务器

## 回滚方案

如果需要回滚到旧版本（不推荐）：

```bash
# 检出之前的提交
git log --oneline  # 找到变更前的 commit
git checkout <commit-hash>

# 重新构建
go build -o socrawler .
```

## 验证清单

上线前请确认：

- [ ] 所有必需的环境变量已设置
- [ ] OSS 密钥有效且权限正确
- [ ] 测试上传功能正常工作
- [ ] 更新了部署文档
- [ ] 通知了相关运维人员
- [ ] 备份了旧配置（如需要）
- [ ] 从代码库中移除了硬编码的密钥

## 相关文档

- [GOLDCAST_INTEGRATION.md](./GOLDCAST_INTEGRATION.md) - Goldcast 集成文档
- [FEED_DOWNLOADER.md](./FEED_DOWNLOADER.md) - Feed 下载器使用文档
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南

## 技术支持

如有问题，请参考：

1. 启用调试模式查看详细日志
2. 检查环境变量是否正确设置
3. 验证 OSS 密钥和权限
4. 查看相关文档

## 版本信息

- **变更版本**: v1.1.0
- **变更日期**: 2025年10月17日
- **变更类型**: 安全改进（破坏性变更）
- **影响**: 需要更新部署配置

