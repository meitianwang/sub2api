# Claude Code 常见问题

本文档汇总 Claude Code 使用过程中的常见问题及解决方案。配置教程请参阅「CLI 配置教程 - Claude Code」。

---

## "Unable to connect to Anthropic services" 错误

**问题**：首次启动 Claude Code 时提示无法连接 Anthropic 服务。

**解决方案**：需要在 `~/.claude.json` 中将 `hasCompletedOnboarding` 设为 `true`。

**macOS / Linux**:

```bash
jq '. + {"hasCompletedOnboarding": true}' ~/.claude.json > /tmp/tmp.json && mv /tmp/tmp.json ~/.claude.json
```

**Windows (PowerShell)**:

```powershell
powershell -Command "$f='%USERPROFILE%\.claude.json';$j=Get-Content $f|ConvertFrom-Json;$j|Add-Member -NotePropertyName 'hasCompletedOnboarding' -NotePropertyValue $true -Force;$j|ConvertTo-Json|Set-Content $f"
```

执行完成后重新启动 Claude Code 即可。

---

## 401 Unauthorized 错误

**问题**：调用时返回 401 认证错误。

**解决方案**：依次检查以下几项：

1. **API Key 是否正确** -- 确认 `settings.json` 中的 `ANTHROPIC_AUTH_TOKEN` 值无误，复制时没有遗漏或多余字符
2. **分组类型是否匹配** -- Claude Code 需要使用 Anthropic / Claude 类型分组的 Key，使用其他类型分组的 Key 会导致认证失败
3. **Key 是否处于启用状态** -- 登录本平台确认该 Key 未被禁用或删除
4. **余额是否充足** -- 余额为零时所有调用都会被拒绝

---

## 如何在 VSCode 插件中使用

**问题**：通过 VSCode 扩展使用 Claude Code 时如何配置。

**解决方案**：VSCode 扩展使用与终端版相同的配置文件，无需额外设置。编辑 `settings.json` 即可：

- **macOS / Linux**: `~/.claude/settings.json`
- **Windows**: `%userprofile%\.claude\settings.json`

配置内容：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{BASE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxxx",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

保存后重新打开 VSCode 中的 Claude Code 面板即可生效。

---

## 如何切换到 200K 上下文模式

**问题**：Claude Code 默认使用 1M 上下文，如何切换到 200K 模式以节省 token 消耗。

**解决方案**：在 `settings.json` 的 `env` 中添加以下配置项：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{BASE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxxx",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_DISABLE_1M_CONTEXT": "1"
  }
}
```

添加 `CLAUDE_CODE_DISABLE_1M_CONTEXT` 并设为 `"1"` 后，Claude Code 将使用 200K 上下文窗口。

---

## 常用命令一览

**问题**：Claude Code 有哪些常用命令。

**解决方案**：以下是常用命令列表：

| 命令 | 说明 |
|------|------|
| `claude` | 启动交互式会话 |
| `claude -p "提示词"` | 非交互模式，直接输出结果 |
| `claude -c` | 恢复上次会话继续对话 |
| `claude update` | 更新 Claude Code 到最新版本 |
| `claude mcp` | 管理 MCP 服务器配置 |
| `claude --model 模型名` | 指定使用的模型 |

---

## 上下文过长导致回复质量下降

**问题**：长时间对话后，上下文窗口占用过多，模型回复质量明显下降。

**解决方案**：

- 使用 `/compact` 命令压缩当前对话上下文，释放窗口空间
- 如果对话已经严重偏离主题或上下文过于庞大，使用 `/clear` 清空上下文并开始新会话
- 建议在上下文占用超过 60% 时主动执行 `/compact`，避免等到接近上限时才处理
