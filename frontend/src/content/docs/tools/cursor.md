# Cursor 配置

本文档介绍如何配置 Cursor IDE 以连接本平台的 API 服务。

## 前置条件

1. 已安装 [Cursor](https://cursor.sh/) 编辑器
2. 已在本平台创建 API Key

## 配置方法

Cursor 原生支持 OpenAI 兼容 API，只需在设置中填写 Base URL 和 API Key 即可。

### 步骤一：打开模型设置

在 Cursor 中依次打开：**Settings** -> **Models** -> **OpenAI API Key**

### 步骤二：填写 API 配置

- **API Key**: 填入你的 `sk-xxxx`
- **Base URL**: 填入 `{{BASE_URL}}/v1`

### 步骤三：选择模型

在模型列表中选择你所属分组内的可用模型。可在本平台的「模型」页面查看当前分组支持哪些模型。

## 支持的模型类型

Cursor 通过 OpenAI 兼容格式（`/v1/chat/completions`）调用模型，因此支持以下模型：

- **Claude 系列**：如 `claude-sonnet-4-6`、`claude-opus-4-6` 等
- **GPT 系列**：如 `gpt-4o`、`gpt-5.4` 等
- **Gemini 系列**：如 `gemini-2.5-pro`、`gemini-3-pro-preview` 等

只要你的分组中包含对应模型，且该分组支持 `chat/completions` 格式，即可在 Cursor 中使用。

## 分组与 Key 选择建议

- 如果主要使用 GPT 系列模型，建议使用 **OpenAI 分组**的 Key
- 如果主要使用 Claude 系列模型，可使用 **Anthropic 分组**的 Key（前提是该分组支持 `chat/completions` 兼容格式）
- 如果需要混合使用多种模型，选择包含所需模型的分组即可

## 验证配置

配置完成后，在 Cursor 的对话窗口中发送任意消息，若能正常收到模型回复，则配置成功。

## 常见问题排查

### 模型列表为空

- 确认 Base URL 格式正确，末尾需要包含 `/v1`
- 确认 API Key 填写无误

### 调用报错

- 检查所选模型是否在你的分组可用模型列表中
- 确认账户余额充足
