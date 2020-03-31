cd ~ && mkdir gansible && cd gansible
#把gansible二进制文件拷到本目录
#把ssh-pubkey 拷到本目录（映射本地) ;COPY id_rsa   /root/.ssh/id_rsa
#把.pwdfile拷到本目录（映射本地);COPY .pwdfile    /root/.pwdfile
#把Dockerfile拷到本地

docker build -t gansible .
docker run --name gansible -it gansible


docker stop gansible
docker rm gansible
docker rmi gansible

问题：
1.出现/bin/sh: ./gansible: not found报错
原因alpine 使用musl libc与gnu libc部分不兼容
https://blog.csdn.net/liumiaocn/article/details/89702529
alpine镜像中加如下内容
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

