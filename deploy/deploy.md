####容器方式安装
###打包
1. cd ~ && mkdir gansible && cd gansible
2. 把gansible二进制文件拷到本目录
3. 把Dockerfile拷到本地
4. 准备ssh私钥及密码文件
   4.1方式1：本地/root目录映射给容器
         启动容器时加 -v /root:/root 参数.
   4.2方式2：把文件打包到镜像里（不推荐）
         把ssh-key 拷到本目录;Dockerfile 添加如下内容 COPY id_rsa   /root/.ssh/id_rsa
         把.pwdfile拷到本目录;Dockerfile 添加如下内容 COPY .pwdfile    /root/.pwdfile
5. docker build -t gansible .
###启动
6. docker run --name gansible -it -v  /root:/root gansible
###停止
docker stop gansible
###删除
docker rm gansible
docker rmi gansible
###问题
1.出现/bin/sh: ./gansible: not found报错
原因:alpine使用musl libc与gnu libc部分不兼容
参考https://blog.csdn.net/liumiaocn/article/details/89702529
解决方案：
alpine镜像中加如下内容
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

