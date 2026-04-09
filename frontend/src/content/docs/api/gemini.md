# Gemini 原生格式

本平台支持 Google Gemini 的原生 v1beta API 格式。Gemini CLI 默认使用此格式进行通信。所有 Gemini 原生端点均位于 `/v1beta` 路径前缀下。

## 端点一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/v1beta/models` | 列出可用模型 |
| GET  | `/v1beta/models/{model}` | 获取指定模型详情 |
| POST | `/v1beta/models/{model}:generateContent` | 生成内容（同步） |
| POST | `/v1beta/models/{model}:streamGenerateContent` | 生成内容（流式） |

## 认证方式

Gemini 原生端点支持以下三种认证方式：

| 方式 | 说明 | 示例 |
|------|------|------|
| `Authorization` 请求头 | Bearer Token | `Authorization: Bearer sk-xxxx` |
| `x-goog-api-key` 请求头 | Google API Key 格式 | `x-goog-api-key: sk-xxxx` |
| `key` 查询参数 | URL 参数传递 | `?key=sk-xxxx` |

三种方式使用同一个 API Key，效果相同。Gemini CLI 默认使用 `x-goog-api-key` 请求头或 `key` 查询参数。

## 列出模型

**端点**: `GET {{BASE_URL}}/v1beta/models`

返回当前 API Key 所属分组下可用的 Gemini 模型列表。

```bash
curl {{BASE_URL}}/v1beta/models \
  -H "x-goog-api-key: sk-xxxx"
```

或使用查询参数：

```bash
curl "{{BASE_URL}}/v1beta/models?key=sk-xxxx"
```

## 获取模型详情

**端点**: `GET {{BASE_URL}}/v1beta/models/{model}`

获取指定模型的详细信息，包括支持的参数、上下文窗口大小等。

```bash
curl {{BASE_URL}}/v1beta/models/gemini-2.5-pro \
  -H "x-goog-api-key: sk-xxxx"
```

## 生成内容（同步）

**端点**: `POST {{BASE_URL}}/v1beta/models/{model}:generateContent`

发送请求并等待完整响应返回。

### curl 示例

```bash
curl -X POST {{BASE_URL}}/v1beta/models/gemini-2.5-pro:generateContent \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: sk-xxxx" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [
          {"text": "你好，请介绍一下自己。"}
        ]
      }
    ],
    "generationConfig": {
      "temperature": 0.7,
      "maxOutputTokens": 1024
    }
  }'
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `contents` | array | 是 | 对话内容数组，每个元素包含 `role` 和 `parts` |
| `generationConfig` | object | 否 | 生成配置，包括 `temperature`、`maxOutputTokens` 等 |
| `systemInstruction` | object | 否 | 系统指令 |

## 生成内容（流式）

**端点**: `POST {{BASE_URL}}/v1beta/models/{model}:streamGenerateContent`

以 SSE（Server-Sent Events）方式返回流式响应。需要在 URL 中添加 `?alt=sse` 查询参数。

### curl 示例

```bash
curl -X POST "{{BASE_URL}}/v1beta/models/gemini-2.5-pro:streamGenerateContent?alt=sse" \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: sk-xxxx" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [
          {"text": "写一首关于春天的诗"}
        ]
      }
    ]
  }'
```

流式响应以 SSE 格式返回，每个事件的 `data` 字段包含一个 JSON 对象，结构与同步响应相同。

## 多轮对话

Gemini 格式通过在 `contents` 数组中包含多条消息来实现多轮对话：

```bash
curl -X POST {{BASE_URL}}/v1beta/models/gemini-2.5-pro:generateContent \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: sk-xxxx" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "什么是机器学习？"}]
      },
      {
        "role": "model",
        "parts": [{"text": "机器学习是人工智能的一个分支..."}]
      },
      {
        "role": "user",
        "parts": [{"text": "它和深度学习有什么区别？"}]
      }
    ]
  }'
```

注意 Gemini 格式中，助手角色使用 `model` 而非 `assistant`。

## 注意事项

- API Key 必须属于 Gemini 类型的分组，否则请求会被拒绝
- 流式请求需要添加 `?alt=sse` 查询参数以启用 SSE 格式
- 使用 `key` 查询参数认证时，仅限 `/v1beta` 路径下的端点有效
- Gemini CLI 的配置方式请参考 CLI 配置教程中的 Gemini CLI 章节
