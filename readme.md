## 功能

1. ws 集群
2. 使用 pprof 和golang trace 进行性能分析
3. 使用 prometheus 进行监控
4. metrics 指标
5. 使用 swagger 进行接口文档管理
6. 使用 jenkins 进行自动化构建,使用k8s进行部署

## 流程

1. 客户端向服务端请求建立长连接，通过 nginx或k8s负载均衡器，将请求转发到任意一个服务端

2. 业务系统将推送的数据发送到服务端，服务端将数据发送到消息队列，消息队列以广播的方式将数据发送到所有的服务端

3. 所有服务端消费消息队列中的数据,如果确定对应的用户在自己的服务端上,将数据发送到客户端

## 接口

### 客户端：

1. 连接接口：`/ws/connect?pid=xxx&uid=xxx&sign=xxx`
   ```
   请求参数
    
    pid(项目 id)
    uid(用户 id)
    sign(签名)
   响应参数
     client_id(客户端 id)
   ```
2. 订阅流：
   ```
   请求参数
    type: subscribe(订阅) unsubscribe(取消订阅)
    request_id(请求 id)
    tag (订阅标签)
   ```

### 业务系统：

1. 推送接口：`/ws/push` 参数：
    ```
   如果指定了uids,client_ids,tags,则会将所有条件求交集
   请求参数 
     pid(项目 id) 必选
     uids(用户 id) 可选 多个用逗号分隔
     client_ids(客户端 id) 可选 多个用逗号分隔
     tags (标签) 可选 多个用逗号分隔
     sign(签名) 必选
     data(数据) 必选
   
     设备类型
   ```
   



