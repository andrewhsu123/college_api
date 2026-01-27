# 人员中心服务中台

基于现有数据库表结构的人员中心服务中台解决方案。

## 📁 文件说明

### 核心文档
- **人员中心服务中台方案.md** - 完整的技术方案文档，包含架构设计、实现细节、部署方案

### 数据库文件
- **departments.sql** - 机构表结构
- **persons.sql** - 人员基础表结构
- **persons_roles.sql** - 角色表结构（包含 JSON 字段存储管辖范围）
- **persons_has_roles.sql** - 用户角色关联表
- **role_has_departments.sql** - 角色机构关联表
- **students.sql** - 学生表结构
- **staff.sql** - 政工表结构

### 优化脚本
- **database_optimization.sql** - 数据库优化脚本（添加索引和主键）
- **database_rollback.sql** - 优化回滚脚本（如果出现问题可回滚）

## 🚀 快速开始

### 1. 数据库优化

**执行前必读：**
- ⚠️ 请先备份数据库！
- ⚠️ 建议在测试环境先执行验证
- ⚠️ 生产环境建议在业务低峰期执行

**执行优化脚本：**
```bash
mysql -u root -p college_db_base < database_optimization.sql
```

**如果需要回滚：**
```bash
mysql -u root -p college_db_base < database_rollback.sql
```

### 2. 验证优化结果

```sql
-- 查看 persons_has_roles 表结构
SHOW CREATE TABLE persons_has_roles;

-- 查看索引使用情况
SHOW INDEX FROM persons;
SHOW INDEX FROM students;
SHOW INDEX FROM staff;
```

## 📊 核心功能

### 1. 权限管理
- 基于角色的权限控制（RBAC）
- 角色管辖机构和人员（JSON 格式存储）
- 灵活的权限配置

### 2. 人员管理
- 统一的人员基础信息管理
- 学生信息管理（单一机构归属）
- 政工信息管理（单一机构归属）

### 3. 机构管理
- 预排序遍历树（Nested Set）
- 高效的树查询（子树、路径）
- 机构人员查询

### 4. 核心查询场景
- **查询我管辖的人员**（最频繁的业务场景）
  - 用户 → 角色 → 管辖范围（机构/人员）
  - 多级缓存优化
  - P99 < 100ms

## 🏗️ 架构设计

### 系统分层
```
业务应用层
    ↓
API Gateway
    ↓
微服务层（人员/机构/角色/权限/搜索）
    ↓
数据层（MySQL/Redis/ES/MQ）
```

### 技术栈
- **数据库**：MySQL 5.7+（主从复制）
- **缓存**：Redis（集群）
- **搜索**：ElasticSearch（集群）
- **消息队列**：RabbitMQ / Kafka
- **开发语言**：PHP / Java / Go（根据团队技术栈选择）

## 📈 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 人员详情查询 | < 50ms | 1000+ | Redis 1小时 |
| 管辖人员查询 | < 100ms | 500+ | Redis 5分钟 |
| 机构树查询 | < 50ms | 1000+ | Redis 1小时 |
| 权限校验 | < 20ms | 2000+ | Redis 30分钟 |
| 搜索接口 | < 200ms | 200+ | ES |

## 🔑 核心优化点

### 1. 数据库优化
- ✅ 添加必要的主键和索引
- ✅ 组合索引优化常见查询
- ✅ 预排序遍历树（Nested Set）
- ✅ 读写分离

### 2. 缓存策略
- ✅ 用户角色缓存（30分钟）
- ✅ 角色管辖范围缓存（1小时）
- ✅ 机构下人员缓存（10分钟）
- ✅ 人员信息缓存（1小时）
- ✅ 管辖人员结果缓存（5分钟）

### 3. 查询优化
- ✅ 批量查询减少数据库往返
- ✅ 预加载避免 N+1 查询
- ✅ 索引覆盖查询
- ✅ 游标分页

## 📝 实施步骤

### 阶段一：数据库优化（1周）
- [x] 编写优化脚本
- [ ] 测试环境验证
- [ ] 生产环境执行
- [ ] 性能对比测试

### 阶段二：基础服务开发（2周）
- [ ] PersonService（人员服务）
- [ ] DepartmentService（机构服务）
- [ ] RoleService（角色服务）
- [ ] StudentService（学生服务）
- [ ] StaffService（政工服务）

### 阶段三：权限服务开发（1周）
- [ ] 权限计算逻辑
- [ ] 权限校验接口
- [ ] 管辖人员/机构查询

### 阶段四：搜索服务开发（1周）
- [ ] ES 索引设计
- [ ] 数据同步（MQ）
- [ ] 搜索接口

### 阶段五：性能优化和压测（1周）
- [ ] 缓存预热
- [ ] SQL 优化
- [ ] 压测调优

### 阶段六：联调和上线（1周）
- [ ] 业务系统联调
- [ ] 数据迁移
- [ ] 灰度发布
- [ ] 全量上线

**总计：7周**

## 🔍 核心查询示例

### 查询用户管辖的人员

```php
// 1. 获取用户角色
$roleIds = getUserRoles($personId);

// 2. 获取角色管辖范围
foreach ($roleIds as $roleId) {
    $scope = getRoleScope($roleId);
    $departmentIds = json_decode($scope['department_ids'], true);
    $personIds = json_decode($scope['person_ids'], true);
}

// 3. 根据机构查询人员
$students = DB::table('students')
    ->whereIn('class_id', $departmentIds)
    ->pluck('person_id');

$staff = DB::table('staff')
    ->whereIn('department_id', $departmentIds)
    ->pluck('person_id');

// 4. 合并结果
$managedPersonIds = array_unique(array_merge(
    $students->toArray(),
    $staff->toArray(),
    $personIds
));
```

### 查询机构树

```sql
-- 查询某个机构的所有子机构
SELECT id, department_name, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_left > (SELECT tree_left FROM departments WHERE id = 55)
  AND tree_right < (SELECT tree_right FROM departments WHERE id = 55)
  AND deleted_at = 0
ORDER BY tree_left;
```

## 🛡️ 风险和应对

### 技术风险
1. **JSON 字段查询性能** → 使用缓存 + 关联表辅助
2. **缓存一致性** → 延迟双删 + 合理过期时间
3. **主从延迟** → 实时查询走主库
4. **ES 同步延迟** → 监控 MQ + 提示用户

### 业务风险
1. **数据迁移失败** → 充分测试 + 备份
2. **权限计算错误** → 单元测试 + 灰度发布
3. **性能不达标** → 提前压测 + 降级方案

## 📞 联系方式

如有问题，请查看 `人员中心服务中台方案.md` 详细文档。

## 📄 License

内部项目，仅供团队使用。
