# Codex CLI 配置

本文档介绍如何配置 OpenAI Codex CLI 以连接本平台的 API 服务。

## 前置条件

1. 已安装 Node.js 环境
2. 安装 Codex CLI：
```bash
npm i -g @openai/codex@latest
```
3. 首次运行 `codex` 命令以生成配置目录

## 配置方法

需要创建或编辑两个配置文件，根据操作系统找到对应路径：

- **macOS / Linux**: `~/.codex/config.toml` 和 `~/.codex/auth.json`
- **Windows**: `%userprofile%\.codex\config.toml` 和 `%userprofile%\.codex\auth.json`

### config.toml

```toml
model_provider = "custom"
model = "gpt-5.4"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"

[model_providers.custom]
name = "custom"
base_url = "{{BASE_URL}}/v1"
wire_api = "responses"
requires_openai_auth = true
```

### auth.json

```json
{
  "OPENAI_API_KEY": "sk-xxxx"
}
```

将 `sk-xxxx` 替换为你的 API Key。请确保该 Key 属于 OpenAI 分组。

## 模型选择

根据你的分组内可用模型，修改 `config.toml` 中的 `model` 字段。常见可选模型包括：

- `gpt-5.4`
- `gpt-5.2`
- `gpt-5.3-codex`

## 验证配置

在终端中运行 `codex`，若能正常启动并交互，则配置成功。

## WebSocket 模式（推荐）

WebSocket 模式可提供更好的流式传输性能。在 `config.toml` 中添加以下配置：

```toml
[model_providers.custom]
name = "custom"
base_url = "{{BASE_URL}}/v1"
wire_api = "responses"
supports_websockets = true
requires_openai_auth = true

[features]
responses_websockets_v2 = true
```

## 常用命令

| 命令 | 说明 |
|------|------|
| `/model` | 切换模型 |
| `/compact` | 压缩对话以释放上下文空间 |
| `/diff` | 查看当前 git diff |
| `/review` | 审查工作区变更 |
| `/undo` | 撤销上一步操作 |

## 常见问题排查

### 401 认证错误

- 检查 `auth.json` 中的 Key 是否正确填写
- 确认系统环境变量中没有设置冲突的 `OPENAI_API_KEY`

### Windows 编码问题

如遇到字符编码异常，请前往：系统设置 -> 区域 -> 更改系统区域设置 -> 勾选"使用 Unicode UTF-8 提供全球语言支持"。
