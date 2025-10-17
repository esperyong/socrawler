# Feed-based Sora Video Downloader

## 概述

新的 feed 下载器通过访问 Sora 的公共 feed 端点来下载视频，相比之前的滚动爬取方式更加高效和稳定。

### 主要特性

- **基于 Feed API**: 直接从 `https://sora.chatgpt.com/backend/public/nf2/feed` 获取视频列表
- **自动去重**: 使用 SQLite 数据库跟踪已下载的视频，避免重复下载
- **元数据存储**: 保存视频的作者、描述、时间戳等信息
- **浏览器模拟**: 使用 stealth 模式绕过反爬虫检测
- **易于调度**: 可通过 cron 或其他调度器定期运行

## 快速开始

### 1. 构建项目

```bash
./run_service.sh build
```

### 2. 基本使用

```bash
# 下载最新的 50 个视频（默认）
./run_service.sh feed

# 下载指定数量的视频
./run_service.sh feed --limit=100

# 非 headless 模式（用于调试）
./run_service.sh feed --headless=false --limit=10
```

### 3. 高级用法

```bash
# 直接使用命令行
./socrawler feed --save-path ./downloads/sora --db-path ./sora.db --limit 50

# 自定义所有参数
./socrawler feed \
  --save-path ./my_videos \
  --db-path ./my_database.db \
  --limit 200 \
  --headless=true \
  --debug
```

## 命令参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--save-path` | `./downloads/sora` | 视频保存目录 |
| `--db-path` | `./sora.db` | SQLite 数据库路径 |
| `--limit` | `50` | 单次运行最大下载数量 |
| `--headless` | `true` | 是否使用无头浏览器 |
| `--debug` | `false` | 是否启用调试日志 |

## 数据库结构

SQLite 数据库包含以下字段：

```sql
CREATE TABLE sora_videos (
    post_id TEXT PRIMARY KEY,           -- 帖子唯一 ID
    generation_id TEXT,                 -- 视频生成 ID
    video_url TEXT,                     -- 视频下载 URL
    thumbnail_url TEXT,                 -- 缩略图 URL
    text TEXT,                          -- 视频描述/prompt
    username TEXT,                      -- 作者用户名
    user_id TEXT,                       -- 作者用户 ID
    posted_at REAL,                     -- 发布时间戳
    downloaded_at DATETIME,             -- 下载时间
    local_video_path TEXT,              -- 本地视频路径
    local_thumbnail_path TEXT,          -- 本地缩略图路径
    width INTEGER,                      -- 视频宽度
    height INTEGER,                     -- 视频高度
    uploaded INTEGER DEFAULT 0          -- 是否已上传到 Goldcast (0=未上传, 1=已上传)
);
```

## 文件组织

下载的文件按 task_id 组织：

```
downloads/sora/
├── task_01k7ma2amre7jsqsx60kpex36s/
│   ├── video.mp4
│   └── thumbnail.webp
├── task_01k7m98aweexhsrzpwgm8cywfs/
│   ├── video.mp4
│   └── thumbnail.webp
└── ...
```

## Goldcast 上传集成

### 概述

Feed 下载器支持将下载的视频自动上传到 Goldcast 媒体 CMS 系统。系统会跟踪哪些视频已经上传，避免重复上传。

### 基本用法

```bash
# 上传所有未上传的视频
./run_service.sh feed-uploadgoldcast

# 限制上传数量（例如每次只上传 10 个）
./run_service.sh feed-uploadgoldcast --upload-limit=10

# 使用自定义 API 配置
./run_service.sh feed-uploadgoldcast --api-key=YOUR_KEY --api-url=YOUR_URL
```

### 命令参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--api-key` | `ucHZBRJ1.w8njpEorJlDgjp0ESnw0qSyOkN6V6VUe` | Goldcast API 密钥 |
| `--api-url` | `https://financial.xiaoyequ9.com/api/v1/external/media/upload` | Goldcast API 地址 |
| `--db-path` | `./sora.db` | SQLite 数据库路径 |
| `--upload-limit` | `0` | 单次上传数量限制 (0=全部) |
| `--debug` | `false` | 启用调试日志 |

### 环境变量支持

**重要**: 从 2025年10月17日 起，OSS 配置为必填项，必须通过环境变量提供。

#### 必需的 OSS 配置

```bash
# 阿里云 OSS 配置（必填）
export OSS_ACCESS_KEY_ID="your-aliyun-access-key-id"
export OSS_ACCESS_KEY_SECRET="your-aliyun-access-key-secret"
export OSS_BUCKET_NAME="your-bucket-name"
export OSS_ENDPOINT="oss-cn-beijing.aliyuncs.com"
export OSS_REGION="cn-beijing"  # 可选，默认为 cn-beijing
```

#### 可选的 Goldcast API 配置

```bash
# Goldcast API 配置（可选）
export GOLDCAST_API_KEY="your-api-key"
export GOLDCAST_API_URL="https://your-goldcast-instance.com/api/v1/external/media/upload"
```

#### 完整示例

```bash
# 设置所有环境变量
export OSS_ACCESS_KEY_ID="LTxxxxxxxxxx"
export OSS_ACCESS_KEY_SECRET="xxxxxxxxxxxxxx"
export OSS_BUCKET_NAME="your-bucket"
export OSS_ENDPOINT="oss-cn-beijing.aliyuncs.com"
export GOLDCAST_API_KEY="your-api-key"
export GOLDCAST_API_URL="https://your-goldcast-instance.com/api/v1/external/media/upload"

# 然后直接运行
./run_service.sh feed-uploadgoldcast
```

**注意**: 如果未设置 OSS 环境变量，上传命令会失败并显示清晰的错误信息。

### 典型工作流

**1. 同步和上传（推荐）**

```bash
# 步骤 1: 下载最新的视频
./run_service.sh feed-sync --limit=100

# 步骤 2: 上传到 Goldcast
./run_service.sh feed-uploadgoldcast
```

**2. 定时任务集成**

```bash
# 在 cron 中设置自动同步和上传
# 每天凌晨 3 点下载，3:30 上传

# /etc/crontab 或 crontab -e
0 3 * * * cd /path/to/socrawler && ./run_service.sh feed-sync --limit=1000
30 3 * * * cd /path/to/socrawler && ./run_service.sh feed-uploadgoldcast
```

**3. 直接使用命令行**

```bash
# 使用二进制直接调用
./socrawler feed uploadgoldcast \
  --db-path ./sora.db \
  --limit 50 \
  --debug
```

### 上传逻辑

1. **自动去重**: 只上传 `uploaded = 0` 的视频
2. **使用原始 URL**: 上传时使用 Sora 的原始视频 URL，由 Goldcast 下载
3. **标题处理**: 自动将视频描述截断到 100 字符作为标题
4. **失败重试**: 上传失败的视频保持 `uploaded = 0` 状态，下次运行时会重试
5. **进度跟踪**: 成功上传后立即标记 `uploaded = 1`

### 视频信息映射

| Sora 字段 | Goldcast 字段 | 说明 |
|-----------|---------------|------|
| `video_url` | `media_url` | 视频下载地址 |
| `text` (截断) | `title` | 标题（最多 100 字符） |
| `text` (完整) | `description` | 描述 |
| 固定值 | `user` | 上传用户信息 |

### 幂等性

上传命令是幂等的，可以安全地多次运行：

```bash
# 运行多次只会上传新的未上传视频
./run_service.sh feed-uploadgoldcast
./run_service.sh feed-uploadgoldcast  # 跳过已上传的
./run_service.sh feed-uploadgoldcast  # 跳过已上传的
```

### 查看上传统计

```bash
# 查看数据库中的上传状态
sqlite3 sora.db "SELECT 
    COUNT(*) as total,
    SUM(uploaded) as uploaded,
    COUNT(*) - SUM(uploaded) as pending
FROM sora_videos;"
```

## 定时任务设置

### 使用 cron（推荐）

```bash
# 每 6 小时运行一次
0 */6 * * * cd /path/to/socrawler && ./run_service.sh feed --limit=100

# 每天凌晨 2 点运行
0 2 * * * cd /path/to/socrawler && ./run_service.sh feed --limit=200

# 每小时运行一次
0 * * * * cd /path/to/socrawler && ./run_service.sh feed --limit=50
```

### 使用 systemd timer

创建 `/etc/systemd/system/socrawler-feed.service`:

```ini
[Unit]
Description=Socrawler Feed Downloader
After=network.target

[Service]
Type=oneshot
User=your-user
WorkingDirectory=/path/to/socrawler
ExecStart=/path/to/socrawler/run_service.sh feed --limit=100
```

创建 `/etc/systemd/system/socrawler-feed.timer`:

```ini
[Unit]
Description=Run Socrawler Feed Downloader every 6 hours

[Timer]
OnCalendar=*-*-* 0/6:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

启用定时器：

```bash
sudo systemctl enable socrawler-feed.timer
sudo systemctl start socrawler-feed.timer
```

## 查询数据库

### 查看所有视频

```bash
sqlite3 sora.db "SELECT post_id, username, text FROM sora_videos;"
```

### 按作者查询

```bash
sqlite3 sora.db "SELECT post_id, text FROM sora_videos WHERE username='karneaux';"
```

### 查看最近下载

```bash
sqlite3 sora.db "SELECT post_id, username, downloaded_at FROM sora_videos ORDER BY downloaded_at DESC LIMIT 10;"
```

### 统计信息

```bash
# 总视频数
sqlite3 sora.db "SELECT COUNT(*) FROM sora_videos;"

# 按作者统计
sqlite3 sora.db "SELECT username, COUNT(*) as count FROM sora_videos GROUP BY username ORDER BY count DESC LIMIT 10;"
```

### 查看表结构
```bash
sqlite3 sora.db ".schema sora_videos"
```

## 测试

运行完整测试套件：

```bash
./test_feed.sh
```

这将：
1. 构建项目
2. 下载少量视频（5个）
3. 验证文件和数据库
4. 测试去重功能
5. 测试 run_service.sh 集成

## 故障排查

### 问题：无法访问 feed 端点

**症状**: 提示无法获取 feed 或返回空数据

**解决方案**:
- 确保网络连接正常
- 尝试非 headless 模式查看浏览器行为：`--headless=false`
- 检查是否被限流，等待一段时间后重试

### 问题：数据库锁定

**症状**: `database is locked` 错误

**解决方案**:
- 确保没有其他进程正在使用数据库
- 检查是否有未关闭的数据库连接

### 问题：下载失败

**症状**: 部分视频下载失败

**解决方案**:
- 检查网络连接
- 查看日志中的详细错误信息
- 视频 URL 可能已过期，重新运行 feed 命令获取最新 URL

## 与旧版爬虫的对比

| 特性 | 旧版（滚动爬取） | 新版（Feed） |
|------|------------------|--------------|
| 速度 | 慢（需要滚动等待） | 快（直接获取列表） |
| 稳定性 | 中（依赖页面渲染） | 高（API 调用） |
| 去重 | 无 | 有（数据库） |
| 元数据 | 有限 | 完整 |
| 可调度 | 不适合 | 非常适合 |

## 示例工作流

### 每日自动下载

```bash
# 1. 设置 cron job
crontab -e

# 2. 添加以下行（每天凌晨 3 点运行）
0 3 * * * cd /path/to/socrawler && ./run_service.sh feed --limit=100 >> /var/log/socrawler-feed.log 2>&1

# 3. 定期清理旧数据（可选）
# 删除 30 天前的视频
find /path/to/downloads/sora -type f -mtime +30 -delete
```

### 手动批量下载

```bash
# 下载尽可能多的视频（分批次）
./run_service.sh feed --limit=100
sleep 60  # 等待 1 分钟避免限流
./run_service.sh feed --limit=100
sleep 60
./run_service.sh feed --limit=100
```

## 技术实现

### 架构

```
┌─────────────────┐
│   CLI Command   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Feed Downloader │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────┐   ┌─────────┐
│ DB  │   │ Fetcher │
└─────┘   └────┬────┘
              │
              ▼
        ┌──────────────┐
        │   Browser    │
        │ (Stealth)    │
        └──────────────┘
```

### 关键文件

- `sora/types.go` - Feed JSON 结构定义
- `sora/database.go` - SQLite 数据库操作
- `sora/feed_fetcher.go` - Feed 获取和解析
- `sora/feed_downloader.go` - 下载逻辑
- `main.go` - CLI 命令定义
- `run_service.sh` - Shell 脚本封装

## 许可证

与主项目相同。

