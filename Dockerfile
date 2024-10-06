FROM docker.io/abbhb/unoserver:1.0.0
# 将你的程序复制到镜像中
COPY file2pdf-node /root/
COPY start_services.sh /root/
# 设置工作目录
WORKDIR /root
CMD ["./start_services.sh"]
