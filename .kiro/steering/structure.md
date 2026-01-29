# 项目结构

## 文档组织

所有文档按服务和类别组织在 `docs/` 目录中：

### 服务文档 (`docs/services/`)

每个微服务都有自己的文档文件：

- **PersonService.md** - 人员管理服务（CRUD、搜索、角色管理）
- **DepartmentService.md** - 组织/部门服务（树查询、嵌套集合）
- **RoleService.md** - 角色管理服务（基于JSON的权限范围）
- **PermissionService.md** - 权限计算和验证服务
- **SearchService.md** - 基于ElasticSearch的搜索服务
- **StudentService.md** - 学生特定操作
- **StaffService.md** - 员工特定操作

### 架构文档 (`docs/architecture/`)

系统级架构和设计：

- **overview.md** - 整体系统架构、微服务设计、数据流
- **caching.md** - Redis缓存策略、缓存失效、多层缓存
- **performance.md** - 性能目标、优化技术、负载测试

### 数据库文档 (`docs/database/`)

数据库模式和优化：

- **schema.md** - 表结构、关系、索引策略

### 部署文档 (`docs/deployment/`)

实施和运维：

- **implementation-plan.md** - 7周分阶段实施计划
- **risks.md** - 风险评估和缓解策略

### API文档 (`docs/api/`)

每个服务的API规范（当前为空，待填充）

## 代码组织

```
.
├── docs/                    # 所有文档
│   ├── services/           # 服务特定文档
│   ├── architecture/       # 系统架构
│   ├── database/           # 数据库模式
│   ├── deployment/         # 实施计划
│   └── api/               # API规范
├── .kiro/                  # Kiro配置
│   └── steering/          # 规则指引
├── *.sql                   # 数据库模式文件
├── main.go                 # 应用程序入口
└── README.md              # 项目概述

```

## 在此项目上工作时

1. **添加新服务**：在 `docs/services/[ServiceName].md` 中创建文档
2. **修改架构**：更新 `docs/architecture/` 中的相关文件
3. **数据库变更**：更新 `docs/database/schema.md` 并创建迁移SQL文件
4. **API变更**：在 `docs/api/[ServiceName].md` 中记录

## 关键原则

- 保持服务文档专注于该服务的职责
- 交叉引用相关服务而不是重复信息
- 进行系统级更改时更新架构文档
- 业务上下文使用中文文档，技术结构使用英文
