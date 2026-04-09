# API 接入文档

---

## 快速开始

只需 3 步即可开始使用：

1. **注册账号** — 访问平台首页，完成注册
2. **创建 API Key** — 在控制台「API Keys」页面创建密钥，选择对应的模型分组
3. **配置工具** — 将 API Key 和接口地址配置到你的开发工具中

> 每个 API Key 绑定一个分组，不同分组可能包含不同的模型和定价。创建 Key 时请选择合适的分组。

---

## 接口地址

在控制台「API Keys」页面顶部可以查看所有可用的接口地址（Base URL）。

如果管理员配置了多个线路，你可以点击测速选择延迟最低的节点。

---

## 认证方式

所有 API 请求通过标准 Bearer Token 认证，与 OpenAI / Anthropic SDK 兼容：

```
Authorization: Bearer sk-xxxx
```

---

## 支持的 API 格式

平台同时支持多种 API 格式，你可以根据使用的工具选择对应的端点：

### OpenAI 兼容格式

| 端点 | 说明 |
|------|------|
| `POST /v1/chat/completions` | Chat Completions（GPT / Claude / Gemini 通用） |
| `POST /v1/responses` | OpenAI Responses API |
| `POST /v1/embeddings` | 文本向量化 |
| `POST /v1/images/generations` | 图像生成 |
| `GET /v1/models` | 获取可用模型列表 |

### Anthropic 原生格式

| 端点 | 说明 |
|------|------|
| `POST /v1/messages` | Claude Messages API |
| `POST /v1/messages/count_tokens` | Token 计数 |

### Gemini 原生格式

| 端点 | 说明 |
|------|------|
| `GET /v1beta/models` | 获取 Gemini 模型列表 |
| `POST /v1beta/models/{model}:generateContent` | Gemini 内容生成 |
| `POST /v1beta/models/{model}:streamGenerateContent` | Gemini 流式生成 |

---

## 调用示例

### cURL — OpenAI 兼容格式

```bash
curl -X POST https://你的域名/v1/chat/completions \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 1024
  }'
```

### cURL — Anthropic 原生格式

```bash
curl -X POST https://你的域名/v1/messages \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 1024
  }'
```

### Python — OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxx",
    base_url="https://你的域名/v1",
)

response = client.chat.completions.create(
    model="claude-sonnet-4-6",
    messages=[{"role": "user", "content": "你好"}],
    max_tokens=1024,
)
print(response.choices[0].message.content)
```

### Python — Anthropic SDK

```python
import anthropic

client = anthropic.Anthropic(
    api_key="sk-xxxx",
    base_url="https://你的域名",
)

message = client.messages.create(
    model="claude-sonnet-4-6",
    max_tokens=1024,
    messages=[{"role": "user", "content": "你好"}],
)
print(message.content[0].text)
```

### Node.js — OpenAI SDK

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-xxxx",
  baseURL: "https://你的域名/v1",
});

const response = await client.chat.completions.create({
  model: "claude-sonnet-4-6",
  messages: [{ role: "user", content: "你好" }],
  max_tokens: 1024,
});
console.log(response.choices[0].message.content);
```

---

## 工具接入指南

### Claude Code

在终端中设置环境变量后启动 Claude Code：

**macOS / Linux：**

```bash
export ANTHROPIC_BASE_URL="https://你的域名"
export ANTHROPIC_AUTH_TOKEN="sk-xxxx"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
```

**Windows CMD：**

```cmd
set ANTHROPIC_BASE_URL=https://你的域名
set ANTHROPIC_AUTH_TOKEN=sk-xxxx
set CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
```

**Windows PowerShell：**

```powershell
$env:ANTHROPIC_BASE_URL="https://你的域名"
$env:ANTHROPIC_AUTH_TOKEN="sk-xxxx"
$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
```

**VSCode 中使用 Claude Code（持久化配置）：**

编辑 `~/.claude/settings.json`（Windows: `%userprofile%\.claude\settings.json`）：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://你的域名",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxxx",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

> 提示：创建 API Key 后，点击「使用」按钮可直接获取已填入你的 Key 的配置代码。

---

### Cursor

Cursor 内置支持 OpenAI 兼容 API。在 Cursor 设置中：

1. 打开 `Settings` → `Models`
2. 配置 `OpenAI API Key` 为你的 `sk-xxxx`
3. 配置 `Base URL` 为 `https://你的域名/v1`
4. 选择需要使用的模型

---

### Codex CLI（OpenAI 官方命令行工具）

编辑配置文件：

**~/.codex/config.toml：**

```toml
model_provider = "OpenAI"
model = "gpt-5.4"
review_model = "gpt-5.4"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "https://你的域名"
wire_api = "responses"
requires_openai_auth = true
```

**~/.codex/auth.json：**

```json
{
  "OPENAI_API_KEY": "sk-xxxx"
}
```

---

### Gemini CLI

```bash
export GOOGLE_GEMINI_BASE_URL="https://你的域名"
export GEMINI_API_KEY="sk-xxxx"
export GEMINI_MODEL="gemini-2.5-pro"
```

---

### OpenCode

在项目根目录创建 `opencode.json`：

```json
{
  "provider": "anthropic",
  "model": "claude-sonnet-4-6",
  "apiKey": "sk-xxxx",
  "baseURL": "https://你的域名/v1"
}
```

---

### Cline / Continue / 其他 IDE 插件

大多数支持 OpenAI 兼容 API 的工具都可以接入，只需配置：

- **API Key**：你的 `sk-xxxx`
- **Base URL**：`https://你的域名/v1`
- **Model**：选择你分组中可用的模型名

---

## 可用模型

登录后访问「模型」页面查看完整的模型列表和定价。

你也可以通过 API 获取当前 Key 可用的模型列表：

```bash
curl https://你的域名/v1/models \
  -H "Authorization: Bearer sk-xxxx"
```

---

## 费用与用量

- **按量计费** — 根据实际使用的 Token 数量扣费，不同模型价格不同
- **余额永不过期** — 充值后余额长期有效
- **实时查看** — 在控制台「活动日志」中可查看每次请求的详细消费记录
- **用量限制** — API Key 支持设置额度上限，用完自动停用，防止超支

---

## 常见问题

### 模型名应该填什么？

使用「模型」页面中显示的模型名称，例如 `claude-sonnet-4-6`、`gpt-5.4`。模型名必须精确匹配。

### 提示 401 Unauthorized？

- 检查 API Key 是否正确
- 确认 Key 状态是否为「活跃」
- 如设置了 IP 白名单，确认当前 IP 在白名单内

### 提示模型不可用？

- 确认你的 API Key 所属分组支持该模型
- 在「模型」页面确认模型名称拼写正确

### 支持流式输出吗？

支持。在请求中设置 `"stream": true` 即可。

### 可以同时使用多个分组吗？

可以。创建多个 API Key，每个绑定不同分组即可。不同 Key 可配置给不同的工具或用途。
