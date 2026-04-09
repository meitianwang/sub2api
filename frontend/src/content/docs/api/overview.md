# API 概述

本平台提供统一的 API 网关，同时支持三种主流协议格式。你可以根据所使用的客户端工具或 SDK，选择最合适的接入方式。

## 基础地址

所有 API 请求的基础地址为：

```
{{BASE_URL}}
```

## 支持的协议格式

| 协议格式 | 说明 | 典型使用场景 |
|---------|------|-------------|
| **OpenAI 兼容格式** | Chat Completions + Responses API | Codex CLI、ChatGPT 兼容客户端、各类 IDE 插件 |
| **Anthropic 原生格式** | Messages API | Claude Code |
| **Gemini 原生格式** | v1beta 格式 | Gemini CLI |

## 支持的端点一览

以下是平台当前支持的全部公开端点：

### OpenAI 兼容端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | Chat Completions（对话补全） |
| POST | `/v1/responses` | Responses API |
| POST | `/v1/responses/*` | Responses API 子路径 |
| GET  | `/v1/responses` | Responses API WebSocket 连接 |
| GET  | `/v1/models` | 获取当前 Key 可用的模型列表 |

### Anthropic 原生端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/messages` | Messages API（消息接口） |
| POST | `/v1/messages/count_tokens` | Token 计数 |

### Gemini 原生端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/v1beta/models` | 列出可用的 Gemini 模型 |
| GET  | `/v1beta/models/:model` | 获取指定模型详情 |
| POST | `/v1beta/models/{model}:generateContent` | 生成内容（同步） |
| POST | `/v1beta/models/{model}:streamGenerateContent` | 生成内容（流式） |

## 认证方式

所有协议格式使用同一个 API Key 进行认证。平台支持以下几种传递方式：

| 认证方式 | 示例 | 适用范围 |
|---------|------|---------|
| `Authorization` 请求头 | `Authorization: Bearer sk-xxxx` | 所有端点 |
| `x-api-key` 请求头 | `x-api-key: sk-xxxx` | 所有端点（Anthropic SDK 常用） |
| `x-goog-api-key` 请求头 | `x-goog-api-key: sk-xxxx` | Gemini 端点 |
| `key` 查询参数 | `?key=sk-xxxx` | 仅 `/v1beta` 端点 |

推荐使用 `Authorization: Bearer` 方式，兼容性最广。

## 自动路由

平台会根据你调用的 API 端点格式自动路由到正确的上游服务商。你使用的 API 格式取决于客户端工具的要求，而非分组本身。例如：

- 使用 Claude Code 时，客户端自动发送 Anthropic 格式请求
- 使用 Codex CLI 时，客户端自动发送 OpenAI Responses 格式请求
- 使用 Gemini CLI 时，客户端自动发送 v1beta 格式请求

你只需确保 API Key 所属的分组与目标服务商匹配即可。
