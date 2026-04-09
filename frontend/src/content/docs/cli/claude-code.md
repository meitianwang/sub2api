# Claude Code 配置

本文档介绍如何配置 Claude Code 以连接本平台的 API 服务。

## 前置条件

1. 已安装 Node.js 环境
2. 安装 Claude Code：
```bash
npm i -g @anthropic-ai/claude-code@latest
```
3. 首次运行 `claude` 命令以生成配置目录

## 配置方法

编辑 `settings.json` 文件，根据操作系统找到对应路径：

- **macOS / Linux**: `~/.claude/settings.json`
- **Windows**: `%userprofile%\.claude\settings.json`

将以下内容写入该文件：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{BASE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-xxxx",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

将 `sk-xxxx` 替换为你的 API Key。请确保该 Key 属于 Anthropic / Claude 分组。

## 验证配置

在终端中运行 `claude`，若能正常启动并交互，则配置成功。

## 常见问题排查

### "Unable to connect to Anthropic services" 错误

需要在 `~/.claude.json` 中将 `hasCompletedOnboarding` 设为 `true`：

**macOS / Linux**:
```bash
jq '. + {"hasCompletedOnboarding": true}' ~/.claude.json > /tmp/tmp.json && mv /tmp/tmp.json ~/.claude.json
```

**Windows**:
```powershell
powershell -Command "$f='%USERPROFILE%\.claude.json';$j=Get-Content $f|ConvertFrom-Json;$j|Add-Member -NotePropertyName 'hasCompletedOnboarding' -NotePropertyValue $true -Force;$j|ConvertTo-Json|Set-Content $f"
```

### 401 认证错误

- 检查 API Key 是否正确填写
- 确认该 Key 所属分组为 Claude / Anthropic 类型

## 环境变量方式（替代方案）

除了编辑 `settings.json`，也可以直接在终端中设置环境变量：

```bash
export ANTHROPIC_BASE_URL="{{BASE_URL}}"
export ANTHROPIC_AUTH_TOKEN="sk-xxxx"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
```

该方式仅对当前终端会话有效。

## VSCode 插件

如果你通过 VSCode 扩展使用 Claude Code，同样适用上述 `settings.json` 配置方式，无需额外设置。

## 提示

创建 API Key 后，点击"使用"按钮可获取预填充的配置内容，方便快速配置。
