#!/bin/bash

# 定义容器名称前缀和后缀
PREFIX="file2pdf_node_"
SUFFIX="file2pdf_node:1.0.3"

# 函数：启动容器
start_containers() {
    for i in {1..5}; do
        docker run -itd --name "${PREFIX}${i}" "${SUFFIX}"
    done
}

# 函数：重启容器
restart_containers() {
    for i in {1..5}; do
        docker restart "${PREFIX}${i}"
    done
}

# 函数：停止容器
stop_containers() {
    for i in {1..5}; do
        docker stop "${PREFIX}${i}"
    done
}

# 函数：删除容器
delete_containers() {
    for i in {1..5}; do
        docker rm -f "${PREFIX}${i}"
    done
}

# 检查参数
case "$1" in
    start)
        echo "Starting containers..."
        start_containers
        ;;
    restart)
        echo "Restarting containers..."
        restart_containers
        ;;
    stop)
        echo "Stopping containers..."
        stop_containers
        ;;
    delete)
        echo "Deleting containers..."
        delete_containers
        ;;
    *)
        echo "Usage: $0 {start|restart|stop|delete}"
        exit 1
esac
