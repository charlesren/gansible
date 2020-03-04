# gansible

#### 介绍

Gansible is a lightweight cli tool designed for system administrator.
可并发在一组服务器上执行命令、上传文件、下载文件、执行本地脚本。可设置并发量，远程登录超时时间，远程执行超时时间

- 特性：
1. 并发在多个设备上执行任务。默认5个并发，可通过--forks参数设定并发数量，最大10000。
2. 可设置ssh连接超时时间。默认30秒，可通过--ssh-timeout参数设定。
2. 以log、csv、json、yaml格式保存任务日志记录。


- 主要功能如下：
1. 尝试使用密码文件中的密码自动登录服务器，并交换操作。支持命令补全、上箭头、信号处理（Ctrl + C ，Ctrl + D）。
2. 上传文件或目录到远程服务器。目标文件夹不存在会自动创建。若源为文件，则把源文件上传到指定文件夹。若源为目录，则把目录下的文件及文件夹上传到指定文件夹。
3. 从远程服务器上下载文件或目录到本地。会在指定的本地文件夹内为每台主机新建一个以该主机IP地址为名称的文件夹。
4. 执行本地脚本。支持指定脚本运行目录、脚本参数。成功执行脚本返回成功，不判断脚本中内容实际执行情况。
5. 执行命令。




#### 安装教程

1.  xxxx
2.  xxxx
3.  xxxx
Gansible会尝试使用密码文件中的密码登录服务器。
默认密码文件位置 ~/.pwdfile.每个密码占一行。


#### 使用说明
查看使用帮助
1. 查看帮助
gansible -h

2. 自动登录到远程服务器，并启动一个窗口交互执行指令。
gansilbe shell 127.0.0.1

3. 执行命令。
有如下两种方式指定一个或多个设备。
- 通过-n 参数指定机器列表。
可以解析10.0.0.1或10.0.0.2-5或10.0.0.6-10.0.0.8格式的ip。三种格式可以自由组合，以;分隔。
如：10.0.0.1;10.0.0.2-5;10.0.0.6-10.0.0.8
- 通过-f 参数指定ip文件。
支持1中的三种格式。行前有#则忽略该行。
文件内容示例如下：
```
# cat /tmp/nodefile.txt
127.0.0.1
127.0.0.2-3
#127.0.0.4
127.0.0.5-127.0.0.6
```
eg1:
```
[root@localhost gansible]# gansible run -n "127.0.0.1-2;127.0.0.3;127.0.0.4-127.0.0.5" -c hostname
127.0.0.3 | Success | rc=0 >>
localhost.localdomain

127.0.0.4 | Success | rc=0 >>
localhost.localdomain

127.0.0.2 | Success | rc=0 >>
localhost.localdomain

127.0.0.1 | Success | rc=0 >>
localhost.localdomain

127.0.0.5 | Success | rc=0 >>
localhost.localdomain


End Time: 2020-03-04 11:09:49
Cost Time: 740.963261ms
Total(5) : Success=5    Failed=0    Unreachable=0    Skipped=0
```
eg2:
```
[root@localhost gansible]# cat /tmp/nodefile.txt
127.0.0.1
127.0.0.2-3
#127.0.0.4
127.0.0.5-127.0.0.6
 
[root@localhost gansible]# gansible run -f /tmp/nodefile.txt -c hostname
127.0.0.2 | Success | rc=0 >>
localhost.localdomain

127.0.0.6 | Success | rc=0 >>
localhost.localdomain

127.0.0.3 | Success | rc=0 >>
localhost.localdomain

127.0.0.1 | Success | rc=0 >>
localhost.localdomain

127.0.0.5 | Success | rc=0 >>
localhost.localdomain


End Time: 2020-03-04 11:17:44
Cost Time: 683.091838ms
Total(5) : Success=5    Failed=0    Unreachable=0    Skipped=0
```

设定任务超时时间
```
[root@localhost gansible]# gansible run -n 127.0.0.1 --timeout 3 -c "sleep 5"
127.0.0.1 | Timeout | rc=1 >>
Task not finished before 3 seconds

End Time: 2020-03-04 11:34:57
Cost Time: 5.399184215s
Total(1) : Success=0    Failed=0    Unreachable=0    Skipped=0
```
4. 执行本地脚本。

本地脚本文件内容如下：
```
[root@localhost gansible]# cat date.sh
#/bin/sh
echo $1 >/tmp/date.log
pwd >>/tmp/date.log
```
-a 参数根据脚本情况，可选。
指定参数执行脚本。
```
[root@localhost gansible]# gansible script -n 127.0.0.1 -a "args" ./date.sh 
127.0.0.1 | Success | rc=0 >>


End Time: 2020-03-04 12:17:50
Cost Time: 443.675611ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0
[root@localhost gansible]# cat /tmp/date.log
args
/root
[root@localhost gansible]# 
```

在指定目录执行脚本。
```
[root@localhost gansible]# gansible script -n 127.0.0.1 -a "args" ./date.sh -d /tmp
127.0.0.1 | Success | rc=0 >>


End Time: 2020-03-04 12:14:45
Cost Time: 428.455598ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0

[root@localhost gansible]# cat /tmp/date.log
args
/tmp
[root@localhost gansible]# 
```
5. 下载文件或目录。需指定dest及src两个参数。

下载文件
```
[root@localhost gansible]# gansible fetch -n 127.0.0.1 -s /data/scm/gansible/date.sh -d /tmp/1
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 12:44:56
Cost Time: 469.315007ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0

[root@localhost gansible]# ls -rtl /tmp/1
total 0
drwxr-xr-x. 2 root root 21 Mar  4 12:44 127.0.0.1
[root@localhost gansible]# ls -rtl /tmp/1/127.0.0.1
total 4
-rw-r--r--. 1 root root 52 Mar  4 12:44 date.sh
[root@localhost gansible]# 
```
下载目录
```
[root@localhost gansible]# ls -rtl /data/scm/gansible/testdir
total 0
-rw-r--r--. 1 root root  0 Mar  4 12:46 test.sh
drwxr-xr-x. 2 root root 18 Mar  4 12:47 a
-rw-r--r--. 1 root root  0 Mar  4 12:47 c.sh
[root@localhost gansible]# gansible fetch -n 127.0.0.1 -s /data/scm/gansible/testdir -d /tmp/2
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 12:49:05
Cost Time: 428.766883ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0

[root@localhost gansible]# ls -rtl /tmp/2
total 0
drwxr-xr-x. 3 root root 42 Mar  4 12:49 127.0.0.1
[root@localhost gansible]# ls -rtl /tmp/2/127.0.0.1
total 0
-rw-r--r--. 1 root root  0 Mar  4 12:49 test.sh
drwxr-xr-x. 2 root root 18 Mar  4 12:49 a
-rw-r--r--. 1 root root  0 Mar  4 12:49 c.sh
[root@localhost gansible]# ls -rtl /tmp/2/127.0.0.1/a
total 0
-rw-r--r--. 1 root root 0 Mar  4 12:49 b.sh
[root@localhost gansible]# 
```
6. 上传文件或目录。需指定dest及src两个参数。
上传文件
```
[root@localhost gansible]# gansible push -n 127.0.0.1 -s /data/scm/gansible/date.sh -d /tmp/1
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 14:39:45
Cost Time: 399.632118ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0

[root@localhost gansible]# ls -rtl /tmp/1
total 4
-rw-r--r--. 1 root root 52 Mar  4 14:39 date.sh
[root@localhost gansible]# 
```
上传目录
```
[root@localhost gansible]# gansible push -n 127.0.0.1 -s /data/scm/gansible/testdir -d /tmp/2
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 14:42:25
Cost Time: 427.764754ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0

[root@localhost gansible]# ls -rtl /tmp/2
total 0
drwxr-xr-x. 2 root root 18 Mar  4 14:42 a
-rw-r--r--. 1 root root  0 Mar  4 14:42 c.sh
-rw-r--r--. 1 root root  0 Mar  4 14:42 test.sh
[root@localhost gansible]# 
```
#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

