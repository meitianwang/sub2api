# Gemini 常见问题

本文档汇总 Gemini CLI 及 Gemini 模型使用过程中的常见问题及解决方案。配置教程请参阅「CLI 配置教程 - Gemini CLI」。

---

## Gemini CLI 无法调用模型

**问题**：运行 Gemini CLI 后无法正常调用模型，提示连接错误或模型不存在。

**解决方案**：依次检查以下几项：

1. **`.env` 文件位置是否正确** -- 确认文件位于以下路径：
   - macOS / Linux: `~/.gemini/.env`
   - Windows: `%userprofile%\.gemini\.env`
2. **API Key 是否属于 Gemini 分组** -- Gemini CLI 需要使用 Gemini 类型分组的 Key，使用其他分组的 Key 会导致调用失败
3. **模型名称是否正确** -- `GEMINI_MODEL` 的值必须与分组内可用的模型名称完全一致，登录本平台的「模型」页面确认
4. **`.env` 文件格式** -- 每行一个配置项，等号两侧不要有多余空格：

```
GOOGLE_GEMINI_BASE_URL={{BASE_URL}}
GEMINI_API_KEY=sk-xxxx
GEMINI_MODEL=gemini-2.5-pro
```

---

## 如何查看可用的 Gemini 模型

**问题**：不确定当前分组下有哪些 Gemini 模型可用。

**解决方案**：

1. 登录本平台
2. 进入「模型」页面
3. 查看你所属分组中标注为 Gemini 类型的模型

常见的 Gemini 模型包括：

- `gemini-2.5-pro`
- `gemini-2.5-flash`
- `gemini-3-pro-preview`

具体可用模型取决于你所属分组的配置，以平台页面显示为准。

---

## 是否可以在其他工具中使用 Gemini 模型

**问题**：除了 Gemini CLI，能否在 Cursor、Cline 等其他工具中调用 Gemini 模型。

**解决方案**：可以。本平台支持通过 OpenAI 兼容格式（`/v1/chat/completions`）调用 Gemini 模型。只要工具支持 OpenAI Compatible API，就可以配置使用。

以下工具均可通过此方式调用 Gemini 模型：

- **Cursor** -- 在 Models 设置中配置 Base URL 和 API Key
- **Cline** -- 选择 OpenAI Compatible Provider 进行配置
- **Continue** -- 在 config.json 中添加 OpenAI 类型的模型配置
- **其他支持 OpenAI Compatible API 的工具**

配置时使用以下参数：

| 配置项 | 值 |
|--------|------|
| Base URL | `{{BASE_URL}}/v1` |
| API Key | `sk-xxxx` |
| Model | Gemini 模型名称，如 `gemini-3-pro-preview` |

注意：通过 OpenAI 兼容格式调用时，API Key 需要属于支持 `chat/completions` 格式的分组。

---

## 如何在 Cline 中使用 Gemini 模型

**问题**：想在 Cline 插件中使用 Gemini 模型，应如何配置。

**解决方案**：在 Cline 中按以下方式配置：

1. 打开 Cline 设置面板
2. **API Provider** 选择 `OpenAI Compatible`
3. 填写配置：
   - **Base URL**: `{{BASE_URL}}/v1`
   - **API Key**: `sk-xxxx`（填入你的实际 Key）
   - **Model ID**: `gemini-3-pro-preview`（或分组内其他可用的 Gemini 模型名称）
4. 保存设置后即可在 Cline 中使用 Gemini 模型

此方式通过 OpenAI 兼容的 `chat/completions` 接口调用，无需 Gemini 原生 API 配置。
