# Flashcard Backend

一个基于Go语言的闪卡学习系统后端服务，使用现代化的技术栈构建。

## 技术栈

- **依赖注入**: [Uber FX](https://uber-go.github.io/fx/) - 强大的依赖注入框架
- **ORM**: [GORM](https://gorm.io/) - Go语言最受欢迎的ORM库
- **RPC协议**: [gRPC](https://grpc.io/) - 高性能的RPC框架
- **数据库**: PostgreSQL
- **日志**: [Zap](https://github.com/uber-go/zap) - 高性能结构化日志库
- **配置管理**: 环境变量 + .env文件

## 项目结构

```
flashcard-backend/
├── cmd/
│   └── server/          # 应用入口
├── internal/
│   ├── app/            # 应用层（FX模块配置）
│   ├── config/         # 配置管理
│   ├── database/       # 数据库连接和仓储层
│   ├── handler/        # gRPC处理器
│   ├── model/          # 数据模型
│   └── service/        # 业务逻辑层
├── pkg/
│   └── logger/         # 日志工具
├── proto/              # Protocol Buffers定义
├── config/             # 配置文件
├── .env.example        # 环境变量示例
├── go.mod              # Go模块定义
├── Makefile            # 构建脚本
└── README.md           # 项目文档
```

## 快速开始

### 1. 环境准备

确保你的系统已安装：
- Go 1.21+
- PostgreSQL 12+
- Protocol Buffers编译器 (protoc)

### 2. 克隆项目

```bash
git clone <repository-url>
cd flashcard-backend
```

### 3. 安装依赖

```bash
make deps
make tools  # 安装protobuf工具
```

### 4. 配置环境变量

```bash
cp .env.example .env
# 编辑.env文件，设置数据库连接等配置
```

### 5. 创建数据库

```bash
make db-create
```

### 6. 生成protobuf代码

```bash
make proto
```

### 7. 运行服务

```bash
# 开发模式
make dev

# 或者构建后运行
make build
make run
```

## API服务

本项目提供以下gRPC服务：

### UserService
- `CreateUser` - 创建用户
- `GetUser` - 获取用户信息
- `UpdateUser` - 更新用户信息
- `DeleteUser` - 删除用户

### DeckService
- `CreateDeck` - 创建卡片组
- `GetDeck` - 获取卡片组
- `GetUserDecks` - 获取用户的卡片组列表
- `UpdateDeck` - 更新卡片组
- `DeleteDeck` - 删除卡片组

### FlashcardService
- `CreateFlashcard` - 创建闪卡
- `GetFlashcard` - 获取闪卡
- `GetDeckFlashcards` - 获取卡片组的闪卡列表
- `UpdateFlashcard` - 更新闪卡
- `DeleteFlashcard` - 删除闪卡

## 数据模型

### User（用户）
- ID, Email, Username, Password
- 创建时间、更新时间
- 关联多个Deck

### Deck（卡片组）
- ID, Title, Description
- 关联User和多个Flashcard
- 创建时间、更新时间

### Flashcard（闪卡）
- ID, Front（正面）, Back（背面）
- 难度等级、复习记录
- 关联Deck
- 创建时间、更新时间

## 开发命令

```bash
# 查看所有可用命令
make help

# 下载依赖
make deps

# 生成protobuf代码
make proto

# 构建应用
make build

# 运行应用
make run

# 开发模式运行
make dev

# 运行测试
make test

# 清理构建产物
make clean
```

## 配置说明

环境变量配置（.env文件）：

```env
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=localhost
APP_ENV=development

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=flashcard_db
DB_SSL_MODE=disable
DB_TIMEZONE=Asia/Shanghai

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
```

## 架构特点

1. **依赖注入**: 使用Uber FX框架实现清晰的依赖管理
2. **分层架构**: Handler -> Service -> Repository 的清晰分层
3. **配置管理**: 支持环境变量和.env文件的灵活配置
4. **结构化日志**: 使用Zap提供高性能的结构化日志
5. **数据库迁移**: GORM自动迁移数据库表结构
6. **gRPC服务**: 高性能的RPC通信协议

## 贡献指南

1. Fork本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。