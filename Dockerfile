
FROM alpine:latest

# 重新声明 ARG，因为 ARG 不会从上一阶段继承
ARG APP_NAME
ARG APP_PORT
ARG CONFIG_PATH
ARG CONFIG_FILE_NAME 
ARG WORKDIR=/app

# 设置环境变量
ENV TZ=Asia/Shanghai
ENV APP_NAME=${APP_NAME}
ENV CONFIG_FILE_NAME=${CONFIG_FILE_NAME}

WORKDIR ${WORKDIR}

# 安装基础工具和证书
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache curl tzdata && \
    cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone && \
    apk del tzdata

# 复制构建产物
COPY bin/${APP_NAME}-linux ${APP_NAME}
COPY ${CONFIG_PATH} ${WORKDIR}/

RUN chmod +x ${APP_NAME}

EXPOSE ${APP_PORT}

# 使用 shell 形式的 ENTRYPOINT 以确保变量被正确展开
ENTRYPOINT ["sh", "-c", "/app/${APP_NAME} $@", "--"]
