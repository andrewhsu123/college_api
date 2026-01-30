# College API 部署文档

## 部署架构

```
┌─────────────────────────────────────────────────────────┐
│                    同一服务器                              │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │staff_member  │  │catch_admin   │  │college_api   │  │
│  │_api (Laravel)│  │_base(Laravel)│  │(Go)          │  │
│  │Port: 8000    │  │Port: 8001    │  │Port: 8080    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │           │
│         └─────────────────┴─────────────────┘           │
│                           │                             │
│                  ┌────────▼────────┐                    │
│                  │  MySQL Database │                    │
│                  │  college_dev_base│                   │
│                  └─────────────────┘                    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 环境要求

### 系统要求
- 操作系统: Linux (推荐 Ubuntu 20.04+) 或 Windows Server
- CPU: 2核心以上
- 内存: 4GB 以上
- 磁盘: 20GB 以上

### 软件要求
- Go 1.21+
- MySQL 5.7+ 或 8.0+
- Nginx (可选，用于反向代理)

## 部署步骤

### 1. 编译应用

#### Linux 环境

```bash
# 在开发机器上编译
cd college_api
GOOS=linux GOARCH=amd64 go build -o college_api main.go

# 或者在服务器上直接编译
go build -o college_api main.go
```

#### Windows 环境

```bash
# 在开发机器上编译
cd college_api
GOOS=windows GOARCH=amd64 go build -o college_api.exe main.go

# 或者在服务器上直接编译
go build -o college_api.exe main.go
```

### 2. 上传文件到服务器

```bash
# 创建部署目录
mkdir -p /opt/college_api

# 上传编译后的文件
scp college_api user@server:/opt/college_api/
scp .env.example user@server:/opt/college_api/

# 或使用 FTP/SFTP 工具上传
```

### 3. 配置环境变量

```bash
cd /opt/college_api

# 复制环境变量文件
cp .env.example .env

# 编辑配置
nano .env
```

配置示例:
```env
PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_secure_password
DB_DATABASE=college_dev_base
```

### 4. 测试运行

```bash
# 赋予执行权限 (Linux)
chmod +x college_api

# 测试运行
./college_api

# 检查服务是否正常
curl http://localhost:8080/health
```

### 5. 配置系统服务 (Linux)

创建 systemd 服务文件:

```bash
sudo nano /etc/systemd/system/college-api.service
```

内容:
```ini
[Unit]
Description=College API Service
After=network.target mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/college_api
ExecStart=/opt/college_api/college_api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 环境变量 (可选，如果不使用 .env 文件)
# Environment="PORT=8080"
# Environment="DB_HOST=127.0.0.1"

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
# 重载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start college-api

# 设置开机自启
sudo systemctl enable college-api

# 查看状态
sudo systemctl status college-api

# 查看日志
sudo journalctl -u college-api -f
```

### 6. 配置 Windows 服务

#### 使用 NSSM (推荐)

```bash
# 下载 NSSM
# https://nssm.cc/download

# 安装服务
nssm install CollegeAPI "C:\college_api\college_api.exe"

# 设置工作目录
nssm set CollegeAPI AppDirectory "C:\college_api"

# 设置日志
nssm set CollegeAPI AppStdout "C:\college_api\logs\stdout.log"
nssm set CollegeAPI AppStderr "C:\college_api\logs\stderr.log"

# 启动服务
nssm start CollegeAPI

# 设置开机自启
nssm set CollegeAPI Start SERVICE_AUTO_START
```

## Nginx 反向代理配置

### 配置文件

```nginx
# /etc/nginx/sites-available/college-api
upstream college_api {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name api.college.com;

    # 日志
    access_log /var/log/nginx/college-api-access.log;
    error_log /var/log/nginx/college-api-error.log;

    # 健康检查
    location /health {
        proxy_pass http://college_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API 路由
    location /api/ {
        proxy_pass http://college_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

启用配置:
```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/college-api /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

### SSL 配置 (HTTPS)

```nginx
server {
    listen 443 ssl http2;
    server_name api.college.com;

    # SSL 证书
    ssl_certificate /etc/ssl/certs/college-api.crt;
    ssl_certificate_key /etc/ssl/private/college-api.key;

    # SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # 其他配置同上...
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name api.college.com;
    return 301 https://$server_name$request_uri;
}
```

## 监控和维护

### 1. 日志管理

#### 查看日志 (Linux)

```bash
# 实时查看日志
sudo journalctl -u college-api -f

# 查看最近 100 行
sudo journalctl -u college-api -n 100

# 查看今天的日志
sudo journalctl -u college-api --since today

# 查看错误日志
sudo journalctl -u college-api -p err
```

#### 日志轮转

创建 logrotate 配置:
```bash
sudo nano /etc/logrotate.d/college-api
```

内容:
```
/var/log/college-api/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 www-data www-data
    sharedscripts
    postrotate
        systemctl reload college-api > /dev/null 2>&1 || true
    endscript
}
```

### 2. 性能监控

#### 使用 systemd 监控

```bash
# 查看服务状态
sudo systemctl status college-api

# 查看资源使用
sudo systemd-cgtop

# 查看进程信息
ps aux | grep college_api
```

#### 使用 htop

```bash
# 安装 htop
sudo apt install htop

# 运行
htop
```

### 3. 数据库连接监控

```sql
-- 查看当前连接数
SHOW PROCESSLIST;

-- 查看连接统计
SHOW STATUS LIKE 'Threads_connected';
SHOW STATUS LIKE 'Max_used_connections';

-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query_log';
```

### 4. 健康检查脚本

```bash
#!/bin/bash
# /opt/college_api/health_check.sh

URL="http://localhost:8080/health"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $URL)

if [ $RESPONSE -eq 200 ]; then
    echo "$(date): Service is healthy"
    exit 0
else
    echo "$(date): Service is down (HTTP $RESPONSE)"
    # 重启服务
    sudo systemctl restart college-api
    exit 1
fi
```

添加到 crontab:
```bash
# 每 5 分钟检查一次
*/5 * * * * /opt/college_api/health_check.sh >> /var/log/college-api-health.log 2>&1
```

## 备份策略

### 1. 数据库备份

```bash
#!/bin/bash
# /opt/scripts/backup_db.sh

BACKUP_DIR="/opt/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="college_dev_base"

mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u root -p$DB_PASSWORD $DB_NAME | gzip > $BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz

# 删除 30 天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "$(date): Database backup completed"
```

添加到 crontab:
```bash
# 每天凌晨 2 点备份
0 2 * * * /opt/scripts/backup_db.sh >> /var/log/db-backup.log 2>&1
```

### 2. 应用备份

```bash
#!/bin/bash
# /opt/scripts/backup_app.sh

BACKUP_DIR="/opt/backups/app"
DATE=$(date +%Y%m%d_%H%M%S)
APP_DIR="/opt/college_api"

mkdir -p $BACKUP_DIR

# 备份应用
tar -czf $BACKUP_DIR/college_api_${DATE}.tar.gz -C $APP_DIR .

# 删除 7 天前的备份
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "$(date): Application backup completed"
```

## 更新部署

### 零停机更新

```bash
#!/bin/bash
# /opt/scripts/update_app.sh

APP_DIR="/opt/college_api"
BACKUP_DIR="/opt/backups/app"
NEW_BINARY="$1"

if [ -z "$NEW_BINARY" ]; then
    echo "Usage: $0 <new_binary_path>"
    exit 1
fi

# 备份当前版本
cp $APP_DIR/college_api $BACKUP_DIR/college_api.backup

# 复制新版本
cp $NEW_BINARY $APP_DIR/college_api
chmod +x $APP_DIR/college_api

# 重启服务
sudo systemctl restart college-api

# 等待服务启动
sleep 5

# 健康检查
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "Update successful"
    exit 0
else
    echo "Update failed, rolling back..."
    cp $BACKUP_DIR/college_api.backup $APP_DIR/college_api
    sudo systemctl restart college-api
    exit 1
fi
```

## 故障排查

### 常见问题

#### 1. 服务无法启动

```bash
# 检查日志
sudo journalctl -u college-api -n 50

# 检查端口占用
sudo netstat -tlnp | grep 8080

# 检查文件权限
ls -la /opt/college_api/

# 手动运行查看错误
cd /opt/college_api
./college_api
```

#### 2. 数据库连接失败

```bash
# 测试数据库连接
mysql -h 127.0.0.1 -u root -p college_dev_base

# 检查 MySQL 状态
sudo systemctl status mysql

# 查看 MySQL 错误日志
sudo tail -f /var/log/mysql/error.log
```

#### 3. 性能问题

```bash
# 查看系统负载
uptime
top

# 查看内存使用
free -h

# 查看磁盘使用
df -h

# 查看网络连接
netstat -an | grep 8080
```

## 安全加固

### 1. 防火墙配置

```bash
# UFW (Ubuntu)
sudo ufw allow 8080/tcp
sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### 2. 限流配置 (Nginx)

```nginx
# 限制请求频率
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

server {
    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://college_api;
    }
}
```

### 3. 访问控制

```nginx
# 仅允许特定 IP 访问
location /api/ {
    allow 192.168.1.0/24;
    deny all;
    proxy_pass http://college_api;
}
```

## 总结

本部署文档涵盖了:
- ✅ 完整的部署流程
- ✅ 系统服务配置
- ✅ Nginx 反向代理
- ✅ 监控和日志管理
- ✅ 备份策略
- ✅ 更新部署
- ✅ 故障排查
- ✅ 安全加固

按照本文档操作，可以将 College API 服务安全、稳定地部署到生产环境。
