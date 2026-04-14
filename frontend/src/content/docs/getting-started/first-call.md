# 发起第一次调用

创建 API Key 后，可以通过以下示例验证接口是否正常工作。根据使用的模型系列选择对应的请求格式：OpenAI 兼容格式和 Anthropic 原生格式用于调用 Claude 模型，Gemini 原生格式用于调用 Gemini 模型。

## OpenAI 兼容格式

```bash
curl -X POST {{BASE_URL}}/v1/chat/completions \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 1024
  }'
```

## Anthropic 原生格式

```bash
curl -X POST {{BASE_URL}}/v1/messages \
  -H "x-api-key: sk-xxxx" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 1024
  }'
```

## Gemini 原生格式

```bash
curl -X POST "{{BASE_URL}}/v1beta/models/gemini-2.5-flash:generateContent?key=sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{"parts": [{"text": "你好"}]}]
  }'
```

将示例中的 `sk-xxxx` 替换为你实际创建的 API Key。

## 验证成功

如果返回了包含模型回复内容的 JSON 响应，说明 Key 配置正确、接口调用正常。接下来可以：

- 查看[「CLI 配置教程」]({{BASE_URL}}/docs/cli/claude-code)，将 Key 配置到 Claude Code、Cursor 等工具中
- 查看[「API 接入文档」]({{BASE_URL}}/docs/api/overview)，了解更多接口参数和用法

## 常见错误排查

如果调用失败，请依次检查以下几项：

1. **API Key 是否正确** -- 确认复制时没有遗漏或多余字符
2. **分组是否支持该模型** -- Key 所属分组必须包含你请求的模型，否则会返回模型不可用的错误
3. **余额是否充足** -- 余额为零时所有调用都会被拒绝，请先充值
4. **请求格式是否正确** -- 检查 URL 路径、请求头和 JSON 格式是否与示例一致
