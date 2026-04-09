# Anthropic 原生格式

本平台支持 Anthropic 原生的 Messages API 格式。Claude Code 等 Anthropic 官方工具默认使用此格式进行通信。

## Messages API

**端点**: `POST {{BASE_URL}}/v1/messages`

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `model` | string | 是 | 模型名称，例如 `claude-sonnet-4-20250514` |
| `messages` | array | 是 | 消息数组，包含 `role` 和 `content` 字段 |
| `max_tokens` | integer | 是 | 最大生成 token 数 |
| `stream` | boolean | 否 | 是否启用流式输出，默认 `false` |
| `system` | string | 否 | 系统提示词 |
| `temperature` | number | 否 | 采样温度，范围 0-1 |

### 必需请求头

| 请求头 | 值 | 说明 |
|-------|-----|------|
| `Content-Type` | `application/json` | 请求体格式 |
| `Authorization` | `Bearer sk-xxxx` | API Key 认证 |
| `anthropic-version` | `2023-06-01` | Anthropic API 版本号，必须包含 |

### curl 示例

```bash
curl -X POST {{BASE_URL}}/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "你好，请介绍一下自己。"}
    ]
  }'
```

流式输出：

```bash
curl -X POST {{BASE_URL}}/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 1024,
    "stream": true,
    "messages": [
      {"role": "user", "content": "写一首关于春天的诗"}
    ]
  }'
```

流式响应采用标准 SSE（Server-Sent Events）协议，事件类型包括 `message_start`、`content_block_start`、`content_block_delta`、`content_block_stop`、`message_delta`、`message_stop` 等。

### Python Anthropic SDK 示例

```python
from anthropic import Anthropic

client = Anthropic(
    api_key="sk-xxxx",
    base_url="{{BASE_URL}}"
)

# 非流式调用
response = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "你好，请介绍一下自己。"}
    ]
)
print(response.content[0].text)

# 流式调用
with client.messages.stream(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "写一首关于春天的诗"}
    ]
) as stream:
    for text in stream.text_stream:
        print(text, end="")
```

> **重要**: 使用 Anthropic SDK 时，`base_url` 应设置为 `{{BASE_URL}}`，**不要**添加 `/v1` 后缀。SDK 会自动拼接 `/v1/messages` 路径。这与 OpenAI SDK 的行为不同。

## Token 计数

**端点**: `POST {{BASE_URL}}/v1/messages/count_tokens`

在不实际发起 API 调用的情况下，计算给定输入的 token 数量。可用于预估请求成本或检查是否超出模型的上下文窗口限制。

### curl 示例

```bash
curl -X POST {{BASE_URL}}/v1/messages/count_tokens \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [
      {"role": "user", "content": "你好，这段文字有多少 token？"}
    ]
  }'
```

> 此端点仅在 API Key 所属分组的平台类型为 Anthropic 时可用。OpenAI 类型的分组调用此端点会返回 404。

## 认证方式

Anthropic 格式的端点支持以下两种认证方式：

| 方式 | 请求头 | 示例 |
|------|-------|------|
| Bearer Token | `Authorization` | `Authorization: Bearer sk-xxxx` |
| API Key | `x-api-key` | `x-api-key: sk-xxxx` |

两种方式效果相同，使用同一个 API Key。Anthropic 官方 SDK 默认使用 `x-api-key` 请求头，平台对此完全兼容。

## 注意事项

- `anthropic-version` 请求头是必需的，否则请求可能被拒绝
- Anthropic SDK 的 `base_url` 设置为 `{{BASE_URL}}`（不含 `/v1`），OpenAI SDK 则需要 `{{BASE_URL}}/v1`（含 `/v1`），请注意区分
- Claude Code 的配置方式请参考 CLI 配置教程中的 Claude Code 章节
