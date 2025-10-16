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
    height INTEGER                      -- 视频高度
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

