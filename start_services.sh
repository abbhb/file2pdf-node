#!/bin/sh

# 启动 unoserver 服务，并将日志重定向到 /root/unoserver.log
unoserver --interface 0.0.0.0 --port 2003 --uno-interface 0.0.0.0 --uno-port 2002 > /root/unoserver.log 2>&1 &

# 启动 unoserver-rest-api 服务，并将日志重定向到 /root/unoserver-rest-api.log
./file2pdf-node > /root/file2pdf-node.log 2>&1 &

wait
