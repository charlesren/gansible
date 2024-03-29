# gansible

## 介绍

Gansible is a lightweight cli tool used to execute commands on multiple devices in parallel.  
可并发在一组服务器上执行命令、上传下载文件或目录、执行本地脚本。

### 主要特性
1. 并发在多个设备上执行任务。默认10个并发，可通过--forks参数设定并发数量，最大10000。
2. 可设置ssh连接超时时间。默认90秒，可通过--ssh-timeout参数设定。
3. 以gansible、json、yaml格式输出任务结果,默认gansible。
4. 以log、csv、json、yaml格式保存任务日志记录。  
   通过--loging指定保存日志（默认不保存）。    
   通过--log-file-format指定日志格式（默认格式为csv）。   
   通过--log-dir 指定日志目录（默认为系统零时文件夹）。  
   通过--log-file-name指定日志文件名称（默认格式为gansible_year-mounth-day_hour:minuter:secondes）。
5. 使用golang 编写，支持windows、linux、mac。
6. 可通过配置文件，修改默认运行参数，即时生效。默认配置文件 ~/.gansible.yaml
7. 支持密码、秘钥、ssh agent 及尝试使用密码文件中一组给定的密码登录。根据输入参数自动选择登录方式，优先顺序（指定的密码>秘钥>密码文件）。
8. 从字符串或文件解析IP,且自动去重。
9. 指定每台设备上任务超时时间。  
   通过--timeout参数指定（默认300s)。
10. 通过引用GAN.NODE关键字，支持在命令行或脚本参数中引用当前所在的设备的IP。

    gansible run -n 127.0.0.1 -c "echo GAN.NODE"    >> 返回  #127.0.0.1
   
    gansible script -n 127.0.0.1 -a "serve GAN.NODE"   /tmp/program.sh

### **计划中**
1. 多颜色展示输出。

### 主要功能
1. 支持密码、秘钥、ssh agent登录，当失败时会尝试使用密码文件中的密码自动登录服务器，并交换操作。支持命令补全、上箭头、信号处理（Ctrl + C ，Ctrl + D）。
2. 上传文件或目录到远程服务器。目标文件夹不存在会自动创建。  
      若源为文件，则把源文件上传到指定文件夹。  
      若源为目录，则把目录下的文件及文件夹上传到指定文件夹。
3. 从远程服务器上下载文件或目录到本地。会在指定的本地文件夹内为每台主机新建一个以该主机IP地址为名称的文件夹。
4. 执行本地脚本。支持指定脚本运行目录、脚本参数。成功执行脚本返回成功，不判断脚本中内容实际执行情况。
5. 执行命令。




## 安装
### 1. 准备二进制文件
```
git clone https://github.com/charlesren/gansible.git
cd gansible
```
*Linux系统*
```
GOOS=linux go build
```
*Windows系统*
```
GOOS=windows go build
```
*Mac系统*
```
GOOS=darwin go build
```

### 2. 安装
- 容器方式
1. cd ~ && mkdir gansible && cd gansible
2. 把gansible二进制文件拷到本目录
3. 把Dockerfile拷到本地
4. 准备ssh私钥及密码文件  
   - 方式1：本地/root目录映射给容器  
         启动容器时加 -v /root:/root 参数.  
   - 方式2：把文件打包到镜像里（不推荐)  
         把ssh-key 拷到本目录;Dockerfile 添加如下内容 COPY id_rsa   /root/.ssh/id_rsa  
         把.pwdfile拷到本目录;Dockerfile 添加如下内容 COPY .pwdfile    /root/.pwdfile
5. docker build -t gansible .
6. docker run --name gansible -it -v /root:/root -v /tmp:/tmp gansible
- 二进制方式安装

*Linux 系统*
1.  cp gansible /usr/bin/gansible
2. 设置默认密码文件。Gansible会尝试使用密码文件中的密码登录服务器。默认密码文件位置 ~/.pwdfile.每个密码占一行。若无默认文件，运行时可通过--pwdfile 参数指定密码文件。

*Windows系统*
1. 把gansible.exe文件拷贝到C:\Windows下。
2. 打开新的命令提示符窗口，输入以下命令生成密码文件(当前用户的根目录下）。编辑密码文件添加相关密码。
```
echo > .pwdfile
```
## 使用说明
**1. 查看帮助**
```
gansible -h
```
**2. 自动登录到远程服务器，并启动一个窗口交互执行指令。**
```
gansilbe shell 127.0.0.1
```

**3. 执行命令。**

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
使用-n 参数指定多个设备:
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
使用-f 参数指定设备文件:
```
[root@localhost gansible]# cat /tmp/nodefile.txt
127.0.0.1
127.0.0.2-3
#127.0.0.4
127.0.0.5-127.0.0.6
```
```
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

设定任务超时时间:
```
[root@localhost gansible]# gansible run -n 127.0.0.1 --timeout 3 -c "sleep 5"
127.0.0.1 | Timeout | rc=1 >>
Task not finished before 3 seconds

End Time: 2020-03-04 11:34:57
Cost Time: 5.399184215s
Total(1) : Success=0    Failed=0    Unreachable=0    Skipped=0
```

保存任务日志：
```
[root@localhost gansible]# gansible run -n 127.0.0.1-3 -c hostname --loging
127.0.0.1 | Success | rc=0 >>
localhost.localdomain

127.0.0.2 | Success | rc=0 >>
localhost.localdomain

127.0.0.3 | Success | rc=0 >>
localhost.localdomain


End Time: 2020-03-05 16:01:45
Cost Time: 794.775125ms
Total(3) : Success=3    Failed=0    Unreachable=0    Skipped=0
save log to file: /tmp/gansible_2020-03-05_16:01:44.csv successfully!
```
**4. 执行本地脚本。**

本地脚本文件内容如下：
```
[root@localhost gansible]# cat date.sh
#/bin/sh
echo $1 >/tmp/date.log
pwd >>/tmp/date.log
```

指定参数执行脚本:
*-a 可指定脚本参数，根据脚本实际情况选用。*
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

在指定目录执行脚本:
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
**5. 下载文件或目录。需指定dest及src两个参数。**

下载文件:
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
下载目录:
```
[root@localhost gansible]# tree /data/scm/gansible/testdir
/data/scm/gansible/testdir
├── a.sh
└── b
    └── c

1 directory, 2 files
[root@localhost gansible]#
```
```
[root@localhost gansible]# gansible fetch -n 127.0.0.1 -s /data/scm/gansible/testdir -d /tmp/2
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 12:49:05
Cost Time: 428.766883ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0
```
```
[root@localhost gansible]# tree /tmp/2
/tmp/2
└── 127.0.0.1
    ├── a.sh
    └── b
        └── c

2 directories, 2 files
[root@localhost gansible]#
```
**6. 上传文件或目录。需指定dest及src两个参数。**

上传文件:
```
[root@localhost gansible]# gansible push -n 127.0.0.1 -s /data/scm/gansible/date.sh -d /tmp/3
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 14:39:45
Cost Time: 399.632118ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0
```
```
[root@localhost gansible]# ls -rtl /tmp/3
total 4
-rw-r--r--. 1 root root 52 Mar  4 14:39 date.sh
[root@localhost gansible]# 
```
上传目录:
```
[root@localhost gansible]# ls /tmp/4
ls: cannot access /tmp/4: No such file or directory
```
```
[root@localhost gansible]# gansible push -n 127.0.0.1 -s /data/scm/gansible/testdir -d /tmp/4
127.0.0.1 | Success | rc=0 >>
upload successfully!

End Time: 2020-03-04 14:42:25
Cost Time: 427.764754ms
Total(1) : Success=1    Failed=0    Unreachable=0    Skipped=0
```
```
[root@localhost gansible]# tree /tmp/4
/tmp/4
├── a.sh
└── b
    └── c

1 directory, 2 files
[root@localhost gansible]# 
```
#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

