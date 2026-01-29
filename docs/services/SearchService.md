# SearchService - 搜索服务

## 职责

- ElasticSearch 索引管理
- 全文搜索
- 复杂条件查询

## ElasticSearch 索引设计

### 索引配置

```json
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "ik_analyzer": {
          "type": "custom",
          "tokenizer": "ik_max_word"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "person_id": {"type": "long"},
      "customer_id": {"type": "long"},
      "person_type": {"type": "byte"},
      "name": {
        "type": "text",
        "analyzer": "ik_max_word",
        "fields": {
          "keyword": {"type": "keyword"}
        }
      },
      "mobile": {"type": "keyword"},
      "email": {"type": "keyword"},
      "gender": {"type": "byte"},
      "status": {"type": "byte"},
      "student_no": {"type": "keyword"},
      "staff_no": {"type": "keyword"},
      "class_id": {"type": "long"},
      "department_id": {"type": "long"},
      "college_id": {"type": "long"},
      "profession_id": {"type": "long"},
      "created_at": {"type": "long"},
      "updated_at": {"type": "long"}
    }
  }
}
```

## 使用场景

- 姓名模糊搜索（支持中文分词）
- 手机号/邮箱精确搜索
- 学号/工号精确搜索
- 多条件组合查询（姓名 + 机构 + 状态）
- 在管辖范围内搜索人员

## 数据同步

### 使用消息队列异步同步

```php
/**
 * 人员数据变更时发送MQ消息
 */
function onPersonChanged($personId, $action) {
    $message = [
        'person_id' => $personId,
        'action' => $action,  // create, update, delete
        'timestamp' => time(),
    ];
    
    // 发送到 RabbitMQ/Kafka
    MessageQueue::publish('person.sync.es', json_encode($message));
}
```

### MQ消费者：同步数据到ES

```php
function syncPersonToES($message) {
    $personId = $message['person_id'];
    $action = $message['action'];
    
    if ($action === 'delete') {
        ES::delete('persons', $personId);
        return;
    }
    
    // 查询完整的人员数据
    $person = DB::table('persons')->find($personId);
    if (!$person) {
        return;
    }
    
    // 查询学生/政工信息
    $extra = [];
    if ($person->person_type == 1) {
        $student = DB::table('students')->where('person_id', $personId)->first();
        if ($student) {
            $extra = [
                'student_no' => $student->student_no,
                'class_id' => $student->class_id,
                'college_id' => $student->college_id,
                'profession_id' => $student->profession_id,
            ];
        }
    } else if ($person->person_type == 2) {
        $staff = DB::table('staff')->where('person_id', $personId)->first();
        if ($staff) {
            $extra = [
                'staff_no' => $staff->staff_no,
                'department_id' => $staff->department_id,
                'college_id' => $staff->college_id,
            ];
        }
    }
    
    // 合并数据
    $doc = array_merge((array)$person, $extra);
    
    // 索引到ES
    if ($action === 'create') {
        ES::index('persons', $personId, $doc);
    } else {
        ES::update('persons', $personId, $doc);
    }
}
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 说明 |
|---------|-------------|-----|------|
| 搜索接口 | < 200ms | 200+ | ES查询 |

## 部署配置

### ElasticSearch 集群

```
┌─────────┐  ┌─────────┐  ┌─────────┐
│  Node1  │  │  Node2  │  │  Node3  │
│ (Master)│  │  (Data) │  │  (Data) │
└─────────┘  └─────────┘  └─────────┘
```

- 3个节点以上
- 索引分片：3个主分片 + 1个副本
- 定期备份快照
