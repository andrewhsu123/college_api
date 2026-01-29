# 项目结构

## 当前项目组织

这是一个基于 Go 的人员中心服务中台项目，采用标准 Go 项目布局。

### 根目录结构

```
.
├── cmd/                    # 应用程序入口（待创建）
├── internal/              # 私有应用代码（待创建）
│   ├── handler/          # HTTP处理器
│   ├── service/          # 业务逻辑
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   └── middleware/       # 中间件
├── pkg/                   # 可复用的公共库（待创建）
├── config/                # 配置文件（待创建）
├── develop/               # 📚 开发参考资料（只读）
│   ├── docs/             # 设计文档和方案
│   └── sql/              # 数据库表结构参考
├── .kiro/                 # Kiro配置
│   └── steering/         # 开发规则指引
├── main.go                # 当前应用入口
├── go.mod                 # Go模块定义
└── README.md             # 项目说明
```

### develop/ 文件夹说明（参考资料）

**重要：develop/ 文件夹是只读参考资料，不能直接使用其中的代码**

#### 文档资料 (`develop/docs/`)

- **人员中心服务中台方案.md** - 完整技术方案
- **项目完成计划表.md** - 实施计划
- **services/** - 各微服务设计文档
  - PersonService.md, DepartmentService.md, RoleService.md 等
- **architecture/** - 架构设计文档
  - overview.md, caching.md, performance.md
- **database/** - 数据库设计文档
  - schema.md
- **deployment/** - 部署文档
  - implementation-plan.md, risks.md

#### 数据库参考 (`develop/sql/`)

- departments.sql - 机构表结构
- persons.sql - 人员基础表
- persons_roles.sql - 角色表
- persons_has_roles.sql - 用户角色关联表
- role_has_departments.sql - 角色机构关联表
- students.sql - 学生表
- staff.sql - 政工表

### 实际开发目录（待创建）

当开始实际开发时，应创建以下标准 Go 项目结构：

```
cmd/
  └── server/
      └── main.go          # 服务启动入口

internal/
  ├── handler/             # HTTP处理器
  │   ├── person.go
  │   ├── department.go
  │   └── role.go
  ├── service/             # 业务逻辑层
  │   ├── person_service.go
  │   ├── department_service.go
  │   └── role_service.go
  ├── repository/          # 数据访问层
  │   ├── person_repo.go
  │   └── department_repo.go
  ├── model/               # 数据模型
  │   ├── person.go
  │   ├── department.go
  │   └── role.go
  └── middleware/          # 中间件
      ├── auth.go
      └── logger.go

pkg/
  ├── cache/               # 缓存工具
  ├── database/            # 数据库工具
  └── response/            # 响应封装

config/
  └── config.yaml          # 配置文件
```

## 开发工作流程

### 1. 查阅参考资料
- 从 `develop/docs/` 查看设计文档
- 从 `develop/sql/` 查看数据库表结构
- **不要直接修改或使用 develop/ 中的文件**

### 2. 创建实际代码
- 在项目根目录创建标准 Go 项目结构
- 参考 develop/ 中的设计，但需要重新编写代码
- 遵循 tech.md 中的编码规范

### 3. 数据库脚本
- 可以从 `develop/sql/` 复制表结构
- 在项目根目录或 `sql/` 目录创建实际使用的脚本
- 根据需要添加优化和索引

## 关键原则

1. **develop/ 是参考资料库**
   - 只读，不修改
   - 查阅设计和方案
   - 理解业务逻辑

2. **实际代码在项目根目录**
   - 遵循标准 Go 项目布局
   - 代码需要重新编写，不能直接复制
   - 保持代码简洁和可维护

3. **文档和代码分离**
   - 设计文档在 develop/docs/
   - 实际代码在 internal/, cmd/, pkg/
   - README.md 保持简洁，指向详细文档

4. **中文业务，英文代码**
   - 业务文档使用中文
   - 代码、注释、变量名使用英文
   - 用户界面文本使用中文
