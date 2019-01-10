#源镜像
FROM golang:latest
#作者
MAINTAINER Gaoxinyuan Dengxin Xiaohuarong "gaoxinyuan@yinhai.com"
#设置工作目录
WORKDIR $GOPATH
#将服务器的go工程代码加入到docker容器中
ADD . $GOPATH
#go构建可执行文件
RUN go build .
#暴露端口
EXPOSE 6064
#最终运行docker的命令
ENTRYPOINT  ["./mygohttp"]