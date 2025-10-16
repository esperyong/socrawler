# Sora Feed 下载器 - 快速使用指南

## 🚀 快速开始

### 1. 构建

```bash
./run_service.sh build
```

### 2. 运行

```bash
# 下载最新的 50 个视频
./run_service.sh feed

# 下载 100 个视频
./run_service.sh feed --limit=100

# 调试模式（显示浏览器）
./run_service.sh feed --headless=false --limit=10
```

## 📝 基本命令

```bash
# 使用 run_service.sh（推荐）
./run_service.sh feed                          # 默认下载 50 个
./run_service.sh feed --limit=100              # 下载 100 个
./run_service.sh feed --headless=false         # 显示浏览器窗口

# 直接使用命令行
./socrawler feed --help                        # 查看帮助
./socrawler feed                               # 使用默认设置
./socrawler feed --save-path ./my_videos       # 自定义保存路径
./socrawler feed --db-path ./my_db.db          # 自定义数据库路径
./socrawler feed --limit 200                   # 下载 200 个视频
```

## ⏰ 定时下载

### 使用 cron（推荐）

```bash
# 编辑 cron
crontab -e

# 每 6 小时下载一次
0 */6 * * * cd /Users/liunig/develop/github_projects/socrawler && ./run_service.sh feed --limit=100

# 每天凌晨 2 点下载
0 2 * * * cd /Users/liunig/develop/github_projects/socrawler && ./run_service.sh feed --limit=200

# 每小时下载
0 * * * * cd /Users/liunig/develop/github_projects/socrawler && ./run_service.sh feed
```

## 🔍 查看下载结果

### 查看视频文件

```bash
# 查看下载的视频
ls -lh downloads/sora/

# 查看某个视频文件夹
ls -lh downloads/sora/task_01k7ma2amre7jsqsx60kpex36s/
```

### 查询数据库

```bash
# 查看所有视频信息
sqlite3 sora.db "SELECT post_id, username, text FROM sora_videos;"

# 查看视频总数
sqlite3 sora.db "SELECT COUNT(*) FROM sora_videos;"

# 查看最近下载的 10 个视频
sqlite3 sora.db "SELECT username, text, downloaded_at FROM sora_videos ORDER BY downloaded_at DESC LIMIT 10;"

# 统计每个作者的视频数量
sqlite3 sora.db "SELECT username, COUNT(*) as count FROM sora_videos GROUP BY username ORDER BY count DESC;"

# 查找特定作者的视频
sqlite3 sora.db "SELECT post_id, text FROM sora_videos WHERE username='karneaux';"
```

## 📊 下载统计

运行完成后会显示：

```
========================================
  Feed Download Results
========================================
Total items fetched:    500          # Feed 中总共有多少视频
New videos found:       50           # 发现多少个新视频
Successfully downloaded: 48          # 成功下载多少个
Skipped:                0            # 跳过多少个
Failed:                 2            # 失败多少个
Duration:               120 seconds  # 耗时
========================================

Videos saved to: ./downloads/sora
Database: ./sora.db
```

## 🛠️ 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--save-path` | `./downloads/sora` | 视频保存目录 |
| `--db-path` | `./sora.db` | 数据库文件路径 |
| `--limit` | `50` | 单次最多下载多少个视频 |
| `--headless` | `true` | 是否使用无头浏览器（不显示窗口） |
| `--debug` | `false` | 是否显示调试信息 |

## 🔄 工作原理

1. **获取 Feed**: 访问 Sora 的 feed 端点获取视频列表
2. **检查数据库**: 查询哪些视频已经下载过
3. **过滤新视频**: 只下载未下载过的视频
4. **下载视频**: 下载视频文件和缩略图
5. **保存记录**: 将视频信息保存到数据库

### 文件组织

```
downloads/sora/
├── task_01k7ma2amre7jsqsx60kpex36s/
│   ├── video.mp4          # 视频文件
│   └── thumbnail.webp     # 缩略图
├── task_01k7m98aweexhsrzpwgm8cywfs/
│   ├── video.mp4
│   └── thumbnail.webp
└── ...
```

## ✅ 测试

运行测试脚本验证功能：

```bash
./test_feed.sh
```

## ❓ 常见问题

### Q: 下载失败了怎么办？

A: 再次运行相同的命令，程序会自动跳过已下载的视频，只下载失败的。

```bash
./run_service.sh feed --limit=100
```

### Q: 如何删除已下载的视频记录？

A: 可以直接删除数据库或清空表：

```bash
# 删除整个数据库（会重新下载所有视频）
rm sora.db

# 清空表但保留结构
sqlite3 sora.db "DELETE FROM sora_videos;"
```

### Q: 可以同时运行多个下载任务吗？

A: 不建议。SQLite 数据库在并发写入时可能出现锁定问题。如果需要加速，可以增加单次下载的 limit。

### Q: 视频 URL 会过期吗？

A: 会的。Feed 中的视频 URL 包含签名和过期时间。如果下载失败，重新运行 feed 命令会获取新的 URL。

### Q: 如何只下载特定作者的视频？

A: 目前暂不支持。可以先下载所有视频，然后通过数据库查询找到特定作者的视频：

```bash
sqlite3 sora.db "SELECT local_video_path FROM sora_videos WHERE username='作者名';"
```

### Q: 占用空间太大怎么办？

A: 可以定期清理旧文件：

```bash
# 删除 30 天前的视频文件
find downloads/sora -type f -mtime +30 -delete

# 只保留数据库记录，删除所有视频文件
rm -rf downloads/sora/*

# 清理数据库中的旧记录
sqlite3 sora.db "DELETE FROM sora_videos WHERE downloaded_at < datetime('now', '-30 days');"
```

## 📖 更多信息

- 详细文档: `FEED_DOWNLOADER.md`
- 实现总结: `IMPLEMENTATION_SUMMARY.md`
- 测试脚本: `test_feed.sh`

## 💡 使用建议

1. **首次使用**: 先用小的 limit（如 10）测试
2. **正式使用**: 建议 limit 设置为 50-100
3. **定时任务**: 每 6 小时运行一次比较合适
4. **存储管理**: 定期清理旧视频或备份到云存储
5. **调试问题**: 使用 `--headless=false` 可以看到浏览器操作过程

## 🎯 典型使用场景

### 场景 1: 每日自动收集

```bash
# 设置 cron 每天凌晨 3 点运行
0 3 * * * cd /path/to/socrawler && ./run_service.sh feed --limit=100 >> /var/log/socrawler.log 2>&1
```

### 场景 2: 手动批量下载

```bash
# 分批下载，避免一次性下载太多
for i in {1..5}; do
    ./run_service.sh feed --limit=100
    sleep 60  # 等待 1 分钟
done
```

### 场景 3: 按作者整理

```bash
# 下载所有视频
./run_service.sh feed --limit=200

# 查询某作者的视频
sqlite3 sora.db "SELECT local_video_path FROM sora_videos WHERE username='特定作者';"

# 复制到单独文件夹
mkdir -p videos_by_author/特定作者
# 根据查询结果复制文件
```

---

**提示**: 第一次运行可能需要较长时间，因为需要启动浏览器和下载文件。后续运行会因为去重而更快。

