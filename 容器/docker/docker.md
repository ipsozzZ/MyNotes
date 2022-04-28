
# docker 操作命令集合
---

## # 镜像操作

- docker images
列出docker本地镜像

- docker search [镜像名]
从网络中搜索镜像，我们可以制作一个镜像上传到docker hub上，然后通过这个命令就能搜索出来
参数：
-s 列出关注度大于某个值的镜像，如：docker search -s 100 centos；
--no-trunc 列出docker search命令搜索到的镜像的详细信息

- docker pull [镜像名]:[标签]
从docker hub上拉去镜像

- docker rm [镜像名]:[标签]  或者 docker rm [镜像ID]
删除单个镜像

- docker rmi `docker images -q`
删除本地所有镜像（反单引号docker images -q：查看本地所有镜像ID）


## # 容器操作
如果想从容器中回到宿主机，而又不停止容器可以使用快捷键：ctrl q p

- docker ps
列出所有正在运行的容器

- docker ps -a
列出所有容器

- docker ps -l
列出最后一次运行的容器

- docker ps -f status=[状态]
列出所有符合条件的容器：如：docker ps -f status=exited，列出所有已停止的容器

- docker run [选项] 镜像:标签
创建容器
选项：
-i 表示交互式运行容器（就是创建容器后马上启动容器并进入容器）通常与-t选项一起用。
-t 启动后会进入其容器的命令行，通常与-i同时使用，使用-it两个选项后，容器创建就能登录进去，即分配一个伪终端。
--name 为创建的容器指定一个名字
-d 创建一个守护式容器在后台运行，并返回容器ID，这样创建容器后不会自动登录容器，如果加-i选项，创建后就会运行容器。
-v 表示目录映射，格式为：-p 宿主机目录:容器目录，**注意：最好做目录映射，在宿主机上修改然后共享到容器上。**
-p 表示端口映射，格式为：-p 宿主机端口:容器端口

- exit
退出并停止容器，这个命令只能在容器命令行中操作，不能在宿主命令行中

- docker stop [镜像名]:[标签]  或者 docker stop [镜像ID]
在宿主机命令行退出某个容器

- docker stop `docker ps -a -q`
停止所有容器，（其中反单引号docker ps -a -q，是查询所有容器ID）

- docker start `docker ps -a -q`
启动所有容器，（其中反单引号docker ps -a -q，是查询所有容器ID）


- docker kill [镜像名]:[标签]  或者 docker kill [镜像ID]
在宿主机命令行退出某个容器，非正常停止

- docker exec -it 容器名|容器ID /bin/bash  ('|'是或者)
进入某个已启动的容器并打开命令行，**注意通过exec进入的容器使用exit退出容器后是不会停止容器的**

- **docker cp 要拷贝的宿主机文件或目录 容器ID:容器文件或目录**
将宿主机的文件或目录复制到容器

- **docker cp 容器ID:要拷贝的容器文件或目录 宿主机文件或目录**
将容器的文件或目录复制到宿主机

- **docker run -id -v /宿主机绝对路径目录:/容器内目录 --name 容器名 镜像名**
我们可以在创建容器的时候，将主机的目录与容器的目录进行映射，这样我们就可以通过修改宿主机某个目录的文件从而去修改容器的文件，使用 -v 选项

- **docker run -id -v /宿主机绝对路径目录:/容器内目录:ro --name 容器名 镜像名**
我们可以在创建容器的时候，将主机的目录与容器的目录进行**容器端只读**映射，这样我们就可以通过修改宿主机某个目录的文件从而去修改容器的文件，而在容器中映射的目录或文件权限为只读，不能对文件做更改    

- docker inspect 容器ID
查看容器的内部信息，比如容器网关、IP地址、与宿主机的挂载信息、容器ID等

- docker rm 容器名称|容器ID
移除一个非运行容器，不能移除正在运行的容器，需要先停止才能移除

## # 实战相关

- 将容器保存为镜像（无目录挂载）
docker commit [-m="提交的描述信息"] [-a="创建者"] 容器名|容器ID 生成的镜像名[:标签名]
这里创建的镜像会保存容器中的项目，这样可以多处复用

- 将容器打包成镜像（有挂载目录）
问题：如果docker对容器挂载了数据目录，将容器打包为镜像时，数据不会被打包到镜像中
原因：因为宿主机与容器做了路径映射，再commit一个新镜像时，该路径下的所有数据都会被抛弃，不会保存到新镜像中，可通过docker inspect 镜像名查看是否有目录挂载。
解决：两种方法
映射方法：先把宿主机的数据备份在某个目录下，在docker run 的时候使用-v参数将宿主机上备份的数据目录映射到容器里的目标路径中（如：tomcat是/usr/local/tomcat/webapps，mysql是 var/lib/mysql）
拷贝方法（推荐使用）：先把在宿主机的数据备份在某个目录下，在创建新容器后使用拷贝(docker cp ...)的方法将备份的数据复制进容器中（如：tomcat是/usr/local/tomcat/webapps，mysql是 var/lib/mysql）

- 创建tomcat容器
下载tomcat镜像（忽略）；使用：docker run -id --name=my_tomcat -p 8888:8080 -v /usr/local/project:/usr/local/tomcat/webapps --privileged=true tomcat:8
解析：
docker run -id --name=my_tomcat: 是创建并启动一个守护进程设置name为my_tomcat；
-p 8888:8080 做端口映射
-v /usr/local/project:/usr/local/tomcat/webapps 做目录映射
--privileged=true  解决权限问题（又是不加会出错，最好都加上）
tomcat:8  镜像名:标签名

- 怎样将镜像交给其它服务器使用？
首先，镜像备份：docker save -o 包名.tar 镜像名:镜像标签，-o指定输出到的文件
其次，数据备份：如果有挂载数据，需要将挂载数据打包与镜像包一起交给其它服务器使用，其它服务器通过镜像包和数据包就可以恢复打包前的容器

- 镜像恢复与迁移：通过以上两部我们可以将一个本地容器和数据打包交给其它服务器使用，接收到镜像包和数据包的服务器怎样恢复容器？
首先，镜像恢复：使用 docker load -i 镜像包.bar 来获取到镜像文件，使用docker run ... 创建容器
其次，数据恢复：使用接收到的数据包，来挂载或者使用 docker cp 将数据恢复到新建的容器中。

## # dockfile
dockerfile 用于构建一个新的Docker镜像的脚本文件，是由一系列命令和参数构成的脚本
构建新镜像的步骤：
1. 编写dockerfile文件
2. 通过docker build命令生成新的镜像
3. 通过docker run命令运行

dockerfile常用命令：
- FROM 镜像名:标签名
- RUN <命令行命令> | ["可执行文件", "参数1", "参数2"]，用于执行后面跟着的命令行命令
- COPY 复制指令，从宿主机目录中复制文件或者目录到容器里指定路径。
格式如：COPY [--chown=<user>:<group>] <源路径1>...  <目标路径>
[--chown=<user>:<group>]：可选参数，用户改变复制到容器内文件的拥有者和属组。
<源路径>：源文件或者源目录，这里可以是通配符表达式，其通配符规则要满足 Go 的 filepath.Match 规则
<目标路径>：容器内的指定路径，该路径不用事先建好，路径不存在的话，会自动创建。

- ADD
ADD 指令和 COPY 的使用格式一致（同样需求下，官方推荐使用 COPY）。功能也类似，不同之处如下：ADD 的优点：在执行 <源文件> 为 tar 压缩文件的话，压缩格式为 gzip, bzip2 以及 xz 的情况下，会自动复制并解压到 <目标路径>。
ADD 的缺点：在不解压的前提下，无法复制 tar 压缩文件。会令镜像构建缓存失效，从而可能会令镜像构建变得比较缓慢。具体是否使用，可以根据是否需要自动解压来决定。

- CMD
格式：CMD ["<可执行文件或命令>","<param1>","<param2>",...] 
类似于 RUN 指令，用于运行程序，但二者运行的时间点不同:
CMD 在docker run 时运行。
RUN 是在 docker build。
作用：为启动的容器指定默认要运行的程序，程序运行结束，容器也就结束。CMD 指令指定的程序可被 docker run 命令行参数中指定要运行的程序所覆盖。**注意：如果 Dockerfile 中如果存在多个 CMD 指令，仅最后一个生效。**

- ENTRYPOINT
格式：ENTRYPOINT ["<executeable>","<param1>","<param2>",...]
类似于 CMD 指令，**但其不会被 docker run 的命令行参数指定的指令所覆盖，而且这些命令行参数会被当作参数送给 ENTRYPOINT 指令指定的程序。** 但是, 如果运行 docker run 时使用了 --entrypoint 选项，此选项的参数可当作要运行的程序覆盖 ENTRYPOINT 指令指定的程序。优点：在执行 docker run 的时候可以指定 ENTRYPOINT 运行所需的参数。**注意：如果 Dockerfile 中如果存在多个 ENTRYPOINT 指令，仅最后一个生效。** 可以搭配 CMD 命令使用：一般是变参才会使用 CMD ，这里的 CMD 等于是在给 ENTRYPOINT 传参。

- ENV
格式：ENV ```<key1>=<value1> <key2>=<value2>...```
设置环境变量，定义了环境变量，那么在后续的指令中，就可以使用这个环境变量。

- ARG
格式：ARG <参数名>[=<默认值>]
构建参数，与 ENV 作用一至。不过作用域不一样。ARG 设置的环境变量仅对 Dockerfile 内有效，也就是说只有 docker build 的过程中有效，构建好的镜像内不存在此环境变量。构建命令 docker build 中可以用 --build-arg <参数名>=<值> 来覆盖。

- VOLUME
格式：VOLUME ["<路径1>", "<路径2>"...] | VOLUME <路径>
定义匿名数据卷。在启动容器时忘记挂载数据卷，会自动挂载到匿名卷。作用：避免重要的数据，因容器重启而丢失，这是非常致命的。避免容器不断变大。在启动容器 docker run 的时候，我们可以通过 -v 参数修改挂载点。

- EXPOSE
格式：EXPOSE <端口1> [<端口2>...]
仅仅只是声明端口。
作用：帮助镜像使用者理解这个镜像服务的守护端口，以方便配置映射。在运行时使用随机端口映射时，也就是 docker run -P 时，会自动随机映射 EXPOSE 的端口。

- WORKDIR
格式：WORKDIR <工作目录路径>
指定工作目录。用 WORKDIR 指定的工作目录，会在构建镜像的每一层中都存在。（WORKDIR 指定的工作目录，必须是提前创建好的）。docker build 构建镜像过程中的，每一个 RUN 命令都是新建的一层。只有通过 WORKDIR 创建的目录才会一直存在。

- USER
格式：USER <用户名>[:<用户组>]
用于指定执行后续命令的用户和用户组，这边只是切换后续命令执行的用户（用户和用户组必须提前已经存在）。

- HEALTHCHECK
格式：
HEALTHCHECK [选项] CMD <命令>：设置检查容器健康状况的命令
HEALTHCHECK NONE：如果基础镜像有健康检查指令，使用这行可以屏蔽掉其健康检查指令
HEALTHCHECK [选项] CMD <命令> : 这边 CMD 后面跟随的命令使用，可以参考 CMD 的用法。
**用于指定某个程序或者指令来监控 docker 容器服务的运行状态。**

- ONBUILD
格式：ONBUILD <其它指令>
用于延迟构建命令的执行。简单的说，就是 Dockerfile 里用 ONBUILD 指定的命令，在本次构建镜像的过程中不会执行（假设镜像为 test-build）。当有新的 Dockerfile 使用了之前构建的镜像 FROM test-build ，这是执行新镜像的 Dockerfile 构建时候，会执行 test-build 的 Dockerfile 里的 ONBUILD 指定的命令。

- MAINTAINER [user_name]
镜像创建者的名称以及邮箱
