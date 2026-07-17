# Apple Health Data Service

基于 Go 的 Apple 健康数据服务，支持导入 Apple 健康 App 导出的 zip 文件到 PostgreSQL 数据库，并可在 Grafana 中展示。

## 功能特性

- 支持 Apple 健康 App 导出的 zip 文件导入（支持中英文版）
  - 英文版：export.xml
  - 中文版：导出.xml
- 自动解析 XML 数据并存储到 PostgreSQL
- 提供 RESTful API 查询健康数据
- 支持 Grafana 可视化展示
- 支持锻炼路线地图显示
- 支持多种健康数据类型：
  - 健康记录（步数、心率、血压等）
  - 锻炼记录
  - 活动摘要
  - 睡眠分析
  - 锻炼路线（GPS 轨迹）

## 配套 App

[apple-health-app](https://github.com/buyfakett/apple-health-app) - 用于后续的健康数据上传

## 技术栈

- Go
- Gin Web Framework
- PostgreSQL
- Redis（可选）
- GORM
- Grafana

## 快速开始

### 1. 准备数据库

创建 PostgreSQL 数据库：

```sql
CREATE DATABASE apple_health;
```

### 2. 配置服务

编辑 `config/default.yaml` 或通过环境变量配置：

```yaml
server:
  port: 8888
  log_level: info
  swagger: true
  zone: Asia/Shanghai
  token: 123qazwsxedc456  # 写接口token

db:
  host: localhost
  port: "5432"
  user: postgres
  password: postgres
  database: apple_health

redis:
  host: ""          # 留空时不启用Redis；启用时可改为 localhost
  port: "6379"
  password: ""
  db: 0
  key_prefix: apple_health
```

或使用环境变量：

```bash
export APPLE_HEALTH_SERVER_TOKEN=123qazwsxedc456
export APPLE_HEALTH_DB_HOST=localhost
export APPLE_HEALTH_DB_PORT=5432
export APPLE_HEALTH_DB_USER=postgres
export APPLE_HEALTH_DB_PASSWORD=postgres
export APPLE_HEALTH_DB_DATABASE=apple_health
export APPLE_HEALTH_REDIS_HOST=localhost
export APPLE_HEALTH_REDIS_PORT=6379
```

### 3. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8888` 启动。

### 4. 导入数据

导入数据 API 需要写 Token 认证，请在请求头中携带 `server.token`。

#### 方式一：上传 zip 文件

```bash
curl -X POST http://localhost:8888/api/health/upload \
  -H "Authorization: Bearer 123qazwsxedc456" \
  -F "file=@apple_health_export.zip"
```

#### 方式二：指定文件路径

```bash
curl -X POST http://localhost:8888/api/health/import \
  -H "Authorization: Bearer 123qazwsxedc456" \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/apple_health_export.zip"}'
```

**注意：** 写 Token 可以在 `config/default.yaml` 的 `server.token` 中配置。

#### Redis 每日缓存

如果配置了 Redis，`/api/health/sync` 会在数据写入 PostgreSQL 后，把请求日期范围内的每日步数统计和锻炼记录写入 Redis。值为 0 的字段不会写入；当天步数为 0 且没有锻炼记录时不会写入 Redis，并会清理该日期可能存在的旧缓存。

Redis 中的每日缓存值只保存原始缓存数据，格式如下：

```json
{
  "date": "2024-01-01",
  "step_count": 12345,
  "workout_records": [
    {
      "start_time": "2024-01-01T08:00:00+08:00",
      "end_time": "2024-01-01T08:30:00+08:00",
      "workout_type": "HKWorkoutActivityTypeRunning",
      "workout_name": "跑步",
      "duration_minutes": 30,
      "distance_km": 5.2,
      "energy_kcal": 320,
      "source_name": "Apple Watch"
    }
  ]
}
```

从 PostgreSQL 回填历史日期缓存：

```bash
curl -X POST http://localhost:8888/api/health/redis/rebuild \
  -H "Authorization: Bearer 123qazwsxedc456" \
  -H "Content-Type: application/json" \
  -d '{"start_date": "2024-01-01", "end_date": "2024-01-31"}'
```

读取前端展示用的近 N 天每日缓存，返回步数数组和详细健身记录数组，最多 30 天：

```bash
curl "http://localhost:8888/api/health/redis/daily?days=7"
```

返回格式见 [Redis Daily 文档](docs/redis_daily.md)。

## API 文档

访问 Swagger 文档：http://localhost:8888/api/swagger/index.html

## Grafana 配置

### 1. 添加 PostgreSQL 数据源

在 Grafana 中添加 PostgreSQL 数据源：

- Host: `localhost:5432`
- Database: `apple_health`
- User: `postgres`
- Password: `postgres`

### 2. 导入 Dashboard

1. 打开 Grafana
2. 点击 "+" -> "Import"
3. 上传 `grafana/dashboard.json` 文件
4. 选择 PostgreSQL 数据源
5. 点击 "Import"

## 数据库表结构

### health_records
存储各种健康指标记录（步数、心率、血压等）

### workouts
存储锻炼记录

### workout_routes
存储锻炼路线信息

### workout_locations
存储锻炼路线的 GPS 位置点

### activity_summaries
存储每日活动摘要

### sleep_analyses
存储睡眠分析数据

### import_logs
存储数据导入日志

## 导出 Apple 健康数据

1. 打开 iPhone 上的"健康"App
2. 点击右上角头像
3. 滚动到底部，点击"导出所有健康数据"
4. 等待导出完成，会生成一个 zip 文件（包含"导出.xml"）
5. 将 zip 文件传输到电脑


**注意：系统会自动识别中文版（导出.xml）和英文版（export.xml）文件**
