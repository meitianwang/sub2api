# OpenCode 配置

本文档介绍如何配置 OpenCode 终端 AI 助手以连接本平台的 API 服务。

## 前置条件

1. 已安装 Node.js 环境
2. 安装 OpenCode：
```bash
npm install -g opencode-ai
```
3. 运行 `opencode` 验证安装成功

## 配置方法

OpenCode 支持多种 Provider 类型，根据你要使用的模型选择对应的配置方式。配置文件为项目根目录或用户目录下的 `opencode.json`。

### Anthropic Provider（Claude 系列模型）

```json
{
  "provider": {
    "anthropic": {
      "apiKey": "sk-xxxx",
      "baseURL": "{{BASE_URL}}"
    }
  },
  "model": "claude-sonnet-4-6"
}
```

请确保 API Key 属于 Anthropic / Claude 分组。

注意 baseURL 格式：Anthropic Provider 使用 `{{BASE_URL}}`，不需要添加 `/v1` 后缀。

### OpenAI Provider（GPT 系列模型）

```json
{
  "provider": {
    "openai": {
      "apiKey": "sk-xxxx",
      "baseURL": "{{BASE_URL}}/v1"
    }
  },
  "model": "gpt-4o"
}
```

请确保 API Key 属于 OpenAI 分组。

注意 baseURL 格式：OpenAI Provider 使用 `{{BASE_URL}}/v1`，需要添加 `/v1` 后缀。

### Google Provider（Gemini 系列模型）

```json
{
  "provider": {
    "google": {
      "apiKey": "sk-xxxx",
      "baseURL": "{{BASE_URL}}"
    }
  },
  "model": "gemini-2.5-pro"
}
```

请确保 API Key 属于 Gemini 分组。

## baseURL 格式对照表

不同 Provider 类型的 baseURL 格式不同，请注意区分：

| Provider | baseURL | 适用模型 |
|----------|---------|----------|
| Anthropic | `{{BASE_URL}}` | Claude 系列 |
| OpenAI | `{{BASE_URL}}/v1` | GPT 系列 |
| Google | `{{BASE_URL}}` | Gemini 系列 |

## CC-Switch 工具

OpenCode 也可以通过 CC-Switch 工具进行配置切换，适合需要在多个 Provider 之间频繁切换的场景。

## 验证配置

配置完成后，在终端中运行 `opencode`，若能正常启动并与模型交互，则配置成功。

## 常见问题排查

### 连接失败

- 检查 baseURL 格式是否与所选 Provider 类型匹配
- 确认 API Key 所属分组与 Provider 类型对应

### 模型不可用

- 确认所填模型名称与分组内的可用模型一致
- 登录本平台的「模型」页面核实可用模型列表
