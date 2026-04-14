# Cline / Continue 配置

本文档介绍如何配置 Cline 和 Continue 两款 VSCode 插件以连接本平台的 API 服务。两者均支持 OpenAI Compatible API 接入方式，配置方法类似。

## 前置条件

1. 已安装 VSCode
2. 已安装 Cline 或 Continue 插件
3. 已在本平台创建 API Key

## 通用配置参数

无论使用 Cline 还是 Continue，核心配置项相同：

| 配置项 | 值 |
|--------|------|
| API Provider | OpenAI Compatible |
| Base URL | `{{BASE_URL}}/v1` |
| API Key | `sk-xxxx`（替换为你的实际 Key） |
| Model ID | 分组内可用的模型名称 |

常用 Model ID 示例：`claude-sonnet-4-6`、`gemini-3-pro-preview`、`gpt-5.4`

## Cline 配置步骤

1. 在 VSCode 中打开 Cline 插件面板
2. 点击设置图标，进入 API 配置页面
3. **API Provider** 选择 `OpenAI Compatible`
4. 填写配置：
   - **Base URL**: `{{BASE_URL}}/v1`
   - **API Key**: `sk-xxxx`
   - **Model ID**: 填入你要使用的模型名称，例如 `claude-sonnet-4-6`
5. 保存设置后，在对话框中发送消息测试是否正常

## Continue 配置步骤

1. 在 VSCode 中打开 Continue 插件面板
2. 点击齿轮图标打开配置文件（`config.json`）
3. 在 `models` 数组中添加以下配置：

```json
{
  "models": [
    {
      "title": "Claude Sonnet",
      "provider": "openai",
      "model": "claude-sonnet-4-6",
      "apiBase": "{{BASE_URL}}/v1",
      "apiKey": "sk-xxxx"
    }
  ]
}
```

4. 保存配置文件后，在 Continue 面板中选择对应模型即可使用

如需配置多个模型，在 `models` 数组中添加多个条目，分别指定不同的 `model` 名称。

## 适用范围

Cline 和 Continue 的配置方式适用于大多数支持 OpenAI Compatible API 的 IDE 插件和工具。如果你使用的工具提供了类似的"OpenAI Compatible"或"自定义 API"选项，通常只需填写以下三项即可接入：

- Base URL: `{{BASE_URL}}/v1`
- API Key: 你的 `sk-xxxx`
- Model: 分组内可用的模型名称

## 常见问题排查

### 无法连接或超时

- 确认 Base URL 格式正确，末尾需要包含 `/v1`
- 检查网络连接是否正常

### 模型不可用

- 确认填写的 Model ID 与分组内的可用模型名称完全一致
- 登录本平台的「模型」页面确认该模型在你的分组中可用

### 401 认证错误

- 检查 API Key 是否正确复制，没有多余的空格或换行
- 确认该 Key 处于启用状态
