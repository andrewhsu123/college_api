# 获取登录信息

## 学校登录
127.0.0.1:8081/api/base/info
Authorization:Bearer 111|cxlho8PuMFjSqUQmWiecuhG2cc3YdVUZ0TK9xF4568f3b872

## 政工登录
127.0.0.1:8081/api/staff/info
Authorization:Bearer 112|au4XBYmV0w8IjGD0KviuLsv4X1oBHyJ1LPAXdpIM259050c6

# 获取信息

## 学校信息
登录后得到学校用户id  即  (user_id=1)
根据user_id 可以得到学校id 即 (customer_id=1)
select customer_id from admin_users where id=1;

## 政工信息
登录后得到政工人员id  即  (person_id=9)
根据person_id 可以在persons表得到 即 (customer_id=1)
select customer_id from persons where id=9;

当前政工可见的部门id
select department_ids 
from persons_roles 
where customer_id=1 and id in (select role_id from persons_has_roles where person_id=9)
查询结果eg:
```
[55,64,56,61,20,46]
[55]
```

# 获取组织机构
树形的第一级为：学校
树形的第二级为：查询的行政机构和组织机构

其中:
-- 查询学校机构
select id, parent_id, recommend_num, department_name, tree_level
from departments where customer_id = 1 and department_type=0;

-- 查询行政机构
select id, parent_id, recommend_num, department_name, tree_level
from departments where customer_id = 1 and tree_level > 2 and department_type=1;

将查出的数据 tree_level=3 的都放到
学校的 items下
eg:
[
  'id'                  => 1,
  'parent_id'           => 0,
  'recommend_num'       => 555,
  'department_name'     => '学校',
  'tree_level'          => 1,
  'items'=>[
    'id'              => 5,
    'parent_id'       => 2,
    'recommend_num'   => 44,
    'department_name' => '学术委员会办公室',
    'tree_level'  => 3,   // tree_level=3 的固定放在 tree_level = 1 的下面
    'items'=>[
      'id' => 6
      'parent_id'   => 5, // tree_level>3时 parent_id 必须等于升级的id
      'recommend_num'   => 44, // recommend_num=0说明没有下级了
    ],
  ],
]

-- 查询组织机构
select id, parent_id, recommend_num, department_name, tree_level
from departments where customer_id = 1 and tree_level > 2 and department_type!=1;

将查出的数据 tree_level=3 的都放到
学校的 items下
eg:
[
  'id'                  => 1,
  'parent_id'           => 0,
  'recommend_num'       => 555,
  'department_name'     => '学校',
  'tree_level'          => 1,
  'items'=>[
    'id'              => 5,
    'parent_id'       => 2,
    'recommend_num'   => 44,
    'department_name' => '计算机科学与技术学院',
    'department_type' => 2, // 2=学院 3=系 4=专业 5=班级
    'tree_level'  => 3,   // tree_level=3 的固定放在 tree_level = 1 的下面
    'items'=>[
      'id' => 6
      'parent_id'   => 5, // tree_level>3时 parent_id 必须等于升级的id
      'recommend_num'   => 44, // recommend_num=0说明没有下级了
    ],
  ],
]

机构类型:0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级',

机构服务提供两种查询
1、查询树形服务
2、机构名称模糊查询、机构类型department_type查询

政工树形：
根据政工组织机构id权限显示可见树形

根据政工组织机构id权限显示搜索结果