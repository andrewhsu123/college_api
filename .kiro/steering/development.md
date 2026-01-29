# 开发工作流程

## develop/ 文件夹使用规则

### 核心原则

**develop/ 文件夹是只读参考资料，不能直接应用其中的代码**

### 正确使用方式

#### ✅ 可以做的事情

1. **查阅设计文档**
   - 阅读 `develop/docs/` 中的架构设计
   - 理解业务需求和技术方案
   - 参考 API 设计和数据流程

2. **参考数据库结构**
   - 查看 `develop/sql/` 中的表结构
   - 理解表关系和字段定义
   - 复制表结构到项目实际使用的 SQL 文件

3. **学习业务逻辑**
   - 理解各个服务的职责
   - 学习权限模型和缓存策略
   - 参考性能优化方案

#### ❌ 不能做的事情

1. **不能直接使用代码**
   - 不能直接引用 develop/ 中的代码文件
   - 不能将 develop/ 中的代码作为项目的一部分
   - 必须在项目根目录重新编写代码

2. **不能修改参考资料**
   - 不能修改 develop/ 中的文件
   - 保持参考资料的完整性
   - 如需调整，在实际代码中实现

3. **不能混淆参考和实现**
   - 实际代码必须在 cmd/, internal/, pkg/ 等标准目录
   - 不能在 develop/ 中创建新文件
   - 保持清晰的边界

## 开发步骤

### 第一步：需求分析

1. 阅读 `develop/docs/人员中心服务中台方案.md`
2. 查看 `develop/docs/services/` 中的服务设计
3. 理解业务场景和性能要求

### 第二步：数据库设计

1. 查看 `develop/sql/` 中的表结构
2. 在项目根目录创建 `sql/` 目录（如需要）
3. 复制并调整表结构，添加必要的索引和优化

```bash
# 示例：复制表结构
# 从 develop/sql/persons.sql 复制内容
# 创建到 sql/persons.sql 或直接使用
```

### 第三步：创建项目结构

```bash
# 创建标准 Go 项目目录
mkdir -p cmd/server
mkdir -p internal/{handler,service,repository,model,middleware}
mkdir -p pkg/{cache,database,response}
mkdir -p config
```

### 第四步：编写代码

1. **定义数据模型** (`internal/model/`)
   - 参考 develop/docs/database/schema.md
   - 使用 Go struct 定义模型
   - 添加 JSON 和数据库标签

2. **实现数据访问层** (`internal/repository/`)
   - 参考 develop/docs/services/ 中的数据操作
   - 实现 CRUD 操作
   - 添加缓存逻辑

3. **实现业务逻辑** (`internal/service/`)
   - 参考 develop/docs/services/ 中的业务规则
   - 实现权限计算、数据聚合等
   - 处理事务和错误

4. **实现 HTTP 处理器** (`internal/handler/`)
   - 参考 develop/docs/api/ 中的 API 设计
   - 实现路由和请求处理
   - 添加参数验证

### 第五步：测试和优化

1. 编写单元测试
2. 进行性能测试
3. 根据 develop/docs/architecture/performance.md 优化

## 代码复用策略

### 可以复制的内容

1. **SQL 表结构**
   ```sql
   -- 可以从 develop/sql/ 复制表定义
   CREATE TABLE persons (
     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
     ...
   );
   ```

2. **配置模板**
   - 数据库连接配置
   - Redis 配置
   - 服务端口配置

3. **常量定义**
   - 错误码
   - 状态码
   - 缓存键前缀

### 必须重写的内容

1. **业务逻辑代码**
   - 所有 Go 代码必须重新编写
   - 不能直接复制粘贴
   - 根据实际需求调整

2. **API 实现**
   - 参考设计文档
   - 使用 Gin 框架实现
   - 遵循 RESTful 规范

3. **测试代码**
   - 根据实际实现编写测试
   - 确保覆盖核心逻辑

## 参考文档索引

### 架构设计
- `develop/docs/architecture/overview.md` - 系统架构
- `develop/docs/architecture/caching.md` - 缓存策略
- `develop/docs/architecture/performance.md` - 性能优化

### 服务设计
- `develop/docs/services/PersonService.md` - 人员服务
- `develop/docs/services/DepartmentService.md` - 部门服务
- `develop/docs/services/RoleService.md` - 角色服务
- `develop/docs/services/PermissionService.md` - 权限服务
- `develop/docs/services/SearchService.md` - 搜索服务

### 数据库设计
- `develop/docs/database/schema.md` - 数据库模式
- `develop/sql/*.sql` - 表结构定义

### 实施计划
- `develop/docs/deployment/implementation-plan.md` - 实施计划
- `develop/docs/deployment/risks.md` - 风险评估
- `develop/docs/项目完成计划表.md` - 项目计划

## 常见问题

### Q: 为什么不能直接使用 develop/ 中的代码？

A: develop/ 是参考资料和设计文档，可能包含示例代码或伪代码。实际项目需要：
- 遵循项目特定的编码规范
- 适配实际的技术栈和框架
- 根据实际需求调整实现
- 保持代码的可维护性和一致性

### Q: 如何快速开始开发？

A: 
1. 先阅读 `develop/docs/人员中心服务中台方案.md` 了解全局
2. 查看 `develop/docs/services/` 了解各服务职责
3. 从最简单的服务开始实现（如 PersonService）
4. 逐步添加缓存、权限等复杂功能

### Q: 数据库脚本可以直接使用吗？

A: 可以复制表结构，但建议：
- 检查字段类型是否符合实际需求
- 添加必要的索引优化
- 根据实际数据量调整表设计
- 在测试环境先验证

### Q: 如何保持代码和文档的一致性？

A: 
- 实现功能时参考对应的设计文档
- 如果实现与设计有差异，记录原因
- 重大变更时更新 README.md
- 保持代码注释与设计意图一致
