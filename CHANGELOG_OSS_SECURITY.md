# Changelog - OSS Security Update

## [1.1.0] - 2025-10-17

### 🔒 Security - Breaking Change

#### OSS 密钥不再硬编码

为了提高安全性，阿里云 OSS 的访问密钥不再作为默认值硬编码在代码中，改为通过环境变量配置的必填项。

### ⚠️ Breaking Changes

1. **`sora/oss_uploader.go`**
   - 移除了硬编码的 OSS 密钥常量：
     - `OSS_ACCESS_KEY_ID`
     - `OSS_ACCESS_KEY_SECRET`
     - `OSS_BUCKET_NAME`
     - `OSS_ENDPOINT`
   - `NewOSSUploader()` 函数签名变更：
     ```go
     // 旧版本
     func NewOSSUploader(db *VideoDatabase) (*OSSUploader, error)
     
     // 新版本
     func NewOSSUploader(config *OSSConfig, db *VideoDatabase) (*OSSUploader, error)
     ```

2. **`sora/goldcast_uploader.go`**
   - `NewGoldcastUploader()` 函数签名变更：
     ```go
     // 旧版本
     func NewGoldcastUploader(config *GoldcastConfig, db *VideoDatabase) (*GoldcastUploader, error)
     
     // 新版本
     func NewGoldcastUploader(config *GoldcastConfig, db *VideoDatabase, ossConfig *OSSConfig) (*GoldcastUploader, error)
     ```

### ✨ New Features

1. **新增函数**: `NewOSSConfigFromEnv()` 
   - 从环境变量读取 OSS 配置
   - 自动验证所有必需字段
   - 提供清晰的错误信息

2. **配置验证**
   - 启动时自动检查 OSS 配置完整性
   - 缺少必需配置时立即报错，避免运行时错误

3. **OSSConfig 结构体增强**
   - 新增 `Region` 字段，支持自定义区域
   - 默认区域：`cn-beijing`

### 📝 Configuration Options

现在运行 `feed uploadgoldcast` 命令时，可以通过以下方式提供 OSS 配置：

**通过命令行参数（推荐）：**
```bash
./socrawler feed uploadgoldcast \
  --oss-access-key-id="your-key-id" \
  --oss-access-key-secret="your-key-secret"
```

**通过环境变量：**

| 变量名 | 必需性 | 默认值 | 说明 |
|--------|--------|--------|------|
| `OSS_ACCESS_KEY_ID` | ✅ 必需 | 无 | 阿里云 Access Key ID |
| `OSS_ACCESS_KEY_SECRET` | ✅ 必需 | 无 | 阿里云 Access Key Secret |
| `OSS_BUCKET_NAME` | ⚪ 可选 | `dreammedias` | OSS Bucket 名称 |
| `OSS_ENDPOINT` | ⚪ 可选 | `oss-cn-beijing.aliyuncs.com` | OSS 端点 |
| `OSS_REGION` | ⚪ 可选 | `cn-beijing` | OSS 区域 |

### 📖 Migration Guide

#### 最简单的方式：命令行参数

```bash
# 只需提供 Access Key 即可，其他使用默认值
./socrawler feed uploadgoldcast \
  --oss-access-key-id="your-access-key-id" \
  --oss-access-key-secret="your-access-key-secret"
```

#### 或者使用环境变量

```bash
# 方式 1: 直接导出（最少配置）
export OSS_ACCESS_KEY_ID="your-access-key-id"
export OSS_ACCESS_KEY_SECRET="your-access-key-secret"
./socrawler feed uploadgoldcast

# 方式 2: 使用配置文件
cp env.example .env
# 编辑 .env 文件，填入 Access Key
source .env
./socrawler feed uploadgoldcast
```

#### 更新现有代码调用

如果你的代码直接调用了这些函数，需要更新：

```go
// 旧代码
ossUploader, err := sora.NewOSSUploader(db)

// 新代码
ossConfig, err := sora.NewOSSConfigFromEnv()
if err != nil {
    return err
}
ossUploader, err := sora.NewOSSUploader(ossConfig, db)
```

### 📁 Modified Files

1. `sora/oss_uploader.go` - 核心修改
2. `sora/goldcast_uploader.go` - 函数签名更新
3. `main.go` - 命令行工具更新

### 📁 New Files

1. `env.example` - 环境变量配置模板
2. `OSS_CONFIG_CHANGE.md` - 详细的变更说明和迁移指南
3. `CHANGELOG_OSS_SECURITY.md` - 本文件

### 📁 Updated Documentation

1. `GOLDCAST_INTEGRATION.md` - 更新了环境变量部分
2. `FEED_DOWNLOADER.md` - 更新了配置说明

### 🔍 Error Messages

新的错误信息更加友好：

```
FATAL Failed to load OSS configuration: OSS_ACCESS_KEY_ID environment variable is required

Please set the following environment variables:
  - OSS_ACCESS_KEY_ID
  - OSS_ACCESS_KEY_SECRET
  - OSS_BUCKET_NAME
  - OSS_ENDPOINT
  - OSS_REGION (optional, defaults to cn-beijing)
```

### ✅ Testing

编译测试通过：
```bash
$ go build -o socrawler .
# 成功，无错误
```

### 🎯 Affected Commands

- ✅ `socrawler feed uploadgoldcast` - 需要 OSS 配置
- ⚪ `socrawler feed fetch` - 不受影响
- ⚪ `socrawler feed download` - 不受影响
- ⚪ `socrawler feed sync` - 不受影响
- ⚪ `socrawler feed export` - 不受影响

### 💡 Benefits

1. **安全性提升**: 密钥不再存储在代码中
2. **灵活性**: 支持多环境配置（开发、测试、生产）
3. **可维护性**: 统一的配置管理方式
4. **合规性**: 符合代码审查和安全规范

### 🔗 Related Issues

- 阿里云 OSS 密钥不应硬编码在代码中
- 需要支持环境变量配置 OSS 凭证

### 📞 Support

详细使用说明请参考：
- [OSS_CONFIG_CHANGE.md](./OSS_CONFIG_CHANGE.md) - 完整的迁移指南
- [GOLDCAST_INTEGRATION.md](./GOLDCAST_INTEGRATION.md) - Goldcast 集成文档
- [FEED_DOWNLOADER.md](./FEED_DOWNLOADER.md) - Feed 下载器文档

---

**注意**: 这是一个破坏性变更（Breaking Change），升级后必须设置环境变量才能使用上传功能。

