# file2pdf-node
依赖unoserver，就加了rocket消费者消费，消费完生产成功的回报消息。[配套项目](https://github.com/abbhb/OA_Helper)

## Usage

启动容器连接上rocketmq自动开始消费消息

## 快速使用
```shell
docker pull abbhb/file2pdf_node:1.0.2
```
直接启动即可，为项目附属，所以消费者逻辑固定，配置就懒得动态了，需要修改自行build


## Build流程
#### 1.拉取代码
```shell
git clone https://github.com/abbhb/file2pdf-node.git
```
#### 2.构建镜像
```shell
docker build
```
#### 3.一键多开
这一步需要具体修改脚本！
```shell
./file2pdf.sh start
```
