FROM alpine
LABEL  maintainer="https://github.com/charlesren/gansible" 
WORKDIR  /data
 
RUN echo "http://mirrors.aliyun.com/alpine/latest-stable/main/" > /etc/apk/repositories
RUN echo "http://mirrors.aliyun.com/alpine/latest-stable/community/" >> /etc/apk/repositories
 
RUN apk update && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 
 
COPY gansible  /sbin/gansible
#EXPOSE 22
#CMD ["/usr/sbin/sshd", "-D"]
