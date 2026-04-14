# OpenAI 兼容格式

本平台完整支持 OpenAI 的 Chat Completions API 和 Responses API。通过这些端点，你可以使用 OpenAI SDK 或任何兼容 OpenAI 格式的客户端来调用平台上的模型，包括 Claude、GPT、Gemini 等 -- 平台会自动完成协议转换。

## Chat Completions

**端点**: `POST {{BASE_URL}}/v1/chat/completions`

这是最广泛支持的 OpenAI 兼容端点，几乎所有 AI 客户端和 IDE 插件都支持此格式。

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `model` | string | 是 | 模型名称，例如 `gpt-5.4`、`claude-sonnet-4-6` |
| `messages` | array | 是 | 消息数组，包含 `role` 和 `content` 字段 |
| `stream` | boolean | 否 | 是否启用流式输出，默认 `false` |
| `temperature` | number | 否 | 采样温度，范围 0-2 |
| `max_tokens` | integer | 否 | 最大生成 token 数 |

### curl 示例

```bash
curl -X POST {{BASE_URL}}/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "system", "content": "你是一个有用的助手。"},
      {"role": "user", "content": "你好，请介绍一下自己。"}
    ],
    "stream": false
  }'
```

流式输出：

```bash
curl -X POST {{BASE_URL}}/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "user", "content": "写一首关于春天的诗"}
    ],
    "stream": true
  }'
```

### Python OpenAI SDK 示例

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxx",
    base_url="{{BASE_URL}}/v1"
)

# 非流式调用
response = client.chat.completions.create(
    model="gpt-5.4",
    messages=[
        {"role": "system", "content": "你是一个有用的助手。"},
        {"role": "user", "content": "你好，请介绍一下自己。"}
    ]
)
print(response.choices[0].message.content)

# 流式调用
stream = client.chat.completions.create(
    model="gpt-5.4",
    messages=[
        {"role": "user", "content": "写一首关于春天的诗"}
    ],
    stream=True
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### Node.js 示例

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-xxxx",
  baseURL: "{{BASE_URL}}/v1",
});

// 非流式调用
const response = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [
    { role: "system", content: "你是一个有用的助手。" },
    { role: "user", content: "你好，请介绍一下自己。" },
  ],
});
console.log(response.choices[0].message.content);

// 流式调用
const stream = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [
    { role: "user", content: "写一首关于春天的诗" },
  ],
  stream: true,
});
for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || "");
}
```

## Responses API

**端点**: `POST {{BASE_URL}}/v1/responses`

这是 OpenAI 较新的 Responses API 格式，Codex CLI 默认使用此端点。

### curl 示例

```bash
curl -X POST {{BASE_URL}}/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxx" \
  -d '{
    "model": "gpt-5.4",
    "input": "用简单的语言解释什么是量子计算。"
  }'
```

### 子路径

Responses API 还支持子路径请求：

```
POST {{BASE_URL}}/v1/responses/*
```

### WebSocket 连接

Responses API 同时支持 WebSocket 方式连接，用于更高效的流式传输：

```
GET {{BASE_URL}}/v1/responses
```

Codex CLI 在启用 WebSocket 模式后会自动使用此连接方式。详细配置请参考 CLI 配置教程中的 Codex CLI 章节。

## 模型列表

**端点**: `GET {{BASE_URL}}/v1/models`

返回当前 API Key 所属分组下可用的模型列表。

### curl 示例

```bash
curl {{BASE_URL}}/v1/models \
  -H "Authorization: Bearer sk-xxxx"
```

### 响应示例

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-5.4",
      "object": "model",
      "owned_by": "openai"
    },
    {
      "id": "claude-sonnet-4-6",
      "object": "model",
      "owned_by": "anthropic"
    }
  ]
}
```

返回的模型列表取决于 API Key 所属分组的配置，不同分组可能包含不同的模型。

## 注意事项

- `base_url` 设置为 `{{BASE_URL}}/v1`（需要包含 `/v1` 后缀）
- 平台支持将 OpenAI 格式请求自动转换为其他服务商的格式，因此即使模型是 Claude 或 Gemini，也可以通过 OpenAI 格式调用
- 流式输出遵循标准 SSE（Server-Sent Events）协议
