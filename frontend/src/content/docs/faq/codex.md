# Codex 常见问题

本文档汇总 Codex CLI 使用过程中的常见问题及解决方案。配置教程请参阅「CLI 配置教程 - Codex CLI」。

---

## 401 认证错误

**问题**：运行 Codex 时返回 401 错误。

**解决方案**：依次检查以下几项：

1. **auth.json 中的 Key 是否正确** -- 确认 `~/.codex/auth.json` 中的 `OPENAI_API_KEY` 值填写正确
2. **环境变量是否冲突** -- 系统环境变量中可能已经设置了 `OPENAI_API_KEY`，该值会覆盖 `auth.json` 中的配置。可通过以下命令检查并清除：

```bash
# 检查是否存在环境变量
echo $OPENAI_API_KEY

# 如果存在，在当前会话中取消设置
unset OPENAI_API_KEY
```

如果环境变量确实存在且指向其他服务，建议将其从 shell 配置文件（如 `~/.bashrc`、`~/.zshrc`）中移除，避免与 Codex 配置冲突。

---

## 403 "Usage not included in your plan" 错误

**问题**：调用时返回 403 错误，提示用量不在当前计划内。

**解决方案**：

1. 该错误偶尔出现时，可以先重试几次
2. 如果持续出现，请检查：
   - 账户余额是否充足
   - 所属分组是否仍然处于可用状态
   - Key 是否被禁用
3. 登录本平台确认分组和 Key 的状态正常

---

## Windows 下字符编码异常 / 乱码

**问题**：在 Windows 终端中运行 Codex 时出现中文乱码或特殊字符显示异常。

**解决方案**：需要启用系统级 UTF-8 支持：

1. 打开 **系统设置**
2. 进入 **区域** (Region)
3. 点击 **管理** (Administrative) 选项卡
4. 点击 **更改系统区域设置** (Change system locale)
5. 勾选 **使用 Unicode UTF-8 提供全球语言支持** (Beta: Use Unicode UTF-8 for worldwide language support)
6. 重启计算机使设置生效

---

## 如何启用 WebSocket 模式

**问题**：如何配置 WebSocket 模式以获得更好的流式传输性能。

**解决方案**：在 `~/.codex/config.toml` 中修改 `[model_providers.custom]` 部分并添加 `[features]` 配置：

```toml
model_provider = "custom"
model = "gpt-5.4"
model_reasoning_effort = "high"
disable_response_storage = true
network_access = "enabled"

[model_providers.custom]
name = "custom"
base_url = "{{BASE_URL}}/v1"
wire_api = "responses"
supports_websockets = true
requires_openai_auth = true

[features]
responses_websockets_v2 = true
```

关键是添加 `supports_websockets = true` 和 `[features]` 下的 `responses_websockets_v2 = true`。

---

## 高效使用 Codex 的建议

**问题**：如何更高效地使用 Codex CLI。

**解决方案**：

1. **拆分任务** -- 将大型任务拆分为小模块逐步完成，避免单次对话承载过多内容
2. **控制上下文占用** -- 避免上下文使用率超过 60%，接近上限时模型理解能力会明显下降
3. **及时压缩上下文** -- 当对话较长时，使用 `/compact` 命令压缩上下文释放空间
4. **善用命令** -- 使用 `/diff` 查看变更、`/review` 审查代码、`/undo` 撤销操作，提高工作效率
5. **选择合适的模型** -- 根据任务复杂度选择模型，简单任务不必使用最大规格的模型
