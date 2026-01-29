# 性能优化与指标

## 性能目标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 人员详情查询 | < 50ms | 1000+ | Redis 1小时 |
| 人员列表查询 | < 100ms | 500+ | 数据库查询 |
| 机构树查询 | < 50ms | 1000+ | Redis 1小时 |
| 权限校验 | < 20ms | 2000+ | Redis 30分钟 |
| 管辖人员查询 | < 100ms | 500+ | Redis 5分钟 |
| 搜索接口 | < 200ms | 200+ | ES |

## 优化手段

### 1. 缓存优化

- Redis 缓存命中率 > 95%
- 多级缓存：本地缓存（Caffeine）+ Redis
- 缓存预热：系统启动时加载热点数据
- 缓存穿透防护：布隆过滤器

### 2. 数据库优化

- 读写分离：写主库，读从库
- 索引优化：覆盖索引、联合索引
- 慢查询优化：< 100ms
- 连接池优化：合理配置连接数

### 3. 查询优化

- 批量查询：减少数据库往返次数
- 分页优化：使用游标分页
- 预加载：避免 N+1 查询
- 异步查询：非关键数据异步加载

### 4. 接口优化

- 接口限流：令牌桶算法
- 接口降级：非核心功能降级
- 接口熔断：防止雪崩
- 异步处理：批量操作异步执行

## 性能优化技巧

### 批量查询优化

```php
// 不好的做法：N+1查询
$persons = Person::where('customer_id', 1)->get();
foreach ($persons as $person) {
    $person->student = Student::where('person_id', $person->id)->first();
}

// 好的做法：预加载
$persons = Person::with('student')->where('customer_id', 1)->get();
```

### 分页查询优化

```sql
-- 不好的做法：OFFSET 很大时性能差
SELECT * FROM persons
WHERE customer_id = 1
ORDER BY id
LIMIT 10000, 20;

-- 好的做法：使用上次查询的最大ID
SELECT * FROM persons
WHERE customer_id = 1
  AND id > 10000  -- 上次查询的最大ID
ORDER BY id
LIMIT 20;
```

### 索引覆盖查询

```sql
-- 只查询需要的字段，利用索引覆盖
SELECT id, name, mobile
FROM persons
WHERE customer_id = 1
  AND person_type = 1
  AND status = 1;

-- 确保有覆盖索引
ALTER TABLE persons
ADD INDEX idx_cover (customer_id, person_type, status, id, name, mobile);
```

## 压测方案

### 压测工具

JMeter / Gatling / wrk

### 压测场景

#### 场景1：查询管辖人员（核心场景）

- 并发用户：1000
- 持续时间：10分钟
- 预期 P99：< 100ms

#### 场景2：人员搜索

- 并发用户：500
- 持续时间：5分钟
- 预期 P99：< 200ms

#### 场景3：机构树查询

- 并发用户：2000
- 持续时间：5分钟
- 预期 P99：< 50ms

#### 场景4：混合场景

- 查询管辖人员：50%
- 人员搜索：20%
- 机构树查询：20%
- 人员详情：10%
- 并发用户：1000
- 持续时间：30分钟

## 监控和日志

### 监控指标

- 接口响应时间（P50、P95、P99）
- 接口 QPS
- 缓存命中率
- 数据库慢查询
- ES 查询性能
- 服务健康状态

### 日志收集

```
应用日志 → Filebeat → Logstash → Elasticsearch → Kibana
```

### 告警规则

- 接口 P99 > 500ms
- 缓存命中率 < 90%
- 数据库慢查询 > 100ms
- 服务不可用
