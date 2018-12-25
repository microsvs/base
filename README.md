base是一个小而美的业务基础框架，它使用graphql作为api通信协议基础。然后提供了一些可以直接使用的后端存储服务。

![tiny-base](https://gewuwei.oss-cn-shanghai.aliyuncs.com/tracelearning/base.png)

## 基础特性

1. 只支持graphql协议通信；提供了graphql-ui；UI即接口。并提供了一些与该协议进行数据转化的API列表;
2. 提供数据访问层DAL，使用`upper/db`库，该库支持：PostgreSQL, MySQL, SQLite, MSSQL, QL and MongoDB;
3. 提供了redis、消息队列、db和配置服务的初始化实例;
4. 支持配置服务的自动watch，并实时更新; 支持基于dns的服务发现；
5. 支持jaeger trace跟踪；支持错误码注册；支持rabbitmq消息队列消费和生产
6. 支持自定义时间规则切割文件，比如：年/月/日/时/分
7. 支持微服务之间的rpc调用和数据解析
8. 轻易支持立即容器化


## 目录结构

```shell
.
├── README.md
├── bindata_assetfs.go
├── cmd/
├── context.go
├── graphql.go
├── graphql_test.go
├── handler.go
├── middlewares.go
├── pkg/
└── plugins/
```

1. cmd目录用于提供对cache, db, discovery和mq服务的client初始化，且不用关心连接释放等问题， 且默认初始化discovery client连接；
2. pkg目录提供了环境变量、日志、消息队列、rpc、系统时间、tracing、types(token, user服务定义)，以及常用的utils包
3. graphql提供了一些gateway的底层调用，graphql的解析、以及与go的类型互相转换(包括：枚举，字符串、时间等)；
4. 提供了基于martini的路由插件，以及处理graphql请求的方法，并支持graphql返回结果的自定义。
5. 提供了访问流控和签名两个服务的api。

## 使用方式

::TODO
