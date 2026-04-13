# Development Workflow Rules

> 此文件定义 LLM 开发工作流的强制规则。
> 所有 LLM 工具在执行任务时必须遵守，不可跳过任何步骤。

## Full Flow (MUST follow, no exceptions)

### feat (新功能)
1. 理解需求，分析影响范围
2. 读取现有代码，理解模式
3. 编写实现代码
4. 编写对应测试
5. 运行测试，修复失败
6. 更新文档（若 API 变更）
7. 自查 lint / type-check

### fix (缺陷修复)
1. 复现问题，确认症状
2. 定位根因
3. 编写失败测试（先有红灯）
4. 修复代码
5. 验证测试通过（变绿灯）
6. 回归测试

### refactor (重构)
1. 确保现有测试通过
2. 小步重构，每步可验证
3. 重构后测试必须全部通过
4. 不改变外部行为

## Context Logging (决策记录)

当你做出以下决策时，MUST 追加到 `.context/current/branches/<当前分支>/session.log`：

1. **方案选择**：选 A 不选 B 时，记录原因
2. **Bug 发现与修复**：根因 + 修复方法 + 教训
3. **API/架构决策**：接口设计选择
4. **放弃的方案**：为什么放弃

追加格式：

```
## <ISO-8601 时间>
**Decision**: <你选择了什么>
**Alternatives**: <被排除的方案>
**Reason**: <为什么>
**Risk**: <潜在风险>
```

## Project-Specific Rules

### SMPP Simulator Workflow

1. **代码修改后必须运行测试**
   ```bash
   # 后端改动
   cd backend && go test ./...

   # 前端改动
   cd frontend && pnpm test

   # E2E 测试（开发完成后）
   ./scripts/test-e2e.sh
   ./scripts/test-smpp.sh
   ```

2. **数据库相关变更**
   - 新增表：在 `backend/internal/repository/database.go` 的 `createTables()` 中添加
   - 必须支持 SQLite/PostgreSQL/MySQL 三种数据库
   - 占位符适配：SQLite `?`, PostgreSQL `$1`, MySQL `?`
   - 测试使用 PostgreSQL 容器（test-e2e.sh 自动启动）

3. **SMPP 协议相关**
   - 新增 PDU 类型：在 `backend/internal/smpp/pdu.go` 添加编码/解码
   - Session 状态管理：修改 `smpp/server.go` 时注意并发安全
   - 测试用 SMPP 客户端：`scripts/smpp_client.py`

4. **前端组件开发**
   - 新页面：在 `src/views/` 创建，添加到 `src/router/index.ts`
   - 新 API：在 `src/api/index.ts` 添加方法
   - 国际化：同时更新 `src/locales/zh.ts` 和 `src/locales/en.ts`
   - 状态管理：复杂状态用 Pinia store (`src/stores/`)

5. **Docker 相关**
   - 后端 Dockerfile: `backend/Dockerfile`
   - 前端 Dockerfile: `frontend/Dockerfile` (使用 pnpm)
   - docker-compose: 项目根目录 `docker-compose.yml`
   - 构建前测试：`docker-compose build` 前确保本地测试通过
