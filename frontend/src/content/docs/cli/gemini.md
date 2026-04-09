# Gemini CLI 配置

本文档介绍如何配置 Google Gemini CLI 以连接本平台的 API 服务。

## 前置条件

1. 已安装 Node.js 环境
2. 安装 Gemini CLI：
```bash
npm i -g @google/gemini-cli@latest
```
3. 首次运行 `gemini` 命令以生成配置目录

## 配置方法

创建或编辑 `.env` 文件，根据操作系统找到对应路径：

- **macOS / Linux**: `~/.gemini/.env`
- **Windows**: `%userprofile%\.gemini\.env`

将以下内容写入该文件：

```
GOOGLE_GEMINI_BASE_URL={{BASE_URL}}
GEMINI_API_KEY=sk-xxxx
GEMINI_MODEL=gemini-2.5-pro
```

将 `sk-xxxx` 替换为你的 API Key。请确保该 Key 属于 Gemini 分组。

## 模型选择

根据你的分组内可用模型，修改 `GEMINI_MODEL` 字段。常见可选模型包括：

- `gemini-2.5-pro`
- `gemini-2.5-flash`
- `gemini-3-pro-preview`

## 验证配置

在终端中运行 `gemini`，若能正常启动并交互，则配置成功。

## 环境变量方式（替代方案）

除了编辑 `.env` 文件，也可以直接在终端中设置环境变量：

```bash
export GOOGLE_GEMINI_BASE_URL="{{BASE_URL}}"
export GEMINI_API_KEY="sk-xxxx"
export GEMINI_MODEL="gemini-2.5-pro"
gemini
```

该方式仅对当前终端会话有效。

## 说明

本平台支持 Gemini 原生 API 格式（`/v1beta/models/{model}:generateContent`），Gemini CLI 默认使用该格式，无需额外适配。

## 常见问题排查

### 模型调用失败

- 确认你的 API Key 所属分组中包含 Gemini 系列模型
- 检查 `GEMINI_MODEL` 填写的模型名称是否与分组内可用模型一致

### 配置不生效

- 确认 `.env` 文件位于正确的目录下（`~/.gemini/.env`）
- 检查文件内容格式是否正确，每行一个配置项，等号两侧不要有多余空格
