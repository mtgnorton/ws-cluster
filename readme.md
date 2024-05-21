## 功能

1. ws 集群
2. 使用 pprof 进行性能分析
   pprof 当启动http服务时，如果设置pprof开启，那么pprof会伴随启动，假设服务运行在 6060
   端口,访问 http://localhost:6060/debug/pprof/
   可以查看性能分析数据参考文档 https://goframe.org/pages/viewpage.action?pageId=17203722
3. 使用 prometheus 进行指标监控
   常用query
    - 5分钟内HTTP请求持续时间的50th百分位数 `histogram_quantile(0.5, sum(rate(request_duration_bucket[5m])) by (le))`
    - 5分钟内HTTP请求持续时间的95th百分位数 `histogram_quantile(0.95, sum(rate(request_duration_bucket[5m])) by (le))`
    - 5m内请求时间<
      =0.3秒的请求数量与总请求数量的比率   `sum(rate(request_duration_bucket{le="0.3"}[5m])) by (job) / sum(rate(request_duration_count[5m])) by (job)`
    - 请求url统计 `request_url_total{job="ws-cluster"}`
    - 请求总数 `request_total{job="ws-cluster"} `
    - ws实时连接数 `ws_connection{job="ws-cluster"}`
5. 使用 swagger 进行接口文档管理
   swagger 访问路径 http://localhost:9092/swagger/index.html
6. 使用 jenkins 进行自动化构建,使用k8s进行部署
7. 使用sentry记录错误日志
   地址： https://docs.sentry.io/

## 使用

./ws-cluster --node 200 --ws_port 8812 --http_port 8912 --queue redis --env dev

## 流程

1. 客户端向服务端请求建立长连接，通过 nginx或k8s负载均衡器，将请求转发到任意一个服务端

2. 业务系统将推送的数据发送到服务端，服务端将数据发送到消息队列，消息队列以广播的方式将数据发送到所有的服务端

3. 所有服务端消费消息队列中的数据,如果确定对应的用户在自己的服务端上,将数据发送到客户端

## todo

1. 接口文档
2. 负载均衡router
3. 日志切割导致日志丢失的问题
4. http接口
5. ~~redis队列读取阻塞问题~~
6. 设备类型上传
7. 压力测试
8. 集成openTelemetry 
9. 心跳检测

### 客户端：

1. 连接接口：`/ws/connect?pid=xxx&uid=xxx&sign=xxx`
   ```
   请求参数
    pid(项目 id)
    uid(用户 id)
    type(连接类型) 1:client 2:server
    sign(签名)
   响应参数
     client_id(客户端 id)
   ```
2. ws消息：
    - 客户端：
      ```
      请求参数
       type: subscribe(订阅)
       tags: 标签
      ```

### 业务系统：

1. 推送接口：`/ws/push` 参数：
    ```
   如果指定了uids,client_ids,tags,则会将所有条件求交集
   请求参数 
     pid(项目 id) 必选
     uids(用户 id) 可选 多个用逗号分隔
     cids(客户端 id) 可选 多个用逗号分隔
     tags (标签) 可选 多个用逗号分隔
     sign(签名) 必选
     data(数据) 必选
   
     设备类型
   ```

## 交互类型

1. 连接
2. 订阅
3. 请求   
   c1->router->s1->router->c1


