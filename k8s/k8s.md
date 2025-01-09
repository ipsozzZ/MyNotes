# k8s

Kubernetes 项目所做的，其实就是将“进程组”的概念映射到了容器技术中，并使其成为了这个云计算“操作系统”里的“一等公民”。


## # 容器（容器本身没有价值，有价值的是“容器编排”。）
容器其实是一种沙盒技术。顾名思义，沙盒就是能够像一个集装箱一样，把你的应用“装”起来的技术。这样，应用与应用之间，就因为有了边界而不至于相互干扰；而被装进集装箱的应用，也可以被方便地搬来搬去，这不就是 PaaS 最理想的状态嘛。

- 容器技术的兴起源于 PaaS 技术的普及；
- Docker 公司发布的 Docker 项目具有里程碑式的意义；
- Docker 项目通过“容器镜像”，解决了应用打包这个根本性难题。

用户运行在容器里的应用进程，跟宿主机上的其他进程一样，都由宿主机操作系统统一管理，只不过这些被隔离的进程拥有额外设置过的 Namespace 参数。而 Docker 项目在这里扮演的角色，更多的是旁路式的辅助和管理工作。

使用虚拟化技术作为应用沙盒，就必须要由 Hypervisor 来负责创建虚拟机，这个虚拟机是真实存在的，并且它里面必须运行一个完整的 Guest OS 才能执行用户的应用进程。这就不可避免地带来了额外的资源消耗和占用。

根据实验，一个运行着 CentOS 的 KVM 虚拟机启动后，在不做优化的情况下，虚拟机自己就需要占用 100~200 MB 内存。此外，用户应用运行在虚拟机里面，它对宿主机操作系统的调用就不可避免地要经过虚拟化软件的拦截和处理，这本身又是一层性能损耗，尤其对计算资源、网络和磁盘 I/O 的损耗非常大。

而相比之下，容器化后的用户应用，却依然还是一个宿主机上的普通进程，这就意味着这些因为虚拟化而带来的性能损耗都是不存在的；而另一方面，使用 Namespace 作为隔离手段的容器并不需要单独的 Guest OS，这就使得容器额外的资源占用几乎可以忽略不计。

#### 容器边界
假如，现在要写一个计算加法的小程序，这个程序需要的输入来自于一个文件，计算完成后的结果则输出到另一个文件中。由于计算机只认识 0 和 1，所以无论用哪种语言编写这段代码，最后都需要通过某种方式翻译成二进制文件，才能在计算机操作系统中运行起来。而为了能够让这些代码正常运行，我们往往还要给它提供数据，比如我们这个加法程序所需要的输入文件。这些数据加上代码本身的二进制文件，放在磁盘上，就是我们平常所说的一个“程序”，也叫代码的可执行镜像（executable image）。然后，我们就可以在计算机上运行这个“程序”了。

首先，操作系统从“程序”中发现输入数据保存在一个文件中，所以这些数据就会被加载到内存中待命。同时，操作系统又读取到了计算加法的指令，这时，它就需要指示 CPU 完成加法操作。而 CPU 与内存协作进行加法计算，又会使用寄存器存放数值、内存堆栈保存执行的命令和变量。同时，计算机里还有被打开的文件，以及各种各样的 I/O 设备在不断地调用中修改自己的状态。

就这样，一旦“程序”被执行起来，它就从磁盘上的二进制文件，变成了计算机内存中的数据、寄存器里的值、堆栈中的指令、被打开的文件，以及各种设备的状态信息的一个集合。像这样一个程序运行起来后的计算机执行环境的总和，就是我们今天的主角：进程。所以，对于进程来说，它的静态表现就是程序，平常都安安静静地待在磁盘上；而一旦运行起来，它就变成了计算机里的数据和状态的总和，这就是它的动态表现。

而容器技术的核心功能，就是通过约束和修改进程的动态表现，从而为其创造出一个“边界”。


#### NameSpace（做隔离）
Namespace 的使用方式非常有意思：它其实只是 Linux 创建新进程的一个可选参数。我们知道，在 Linux 系统中创建进程的系统调用是 clone()，比如：
```c
int pid = clone(main_function, stack_size, SIGCHLD, NULL); 
```
这个系统调用就会为我们创建一个新的进程，并且返回它的进程号 pid。

而当我们用 clone() 系统调用创建一个新进程时，就可以在参数中指定 CLONE_NEWPID 参数，比如：
```c
int pid = clone(main_function, stack_size, CLONE_NEWPID | SIGCHLD, NULL); 
```

这时，新创建的这个进程将会“看到”一个全新的进程空间，在这个进程空间里，它的 PID 是 1。之所以说“看到”，是因为这只是一个“障眼法”，在宿主机真实的进程空间里，这个进程的 PID 还是真实的数值，比如 100。

当然，我们还可以多次执行上面的 clone() 调用，这样就会创建多个 PID Namespace，而每个 Namespace 里的应用进程，都会认为自己是当前容器里的第 1 号进程，它们既看不到宿主机里真正的进程空间，也看不到其他 PID Namespace 里的具体情况。

**而除了我们刚刚用到的 PID Namespace，Linux 操作系统还提供了 Mount、UTS、IPC、Network 和 User 这些 Namespace，用来对各种不同的进程上下文进行“障眼法”操作。**

比如，Mount Namespace，用于让被隔离进程只看到当前 Namespace 里的挂载点信息；Network Namespace，用于让被隔离进程看到当前 Namespace 里的网络设备和配置。

这，就是 Linux 容器最基本的实现原理了。

所以，Docker 容器这个听起来玄而又玄的概念，实际上是在创建容器进程时，指定了这个进程所需要启用的一组 Namespace 参数。这样，容器就只能“看”到当前 Namespace 所限定的资源、文件、设备、状态，或者配置。而对于宿主机以及其他不相关的程序，它就完全看不到了。所以说，容器，其实是一种特殊的进程而已。

#### linux cgroups
不是已经通过 Linux Namespace 创建了一个“容器”吗，为什么还需要对容器做“限制”呢？

虽然容器内的第 1 号进程在“障眼法”的干扰下只能看到容器里的情况，但是宿主机上，它作为第 100 号进程与其他所有进程之间依然是平等的竞争关系。这就意味着，虽然第 100 号进程表面上被隔离了起来，但是它所能够使用到的资源（比如 CPU、内存），却是可以随时被宿主机上的其他进程（或者其他容器）占用的。当然，这个 100 号进程自己也可能把所有资源吃光。这些情况，显然都不是一个“沙盒”应该表现出来的合理行为。

Linux Cgroups 的全称是 Linux Control Group。它最主要的作用，就是限制一个进程组能够使用的资源上限，包括 CPU、内存、磁盘、网络带宽等等。

此外，Cgroups 还能够对进程进行优先级设置、审计，以及将进程挂起和恢复等操作。

“敏捷”和“高性能”是容器相较于虚拟机最大的优势，也是它能够在 PaaS 这种更细粒度的资源管理平台上大行其道的重要原因。

不过，有利就有弊，基于 Linux Namespace 的隔离机制相比于虚拟化技术也有很多不足之处，其中最主要的问题就是：隔离得不彻底。

首先，既然容器只是运行在宿主机上的一种特殊的进程，那么多个容器之间使用的就还是同一个宿主机的操作系统内核。

其次，在 Linux 内核中，有很多资源和对象是不能被 Namespace 化的，最典型的例子就是：时间。这就意味着，如果你的容器中的程序使用 settimeofday(2) 系统调用修改了时间，整个宿主机的时间都会被随之修改，这显然不符合用户的预期。相比于在虚拟机里面可以随便折腾的自由度，在容器里部署应用的时候，“什么能做，什么不能做”，就是用户必须考虑的一个问题。

此外，由于上述问题，尤其是共享宿主机内核的事实，容器给应用暴露出来的攻击面是相当大的，应用“越狱”的难度自然也比虚拟机低得多。

更为棘手的是，尽管在实践中我们确实可以使用 Seccomp 等技术，对容器内部发起的所有系统调用进行过滤和甄别来进行安全加固，但这种方法因为多了一层对系统调用的过滤，必然会拖累容器的性能。何况，默认情况下，谁也不知道到底该开启哪些系统调用，禁止哪些系统调用。

所以，在生产环境中，没有人敢把运行在物理机上的 Linux 容器直接暴露到公网上。当然，下面会讲到的基于虚拟化或者独立内核技术的容器实现，则可以比较好地在隔离与性能之间做出平衡。

#### chroot (容器镜像)
即使开启了 Mount Namespace，容器进程看到的文件系统也跟宿主机完全一样。

这是怎么回事呢？

仔细思考一下，会发现这其实并不难理解：Mount Namespace 修改的，是容器进程对文件系统“挂载点”的认知。但是，这也就意味着，只有在“挂载”这个操作发生之后，进程的视图才会被改变。而在此之前，新创建的容器会直接继承宿主机的各个挂载点。

在 Linux 操作系统里，有一个名为 chroot 的命令可以帮助你在 shell 中方便地完成这个工作。顾名思义，它的作用就是帮你“change root file system”，即改变进程的根目录到你指定的位置。它的用法也非常简单。

假设，现在有一个 $HOME/test 目录，想要把它作为一个 /bin/bash 进程的根目录。

首先，创建一个 test 目录和几个 lib 文件夹：
```sh
mkdir -p $HOME/test
mkdir -p $HOME/test/{bin,lib64,lib}
cd $T
# 然后，把 bash 命令拷贝到 test 目录对应的 bin 路径下：
cp -v /bin/{bash,ls} $HOME/test/bin

# 接下来，把 bash 命令需要的所有 so 文件，也拷贝到 test 目录对应的 lib 路径下。找到 so 文件可以用 ldd 命令：
T=$HOME/test
list="$(ldd /bin/ls | egrep -o '/lib.*\.[0-9]')"
for i in $list; do cp -v "$i" "${T}${i}"; done

# 最后，执行 chroot 命令，告诉操作系统，我们将使用 $HOME/test 目录作为 /bin/bash 进程的根目录：
chroot $HOME/test /bin/bash

```
这时，你如果执行 "ls /"，就会看到，它返回的都是 $HOME/test 目录下面的内容，而不是宿主机的内容。更重要的是，对于被 chroot 的进程来说，它并不会感受到自己的根目录已经被“修改”成 $HOME/test 了。

实际上，Mount Namespace 正是基于对 chroot 的不断改良才被发明出来的，它也是 Linux 操作系统里的第一个 Namespace。


当然，为了能够让容器的这个根目录看起来更“真实”，我们一般会在这个容器的根目录下挂载一个完整操作系统的文件系统，比如 Ubuntu16.04 的 ISO。这样，在容器启动之后，我们在容器里通过执行 "ls /" 查看根目录下的内容，就是 Ubuntu 16.04 的所有目录和文件。

而这个挂载在容器根目录上、用来为容器进程提供隔离后执行环境的文件系统，就是所谓的“容器镜像”。它还有一个更为专业的名字，叫作：rootfs（根文件系统）。

所以，一个最常见的 rootfs，或者说容器镜像，会包括如下所示的一些目录和文件，比如 /bin，/etc，/proc 等等：
```sh
ls /
# bin dev etc home lib lib64 mnt opt proc root run sbin sys tmp usr var
```

而你进入容器之后执行的 /bin/bash，就是 /bin 目录下的可执行文件，与宿主机的 /bin/bash 完全不同。

现在，你应该可以理解，对 Docker 项目来说，它最核心的原理实际上就是为待创建的用户进程：
1. 启用 Linux Namespace 配置；
2. 设置指定的 Cgroups 参数；
3. 切换进程的根目录（Change Root）。

这样，一个完整的容器就诞生了。不过，Docker 项目在最后一步的切换上会优先使用 pivot_root 系统调用，如果系统不支持，才会使用 chroot。这两个系统调用虽然功能类似，但是也有细微的区别，可以自行查阅资料。

另外，需要明确的是，rootfs 只是一个操作系统所包含的文件、配置和目录，并不包括操作系统内核。在 Linux 操作系统中，这两部分是分开存放的，操作系统只有在开机启动时才会加载指定版本的内核镜像。

所以说，rootfs 只包括了操作系统的“躯壳”，并没有包括操作系统的“灵魂”。

那么，对于容器来说，这个操作系统的“灵魂”又在哪里呢？实际上，同一台机器上的所有容器，都共享宿主机操作系统的内核。

这就意味着，如果你的应用程序需要配置内核参数、加载额外的内核模块，以及跟内核进行直接的交互，你就需要注意了：这些操作和依赖的对象，都是宿主机操作系统的内核，它对于该机器上的所有容器来说是一个“全局变量”，牵一发而动全身。

这也是容器相比于虚拟机的主要缺陷之一：毕竟后者不仅有模拟出来的硬件机器充当沙盒，而且每个沙盒里还运行着一个完整的 Guest OS 给应用随便折腾。

由于云端与本地服务器环境不同，应用的打包过程，一直是使用 PaaS 时最“痛苦”的一个步骤。但有了容器之后，更准确地说，有了容器镜像（即 rootfs）之后，这个问题被非常优雅地解决了。

由于 rootfs 里打包的不只是应用，而是整个操作系统的文件和目录，也就意味着，应用以及它运行所需要的所有依赖，都被封装在了一起。

事实上，对于大多数开发者而言，他们对应用依赖的理解，一直局限在编程语言层面。比如应用的业务配置、服务配置等。但实际上，一个一直以来很容易被忽视的事实是，对一个应用来说，操作系统本身才是它运行所需要的最完整的“依赖库”。

有了容器镜像“打包操作系统”的能力，这个最基础的依赖环境也终于变成了应用沙盒的一部分。这就赋予了容器所谓的一致性：无论在本地、云端，还是在一台任何地方的机器上，用户只需要解压打包好的容器镜像，那么这个应用运行所需要的完整的执行环境就被重现出来了。

这种深入到操作系统级别的运行环境一致性，打通了应用在本地开发和远端执行环境之间难以逾越的鸿沟。

#### 镜像分层
容器镜像 = 镜像层 + 容器层

所有的镜像层和容器层共同构成了容器的rootfs(根目录)

不过，这时你可能已经发现了另一个非常棘手的问题：难道我每开发一个应用，或者升级一下现有的应用，都要重复制作一次 rootfs 吗？

比如，我现在用 Ubuntu 操作系统的 ISO 做了一个 rootfs，然后又在里面安装了 Java 环境，用来部署我的 Java 应用。那么，我的另一个同事在发布他的 Java 应用时，显然希望能够直接使用我安装过 Java 环境的 rootfs，而不是重复这个流程。

这些修改都基于一个旧的 rootfs，我们以增量的方式去做这些修改，这样做的好处是，所有人都只需要维护相对于 base rootfs 修改的增量内容，而不是每次修改都制造一个“fork”。

这也正是为何，Docker 公司在实现 Docker 镜像时并没有沿用以前制作 rootfs 的标准流程，而是做了一个小小的创新：Docker 在镜像的设计中，引入了层（layer）的概念。也就是说，用户制作镜像的每一步操作，都会生成一个层，也就是一个增量 rootfs。

当然，这个想法不是凭空臆造出来的，而是用到了一种叫作联合文件系统（Union File System）的能力。Union File System 也叫 UnionFS，最主要的功能是将多个不同位置的目录联合挂载（union mount）到同一个目录下。

那么，在 Docker 项目中，又是如何使用这种 Union File System 的呢？

Docker运行时由镜像层(image layers)和容器层(container layer)组成。镜像层就是执行Docker image命令时看到的镜像，这些层是只读的。容器层则是创建Docker容器后添加的可写层，所有创建的数据放在可写层中。

Docker Storage Driver(后文统称：Docker存储驱动)把Docker的镜像层和容器层进行组合，得到我们所看到的文件系统，即进入容器后看到的文件系统。

由于不同操作系统(如：CentOS、Ubuntu等)对Docker的支持不一致，所以存在多种存储驱动。接下来我们将深入介绍一下，常见的Docker存储驱动：Aufs、OverlayFs、DeviceMapper，这里主要介绍Aufs和OverlayFs。

##### 一. AUFS

1. 镜像层结构
AUFS是一个联合文件系统，在linux主机上将多个目录分层，并将它们显示为单个目录。这些目录在AUFS术语中称为分支，对应Docker镜像中的层。

接下来，我们以centos镜像为例，说明Aufs文件系统

diff：对应容器的镜像每一层的内容，每个文件夹代表一个层

layers：镜像各层之间依赖关系，每个文件代表一个层，每个文件中按照自上而下的顺序存储着其他镜像层的ID

mnt: 容器最终看到的文件系统的样子，每个正在运行的容器的统一文件系统的挂载点，与容器内完全相同

2. 启动容器
容器启动后会在mnt目录下生成2个新文件夹(简称xxxx和xxxx-init)，init这一层用来记录容器启动时根据系统环境和用户配置自动生成的内容，然后将镜像层+xxxx-init层联合挂载到mnt目录下的xxxx文件夹下。

启动容器后 mnt、layers、diff 目录下可以看到，在基础镜像层上，多了init层和读写层

3. 容器读写
容器读写文件时，根据不同的情况，做出相应的策略：
- 3.1 容器层中不存在该文件，镜像层中存在该文件
读操作：当容器需要读取某文件，且容器层中尚不存在该文件时，则存储驱动程序将从容器层下方的镜像层自上而下开始逐层查找文件，并读取。

写操作：从容器层下方的镜像层自上而下开始逐层查找文件，最后将找到的文件拷贝到容器读写层，然后将更改后的文件存储在容器层

删除操作：删除容器本身的文件，文件并没有被删除，在宿主机diff目录下，相对应的容器目录中，文件以.wh.filename的形式被隐藏，当删除该隐藏文件后，容器中又可以查找到该文件

例如:centos容器中有一自带文件/root/test.sh

当删除该文件后，在宿主机查看对应的diff目录，有.wh.test.sh文件

在宿主机上删除该隐藏文件后，容器中test.sh文件被还原

- 3.2 文件同时位于容器层和镜像层
读写操作：如果文件同时存在于容器层和镜像层，将直接从容器层进行读写。

删除操作：使用3.1中方法对镜像层中要删除的文件进行标记，同时删除容器层中的文件。

- 3.3 文件仅位于容器层中
读写操作：如果该文件仅存在于容器层中，直接在容器层进行读写。

删除操作：直接删除容器层中的文件。

4. Aufs特点

- 1.对于同一个文件的修改只需要复制一次，对同一文件的后续写入操作将对已经复制到容器的文件的副本进行操作，主要区别于devicemapper。

- 2.当许多文件存在于镜像的底层，或者文件很大时将影响aufs的性能。

- 3.因为是文件级存储，当修改文件时会将整个文件复制到读写层，即使只修改其中的一小部分。

##### 二. Overlay2
OverlayFS是一个官方默认并推荐的联合文件系统，类似于AUFS，但速度更快，实现更简单。Docker为OverlayFS提供了两个存储驱动程序：原始版本overlay，升级版本为overlay2。

overlay在linux主机上有两层目录，镜像层和容器层，镜像层又叫lowerdir，容器层又叫upperdir, 容器内的文件系统通过一个名为merged的目录公开，该目录实际上是容器挂载点。

Overlay2的读写与aufs类似，但是由于overlayfs只有两层，lowerdir和upperdir，在读取文件时可以直接从镜像层读取文件，相比aufs的逐层查找要快很多。

1. 镜像层结构

OverlayFS中的目录类型有：
   1. merged：挂载点（mount point）目录，即以用户视角看到的目录。用户的实际文件操作在这里进行，即mount操作的目标目录。
   2. lower：这一层里的文件是不会被修改的，即lower层是只读的。并且OverlayFS支持一次挂载多个lower层的目录。
   3. upper：如果有文件的创建，修改，删除操作，这些操作的执行结果都会在这一层反映出来，即upper是可读写的。upper层内保存了所有对于目标目录的写操作内容。
   4. work：用于存放临时文件的目录。与liunx中的/tmp目录类似。

目录层级关系：lower和upper层的文件会映射到merged层中。此时如果在两个层中存在相同的文件，则在merged层中只会看到upper层的对应文件，即upper层会覆盖lower层。

在/var/lib/Docker/overlay2目录下的显示，overlay只有两层，意味着多层的镜像不能实现为多个overlay层，所以每个镜像的每一层在/var/lib/Docker/overlay2都有对应的目录。

目录var/lib/Docker/overlay2下存放镜像层文件和l目录，var/lib/Docker/overlay2/l存放的都是var/lib/Docker/overlay2中镜像层的软连接。

查看软连接与diff同级的 lower 文件内容：
```log
l/BJRQBKDTPUTSNEPRT3PLWL3CXS:l/JIGVRGJJJBNV7KYFIZ27T3FWSR:l/E7IG6Q3HNVNZQSNCW3PNW6VU6M:l/P3BZFL6RJLC3NRZH7ZOPZRX7BU:l/RMQHN5EPD3Z5ESURVQ5PJPPZYR:l/E4Q4BJX22HSNZFKOOLCZOKTC3I:l/P7G4YO647WYB5RA57TBTDIGJKD:l/322XGYX54K4KCBV7DZMP2JMCWZ:l/HLJUBVY24U7W7OMC7MGI3TLY2W:l/K4OTGXYXHXECV4T7ZH6AVL7DZU:l/RTQ43MFQXIMZRGXQZUX5OQUW4Y:l/VVKMYRBI4EAK7BJPGJUQUNJEHC:l/VOQEOBTZFPJ3WFELOJBGQUH2GO:l/LCKLRCUF2BSZJIXZNOQ2NP5POO:l/OJJ5NSH5OP3GI5L3ZHHLXIU4Y2:l/77EEZUIN2UM4IQZGUK2N2YJOZ4:l/W4DVB2U33K4MGZ6T4LFH467TNR:l/KTI753NY32LLXXBKKRXZS2QGQJ:l/4LTO4AJZHZEY2AWW5JEOLDFCVF
```

Lower文件内容为lowerdir的镜像层级关系。

2. 启动容器
overlay2在linux主机上有两层目录，镜像层和容器层，镜像层又叫lowerdir，容器层又叫upperdir,统一视图通过一个名为merged的目录公开，该目录实际上是容器挂载点，路径在容器层目录下的merged目录中。

3. 容器读写
Overlay2的读写方式和aufs类似。

- 3.1 如果该文件在容器层不存在，镜像层中存在

读操作：即文件不存在upperdir中，则从lowdir中读取。

写操作：将文件拷贝至diff层，并写入数据。

删除操作：在diff目录下对删除文件进行标记。

- 3.2 文件同时位于容器层和镜像层
读写操作：容器层upperdir会覆盖镜像层lowdir中的文件，直接在进行upperdir读写。

删除操作：删除upperdir中的文件，同时在diff目录下对删除文件进行标记。

- 3.3 文件仅位于容器层中

文件只存在upperdir中，则直接从容器中读、写、删除文件。

4. 文件操作实现原理（第三点的另一种解释）
   1. 新建文件
      在merged层中新建的文件会出现在upper层对应的目录中。
   2. 删除文件，在删除文件时，OverlayFS会在所有目录层级中搜索目标文件并对其进行删除操作：
      - 目标文件存在upper层中：文件会在upper目录中被删除。
      - 目标文件存在lower层中：此时在lower目录里的文件不会有变化，OverlayFS会在upper目录中增加了一个特殊文件来标识该文件文件不能出现在 merged/ 里了，即表示文件已经被删除。
    3. 修改操作，OverlayFS会对处于不同目录层级中的目标文件进行不同的修改操作：
       - 文件存在于upper层中：此时只需直接在upper层对应的目录中修改指定文件即可。
       - 文件存在于lower层中：那么就会在upper目录中新建一个文件，新文件中包含更新的内容。而在lower中的原文件不会改变。实际上此时是先将存在于lower层中的文件拷贝到upper层中，然后再修改upper层中的文件。（这种修改数据的方式就是 copy-on-write 写时复制。）
    4. 读取操作
       在查找需要读取的文件时，overlayFS会先从所有的upper层中查找所需的文件。此时overlayFS会按照这些upper层的顺序来自上而下的查找文件。如果查找到目标文件，则直接返回了否则继续向下逐层查找，直到找到文件或者查找完所有upper层。
       如果在所有的upper层中都没有查找到文件，则会在所有的lower层中进行查找。同样,此时overlayFS会按照这些lower层的顺序来自上而下的查找文件。如果查找到目标文件，则直接返回了否则继续向下逐层查找，直到找到文件或者查找完所有lower层。

5. Overlay2特性
- overlayfs支持页缓存共享，也就是说如果多个容器访问同一个文件，可以共享同一个页缓存。
- 但是overlayfs是文件级别的，不是块级别的，这就意味着即使文件只有很少的一点修改也要复制整个文件到容器的读写层，尤其是大文件，会导致写延迟。
- overlay2 是所有当前支持的Linux发行版的首选存储驱动程序，无需额外配置。


#### docker 容器启动流程

1. 启用 Linux Namespace 配置。（即对容器的资源进行隔离）

2. 设置指定的 Cgroups 参数。（即对容器进行资源限制配置）

3. 切换进程的根目录（Change Root）。（使内核能够加载到 rootfs 中的系统数据）

#### 容器数据卷(Data Volume)
在 Docker 项目里，它支持两种 Volume 声明方式，可以把宿主机目录挂载进容器的 /test 目录当中：
```sh
$ docker run -v /test ...
$ docker run -v /home:/test ...
```
而这两种声明方式的本质，实际上是相同的：都是把一个宿主机的目录挂载进了容器的 /test 目录。只不过，在第一种情况下，由于你并没有显示声明宿主机目录，那么 Docker 就会默认在宿主机上创建一个临时目录 /var/lib/docker/volumes/[VOLUME_ID]/_data，然后把它挂载到容器的 /test 目录上。而在第二种情况下，Docker 就直接把宿主机的 /home 目录挂载到容器的 /test 目录上。

Data Volume本质上依旧是host或者host所挂载外部存储的存储空间。

基于上述事实，导致Docker容器无法针对Data Volume进行容量以及权限的调整。即只能在host一侧进行调整。

Data Volume的特点有以下几点：
1. 必须是目录或者文件，不能是裸磁盘/物理设备
2. 容器对Data Volume有读写权限
3. 生命周期与容器独立，容器销毁后Data Volume不会跟随其一同销毁，而是永久保留数据。

对于Data Volume，Docker提供了两种使用方式：bind mount和Docker managed volume。

两种挂载方式实际上是利用了Linux Bind Mount机制。

- Linux Bind Mount机制
Linux Bind Mount的特点有两个：
   - 允许将一个目录或者文件(而不是整个设备)挂载到一个指定目录(挂载点)上
   - 在该挂载点上进行的任何操作只会发生在被挂载的目录或者文件上，而`原挂载点的内容则会被隐藏起来且不受影响。

其实现原理实际上是一个 inode 替换的过程。在 Linux 操作系统中，inode用于保存文件的元数据，可以将其理解为存放文件内容的“对象”。而访问这个inode 所使用的“指针”叫目录项，即dentry。
系统命令 ```mount --bind /home /test```，会将 /home 挂载到 /test 上。其实相当于将 /test 的 dentry的指向目标改为指向/home 的 inode。这样当修改 /test 目录时，实际修改的是 /home 目录的 inode。一旦执行 umount 命令，/test 目录原先的内容就会恢复；因为修改真正发生在的是在 /home 目录里。

- bind mount
这种方式是将host已经存在的目录mount到容器内的指定目录上，因此需要先在host上建立相应的目录或文件后才能进行mount操作并存储数据。由于其实现原理为Linux Bind Mount机制，所以使用bind mount所挂载的容器内目录会被host目录中的数据所覆盖(而非合并)。

在使用细节上，一个host上的目录可以被多个容器进行bind mount。同时，bind mount还支持挂载指定的单一文件；这种操作一般适用于”只需要往容器中添加文件，而不希望覆盖整个目录”的场景。另外bind mount支持更改Data Volume的读写权限，默认是读写权限，可更改为只读权限。

这种类型的Data Volume可以实现host与容器之间的数据共享以及容器间的数据共享。同时基于这些特性，bind mount可看作为是一种静态挂载的Data Volume的操作。

而bind mount的缺点在于可移植性较差。因为其这种静态绑定的特性，需要在每一台host上都要按照相应的规则建立对应的目录或文件才可支持容器使用Data Volume，这会大大增加维护集群的成本。

- Docker managed volume
Docker managed volume与bind mount的唯一区别就是只需要指定容器内的挂载点即可，即Docker来负责维护挂载源。Docker会为容器在host上建立和分配相应的目录来作为挂载源，一般会建立在/var/lib/Docker/volumes/目录下。

实现细节上，如果指定容器内挂载点(目录)中已经存在数据(即容器镜像内自带的数据)，此时Docker会自动将这部分数据复制到Data Volume中。其次，Docker managed volume没有实现对于目录权限的管控，因此Data Volume默认为“全放开”的读写权限。

基于上述特性，Docker managed volume可以看作是一种动态挂载Data Volume的操作。

而Docker managed volume方式的可移植性同样也是较差的。因为在迁移前，需要先查清Docker managed volume动态建立的哪些目录并需要手动备份这些数据。之后还需在新的host上或者是同host上中找到容器对应的新的Data Volume，并将备份的数据拷贝到新的Data Volume中。

- 数据共享
在Docker中数据共享可分为两个方面：
   1. 容器与host的数据共享
   2. 容器之间的数据共享

容器与host的数据共享:
   - bind mount：直接在host上针对Data Volume所对应的源数据目录进行相应的读写操作即可。
   - Docker managed volume：需要使用到Docker cp工具或者是自行查找对应的数据源目录进行cp操作。

容器之间的数据共享:
   - bind mount：使多个容器都mount同一个host目录即可。
   - 其次，还可使用volume container来进行数据共享。volume container实际上就是将一个或多个Data Volume提前挂载在一个容器上，并允许其他的容器与该容器共享这些volume。基于其这种特性，使得volume container能够统一管理Data Volume(因为volume都mount在它上),并且其他容器只需要与volume container进行关联，而非与相关的volume进行关联；即实现了容器与host的解耦。
   - 另外，可使用data-packed volume container来进行数据共享。这种volume container实际上就是提前将数据都压缩到了容器镜像中。

- 数据卷的生命周期管理
   - 数据备份与恢复:
   对于Data Volume的数据备份实际上就是针对相应的host文件系统中的指定目录进行备份。而恢复数据则就是将数据拷贝到Data Volume所对应的host文件系统中的目录内即可。
   - 迁移
   使新的容器在挂载Data Volume时，还保证其能够对应原有的host文件系统中的目录即可。
   - 删除
   bind mount: Docker在删除容器时不会删除相应的host文件目录，因此需要从host层面介入删除数据。
   Docker managed volume：可以在删除容器时加入-v的参数，这样Docker就会自动删除相应的host文件目录了。

#### docker 以及 dockerfile

```py
from flask import Flask
import socket
import os

app = Flask(__name__)

@app.route('/')
def hello():
    html = "<h3>Hello {name}!</h3>" \
           "<b>Hostname:</b> {hostname}<br/>"           
    return html.format(name=os.getenv("NAME", "world"), hostname=socket.gethostname())
    
if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)
```

在这段代码中，我使用 Flask 框架启动了一个 Web 服务器，而它唯一的功能是：如果当前环境中有“NAME”这个环境变量，就把它打印在“Hello”之后，否则就打印“Hello world”，最后再打印出当前环境的 hostname。

这个应用的依赖，则被定义在了同目录下的 requirements.txt 文件里，内容如下所示：
```txt
Flask
```

而将这样一个应用容器化的第一步，是制作容器镜像。不过，相较于之前介绍的制作 rootfs 的过程，Docker 提供了一种更便捷的方式，叫作 Dockerfile，如下所示：
```dockerfile
# 使用官方提供的Python开发镜像作为基础镜像
FROM python:3.7-slim

# 将工作目录切换为/app
WORKDIR /app

# 将当前目录下的所有内容复制到/app下
ADD . /app

# 使用pip命令安装这个应用所需要的依赖
RUN pip install --trusted-host pypi.python.org -r requirements.txt

# 允许外界访问容器的80端口
EXPOSE 80

# 设置环境变量
ENV NAME PY3.7

# 设置容器进程为：python app.py，即：这个Python应用的启动命令
CMD ["python", "app.py"]
```

通过这个文件的内容，可以看到 Dockerfile 的设计思想，是使用一些标准的原语（即大写高亮的词语），描述我们所要构建的 Docker 镜像。并且这些原语，都是按顺序处理的。

比如 FROM 原语，指定了“python:3.7-slim”这个官方维护的基础镜像，从而免去了安装 Python 等语言环境的操作。否则，这一段我们就得这么写了：
```dockerfile
FROM ubuntu:latest
RUN apt-get update -yRUN apt-get install -y python-pip python-dev build-essential
# ...
```

其中，RUN 原语就是在容器里执行 shell 命令的意思。

而 WORKDIR，意思是在这一句之后，Dockerfile 后面的操作都以这一句指定的 /app 目录作为当前目录。

所以，到了最后的 CMD，意思是 Dockerfile 指定 python app.py 为这个容器的进程。这里，app.py 的实际路径是 /app/app.py。所以，CMD ["python", "app.py"]等价于"docker run python app.py"。

另外，在使用 Dockerfile 时，你可能还会看到一个叫作 ENTRYPOINT 的原语。实际上，它和 CMD 都是 Docker 容器进程启动所必需的参数，完整执行格式是：“ENTRYPOINT CMD”。

但是，默认情况下，Docker 会为你提供一个隐含的 ENTRYPOINT，即：/bin/sh -c。所以，在不指定 ENTRYPOINT 时，比如在我们这个例子里，实际上运行在容器里的完整进程是：/bin/sh -c "python app.py"，即 CMD 的内容就是 ENTRYPOINT 的参数。

基于以上原因，下面会统一称 Docker 容器的启动进程为 ENTRYPOINT，而不是 CMD。

需要注意的是，Dockerfile 里的原语并不都是指对容器内部的操作。就比如 ADD，它指的是把当前目录（即 Dockerfile 所在的目录）里的文件，复制到指定容器内的目录当中。

接下来，我就可以让 Docker 制作这个镜像了，在当前目录执行：
```sh
docker build -t hellopy .
```

其中，-t 的作用是给这个镜像加一个 Tag，即：起一个好听的名字。docker build 会自动加载当前目录下的 Dockerfile 文件，然后按照顺序，执行文件中的原语。而这个过程，实际上可以等同于 Docker 使用基础镜像启动了一个容器，然后在容器中依次执行 Dockerfile 中的原语。

需要注意的是，Dockerfile 中的每个原语执行后，都会生成一个对应的镜像层。即使原语本身并没有明显地修改文件的操作（比如，ENV 原语），它对应的层也会存在。只不过在外界看来，这个层是空的。

接下来，我使用这个镜像，通过 docker run 命令启动容器：
```sh
docker run -p 4000:80 hellopy
```

在这一句命令中，镜像名 hellopy 后面，我什么都不用写，因为在 Dockerfile 中已经指定了 CMD。否则，我就得把进程的启动命令加在后面：
```sh
docker run -p 4000:80 hellopy python app.py
```

容器启动之后，我可以使用 docker ps 命令看到运行中的容器

同时，我已经通过 -p 4000:80 告诉了 Docker，请把容器内的 80 端口映射在宿主机的 4000 端口上。

这样做的目的是，只要访问宿主机的 4000 端口，我就可以看到容器里应用返回的结果

否则，就得先用 docker inspect 命令查看容器的 IP 地址，然后访问“http://< 容器 IP 地址 >:80”才可以看到容器内应用的返回。

为了能够上传镜像，我首先需要注册一个 Docker Hub 账号（本人一般上传至阿里云的个人仓库中），然后使用 docker login 命令登录。

接下来，要用 docker tag 命令给容器镜像起一个完整的名字：

```sh
docker tag hellopy ipso/hellopy:v1
```

其中，ipso 是我在 Docker Hub 上的用户名，它的“学名”叫镜像仓库（Repository）；“/”后面的 hellopy 是这个镜像的名字，而“v1”则是我给这个镜像分配的版本号。

这样，我就可以把这个镜像上传到 Docker Hub 上了。

**此外，我还可以使用 docker commit 指令，把一个正在运行的容器，直接提交为一个镜像。一般来说，需要这么操作原因是：这个容器运行起来后，我又在里面做了一些操作，并且要把操作结果保存到镜像里，比如：**
```sh
$ docker exec -it 4ddf4638572c /bin/sh
# 在容器内部新建了一个文件
root@4ddf4638572d:/app# touch test.txt
root@4ddf4638572d:/app# exit

#将这个新建的文件提交到镜像中保存
$ docker commit 4ddf4638572c ipso/hellopy:v2
```

这里，我使用了 docker exec 命令进入到了容器当中。在了解了 Linux Namespace 的隔离机制后，你应该会很自然地想到一个问题：docker exec 是怎么做到进入容器里的呢？

实际上，Linux Namespace 创建的隔离空间虽然看不见摸不着，但一个进程的 Namespace 信息在宿主机上是确确实实存在的，并且是以一个文件的方式存在。

比如，通过如下指令，你可以看到当前正在运行的 Docker 容器的进程号（PID）是 2290900
```sh
$ docker inspect --format '{{ .State.Pid }}' 4ddf4638572c
2290900
```

这时，你可以通过查看宿主机的 proc 文件，看到这个 2290900 进程的所有 Namespace 对应的文件：
```sh
[root@ttt ~]# ls -l /proc/2290900/ns
total 0
lrwxrwxrwx 1 root root 0 Dec 13 16:17 cgroup -> 'cgroup:[4026532678]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 ipc -> 'ipc:[4026532621]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 mnt -> 'mnt:[4026532619]'
lrwxrwxrwx 1 root root 0 Oct 17 14:34 net -> 'net:[4026532623]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 pid -> 'pid:[4026532622]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 pid_for_children -> 'pid:[4026532622]'  
lrwxrwxrwx 1 root root 0 Dec 13 16:17 time -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0 Dec 13 16:17 uts -> 'uts:[4026532620]'
```

可以看到，一个进程的每种 Linux Namespace，都在它对应的 /proc/[进程号]/ns 下有一个对应的虚拟文件，并且链接到一个真实的 Namespace 文件上。

有了这样一个可以“hold 住”所有 Linux Namespace 的文件，我们就可以对 Namespace 做一些很有意义事情了，比如：加入到一个已经存在的 Namespace 当中。

**这也就意味着：一个进程，可以选择加入到某个进程已有的 Namespace 当中，从而达到“进入”这个进程所在容器的目的，这正是 docker exec 的实现原理。**

而这个操作所依赖的，乃是一个名叫 setns() 的 Linux 系统调用
```c
#define _GNU_SOURCE
#include <fcntl.h>
#include <sched.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>

#define errExit(msg) do { perror(msg); exit(EXIT_FAILURE);} while (0)

int main(int argc, char *argv[]) {
    int fd;
    
    fd = open(argv[1], O_RDONLY);
    if (setns(fd, 0) == -1) {
        errExit("setns");
    }
    execvp(argv[2], &argv[2]); 
    errExit("execvp");
}
```

上面代码，请使用 golang 实现

这段代码功能非常简单：它一共接收两个参数，第一个参数是 argv[1]，即当前进程要加入的 Namespace 文件的路径，比如 /proc/25686/ns/net；而第二个参数，则是你要在这个 Namespace 里运行的进程，比如 /bin/bash。

这段代码的核心操作，则是通过 open() 系统调用打开了指定的 Namespace 文件，并把这个文件的描述符 fd 交给 setns() 使用。在 setns() 执行后，当前进程就加入了这个文件对应的 Linux Namespace 当中了。

上面代码使用 golang 实现：
```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: <program> <namespace-file> <command> [args...]")
		os.Exit(1)
	}

	// 打开命名空间文件
	fd, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening namespace file: %v\n", err)
		os.Exit(1)
	}
	defer fd.Close()

	// 调用 setns 系统调用加入命名空间
	if err := unix.Setns(int(fd.Fd()), unix.CLONE_NEWNS); err != nil {
		fmt.Fprintf(os.Stderr, "Error calling setns: %v\n", err)
		os.Exit(1)
	}

	// 执行新的程序
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
```

- 提交镜像的操作 docker commit 上来吧
docker commit，实际上就是在容器运行起来后，把最上层的“可读写层”，加上原先容器镜像的只读层，打包组成了一个新的镜像。当然，下面这些只读层在宿主机上是共享的，不会占用额外的空间。

而由于使用了联合文件系统，你在容器里对镜像 rootfs 所做的任何修改，都会被操作系统先复制到这个可读写层，然后再修改。这就是所谓的：Copy-on-Write。

而正如前所说，Init 层的存在，就是为了避免你执行 docker commit 时，把 Docker 自己对 /etc/hosts 等文件做的修改，也一起提交掉。


## # 从容器到容器云：理解 Kubernetes 的本质

#### 首先，Kubernetes 项目要解决的问题是什么？
编排？调度？容器云？还是集群管理？

实际上，这个问题到目前为止都没有固定的答案。因为在不同的发展阶段，Kubernetes 需要着重解决的问题是不同的。

但是，对于大多数用户来说，他们希望 Kubernetes 项目带来的体验是确定的：现在我有了应用的容器镜像，请帮我在一个给定的集群上把这个应用运行起来。

更进一步地说，我还希望 Kubernetes 能给我提供路由网关、水平扩展、监控、备份、灾难恢复等一系列运维能力。

等一下，这些功能听起来好像有些耳熟？这不就是经典 PaaS（比如，Cloud Foundry）项目的能力吗？

而且，有了 Docker 之后，我根本不需要什么 Kubernetes、PaaS，只要使用 Docker 公司的 Compose+Swarm 项目，就完全可以很方便地 DIY 出这些功能了！

所以说，如果 Kubernetes 项目只是停留在拉取用户镜像、运行容器，以及提供常见的运维功能的话，那么别说跟“原生”的 Docker Swarm 项目竞争了，哪怕跟经典的 PaaS 项目相比也难有什么优势可言。

而实际上，在定义核心功能的过程中，Kubernetes 项目正是依托着 Borg 项目的理论优势，才在短短几个月内迅速站稳了脚跟。

Kubernetes 项目的架构，跟它的原型项目 Borg 非常类似，都由 Master 和 Node 两种节点组成，而这两种角色分别对应着控制节点和计算节点。

其中，控制节点，即 Master 节点，由三个紧密协作的独立组件组合而成，它们分别是负责 API 服务的 kube-apiserver、负责调度的 kube-scheduler，以及负责容器编排的 kube-controller-manager。整个集群的持久化数据，则由 kube-apiserver 处理后保存在 Etcd 中。

而计算节点上最核心的部分，则是一个叫作 kubelet 的组件。

在 Kubernetes 项目中，kubelet 主要负责同容器运行时（比如 Docker 项目）打交道。而这个交互所依赖的，是一个称作 CRI（Container Runtime Interface）的远程调用接口，这个接口定义了容器运行时的各项核心操作，比如：启动一个容器需要的所有参数。

这也是为何，Kubernetes 项目并不关心你部署的是什么容器运行时、使用的什么技术实现，只要你的这个容器运行时能够运行标准的容器镜像，它就可以通过实现 CRI 接入到 Kubernetes 项目当中。

而具体的容器运行时，比如 Docker 项目，则一般通过 OCI 这个容器运行时规范同底层的 Linux 操作系统进行交互，即：把 CRI 请求翻译成对 Linux 操作系统的调用（操作 Linux Namespace 和 Cgroups 等）。

此外，kubelet 还通过 gRPC 协议同一个叫作 Device Plugin 的插件进行交互。这个插件，是 Kubernetes 项目用来管理 GPU 等宿主机物理设备的主要组件，也是基于 Kubernetes 项目进行机器学习训练、高性能作业支持等工作必须关注的功能。

而 kubelet 的另一个重要功能，则是调用网络插件和存储插件为容器配置网络和持久化存储。这两个插件与 kubelet 进行交互的接口，分别是 CNI（Container Networking Interface）和 CSI（Container Storage Interface）。

实际上，kubelet 这个奇怪的名字，来自于 Borg 项目里的同源组件 Borglet。不过，如果你浏览过 Borg 论文的话，就会发现，这个命名方式可能是 kubelet 组件与 Borglet 组件的唯一相似之处。因为 Borg 项目，并不支持我们这里所讲的容器技术，而只是简单地使用了 Linux Cgroups 对进程进行限制。这就意味着，像 Docker 这样的“容器镜像”在 Borg 中是不存在的，Borglet 组件也自然不需要像 kubelet 这样考虑如何同 Docker 进行交互、如何对容器镜像进行管理的问题，也不需要支持 CRI、CNI、CSI 等诸多容器技术接口。

可以说，kubelet 完全就是为了实现 Kubernetes 项目对容器的管理能力而重新实现的一个组件，与 Borg 之间并没有直接的传承关系。

那么，Borg 对于 Kubernetes 项目的指导作用又体现在哪里呢？

答案是，Master 节点。虽然在 Master 节点的实现细节上 Borg 项目与 Kubernetes 项目不尽相同，但它们的出发点却高度一致，即：如何编排、管理、调度用户提交的作业？

所以，Borg 项目完全可以把 Docker 镜像看作一种新的应用打包方式。这样，Borg 团队过去在大规模作业管理与编排上的经验就可以直接“套”在 Kubernetes 项目上了。

这些经验最主要的表现就是，从一开始，Kubernetes 项目就没有像同时期的各种“容器云”项目那样，把 Docker 作为整个架构的核心，而仅仅把它作为最底层的一个容器运行时实现。（太高明的设计。不要依赖细节，要依赖抽象）

而 Kubernetes 项目要着重解决的问题，则来自于 Borg 的研究人员在论文中提到的一个非常重要的观点：运行在大规模集群中的各种任务之间，实际上存在着各种各样的关系。这些关系的处理，才是作业编排和管理系统最困难的地方。

Kubernetes 项目最主要的设计思想是，从更宏观的角度，以统一的方式来定义任务之间的各种关系，并且为将来支持更多种类的关系留有余地。

比如，Kubernetes 项目对容器间的“访问”进行了分类，首先总结出了一类非常常见的“紧密交互”的关系，即：这些应用之间需要非常频繁的交互和访问；又或者，它们会直接通过本地文件进行信息交换。

在常规环境下，这些应用往往会被直接部署在同一台机器上，通过 Localhost 通信，通过本地磁盘目录交换文件。而在 Kubernetes 项目中，这些容器则会被划分为一个“Pod”，Pod 里的容器共享同一个 Network Namespace、同一组数据卷，从而达到高效率交换信息的目的。

而对于另外一种更为常见的需求，比如 Web 应用与数据库之间的访问关系，Kubernetes 项目则提供了一种叫作“Service”的服务。像这样的两个应用，往往故意不部署在同一台机器上，这样即使 Web 应用所在的机器宕机了，数据库也完全不受影响。可是，我们知道，对于一个容器来说，它的 IP 地址等信息不是固定的，那么 Web 应用又怎么找到数据库容器的 Pod 呢？

所以，Kubernetes 项目的做法是给 Pod 绑定一个 Service 服务，而 Service 服务声明的 IP 地址等信息是“终生不变”的。这个 Service 服务的主要作用，就是作为 Pod 的代理入口（Portal），从而代替 Pod 对外暴露一个固定的网络地址。

这样，对于 Web 应用的 Pod 来说，它需要关心的就是数据库 Pod 的 Service 信息。不难想象，Service 后端真正代理的 Pod 的 IP 地址、端口等信息的自动更新、维护，则是 Kubernetes 项目的职责。

我们从容器这个最基础的概念出发，首先遇到了容器间“紧密协作”关系的难题，于是就扩展到了 Pod；有了 Pod 之后，我们希望能一次启动多个应用的实例，这样就需要 Deployment 这个 Pod 的多实例管理器；而有了这样一组相同的 Pod 后，我们又需要通过一个固定的 IP 地址和端口以负载均衡的方式访问它，于是就有了 Service。

可是，如果现在两个不同 Pod 之间不仅有“访问关系”，还要求在发起时加上授权信息。最典型的例子就是 Web 应用对数据库访问时需要 Credential（数据库的用户名和密码）信息。那么，在 Kubernetes 中这样的关系又如何处理呢？

Kubernetes 项目提供了一种叫作 Secret 的对象，它其实是一个保存在 Etcd 里的键值对数据。这样，你把 Credential 信息以 Secret 的方式存在 Etcd 里，Kubernetes 就会在你指定的 Pod（比如，Web 应用的 Pod）启动时，自动把 Secret 里的数据以 Volume 的方式挂载到容器里。这样，这个 Web 应用就可以访问数据库了。

除了应用与应用之间的关系外，应用运行的形态是影响“如何容器化这个应用”的第二个重要因素。

为此，Kubernetes 定义了新的、基于 Pod 改进后的对象。比如 Job，用来描述一次性运行的 Pod（比如，大数据任务）；再比如 DaemonSet，用来描述每个宿主机上必须且只能运行一个副本的守护进程服务；又比如 CronJob，则用于描述定时任务等等。

Kubernetes 项目中，我们所推崇的使用方法是：
- 首先，通过一个“编排对象”，比如 Pod、Job、CronJob 等，来描述你试图管理的应用；
- 然后，再为它定义一些“服务对象”，比如 Service、Secret、Horizontal Pod Autoscaler（自动水平扩展器）等。这些对象，会负责具体的平台级功能。

**这种使用方法，就是所谓的“声明式 API”。这种 API 对应的“编排对象”和“服务对象”，都是 Kubernetes 项目中的 API 对象（API Object）。**

实际上，过去很多的集群管理项目（比如 Yarn、Mesos，以及 Swarm）所擅长的，都是把一个容器，按照某种规则，放置在某个最佳节点上运行起来。这种功能，我们称为“调度”。

而 Kubernetes 项目所擅长的，是按照用户的意愿和整个系统的规则，完全自动化地处理好容器之间的各种关系。这种功能，就是我们经常听到的一个概念：编排。

所以说，Kubernetes 项目的本质，是为用户提供一个具有普遍意义的容器编排工具。不过，更重要的是，Kubernetes 项目为用户提供的不仅限于一个工具。它真正的价值，乃在于提供了一套基于容器构建分布式系统的基础依赖。

#### yaml 配置文件

Kubernetes 跟 Docker 等很多项目最大的不同，就在于它不推荐你使用命令行的方式直接运行容器（虽然 Kubernetes 项目也支持这种方式，比如：kubectl run），而是希望你用 YAML 文件的方式，即：把容器的定义、参数、配置，统统记录在一个 YAML 文件中，然后用这样一句指令把它运行起来：```kubectl create/apply -f 我的配置文件```

这么做最直接的好处是，你会有一个文件能记录下 Kubernetes 到底“run”了什么。比如下面这个例子：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
```

像这样的一个 YAML 文件，对应到 Kubernetes 中，就是一个 API Object（API 对象）。当你为这个对象的各个字段填好值并提交给 Kubernetes 之后，Kubernetes 就会负责创建出这些对象所定义的容器或者其他类型的 API 资源。

可以看到，这个 YAML 文件中的 Kind 字段，指定了这个 API 对象的类型（Type），是一个 Deployment。

所谓 Deployment，是一个定义多副本应用（即多个副本 Pod）的对象。此外，Deployment 还负责在 Pod 定义发生变化时，对每个副本进行滚动更新（Rolling Update）。

需要注意的是，像这样使用一种 API 对象（Deployment）管理另一种 API 对象（Pod）的方法，在 Kubernetes 中，叫作“控制器”模式（controller pattern）。在我们的例子中，Deployment 扮演的正是 Pod 的控制器的角色。

你可能还注意到，这样的每一个 API 对象都有一个叫作 Metadata 的字段，这个字段就是 API 对象的“标识”，即元数据，它也是我们从 Kubernetes 里找到这个对象的主要依据。这其中最主要使用到的字段是 Labels。

顾名思义，Labels 就是一组 key-value 格式的标签。而像 Deployment 这样的控制器对象，就可以通过这个 Labels 字段从 Kubernetes 中过滤出它所关心的被控制对象。

比如，在上面这个 YAML 文件中，Deployment 会把所有正在运行的、携带“app: nginx”标签的 Pod 识别为被管理的对象，并确保这些 Pod 的总数严格等于两个。

而这个过滤规则的定义，是在 Deployment 的“spec.selector.matchLabels”字段。我们一般称之为：Label Selector。

另外，在 Metadata 中，还有一个与 Labels 格式、层级完全相同的字段叫 Annotations，它专门用来携带 key-value 格式的内部信息。所谓内部信息，指的是对这些信息感兴趣的，是 Kubernetes 组件本身，而不是用户。所以大多数 Annotations，都是在 Kubernetes 运行过程中，被自动加在这个 API 对象上。

**一个 Kubernetes 的 API 对象的定义，大多可以分为 Metadata 和 Spec 两个部分。前者存放的是这个对象的元数据，对所有 API 对象来说，这一部分的字段和格式基本上是一样的；而后者存放的，则是属于这个对象独有的定义，用来描述它所要表达的功能。**

使用```kubectl get pods -l app=nginx```

kubectl get 指令的作用，就是从 Kubernetes 里面获取（GET）指定的 API 对象。可以看到，在这里我还加上了一个 -l 参数，即获取所有匹配 app: nginx 标签的 Pod。需要注意的是，在命令行中，所有 key-value 格式的参数，都使用“=”而非“:”表示。

在 Kubernetes 执行的过程中，对 API 对象的所有重要操作，都会被记录在这个对象的 Events 里，并且显示在 kubectl describe 指令返回的结果中。

比如，对于这个 Pod，我们可以看到它被创建之后，被调度器调度（Successfully assigned）到了 node-1，拉取了指定的镜像（pulling image），然后启动了 Pod 里定义的容器（Started container）。

所以，这个部分正是我们将来进行 Debug 的重要依据。如果有异常发生，你一定要第一时间查看这些 Events，往往可以看到非常详细的错误信息。

推荐使用 kubectl apply 命令，来统一进行 Kubernetes 对象的创建和更新操作。这样的操作方法，是 Kubernetes“声明式 API”所推荐的使用方法。也就是说，作为用户，你不必关心当前的操作是创建，还是更新，你执行的命令始终是 kubectl apply，而 Kubernetes 则会根据 YAML 文件的内容变化，自动进行具体的处理。

所以说，如果通过容器镜像，我们能够保证应用本身在开发与部署环境里的一致性的话，那么现在，Kubernetes 项目通过这些 YAML 文件，就保证了应用的“部署参数”在开发与部署环境中的一致性。

而当应用本身发生变化时，开发人员和运维人员可以依靠容器镜像来进行同步；当应用部署参数发生变化时，这些 YAML 文件就是他们相互沟通和信任的媒介。


## # Kubernetes 一键部署利器之 kubeadm

个项目的目的，就是要让用户能够通过这样两条指令完成一个 Kubernetes 集群的部署：
```sh
# 创建一个Master节点
kubeadm init

# 将一个Node节点加入到当前集群中
kubeadm join <Master节点的IP和端口>
```

是不是非常方便呢？不过，你可能也会有所顾虑：Kubernetes 的功能那么多，这样一键部署出来的集群，能用于生产环境吗？

#### kubeadm 的工作原理
Kubernetes 的架构和它的组件。在部署时，它的每一个组件都是一个需要被执行的、单独的二进制文件。所以不难想象，SaltStack 这样的运维工具或者由社区维护的脚本的功能，就是要把这些二进制文件传输到指定的机器当中，然后编写控制脚本来启停这些组件。

1. 为什么不用容器部署 Kubernetes 呢？
这样做会带来一个很麻烦的问题，即：如何容器化 kubelet。

kubelet 是 Kubernetes 项目用来操作 Docker 等容器运行时的核心组件。可是，除了跟容器运行时打交道外，kubelet 在配置容器网络、管理容器数据卷时，都需要直接操作宿主机。

而如果现在 kubelet 本身就运行在一个容器里，那么直接操作宿主机就会变得很麻烦。对于网络配置来说还好，kubelet 容器可以通过不开启 Network Namespace（即 Docker 的 host network 模式）的方式，直接共享宿主机的网络栈。可是，要让 kubelet 隔着容器的 Mount Namespace 和文件系统，操作宿主机的文件系统，就有点儿困难了。

正因为如此，kubeadm 选择了一种妥协方案：把 kubelet 直接运行在宿主机上，然后使用容器部署其他的 Kubernetes 组件。

所以，你使用 kubeadm 的第一步，是在机器上手动安装 kubeadm、kubelet 和 kubectl 这三个二进制文件。当然，kubeadm 的作者已经为各个发行版的 Linux 准备好了安装包，所以你只需要执行：```apt-get install kubeadm```就可以了。

接下来，你就可以使用“kubeadm init”部署 Master 节点了。

#### kubeadm init 的工作流程
当你执行 kubeadm init 指令后，kubeadm 首先要做的，是一系列的检查工作，以确定这台机器可以用来部署 Kubernetes。这一步检查，我们称为“Preflight Checks”，它可以为你省掉很多后续的麻烦。其实，Preflight Checks 包括了很多方面，比如：
- Linux 内核的版本必须是否是 3.10 以上？
- Linux Cgroups 模块是否可用？
- 机器的 hostname 是否标准？在 Kubernetes 项目里，机器的名字以及一切存储在 Etcd 中的 API 对象，都必须使用标准的 DNS 命名（RFC 1123）。
- 用户安装的 kubeadm 和 kubelet 的版本是否匹配？
- 机器上是不是已经安装了 Kubernetes 的二进制文件？
- Kubernetes 的工作端口 10250/10251/10252 端口是不是已经被占用？
- ip、mount 等 Linux 指令是否存在？
- 容器运行时是否已安装？
- ...

在通过了 Preflight Checks 之后，kubeadm 要为你做的，是生成 Kubernetes 对外提供服务所需的各种证书和对应的目录。

**Kubernetes 对外提供服务时，除非专门开启“不安全模式”，否则都要通过 HTTPS 才能访问 kube-apiserver。这就需要为 Kubernetes 集群配置好证书文件。**

kubeadm 为 Kubernetes 项目生成的证书文件都放在 Master 节点的 /etc/kubernetes/pki 目录下。在这个目录下，最主要的证书文件是 ca.crt 和对应的私钥 ca.key。

此外，用户使用 kubectl 获取容器日志等 streaming 操作时，需要通过 kube-apiserver 向 kubelet 发起请求，这个连接也必须是安全的。kubeadm 为这一步生成的是 apiserver-kubelet-client.crt 文件，对应的私钥是 apiserver-kubelet-client.key。

除此之外，Kubernetes 集群中还有 Aggregate APIServer 等特性，也需要用到专门的证书，这里就不再一一列举了。需要指出的是，你可以选择不让 kubeadm 为你生成这些证书，而是拷贝现有的证书到如下证书的目录里：
```sh
/etc/kubernetes/pki/ca.{crt,key}
```

这时，kubeadm 就会跳过证书生成的步骤，把它完全交给用户处理。

证书生成后，kubeadm 接下来会为其他组件生成访问 kube-apiserver 所需的配置文件。这些文件的路径是：/etc/kubernetes/xxx.conf：
```sh
ls /etc/kubernetes/
admin.conf  controller-manager.conf  kubelet.conf  scheduler.conf
```

这些文件里面记录的是，当前这个 Master 节点的服务器地址、监听端口、证书目录等信息。这样，对应的客户端（比如 scheduler，kubelet 等），可以直接加载相应的文件，使用里面的信息与 kube-apiserver 建立安全连接。(这些文件内容格式居然也是yaml，而且也是api对象配置格式)

接下来，kubeadm 会为 Master 组件生成 Pod 配置文件。Kubernetes 有三个 Master 组件 kube-apiserver、kube-controller-manager、kube-scheduler，而它们都会被使用 Pod 的方式部署起来。

你可能会有些疑问：这时，Kubernetes 集群尚不存在，难道 kubeadm 会直接执行 docker run 来启动这些容器吗？

当然不是。

在 Kubernetes 中，有一种特殊的容器启动方法叫做“Static Pod”。它允许你把要部署的 Pod 的 YAML 文件放在一个指定的目录里。这样，当这台机器上的 kubelet 启动时，它会自动检查这个目录，加载所有的 Pod YAML 文件，然后在这台机器上启动它们。

从这一点也可以看出，kubelet 在 Kubernetes 项目中的地位非常高，在设计上它就是一个完全独立的组件，而其他 Master 组件，则更像是辅助性的系统容器。

在 kubeadm 中，Master 组件的 YAML 文件会被生成在 /etc/kubernetes/manifests 路径下：
```sh
ls /etc/kubernetes/manifests
etcd.yaml  kube-apiserver.yaml  kube-controller-manager.yaml  kube-scheduler.yaml
```

如etcd.yaml:
```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubeadm.kubernetes.io/etcd.advertise-client-urls: https://172.31.212.179:2379
  creationTimestamp: null
  labels:
    component: etcd
    tier: control-plane
  name: etcd
  namespace: kube-system
spec:
  containers:
  - command:
    - etcd
    - --advertise-client-urls=https://172.31.212.179:2379
    - --cert-file=/etc/kubernetes/pki/etcd/server.crt
    - --client-cert-auth=true
    - --data-dir=/var/lib/etcd
    - --experimental-initial-corrupt-check=true
    - --experimental-watch-progress-notify-interval=5s
    - --initial-advertise-peer-urls=https://172.31.212.179:2380
    - --initial-cluster=k8s-master=https://172.31.212.179:2380
    - --key-file=/etc/kubernetes/pki/etcd/server.key
    - --listen-client-urls=https://127.0.0.1:2379,https://172.31.212.179:2379
    - --listen-metrics-urls=http://127.0.0.1:2381
    - --listen-peer-urls=https://172.31.212.179:2380
    - --name=k8s-master
    - --peer-cert-file=/etc/kubernetes/pki/etcd/peer.crt
    - --peer-client-cert-auth=true
    - --peer-key-file=/etc/kubernetes/pki/etcd/peer.key
    - --peer-trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt
    - --snapshot-count=10000
    - --trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt
    image: registry.cn-hangzhou.aliyuncs.com/google_containers/etcd:3.5.7-0
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 8
      httpGet:
        host: 127.0.0.1
        path: /health?exclude=NOSPACE&serializable=true
        port: 2381
        scheme: HTTP
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 15
    name: etcd
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
    startupProbe:
      failureThreshold: 24
      httpGet:
        host: 127.0.0.1
        path: /health?serializable=false
        port: 2381
        scheme: HTTP
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 15
    volumeMounts:
    - mountPath: /var/lib/etcd
      name: etcd-data
    - mountPath: /etc/kubernetes/pki/etcd
      name: etcd-certs
  hostNetwork: true
  priority: 2000001000
  priorityClassName: system-node-critical
  securityContext:
    seccompProfile:
      type: RuntimeDefault
  volumes:
  - hostPath:
      path: /etc/kubernetes/pki/etcd
      type: DirectoryOrCreate
    name: etcd-certs
  - hostPath:
      path: /var/lib/etcd
      type: DirectoryOrCreate
    name: etcd-data
status: {}
```

需要注意：
1. 这个 Pod 里只定义了一个容器。
2. 这个容器的启动命令（commands）非常长的命令。其实，它就是容器里 etcd 这个二进制文件再加上指定的配置参数而已。
3. 如果你要修改一个已有集群的 etcd 的配置，需要修改这个 YAML 文件。
4. 这些组件的参数也可以在部署时指定。


而一旦这些 YAML 文件出现在被 kubelet 监视的 /etc/kubernetes/manifests 目录下，kubelet 就会自动创建这些 YAML 文件中定义的 Pod，即 Master 组件的容器。

Master 容器启动后，kubeadm 会通过检查 localhost:6443/healthz 这个 Master 组件的健康检查 URL，等待 Master 组件完全运行起来。

然后，kubeadm 就会为集群生成一个 bootstrap token。在后面，只要持有这个 token，任何一个安装了 kubelet 和 kubadm 的节点，都可以通过 kubeadm join 加入到这个集群当中。

这个 token 的值和使用方法，会在 kubeadm init 结束后被打印出来。

在 token 生成之后，kubeadm 会将 ca.crt 等 Master 节点的重要信息，通过 ConfigMap 的方式保存在 Etcd 当中，供后续部署 Node 节点使用。这个 ConfigMap 的名字是 cluster-info。

kubeadm init 的最后一步，就是安装默认插件。Kubernetes 默认 kube-proxy 和 DNS 这两个插件是必须安装的。它们分别用来提供整个集群的服务发现和 DNS 功能。其实，这两个插件也只是两个容器镜像而已，所以 kubeadm 只要用 Kubernetes 客户端创建两个 Pod 就可以了。


#### kubeadm join 的工作流程
这个流程其实非常简单，kubeadm init 生成 bootstrap token 之后，你就可以在任意一台安装了 kubelet 和 kubeadm 的机器上执行 kubeadm join 了。

可是，为什么执行 kubeadm join 需要这样一个 token 呢？

因为，任何一台机器想要成为 Kubernetes 集群中的一个节点，就必须在集群的 kube-apiserver 上注册。可是，要想跟 apiserver 打交道，这台机器就必须要获取到相应的证书文件（CA 文件）。可是，为了能够一键安装，我们就不能让用户去 Master 节点上手动拷贝这些文件。

所以，kubeadm 至少需要发起一次“不安全模式”的访问到 kube-apiserver，从而拿到保存在 ConfigMap 中的 cluster-info（它保存了 APIServer 的授权信息）。而 bootstrap token，扮演的就是这个过程中的安全验证的角色。

只要有了 cluster-info 里的 kube-apiserver 的地址、端口、证书，kubelet 就可以以“安全模式”连接到 apiserver 上，这样一个新的节点就部署完成了。


#### 配置 kubeadm 的部署参数
前面讲了 kubeadm 部署 Kubernetes 集群最关键的两个步骤，kubeadm init 和 kubeadm join。相信你一定会有这样的疑问：kubeadm 确实简单易用，可是我又该如何定制我的集群组件参数呢？

比如，我要指定 kube-apiserver 的启动参数，该怎么办？

在这里，我强烈推荐你在使用 kubeadm init 部署 Master 节点时，使用下面这条指令：
```sh
$ kubeadm init --config kubeadm.yaml
```

这时，你就可以给 kubeadm 提供一个 YAML 文件（比如，kubeadm.yaml），它的内容如下所示（这里仅列举了主要部分）：
```yaml
apiVersion: kubeadm.k8s.io/v1alpha2
kind: MasterConfiguration
kubernetesVersion: v1.27.2
api:
  advertiseAddress: 192.168.0.102
  bindPort: 6443
  # ...
etcd:
  local:
    dataDir: /var/lib/etcd
    image: ""
imageRepository: k8s.gcr.io
kubeProxy:
  config:
    bindAddress: 0.0.0.0
    # ...
kubeletConfiguration:
  baseConfig:
    address: 0.0.0.0
    # ...
networking:
  dnsDomain: cluster.local
  podSubnet: ""
  serviceSubnet: 10.96.0.0/12
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  # ...
```

通过制定这样一个部署参数配置文件，你就可以很方便地在这个文件里填写各种自定义的部署参数了。比如，我现在要指定 kube-apiserver 的参数，那么我只要在这个文件里加上这样一段信息：
```yaml
# ...
apiServerExtraArgs:
  advertise-address: 192.168.0.103
  anonymous-auth: false
  enable-admission-plugins: AlwaysPullImages,DefaultStorageClass
  audit-log-path: /home/johndoe/audit.log
```

然后，kubeadm 就会使用上面这些信息替换 /etc/kubernetes/manifests/kube-apiserver.yaml 里的 command 字段里的参数了。

而这个 YAML 文件提供的可配置项远不止这些。比如，你还可以修改 kubelet 和 kube-proxy 的配置，修改 Kubernetes 使用的基础镜像的 URL（默认的k8s.gcr.io/xxx镜像 URL 在国内访问是有困难的），指定自己的证书文件，指定特殊的容器运行时等等。

**其实国内用kubeadm安装集群最大的拦路虎在于有几个镜像没法下载，建议大家先手动把镜像pull 下来，从阿里的镜像源上，然后tag成安装所需的镜像名称，这样你发现安装过程会异常顺利**

#### 安装一般需要准备的工作
首先，准备机器。最直接的办法，自然是到公有云上申请几个虚拟机。当然，如果条件允许的话，拿几台本地的物理服务器来组集群是最好不过了。这些机器只要满足如下几个条件即可：

1. 满足安装 Docker 项目所需的要求，比如 64 位的 Linux 操作系统、3.10 及以上的内核版本；
2. x86 或者 ARM 架构均可；
3. 机器之间网络互通，这是将来容器之间网络互通的前提；
4. 有外网访问权限，因为需要拉取镜像；
5. 能够访问到gcr.io、quay.io这两个 docker registry，因为有小部分镜像需要在这里拉取；
6. 单机可用资源建议 2 核 CPU、8 GB 内存或以上，再小的话问题也不大，但是能调度的 Pod 数量就比较有限了；
7. 30 GB 或以上的可用磁盘空间，这主要是留给 Docker 镜像和日志文件用的。


## # pod
Pod 就是 Kubernetes 世界里的“应用”；而一个应用，可以由多个容器组成。

Pod，是 Kubernetes 项目中最小的 API 对象。如果换一个更专业的说法，我们可以这样描述：Pod，是 Kubernetes 项目的原子调度单位。

为什么我们会需要 Pod？

是啊，我们在前面已经花了很多精力去解读 Linux 容器的原理、分析了 Docker 容器的本质，终于，“Namespace 做隔离，Cgroups 做限制，rootfs 做文件系统”这样的“三句箴言”可以朗朗上口了，为什么 Kubernetes 项目又突然搞出一个 Pod 来呢？

要回答这个问题，我们需要明白：容器的本质到底是什么？容器的本质是进程。没错。容器，就是未来云计算系统中的进程；容器镜像就是这个系统里的“.exe”安装包。那么 Kubernetes 呢？Kubernetes 就是操作系统！

现在，就让我们登录到一台 Linux 机器里，执行一条如下所示的命令：```pstree -g```，这条命令的作用，是展示当前系统中正在运行的进程的树状结构。

不难发现，在一个真正的操作系统里，进程并不是“孤苦伶仃”地独自运行的，而是以进程组的方式，“有原则地”组织在一起。比如，这里有一个叫作 rsyslogd 的程序，它负责的是 Linux 操作系统里的日志处理。可以看到，rsyslogd 的主程序 main，和它要用到的内核日志模块 imklog 等，同属于一个进程组。这些进程相互协作，共同完成 rsyslogd 程序的职责。（这里的进程严格意义上来说，其实是 Linux 操作系统语境下的“线程”）

而 Kubernetes 项目所做的，其实就是将“进程组”的概念映射到了容器技术中，并使其成为了这个云计算“操作系统”里的“一等公民”

Kubernetes 项目之所以要这么做的原因，是 google 开发者在 Borg 项目的开发和实践过程中，工程师们发现，他们部署的应用，往往都存在着类似于“进程和进程组”的关系。更具体地说，就是这些应用之间有着密切的协作关系，使得它们必须部署在同一台机器上。

而如果事先没有“组”的概念，像这样的运维关系就会非常难以处理。已知 rsyslogd 由三个进程组成：一个 imklog 模块，一个 imuxsock 模块，一个 rsyslogd 自己的 main 函数主进程。这三个进程一定要运行在同一台机器上，否则，它们之间基于 Socket 的通信和文件交换，都会出现问题。

现在，要把 rsyslogd 这个应用给容器化，由于受限于容器的“单进程模型”，这三个模块必须被分别制作成三个不同的容器。而在这三个容器运行的时候，它们设置的内存配额都是 1 GB。（注意：**容器的“单进程模型”，并不是指容器里只能运行“一个”进程，而是指容器没有管理多个进程的能力。这是因为容器里 PID=1 的进程就是应用本身，其他的进程都是这个 PID=1 进程的子进程。可是，用户编写的应用，并不能够像正常操作系统里的 init 进程或者 systemd 那样拥有进程管理的功能。**）

这时，假设我要用 Docker Swarm 来运行这个 rsyslogd 程序。为了能够让这三个容器都运行在同一台机器上，我就必须在另外两个容器上设置一个 affinity=main（与 main 容器有亲密性）的约束，即：它们俩必须和 main 容器运行在同一台机器上。然后，我顺序执行：“docker run main”“docker run imklog”和“docker run imuxsock”，创建这三个容器。这样，这三个容器都会进入 Swarm 的待调度队列。然后，main 容器和 imklog 容器都先后出队并被调度到了 node-2 上（这个情况是完全有可能的）。可是，当 imuxsock 容器出队开始被调度时，Swarm 就有点懵了：node-2 上的可用资源只有 0.5 GB 了，并不足以运行 imuxsock 容器；可是，根据 affinity=main 的约束，imuxsock 容器又只能运行在 node-2 上。这就是一个典型的成组调度（gang scheduling）没有被妥善处理的例子。k8s之前有很多处理方法，但是都不是很完美。

但是，到了 Kubernetes 项目里，这样的问题就迎刃而解了：Pod 是 Kubernetes 里的原子调度单位。这就意味着，Kubernetes 项目的调度器，是统一按照 Pod 而非容器的资源需求进行计算的。

所以，像 imklog、imuxsock 和 main 函数主进程这样的三个容器，正是一个典型的由三个容器组成的 Pod。Kubernetes 项目在调度时，自然就会去选择可用内存等于 3 GB 的 node-1 节点进行绑定，而根本不会考虑 node-2。

像这样容器间的紧密协作，我们可以称为“超亲密关系”。这些具有“超亲密关系”容器的典型特征包括但不限于：互相之间会发生直接的文件交换、使用 localhost 或者 Socket 文件进行本地通信、会发生非常频繁的远程调用、需要共享某些 Linux Namespace（比如，一个容器要加入另一个容器的 Network Namespace）等等。

这也就意味着，并不是所有有“关系”的容器都属于同一个 Pod。比如，PHP 应用容器和 MySQL 虽然会发生访问关系，但并没有必要、也不应该部署在同一台机器上，它们更适合做成两个 Pod。

更重要的，Pod 在 Kubernetes 项目里还有更重要的意义，那就是：容器设计模式。

#### Pod 的实现原理
首先，关于 Pod 最重要的一个事实是：它只是一个逻辑概念。

也就是说，Kubernetes 真正处理的，还是宿主机操作系统上 Linux 容器的 Namespace 和 Cgroups，而并不存在一个所谓的 Pod 的边界或者隔离环境。

那么，Pod 又是怎么被“创建”出来的呢？

答案是：Pod，其实是一组共享了某些资源的容器。

具体的说：Pod 里的所有容器，共享的是同一个 Network Namespace，并且可以声明共享同一个 Volume。

那这么来看的话，一个有 A、B 两个容器的 Pod，不就是等同于一个容器（容器 A）共享另外一个容器（容器 B）的网络和 Volume 的玩儿法么？

这好像通过 docker run --net --volumes-from 这样的命令就能实现嘛。

但是，你有没有考虑过，如果真这样做的话，容器 B 就必须比容器 A 先启动，这样一个 Pod 里的多个容器就不是对等关系，而是拓扑关系了。

所以，在 Kubernetes 项目里，Pod 的实现需要使用一个中间容器，这个容器叫作 Infra 容器。在这个 Pod 中，Infra 容器永远都是第一个被创建的容器，而其他用户定义的容器，则通过 Join Network Namespace 的方式，与 Infra 容器关联在一起。

**所以这个 Pod 里有两个用户容器 A 和 B，还有一个 Infra 容器。很容易理解，在 Kubernetes 项目里，Infra 容器一定要占用极少的资源，所以它使用的是一个非常特殊的镜像，叫作：k8s.gcr.io/pause。这个镜像是一个用汇编语言编写的、永远处于“暂停”状态的容器，解压后的大小也只有 100~200 KB 左右。**

而在 Infra 容器“Hold 住”Network Namespace 后，用户容器就可以加入到 Infra 容器的 Network Namespace 当中了。所以，如果你查看这些容器在宿主机上的 Namespace 文件，它们指向的值一定是完全一样的。

**这也就意味着，对于 Pod 里的容器 A 和容器 B 来说**：
- 它们可以直接使用 localhost 进行通信；
- 它们看到的网络设备跟 Infra 容器看到的完全一样；
- 一个 Pod 只有一个 IP 地址，也就是这个 Pod 的 Network Namespace 对应的 IP 地址；
- 当然，其他的所有网络资源，都是一个 Pod 一份，并且被该 Pod 中的所有容器共享；
- Pod 的生命周期只跟 Infra 容器一致，而与容器 A 和 B 无关。

而对于同一个 Pod 里面的所有用户容器来说，它们的进出流量，也可以认为都是通过 Infra 容器完成的。这一点很重要，因为将来如果你要为 Kubernetes 开发一个网络插件时，应该重点考虑的是如何配置这个 Pod 的 Network Namespace，而不是每一个用户容器如何使用你的网络配置，这是没有意义的。

这就意味着，如果你的网络插件需要在容器里安装某些包或者配置才能完成的话，是不可取的：Infra 容器镜像的 rootfs 里几乎什么都没有，没有你随意发挥的空间。当然，这同时也意味着你的网络插件完全不必关心用户容器的启动与否，而只需要关注如何配置 Pod，也就是 Infra 容器的 Network Namespace 即可。

有了这个设计之后，共享 Volume 就简单多了：Kubernetes 项目只要把所有 Volume 的定义都设计在 Pod 层级即可。

这样，一个 Volume 对应的宿主机目录对于 Pod 来说就只有一个，Pod 里的容器只要声明挂载这个 Volume，就一定可以共享这个 Volume 对应的宿主机目录。比如下面这个例子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: two-containers
spec:
  restartPolicy: Never
  volumes:
  - name: shared-data
    hostPath:      
      path: /data
  containers:
  - name: nginx-container
    image: nginx
    volumeMounts:
    - name: shared-data
      mountPath: /usr/share/nginx/html
  - name: debian-container
    image: debian
    volumeMounts:
    - name: shared-data
      mountPath: /pod-data
    command: ["/bin/sh"]
    args: ["-c", "echo Hello from the debian container > /pod-data/index.html"]
```
在这个例子中，debian-container 和 nginx-container 都声明挂载了 shared-data 这个 Volume。而 shared-data 是 hostPath 类型。所以，它对应在宿主机上的目录就是：/data。而这个目录，其实就被同时绑定挂载进了上述两个容器当中。

这就是为什么，nginx-container 可以从它的 /usr/share/nginx/html 目录中，读取到 debian-container 生成的 index.html 文件的原因。

**在 Pod 中，所有 Init Container 定义的容器，都会比 spec.containers 定义的用户容器先启动。并且，Init Container 容器会按顺序逐一启动，而直到它们都启动并且退出了，用户容器才会启动。**

#### sidecar
像这样，我们就用一种“组合”方式，解决容器之间耦合关系的问题。实际上，这个所谓的“组合”操作，正是容器设计模式里最常用的一种模式，它的名字叫：sidecar。顾名思义，sidecar 指的就是我们可以在一个 Pod 中，启动一个辅助容器，来完成一些独立于主进程（主容器）之外的工作。

举个例子：

我现在有一个应用，需要不断地把日志文件输出到容器的 /var/log 目录中。这时，我就可以把一个 Pod 里的 Volume 挂载到应用容器的 /var/log 目录上。

然后，我在这个 Pod 里同时运行一个 sidecar 容器，它也声明挂载同一个 Volume 到自己的 /var/log 目录上。

这样，接下来 sidecar 容器就只需要做一件事，那就是不断地从自己的 /var/log 目录里读取日志文件，转发到 MongoDB 或者 Elasticsearch 中存储起来。这样，一个最基本的日志收集工作就完成了。

这个例子中的 sidecar 的主要工作也是使用共享的 Volume 来完成对文件的操作。

但不要忘记，Pod 的另一个重要特性是，它的所有容器都共享同一个 Network Namespace。这就使得很多与 Pod 网络相关的配置和管理，也都可以交给 sidecar 完成，而完全无须干涉用户容器。这里最典型的例子莫过于 Istio 这个微服务治理项目了。Istio 项目使用 sidecar 容器完成微服务治理的原理


## # pod基本概念
现在，你已经非常清楚：Pod，而不是容器，才是 Kubernetes 项目中的最小编排单位。将这个设计落实到 API 对象上，容器（Container）就成了 Pod 属性里的一个普通的字段。那么，一个很自然的问题就是：到底哪些属性属于 Pod 对象，而又有哪些属性属于 Container 呢？

要彻底理解这个问题，你就一定要牢记一个结论：Pod 扮演的是传统部署环境里“虚拟机”的角色。这样的设计，是为了使用户从传统环境（虚拟机环境）向 Kubernetes（容器环境）的迁移，更加平滑。

而如果你能把 Pod 看成传统环境里的“机器”、把容器看作是运行在这个“机器”里的“用户程序”，那么很多关于 Pod 对象的设计就非常容易理解了。

**比如，凡是调度、网络、存储，以及安全相关的属性，基本上是 Pod 级别的。**

这些属性的共同特征是，它们描述的是“机器”这个整体，而不是里面运行的“程序”。比如，配置这个“机器”的网卡（即：Pod 的网络定义），配置这个“机器”的磁盘（即：Pod 的存储定义），配置这个“机器”的防火墙（即：Pod 的安全定义）。更不用说，这台“机器”运行在哪个服务器之上（即：Pod 的调度）。

#### Pod 中几个重要字段的含义和用法

1. NodeSelector：是一个供用户将 Pod 与 Node 进行绑定的字段，用法如下所示：
```yaml
apiVersion: v1
kind: Pod
...
spec:
 nodeSelector:
   disktype: ssd
```
这样的一个配置，意味着这个 Pod 永远只能运行在携带了“disktype: ssd”标签（Label）的节点上；否则，它将调度失败。

2. NodeName
一旦 Pod 的这个字段被赋值，Kubernetes 项目就会被认为这个 Pod 已经经过了调度，调度的结果就是赋值的节点名字。所以，这个字段一般由调度器负责设置，但用户也可以设置它来“骗过”调度器，当然这个做法一般是在测试或者调试的时候才会用到。


3. HostAliases 定义了 Pod 的 hosts 文件（比如 /etc/hosts）里的内容，用法如下：
```yaml
apiVersion: v1
kind: Pod
# ...
spec:
  hostAliases:
  - ip: "10.1.2.3"
    hostnames:
    - "foo.remote"
    - "bar.remote"
# ...
```
在这个 Pod 的 YAML 文件中，我设置了一组 IP 和 hostname 的数据。这样，这个 Pod 启动后，/etc/hosts 文件的内容将如下所示：
```sh
cat /etc/hosts
# Kubernetes-managed hosts file.
127.0.0.1 localhost
# ...
10.244.135.10 hostaliases-pod
10.1.2.3 foo.remote
10.1.2.3 bar.remote
```
其中，最下面两行记录，就是我通过 HostAliases 字段为 Pod 设置的。**需要指出的是，在 Kubernetes 项目中，如果要设置 hosts 文件里的内容，一定要通过这种方法。否则，如果直接修改了 hosts 文件的话，在 Pod 被删除重建之后，kubelet 会自动覆盖掉被修改的内容。**

4. 除了上述跟“机器”相关的配置外，你可能也会发现，凡是跟容器的 Linux Namespace 相关的属性，也一定是 Pod 级别的。这个原因也很容易理解：Pod 的设计，就是要让它里面的容器尽可能多地共享 Linux Namespace，仅保留必要的隔离和限制能力。这样，Pod 模拟出的效果，就跟虚拟机里程序间的关系非常类似了。

**tty 和 stdin。在 Pod 的 YAML 文件里声明开启它们俩，其实等同于设置了 docker run 里的 -it（-i 即 stdin，-t 即 tty）参数。**

如果你还是不太理解它们俩的作用的话，可以直接认为 tty 就是 Linux 给用户提供的一个常驻小程序，用于接收用户的标准输入，返回操作系统的标准输出。当然，为了能够在 tty 中输入信息，你还需要同时开启 stdin（标准输入流）。

5. 类似地，凡是 Pod 中的容器要共享宿主机的 Namespace，也一定是 Pod 级别的定义，比如：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  hostNetwork: true
  hostIPC: true
  hostPID: true
  containers:
  - name: nginx
    image: nginx
  - name: shell
    image: busybox
    stdin: true
    tty: true
```
在这个 Pod 中，我定义了共享宿主机的 Network、IPC 和 PID Namespace。这就意味着，这个 Pod 里的所有容器，会直接使用宿主机的网络、直接与宿主机进行 IPC 通信、看到宿主机里正在运行的所有进程。

6. 当然，除了这些属性，Pod 里最重要的字段当属“Containers”、“Init Containers”。其实，这两个字段都属于 Pod 对容器的定义，内容也完全相同，只是 Init Containers 的生命周期，会先于所有的 Containers，并且严格按照定义的顺序执行。

Kubernetes 项目中对 Container 的定义，和 Docker 相比并没有什么太大区别。前面的容器技术中的 Image（镜像）、Command（启动命令）、workingDir（容器的工作目录）、Ports（容器要开发的端口），以及 volumeMounts（容器要挂载的 Volume）都是构成 Kubernetes 项目中 Container 的主要字段。不过在这里，还有这么几个属性值得你额外关注。

- 6.1 首先，是 ImagePullPolicy 字段。它定义了镜像拉取的策略。而它之所以是一个 Container 级别的属性，是因为容器镜像本来就是 Container 定义中的一部分。

ImagePullPolicy 的值默认是 Always，即每次创建 Pod 都重新拉取一次镜像。另外，当容器的镜像是类似于 nginx 或者 nginx:latest 这样的名字时，ImagePullPolicy 也会被认为 Always。

而如果它的值被定义为 Never 或者 IfNotPresent，则意味着 Pod 永远不会主动拉取这个镜像，或者只在宿主机上不存在这个镜像时才拉取。


- 6.2 其次，是 Lifecycle 字段。它定义的是 Container Lifecycle Hooks。顾名思义，Container Lifecycle Hooks 的作用，是在容器状态发生变化时触发一系列“钩子”。我们来看这样一个例子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: lifecycle-demo
spec:
  containers:
  - name: lifecycle-demo-container
    image: nginx
    lifecycle:
      postStart:
        exec:
          command: ["/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"]
      preStop:
        exec:
          command: ["/usr/sbin/nginx","-s","quit"]
```

这是一个来自 Kubernetes 官方文档的 Pod 的 YAML 文件。它其实非常简单，只是定义了一个 nginx 镜像的容器。不过，在这个 YAML 文件的容器（Containers）部分，你会看到这个容器分别设置了一个 postStart 和 preStop 参数。这是什么意思呢？

先说 postStart 吧。它指的是，在容器启动后，立刻执行一个指定的操作。需要明确的是，postStart 定义的操作，虽然是在 Docker 容器 ENTRYPOINT 执行之后，但它并不严格保证顺序。也就是说，在 postStart 启动时，ENTRYPOINT 有可能还没有结束。

当然，如果 postStart 执行超时或者错误，Kubernetes 会在该 Pod 的 Events 中报出该容器启动失败的错误信息，导致 Pod 也处于失败的状态。


而类似地，preStop 发生的时机，则是容器被杀死之前（比如，收到了 SIGKILL 信号）。而需要明确的是，preStop 操作的执行，是同步的。所以，它会阻塞当前的容器杀死流程，直到这个 Hook 定义操作完成之后，才允许容器被杀死，这跟 postStart 不一样。

所以，在这个例子中，我们在容器成功启动之后，在 /usr/share/message 里写入了一句“欢迎信息”（即 postStart 定义的操作）。而在这个容器被删除之前，我们则先调用了 nginx 的退出指令（即 preStop 定义的操作），从而实现了容器的“优雅退出”。

7. 在熟悉了 Pod 以及它的 Container 部分的主要字段之后，再分享一下这样一个的 Pod 对象在 Kubernetes 中的生命周期。
Pod 生命周期的变化，主要体现在 Pod API 对象的 Status 部分，这是它除了 Metadata 和 Spec 之外的第三个重要字段。其中，pod.status.phase，就是 Pod 的当前状态，它有如下几种可能的情况：
   - Pending。这个状态意味着，Pod 的 YAML 文件已经提交给了 Kubernetes，API 对象已经被创建并保存在 Etcd 当中。但是，这个 Pod 里有些容器因为某种原因而不能被顺利创建。比如，调度不成功。
   - Running。这个状态下，Pod 已经调度成功，跟一个具体的节点绑定。它包含的容器都已经创建成功，并且至少有一个正在运行中。
   - Succeeded。这个状态意味着，Pod 里的所有容器都正常运行完毕，并且已经退出了。这种情况在运行一次性任务时最为常见。
   - Failed。这个状态下，Pod 里至少有一个容器以不正常的状态（非 0 的返回码）退出。这个状态的出现，意味着你得想办法 Debug 这个容器的应用，比如查看 Pod 的 Events 和日志。
   - Unknown。这是一个异常状态，意味着 Pod 的状态不能持续地被 kubelet 汇报给 kube-apiserver，这很有可能是主从节点（Master 和 Kubelet）间的通信出现了问题。

更进一步地，Pod 对象的 Status 字段，还可以再细分出一组 Conditions。这些细分状态的值包括：PodScheduled、Ready、Initialized，以及 Unschedulable。它们主要用于描述造成当前 Status 的具体原因是什么。

比如，Pod 当前的 Status 是 Pending，对应的 Condition 是 Unschedulable，这就意味着它的调度出现了问题。

而其中，Ready 这个细分状态非常值得我们关注：它意味着 Pod 不仅已经正常启动（Running 状态），而且已经可以对外提供服务了。这两者之间（Running 和 Ready）是有区别的，你不妨仔细思考一下。

Pod 的这些状态信息，是我们判断应用运行情况的重要标准，尤其是 Pod 进入了非“Running”状态后，你一定要能迅速做出反应，根据它所代表的异常情况开始跟踪和定位，而不是去手忙脚乱地查阅文档。


## # pod 使用进阶
作为 Kubernetes 项目里最核心的编排对象，Pod 携带的信息非常丰富。其中，资源定义（比如 CPU、内存等），以及调度相关的字段。在这里，就先从一种特殊的 Volume 开始，来深入地理解 Pod 对象各个重要字段的含义。

这种特殊的 Volume，叫作 Projected Volume，可以把它翻译为“投射数据卷”。（备注：Projected Volume 是 Kubernetes v1.11 之后的新特性）

在 Kubernetes 中，有几种特殊的 Volume，它们存在的意义不是为了存放容器里的数据，也不是用来进行容器和宿主机之间的数据交换。这些特殊 Volume 的作用，是为容器提供预先定义好的数据。所以，从容器的角度来看，这些 Volume 里的信息就是仿佛是被 Kubernetes“投射”（Project）进入容器当中的。这正是 Projected Volume 的含义。

到目前为止，Kubernetes 支持的 Projected Volume 一共有五种：
1. Secret；
2. ConfigMap；
3. Downward API；
4. ServiceAccountToken。
5. clusterTrustBundle

#### Secret
它的作用，是帮你把 Pod 想要访问的加密数据，存放到 Etcd 中。然后，你就可以通过在 Pod 的容器里挂载 Volume 的方式，访问到这些 Secret 里保存的信息了。

Secret 最典型的使用场景，莫过于存放数据库的 Credential 信息，比如下面这个例子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-projected-volume 
spec:
  containers:
  - name: test-secret-volume
    image: busybox
    args:
    - sleep
    - "86400"
    volumeMounts:
    - name: mysql-cred
      mountPath: "/projected-volume"
      readOnly: true
  volumes:
  - name: mysql-cred
    projected:
      sources:
      - secret:
          name: user
      - secret:
          name: pass
```
在这个 Pod 中，我定义了一个简单的容器。它声明挂载的 Volume，并不是常见的 emptyDir 或者 hostPath 类型，而是 projected 类型。而这个 Volume 的数据来源（sources），则是名为 user 和 pass 的 Secret 对象，分别对应的是数据库的用户名和密码。

这里用到的数据库的用户名、密码，正是以 Secret 对象的方式交给 Kubernetes 保存的。完成这个操作的指令，如下所示：
```sh
$ cat ./username.txt
admin
$ cat ./password.txt
c1oudc0w!

$ kubectl create secret generic user --from-file=./username.txt
$ kubectl create secret generic pass --from-file=./password.txt
```
其中，username.txt 和 password.txt 文件里，存放的就是用户名和密码；而 user 和 pass，则是我为 Secret 对象指定的名字。而我想要查看这些 Secret 对象的话，只要执行一条 kubectl get 命令就可以了：
```sh
$ kubectl get secrets
NAME           TYPE                                DATA      AGE
user          Opaque                                1         51s
pass          Opaque                                1         51s
```

当然，除了使用 kubectl create secret 指令外，我也可以直接通过编写 YAML 文件的方式来创建这个 Secret 对象，比如：
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  user: YWRtaW4=
  pass: MWYyZDFlMmU2N2Rm
```
可以看到，通过编写 YAML 文件创建出来的 Secret 对象只有一个。但它的 data 字段，却以 Key-Value 的格式保存了两份 Secret 数据。其中，“user”就是第一份数据的 Key，“pass”是第二份数据的 Key。

需要注意的是，Secret 对象要求这些数据必须是经过 Base64 转码的，以免出现明文密码的安全隐患。这个转码操作也很简单，比如：
```sh
$ echo -n 'admin' | base64
YWRtaW4=
$ echo -n '1f2d1e2e67df' | base64
MWYyZDFlMmU2N2Rm
```

这里需要注意的是，像这样创建的 Secret 对象，它里面的内容仅仅是经过了转码，而并没有被加密。在真正的生产环境中，你需要在 Kubernetes 中开启 Secret 的加密插件，增强数据的安全性。关于开启 Secret 加密插件的内容，参考下面 Secret 的相关的内容。

当 Pod 变成 Running 状态之后，我们再验证一下这些 Secret 对象是不是已经在容器里了

从返回结果中，我们可以看到，保存在 Etcd 里的用户名和密码信息，已经以文件的形式出现在了容器的 Volume 目录里。而这个文件的名字，就是 kubectl create secret 指定的 Key，或者说是 Secret 对象的 data 字段指定的 Key。

**更重要的是，像这样通过挂载方式进入到容器里的 Secret，一旦其对应的 Etcd 里的数据被更新，这些 Volume 里的文件内容，同样也会被更新。其实，这是 kubelet 组件在定时维护这些 Volume。**

需要注意的是，这个更新可能会有一定的延时。所以在编写应用程序时，在发起数据库连接的代码处写好重试和超时的逻辑，绝对是个好习惯。

#### 与 Secret 类似的是 ConfigMap
它与 Secret 的区别在于，ConfigMap 保存的是不需要加密的、应用所需的配置信息。而 ConfigMap 的用法几乎与 Secret 完全相同：你可以使用 kubectl create configmap 从文件或者目录创建 ConfigMap，也可以直接编写 ConfigMap 对象的 YAML 文件。

**kubectl get -o yaml 这样的参数，会将指定的 Pod API 对象以 YAML 的方式展示出来。**

#### 接下来是 Downward API
它的作用是：让 Pod 里的容器能够直接获取到这个 Pod API 对象本身的信息。

举个栗子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-downwardapi-volume
  labels:
    zone: us-est-coast
    cluster: test-cluster1
    rack: rack-22
spec:
  containers:
    - name: client-container
      image: k8s.gcr.io/busybox
      command: ["sh", "-c"]
      args:
      - while true; do
          if [[ -e /etc/podinfo/labels ]]; then
            echo -en '\n\n'; cat /etc/podinfo/labels; fi;
          sleep 5;
        done;
      volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
          readOnly: false
  volumes:
    - name: podinfo
      projected:
        sources:
        - downwardAPI:
            items:
              - path: "labels"
                fieldRef:
                  fieldPath: metadata.labels
```

在这个 Pod 的 YAML 文件中，我定义了一个简单的容器，声明了一个 projected 类型的 Volume。只不过这次 Volume 的数据来源，变成了 Downward API。而这个 Downward API Volume，则声明了要暴露 Pod 的 metadata.labels 信息给容器。

通过这样的声明方式，当前 Pod 的 Labels 字段的值，就会被 Kubernetes 自动挂载成为容器里的 /etc/podinfo/labels 文件。

目前，Downward API 支持的字段已经非常丰富了，比如：
```md
1. 使用fieldRef可以声明使用:
spec.nodeName - 宿主机名字
status.hostIP - 宿主机IP
metadata.name - Pod的名字
metadata.namespace - Pod的Namespace（这个目前好像是默认的）
status.podIP - Pod的IP
spec.serviceAccountName - Pod的Service Account的名字
metadata.uid - Pod的UID
metadata.labels['<KEY>'] - 指定<KEY>的Label值
metadata.annotations['<KEY>'] - 指定<KEY>的Annotation值
metadata.labels - Pod的所有Label
metadata.annotations - Pod的所有Annotation

2. 使用resourceFieldRef可以声明使用:
容器的CPU limit
容器的CPU request
容器的memory limit
容器的memory request
```
上面这个列表的内容，随着 Kubernetes 项目的发展肯定还会不断增加。所以这里列出来的信息仅供参考，在使用 Downward API 时，还是要记得去查阅一下官方文档。

**不过，需要注意的是，Downward API 能够获取到的信息，一定是 Pod 里的容器进程启动之前就能够确定下来的信息。而如果你想要获取 Pod 容器运行后才会出现的信息，比如，容器进程的 PID，那就肯定不能使用 Downward API 了，而应该考虑在 Pod 里定义一个 sidecar 容器。**

其实，Secret、ConfigMap，以及 Downward API 这三种 Projected Volume 定义的信息，大多还可以通过环境变量的方式出现在容器里。但是，通过环境变量获取这些信息的方式，不具备自动更新的能力。所以，一般情况下，都建议你使用 Volume 文件的方式获取这些信息。

#### Service Account
在明白了 Secret 之后，我再为你讲解 Pod 中一个与它密切相关的概念：Service Account。

相信你一定有过这样的想法：我现在有了一个 Pod，我能不能在这个 Pod 里安装一个 Kubernetes 的 Client，这样就可以从容器里直接访问并且操作这个 Kubernetes 的 API 了呢？

这当然是可以的。（这就是 space 服务器的使用场景，需要动态创建pod）

不过，你首先要解决 API Server 的授权问题。

Service Account 对象的作用，就是 Kubernetes 系统内置的一种“服务账户”，它是 Kubernetes 进行权限分配的对象。比如，Service Account A，可以只被允许对 Kubernetes API 进行 GET 操作，而 Service Account B，则可以有 Kubernetes API 的所有操作权限。

**像这样的 Service Account 的授权信息和文件，实际上保存在它所绑定的一个特殊的 Secret 对象里的。这个特殊的 Secret 对象，就叫作 ServiceAccountToken。任何运行在 Kubernetes 集群上的应用，都必须使用这个 ServiceAccountToken 里保存的授权信息，也就是 Token，才可以合法地访问 API Server。**

所以说，Kubernetes 项目的 Projected Volume 其实只有三种，因为第四种 ServiceAccountToken，只是一种特殊的 Secret 而已。

另外，为了方便使用，Kubernetes 已经为你提供了一个默认“服务账户”（default Service Account）。并且，任何一个运行在 Kubernetes 里的 Pod，都可以直接使用这个默认的 Service Account，而无需显示地声明挂载它。

这是如何做到的呢？

当然还是靠 Projected Volume 机制。

如果你查看一下任意一个运行在 Kubernetes 集群里的 Pod，就会发现，每一个 Pod，都已经自动声明一个类型是 Secret、名为 default-token-xxxx 的 Volume，然后 自动挂载在每个容器的一个固定目录上。（自Kubernetes v1.22 stable 版本之后，ServiceAccount 准入控制器将添加投射卷， 而不是为令牌控制器所生成的不过期的服务账号令牌而创建的基于 Secret 的卷。官方文档 https://kubernetes.io/zh-cn/docs/concepts/storage/projected-volumes/#serviceaccounttoken）

Kubernetes 其实在每个 Pod 创建的时候，自动在它的 spec.volumes 部分添加上了默认 ServiceAccountToken 的定义，然后自动给每个容器加上了对应的 volumeMounts 字段。这个过程对于用户来说是完全透明的。

这样，一旦 Pod 创建完成，容器里的应用就可以直接从这个默认 ServiceAccountToken 的挂载目录里访问到授权信息和文件。这个容器内的路径在 Kubernetes 里是固定的，即：/var/run/secrets/kubernetes.io/serviceaccount ，而这个 Secret 类型的 Volume 里面的内容如下所示：
```sh
$ ls /var/run/secrets/kubernetes.io/serviceaccount 
ca.crt namespace  token
```
所以，你的应用程序只要直接加载这些授权文件，就可以访问并操作 Kubernetes API 了。而且，如果你使用的是 Kubernetes 官方的 Client 包（k8s.io/client-go）的话，它还可以自动加载这个目录下的文件，你不需要做任何配置或者编码操作。

**这种把 Kubernetes 客户端以容器的方式运行在集群里，然后使用 default Service Account 自动授权的方式，被称作“InClusterConfig”，也是我最推荐的进行 Kubernetes API 编程的授权方式。**（open api的机制，这样可以定制化client操作，提高了操作k8s集群的灵活性）

当然，考虑到自动挂载默认 ServiceAccountToken 的潜在风险，Kubernetes 允许你设置默认不为 Pod 里的容器自动挂载这个 Volume。

除了这个默认的 Service Account 外，我们很多时候还需要创建一些我们自己定义的 Service Account，来对应不同的权限设置。这样，我们的 Pod 里的容器就可以通过挂载这些 Service Account 对应的 ServiceAccountToken，来使用这些自定义的授权信息。参考下面 Kubernetes 开发插件相关内容。

#### 容器健康检查和恢复机制配置
再来看 Pod 的另一个重要的配置：容器健康检查和恢复机制。

在 Kubernetes 中，你可以为 Pod 里的容器定义一个健康检查“探针”（Probe）。这样，kubelet 就会根据这个 Probe 的返回值决定这个容器的状态，而不是直接以容器镜像是否运行（来自 Docker 返回的信息）作为依据。这种机制，是生产环境中保证应用健康存活的重要手段。

来看一个 Kubernetes 文档中的例子。
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    test: liveness
  name: test-liveness-exec
spec:
  containers:
  - name: liveness
    image: busybox
    args:
    - /bin/sh
    - -c
    - touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600
    livenessProbe:
      exec:
        command:
        - cat
        - /tmp/healthy
      initialDelaySeconds: 5
      periodSeconds: 5
```

在这个 Pod 中，我们定义了一个有趣的容器。它在启动之后做的第一件事，就是在 /tmp 目录下创建了一个 healthy 文件，以此作为自己已经正常运行的标志。而 30 s 过后，它会把这个文件删除掉。

与此同时，**我们定义了一个这样的 livenessProbe（健康检查）。它的类型是 exec，这意味着，它会在容器启动后，在容器里面执行一条我们指定的命令**，比如：“cat /tmp/healthy”。这时，如果这个文件存在，这条命令的返回值就是 0，Pod 就会认为这个容器不仅已经启动，而且是健康的。这个健康检查，在容器启动 5 s 后开始执行（initialDelaySeconds: 5），每 5 s 执行一次（periodSeconds: 5）。

现在，让我们来具体实践一下这个过程。

首先，创建这个 Pod，然后，查看这个 Pod 的状态，可以看到，由于已经通过了健康检查，这个 Pod 就进入了 Running 状态。而 30 s 之后，我们再查看一下 Pod 的 Events。你会发现，这个 Pod 在 Events 报告了一个异常。显然，这个健康检查探查到 /tmp/healthy 已经不存在了，所以它报告容器是不健康的。那么接下来会发生什么呢？

我们不妨再次查看一下这个 Pod 的状态，这时我们发现，Pod 并没有进入 Failed 状态，而是保持了 Running 状态。这是为什么呢？

其实，如果你注意到 RESTARTS 字段从 0 到 1 的变化，就明白原因了：这个异常的容器已经被 Kubernetes 重启了。在这个过程中，Pod 保持 Running 状态不变。

**需要注意的是：Kubernetes 中并没有 Docker 的 Stop 语义。所以虽然是 Restart（重启），但实际却是重新创建了容器。**

**这个功能就是 Kubernetes 里的 Pod 恢复机制，也叫 restartPolicy。它是 Pod 的 Spec 部分的一个标准字段（pod.spec.restartPolicy），默认值是 Always，即：任何时候这个容器发生了异常，它一定会被重新创建。**

**但一定要强调的是，Pod 的恢复过程，永远都是发生在当前节点上，而不会跑到别的节点上去。事实上，一旦一个 Pod 与一个节点（Node）绑定，除非这个绑定发生了变化（pod.spec.node 字段被修改），否则它永远都不会离开这个节点。这也就意味着，如果这个宿主机宕机了，这个 Pod 也不会主动迁移到其他节点上去。**

**而如果你想让 Pod 出现在其他的可用节点上，就必须使用 Deployment 这样的“控制器”来管理 Pod，哪怕你只需要一个 Pod 副本。这就是一个单 Pod 的 Deployment 与一个 Pod 最主要的区别。**


而作为用户，你还可以通过设置 restartPolicy，改变 Pod 的恢复策略。除了 Always，它还有 OnFailure 和 Never 两种情况：
1. Always：在任何情况下，只要容器不在运行状态，就自动重启容器；
2. OnFailure: 只在容器 异常时才自动重启容器；
3. Never: 从来不重启容器。

在实际使用时，我们需要根据应用运行的特性，合理设置这三种恢复策略。

比如，一个 Pod，它只计算 1+1=2，计算完成输出结果后退出，变成 Succeeded 状态。这时，你如果再用 restartPolicy=Always 强制重启这个 Pod 的容器，就没有任何意义了。

而如果你要关心这个容器退出后的上下文环境，比如容器退出后的日志、文件和目录，就需要将 restartPolicy 设置为 Never。因为一旦容器被自动重新创建，这些内容就有可能丢失掉了（被垃圾回收了）。

值得一提的是，Kubernetes 的官方文档，把 restartPolicy 和 Pod 里容器的状态，以及 Pod 状态的对应关系，总结了非常复杂的一大堆情况。实际上，你根本不需要死记硬背这些对应关系，只要记住如下两个基本的设计原理即可：
1. 只要 Pod 的 restartPolicy 指定的策略允许重启异常的容器（比如：Always），那么这个 Pod 就会保持 Running 状态，并进行容器重启。否则，Pod 就会进入 Failed 状态 。
2. 对于包含多个容器的 Pod，只有它里面所有的容器都进入异常状态后，Pod 才会进入 Failed 状态。在此之前，Pod 都是 Running 状态。此时，Pod 的 READY 字段会显示正常容器的个数，比如：
```sh
$ kubectl get pod test-liveness-exec
NAME           READY     STATUS    RESTARTS   AGE
liveness-exec   0/1       Running   1          1m
```
所以，假如一个 Pod 里只有一个容器，然后这个容器异常退出了。那么，只有当 restartPolicy=Never 时，这个 Pod 才会进入 Failed 状态。而其他情况下，由于 Kubernetes 都可以重启这个容器，所以 Pod 的状态保持 Running 不变。

而如果这个 Pod 有多个容器，仅有一个容器异常退出，它就始终保持 Running 状态，哪怕即使 restartPolicy=Never。只有当所有容器也异常退出之后，这个 Pod 才会进入 Failed 状态。

回到前面提到的 livenessProbe 上来。除了在容器中执行命令外，livenessProbe 也可以定义为发起 HTTP 或者 TCP 请求的方式，定义格式如下：
```yaml
# ...
livenessProbe:
    httpGet:
       path: /healthz
       port: 8080
       httpHeaders:
       - name: X-Custom-Header
         value: Awesome
       initialDelaySeconds: 3
       periodSeconds: 3
```
```yaml
    ...
livenessProbe:
    tcpSocket:
      port: 8080
    initialDelaySeconds: 15
    periodSeconds: 20
```

所以，你的 Pod 其实可以暴露一个健康检查 URL（比如 /healthz），或者直接让健康检查去检测应用的监听端口。这两种配置方法，在 Web 服务类的应用中非常常用。

**在 Kubernetes 的 Pod 中，还有一个叫 readinessProbe 的字段。虽然它的用法与 livenessProbe 类似，但作用却大不一样。readinessProbe 检查结果的成功与否，决定的这个 Pod 是不是能被通过 Service 的方式访问到，而并不影响 Pod 的生命周期。这部分内容，具体参考下面 Service 部分。**

Pod 的字段这么多，我又不可能全记住，Kubernetes 能不能自动给 Pod 填充某些字段呢？

这个需求实际上非常实用。比如，开发人员只需要提交一个基本的、非常简单的 Pod YAML，Kubernetes 就可以自动给对应的 Pod 对象加上其他必要的信息，比如 labels，annotations，volumes 等等。而这些信息，可以是运维人员事先定义好的。这就是 PodPreset （v1.11 版本开始，1.20后已经去掉这个功能了）


#### clusterTrustBundle 投射卷
要在 Kubernetes 1.32 中使用此特性，你必须通过 ClusterTrustBundle 特性门控和 --runtime-config=certificates.k8s.io/v1alpha1/clustertrustbundles=true kube-apiserver 标志启用对 ClusterTrustBundle 对象的支持，然后才能启用 ClusterTrustBundleProjection 特性门控。

clusterTrustBundle 投射卷源将一个或多个 ClusterTrustBundle 对象的内容作为一个自动更新的文件注入到容器文件系统中。


## # 编排其实很简单：谈谈“控制器”模型
实际上，你可能已经有所感悟：Pod 这个看似复杂的 API 对象，实际上就是对容器的进一步抽象和封装而已。

所以，Pod 对象，其实就是容器的升级版。它对容器进行了组合，添加了更多的属性和字段。这就好比给集装箱四面安装了吊环，使得 Kubernetes 这架“吊车”，可以更轻松地操作它。

而 Kubernetes 操作这些“集装箱”的逻辑，都由控制器（Controller）完成。在上面我们曾经使用过 Deployment 这个最基本的控制器对象：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
```
这个 Deployment 定义的编排动作非常简单，即：确保携带了 app=nginx 标签的 Pod 的个数，永远等于 spec.replicas 指定的个数，即 2 个。

这就意味着，如果在这个集群中，携带 app=nginx 标签的 Pod 的个数大于 2 的时候，就会有旧的 Pod 被删除；反之，就会有新的 Pod 被创建。

究竟是 Kubernetes 项目中的哪个组件，在执行这些操作呢？

在前面介绍 Kubernetes 架构的时候，曾经提到过一个叫作 kube-controller-manager 的组件。

实际上，这个组件，就是一系列控制器的集合。我们可以查看一下 Kubernetes 项目的 pkg/controller 目录：(https://github.com/kubernetes/kubernetes/tree/master/pkg/controller)
```sh
$ cd kubernetes/pkg/controller/
$ ls -d */              
deployment/             job/                    podautoscaler/          
cloud/                  disruption/             namespace/              
replicaset/             serviceaccount/         volume/
cronjob/                garbagecollector/       nodelifecycle/          replication/            statefulset/            daemon/
# ...
```

这个目录下面的每一个控制器，都以独有的方式负责某种编排功能。而我们的 Deployment，正是这些控制器中的一种。

实际上，这些控制器之所以被统一放在 pkg/controller 目录下，就是因为它们都遵循 Kubernetes 项目中的一个通用编排模式，即：控制循环（control loop）。（循环对比实际状态和期望状态，然后执行某种动作以达到期望状态）

比如，现在有一种待编排的对象 X，它有一个对应的控制器。那么，我就可以用一段 Go 语言风格的伪代码，为你描述这个控制循环：
```go
for {
  实际状态 := 获取集群中对象X的实际状态（Actual State）
  期望状态 := 获取集群中对象X的期望状态（Desired State）
  if 实际状态 == 期望状态{
    什么都不做
  } else {
    执行编排动作，将实际状态调整为期望状态
  }
}
```

在具体实现中，实际状态往往来自于 Kubernetes 集群本身。

比如，kubelet 通过心跳汇报的容器状态和节点状态，或者监控系统中保存的应用监控数据，或者控制器主动收集的它自己感兴趣的信息，这些都是常见的实际状态的来源。

而期望状态，一般来自于用户提交的 YAML 文件。

比如，Deployment 对象中 Replicas 字段的值。很明显，这些信息往往都保存在 Etcd 中。

接下来，以 Deployment 为例，简单描述一下它对控制器模型的实现：
1. Deployment 控制器从 Etcd 中获取到所有携带了“app: nginx”标签的 Pod，然后统计它们的数量，这就是实际状态；
2. Deployment 对象的 Replicas 字段的值就是期望状态；
3. Deployment 控制器将两个状态做比较，然后根据比较结果，确定是创建 Pod，还是删除已有的 Pod（具体如何操作 Pod 对象，参考下面作业副本与水平扩展部分）。

可以看到，一个 Kubernetes 对象的主要编排逻辑，实际上是在第三步的“对比”阶段完成的。

这个操作，通常被叫作调谐（Reconcile）。这个调谐的过程，则被称作“Reconcile Loop”（调谐循环）或者“Sync Loop”（同步循环），所以，如果你以后在文档或者社区中碰到这些词，都不要担心，它们其实指的都是同一个东西：控制循环。

而调谐的最终结果，往往都是对被控制对象的某种写操作。

比如，增加 Pod，删除已有的 Pod，或者更新 Pod 的某个字段。这也是 Kubernetes 项目“面向 API 对象编程”的一个直观体现。

其实，像 Deployment 这种控制器的设计原理，就是我们前面提到过的，“用一种对象管理另一种对象”的“艺术”。

其中，这个控制器对象本身，负责定义被管理对象的期望状态。比如，Deployment 里的 replicas=2 这个字段。

而被控制对象的定义，则来自于一个“模板”。比如，Deployment 里的 template 字段。

可以看到，Deployment 这个 template 字段里的内容，跟一个标准的 Pod 对象的 API 定义，丝毫不差。而所有被这个 Deployment 管理的 Pod 实例，其实都是根据这个 template 字段的内容创建出来的。

像 Deployment 定义的 template 字段，在 Kubernetes 项目中有一个专有的名字，叫作 PodTemplate（Pod 模板）。

这个概念非常重要，因为后面我要讲解到的大多数控制器，都会使用 PodTemplate 来统一定义它所要管理的 Pod。更有意思的是，我们还会看到其他类型的对象模板，比如 Volume 的模板。

类似 Deployment 这样的一个控制器，实际上都是由上半部分的控制器定义（包括期望状态），加上 template 下半部分的被控制对象的模板组成的。

这就是为什么，在所有 API 对象的 Metadata 里，都有一个字段叫作 ownerReference，用于保存当前这个 API 对象的拥有者（Owner）的信息。

那么，对于我们这个 nginx-deployment 来说，它创建出来的 Pod 的 ownerReference 就是 nginx-deployment 吗？或者说，nginx-deployment 所直接控制的，就是 Pod 对象么？（不是，是replica set）


## # 作业副本与水平扩展
上面详细介绍了 Kubernetes 项目中第一个重要的设计思想：控制器模式。而现在就详细了解一下，Kubernetes 里第一个控制器模式的完整实现：Deployment。

Deployment 看似简单，但实际上，它实现了 Kubernetes 项目中一个非常重要的功能：Pod 的“水平扩展 / 收缩”（horizontal scaling out/in）。这个功能，是从 PaaS 时代开始，一个平台级项目就必须具备的编排能力。

举个例子，如果你更新了 Deployment 的 Pod 模板（比如，修改了容器的镜像），那么 Deployment 就需要遵循一种叫作“滚动更新”（rolling update）的方式，来升级现有的容器。

而这个能力的实现，依赖的是 Kubernetes 项目中的一个非常重要的概念（API 对象）：ReplicaSet。

ReplicaSet 的结构非常简单，我们可以通过这个 YAML 文件查看一下：
```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-set
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
```

从这个 YAML 文件中，我们可以看到，一个 ReplicaSet 对象，其实就是由副本数目的定义和一个 Pod 模板组成的。不难发现，它的定义其实是 Deployment 的一个子集。

**更重要的是，Deployment 控制器实际操纵的，正是这样的 ReplicaSet 对象，而不是 Pod 对象。**


```md
ReplicaSet是Kubernetes中的一种资源对象（也是控制器），用于定义和管理Pod副本的集合。它确保指定数量的Pod副本在任何时候都在运行，并在Pod生故障或被删除时自动进行替换。

ReplicaSet使用标签器来确定要管理的Pod副本。它可以根据指定的标签选择器创建、删除和调整Pod副本的数量，以保持期望的副本数量。

当创建一个ReplicaSet时，您需要指定以下几个重要的属性：

selector：用于选择要管理的Pod副本的标签选择器。
replicas：指定要创建的Pod副本的数量。
template：定义要创建的Pod副本的模板。
ReplicaSet的工作原理是通过持续监控集群中的Pod副本数量，并根据需要创建或删除副本来维持期望的状态。如果有副本数量少于期望数量，ReplicaSet将创建新的副本来替补；如果有副本数量多于期望数量，ReplicaSet将删除多余的副本。

请注意，ReplicaSet本身并不支持水平扩展或滚动更新，它只负责维护指定数量的Pod副本。如果您需要进行水平扩展或滚动更新，可以使用Deployment对象，它是在ReplicaSet之上的一个更高级别的抽象。
```

明白了这个原理，再来分析一个如下所示的 Deployment：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
```
可以看到，这就是一个我们常用的 nginx-deployment，它定义的 Pod 副本个数是 3（spec.replicas=3）。

那么，在具体的实现上，这个 Deployment，与 ReplicaSet，以及 Pod 的关系是怎样的呢？

一个定义了 replicas=3 的 Deployment，与它的 ReplicaSet，以及 Pod 的关系，实际上是一种“层层控制”的关系。

其中，ReplicaSet 负责通过“控制器模式”，保证系统中 Pod 的个数永远等于指定的个数（比如，3 个）。这也正是 Deployment 只允许容器的 restartPolicy=Always 的主要原因：只有在容器能保证自己始终是 Running 状态的前提下，ReplicaSet 调整 Pod 的个数才有意义。

而在此基础上，Deployment 同样通过“控制器模式”，来操作 ReplicaSet 的个数和属性，进而实现“水平扩展 / 收缩”和“滚动更新”这两个编排动作。

其中，“水平扩展 / 收缩”非常容易实现，Deployment Controller 只需要修改它所控制的 ReplicaSet 的 Pod 副本个数就可以了。

比如，把这个值从 3 改成 4，那么 Deployment 所对应的 ReplicaSet，就会根据修改后的值自动创建一个新的 Pod。这就是“水平扩展”了；“水平收缩”则反之。

而用户想要执行这个操作的指令也非常简单，就是 kubectl scale，比如：
```sh
$ kubectl scale deployment nginx-deployment --replicas=4
deployment.apps/nginx-deployment scaled
```

那么，“滚动更新”又是什么意思，是如何实现的呢？

接下来，以这个 Deployment 为例，来为你讲解“滚动更新”的过程。

首先，我们来创建这个 nginx-deployment：
```sh
$ kubectl create -f nginx-deployment.yaml --record
```
注意，在这里，我额外加了一个–record 参数。它的作用，是记录下你每次操作所执行的命令，以方便后面查看。

然后，我们来检查一下 nginx-deployment 创建后的状态信息：
```sh
$ kubectl get deployments
NAME               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3         0         0            0           1s
```

在返回结果中，我们可以看到四个状态字段，它们的含义如下:
1. DESIRED：用户期望的 Pod 副本个数（spec.replicas 的值）；
2. CURRENT：当前处于 Running 状态的 Pod 的个数；
3. UP-TO-DATE：当前处于最新版本的 Pod 的个数，所谓最新版本指的是 Pod 的 Spec 部分与 Deployment 里 Pod 模板里定义的完全一致；
4. AVAILABLE：当前已经可用的 Pod 的个数，即：既是 Running 状态，又是最新版本，并且已经处于 Ready（健康检查正确）状态的 Pod 的个数。

可以看到，只有这个 AVAILABLE 字段，描述的才是用户所期望的最终状态。

而 Kubernetes 项目还为我们提供了一条指令，让我们可以实时查看 Deployment 对象的状态变化。这个指令就是 kubectl rollout status：
```sh
$ kubectl rollout status deployment/nginx-deployment
Waiting for rollout to finish: 2 out of 3 new replicas have been updated...
deployment.apps/nginx-deployment successfully rolled out
```
在这个返回结果中，“2 out of 3 new replicas have been updated”意味着已经有 2 个 Pod 进入了 UP-TO-DATE 状态。

继续等待一会儿，我们就能看到这个 Deployment 的 3 个 Pod，就进入到了 AVAILABLE 状态：
```sh
NAME               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3         3         3            3           20s
```

此时，你可以尝试查看一下这个 Deployment 所控制的 ReplicaSet：
```sh
$ kubectl get rs
NAME                          DESIRED   CURRENT   READY   AGE
nginx-deployment-3167673210   3         3         3       20s
```
如上所示，在用户提交了一个 Deployment 对象后，Deployment Controller 就会立即创建一个 Pod 副本个数为 3 的 ReplicaSet。这个 ReplicaSet 的名字，则是由 Deployment 的名字和一个随机字符串共同组成。

这个随机字符串叫作 pod-template-hash，在我们这个例子里就是：3167673210。ReplicaSet 会把这个随机字符串加在它所控制的所有 Pod 的标签里，从而保证这些 Pod 不会与集群里的其他 Pod 混淆。

而 ReplicaSet 的 DESIRED、CURRENT 和 READY 字段的含义，和 Deployment 中是一致的。所以，相比之下，Deployment 只是在 ReplicaSet 的基础上，添加了 UP-TO-DATE 这个跟版本有关的状态字段。

这个时候，如果我们修改了 Deployment 的 Pod 模板，“滚动更新”就会被自动触发。

修改 Deployment 有很多方法。比如，我可以直接使用 kubectl edit 指令编辑 Etcd 里的 API 对象。
```sh
$ kubectl edit deployment/nginx-deployment
... 
    spec:
      containers:
      - name: nginx
        image: nginx:1.9.1 # 1.7.9 -> 1.9.1
        ports:
        - containerPort: 80
...
deployment.extensions/nginx-deployment edited
```

这个 kubectl edit 指令，会帮你直接打开 nginx-deployment 的 API 对象。然后，你就可以修改这里的 Pod 模板部分了。比如，在这里，我将 nginx 镜像的版本升级到了 1.9.1。（**kubectl edit 并不神秘，它不过是把 API 对象的内容下载到了本地文件，让你修改完成后再提交上去。**）

kubectl edit 指令编辑完成后，保存退出，Kubernetes 就会立刻触发“滚动更新”的过程。你还可以通过 kubectl rollout status 指令查看 nginx-deployment 的状态变化，这时，你可以通过查看 Deployment 的 Events，看到这个“滚动更新”的流程。

这种“滚动更新”的好处是显而易见的。

比如，在升级刚开始的时候，集群里只有 1 个新版本的 Pod。如果这时，新版本 Pod 有问题启动不起来，那么“滚动更新”就会停止，从而允许开发和运维人员介入。而在这个过程中，由于应用本身还有两个旧版本的 Pod 在线，所以服务并不会受到太大的影响。

当然，这也就要求你一定要使用 Pod 的 Health Check 机制检查应用的运行状态，而不是简单地依赖于容器的 Running 状态。要不然的话，虽然容器已经变成 Running 了，但服务很有可能尚未启动，“滚动更新”的效果也就达不到了。

而为了进一步保证服务的连续性，Deployment Controller 还会确保，在任何时间窗口内，只有指定比例的 Pod 处于离线状态。同时，它也会确保，在任何时间窗口内，只有指定比例的新 Pod 被创建出来。这两个比例的值都是可以配置的，默认都是 DESIRED 值的 25%。

所以，在上面这个 Deployment 的例子中，它有 3 个 Pod 副本，那么控制器在“滚动更新”的过程中永远都会确保至少有 2 个 Pod 处于可用状态，至多只有 4 个 Pod 同时存在于集群中。这个策略，是 Deployment 对象的一个字段，名叫 RollingUpdateStrategy，如下所示：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
...
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
```
在上面这个 RollingUpdateStrategy 的配置中，maxSurge 指定的是除了 DESIRED 数量之外，在一次“滚动”中，Deployment 控制器还可以创建多少个新 Pod；而 maxUnavailable 指的是，在一次“滚动”中，Deployment 控制器可以删除多少个旧 Pod。

同时，这两个配置还可以用前面我们介绍的百分比形式来表示，比如：maxUnavailable=50%，指的是我们最多可以一次删除“50%*DESIRED 数量”个 Pod。

Deployment 的控制器，实际上控制的是 ReplicaSet 的数目，以及每个 ReplicaSet 的属性。

而一个应用的版本，对应的正是一个 ReplicaSet；这个版本应用的 Pod 数量，则由 ReplicaSet 通过它自己的控制器（ReplicaSet Controller）来保证。

而明白了“应用版本和 ReplicaSet 一一对应”的设计思想之后，就可以为了解一下 Deployment 对应用进行版本控制的具体原理了。

这一次，我会使用一个叫 kubectl set image 的指令，直接修改 nginx-deployment 所使用的镜像。这个命令的好处就是，你可以不用像 kubectl edit 那样需要打开编辑器。

不过这一次，我把这个镜像名字修改成为了一个错误的名字，比如：nginx:1.91。这样，这个 Deployment 就会出现一个升级失败的版本。

由于这个 nginx:1.91 镜像在 Docker Hub 中并不存在，所以这个 Deployment 的“滚动更新”被触发后，会立刻报错并停止。

通过```kubectl get rs```这个返回结果，我们可以看到，新版本的 ReplicaSet（hash=2156724341）的“水平扩展”已经停止。而且此时，它已经创建了两个 Pod，但是它们都没有进入 READY 状态。这当然是因为这两个 Pod 都拉取不到有效的镜像。

与此同时，旧版本的 ReplicaSet（hash=1764197365）的“水平收缩”，也自动停止了。此时，已经有一个旧 Pod 被删除，还剩下两个旧 Pod。

**那么问题来了， 我们如何让这个 Deployment 的 3 个 Pod，都回滚到以前的旧版本呢？**

**我们只需要执行一条 kubectl rollout undo 命令，就能把整个 Deployment 回滚到上一个版本：**
```sh
$ kubectl rollout undo deployment/nginx-deployment
deployment.extensions/nginx-deployment
```
很容易想到，在具体操作上，Deployment 的控制器，其实就是让这个旧 ReplicaSet（hash=1764197365）再次“扩展”成 3 个 Pod，而让新的 ReplicaSet（hash=2156724341）重新“收缩”到 0 个 Pod。

更进一步地，如果我想回滚到更早之前的版本，要怎么办呢？

首先，我需要使用 kubectl rollout history 命令，查看每次 Deployment 变更对应的版本。而由于我们在创建这个 Deployment 的时候，指定了–record 参数，所以我们创建这些版本时执行的 kubectl 命令，都会被记录下来。这个操作的输出可以看到，我们前面执行的创建和更新操作，分别对应了版本 1 和版本 2，而那次失败的更新操作，则对应的是版本 3。

当然，你还可以通过这个 kubectl rollout history 指令，看到每个版本对应的 Deployment 的 API 对象的细节，具体命令如下所示：
```sh
kubectl rollout history deployment/nginx-deployment --revision=2
```
然后，我们就可以在 kubectl rollout undo 命令行最后，加上要回滚到的指定版本的版本号，就可以回滚到指定版本了。这个指令的用法如下：
```sh
kubectl rollout undo deployment/nginx-deployment --to-revision=2
```
这样，Deployment Controller 还会按照“滚动更新”的方式，完成对 Deployment 的降级操作。

不过，你可能已经想到了一个问题：我们对 Deployment 进行的每一次更新操作，都会生成一个新的 ReplicaSet 对象，是不是有些多余，甚至浪费资源呢？

没错。所以，Kubernetes 项目还提供了一个指令，使得我们对 Deployment 的多次更新操作，最后 只生成一个 ReplicaSet。

具体的做法是，在更新 Deployment 前，你要先执行一条 kubectl rollout pause 指令。它的用法如下所示：
```sh
kubectl rollout pause deployment/nginx-deployment
```
这个 kubectl rollout pause 的作用，是让这个 Deployment 进入了一个“暂停”状态。

所以接下来，你就可以随意使用 kubectl edit 或者 kubectl set image 指令，修改这个 Deployment 的内容了。

由于此时 Deployment 正处于“暂停”状态，所以我们对 Deployment 的所有修改，都不会触发新的“滚动更新”，也不会创建新的 ReplicaSet。

而等到我们对 Deployment 修改操作都完成之后，只需要再执行一条 kubectl rollout resume 指令，就可以把这个 Deployment“恢复”回来。而在这个 kubectl rollout resume 指令执行之前，在 kubectl rollout pause 指令之后的这段时间里，我们对 Deployment 进行的所有修改，最后只会触发一次“滚动更新”。

不过，即使你像上面这样小心翼翼地控制了 ReplicaSet 的生成数量，随着应用版本的不断增加，Kubernetes 中还是会为同一个 Deployment 保存很多很多不同的 ReplicaSet。

**那么，我们又该如何控制这些“历史”ReplicaSet 的数量呢？**

**很简单，Deployment 对象有一个字段，叫作 spec.revisionHistoryLimit，就是 Kubernetes 为 Deployment 保留的“历史版本”个数。所以，如果把它设置为 0，你就再也不能做回滚操作了。**


## # 深入理解StatefulSet（一）：拓扑状态

#### StatefulSet 可以说是 Kubernetes 项目中最为复杂的编排对象（一定要认真对待）

上面的 Deployment 实际上并不足以覆盖所有的应用编排问题。造成这个问题的根本原因，在于 Deployment 对应用做了一个简单化假设。

它认为，一个应用的所有 Pod，是完全一样的。所以，它们互相之间没有顺序，也无所谓运行在哪台宿主机上。需要的时候，Deployment 就可以通过 Pod 模板创建新的 Pod；不需要的时候，Deployment 就可以“杀掉”任意一个 Pod。

但是，在实际的场景中，并不是所有的应用都可以满足这样的要求。尤其是分布式应用，它的多个实例之间，往往有依赖关系，比如：主从关系、主备关系。

还有就是数据存储类应用，它的多个实例，往往都会在本地磁盘上保存一份数据。而这些实例一旦被杀掉，即便重建出来，实例与数据之间的对应关系也已经丢失，从而导致应用失败。

所以，这种实例之间有不对等关系，以及实例对外部数据有依赖关系的应用，就被称为“有状态应用”（Stateful Application）。

容器技术诞生后，大家很快发现，它用来封装“无状态应用”（Stateless Application），尤其是 Web 服务，非常好用。但是，一旦你想要用容器运行“有状态应用”，其困难程度就会直线上升。而且，这个问题解决起来，单纯依靠容器技术本身已经无能为力，这也就导致了很长一段时间内，“有状态应用”几乎成了容器技术圈子的“忌讳”，大家一听到这个词，就纷纷摇头。

不过，Kubernetes 项目还是成为了“第一个吃螃蟹的人”。

**得益于“控制器模式”的设计思想，Kubernetes 项目很早就在 Deployment 的基础上，扩展出了对“有状态应用”的初步支持。这个编排功能，就是：StatefulSet。**

StatefulSet 的设计其实非常容易理解。它把真实世界里的应用状态，抽象为了两种情况：
1. 拓扑状态。这种情况意味着，应用的多个实例之间不是完全对等的关系。这些应用实例，必须按照某些顺序启动，比如应用的主节点 A 要先于从节点 B 启动。而如果你把 A 和 B 两个 Pod 删除掉，它们再次被创建出来时也必须严格按照这个顺序才行。并且，新创建出来的 Pod，必须和原来 Pod 的网络标识一样，这样原先的访问者才能使用同样的方法，访问到这个新 Pod。
2. 存储状态。这种情况意味着，应用的多个实例分别绑定了不同的存储数据。对于这些应用实例来说，Pod A 第一次读取到的数据，和隔了十分钟之后再次读取到的数据，应该是同一份，哪怕在此期间 Pod A 被重新创建过。这种情况最典型的例子，就是一个数据库应用的多个存储实例。

所以，StatefulSet 的核心功能，就是通过某种方式记录这些状态，然后在 Pod 被重新创建时，能够为新 Pod 恢复这些状态。

在开始了解 StatefulSet 的工作原理之前，就必须先了解一个 Kubernetes 项目中非常实用的概念：Headless Service。

#### Headless Service

在讨论 Kubernetes 架构的时候就曾介绍过，Service 是 Kubernetes 项目中用来将一组 Pod 暴露给外界访问的一种机制。比如，一个 Deployment 有 3 个 Pod，那么我就可以定义一个 Service。然后，用户只要能访问到这个 Service，它就能访问到某个具体的 Pod。

那么，这个 Service 又是如何被访问的呢？
1. 第一种方式，是以 Service 的 VIP（Virtual IP，即：虚拟 IP）方式。比如：当我访问 10.0.23.1 这个 Service 的 IP 地址时，10.0.23.1 其实就是一个 VIP，它会把请求转发到该 Service 所代理的某一个 Pod 上。这里的具体原理，参考下面的 Service 部分。

2. 第二种方式，就是以 Service 的 DNS 方式。比如：这时候，只要我访问“my-svc.my-namespace.svc.cluster.local”这条 DNS 记录，就可以访问到名叫 my-svc 的 Service 所代理的某一个 Pod。

而在第二种 Service DNS 的方式下，具体还可以分为两种处理方法：
1. 第一种处理方法，是 Normal Service。这种情况下，你访问“my-svc.my-namespace.svc.cluster.local”解析到的，正是 my-svc 这个 Service 的 VIP，后面的流程就跟 VIP 方式一致了。
2. 而第二种处理方法，正是 Headless Service。这种情况下，你访问“my-svc.my-namespace.svc.cluster.local”解析到的，直接就是 my-svc 代理的某一个 Pod 的 IP 地址。可以看到，这里的区别在于，Headless Service 不需要分配一个 VIP，而是可以直接以 DNS 记录的方式解析出被代理 Pod 的 IP 地址。

那么，这样的设计又有什么作用呢？

想要回答这个问题，我们需要从 Headless Service 的定义方式看起。

下面是一个标准的 Headless Service 对应的 YAML 文件：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: nginx
```

可以看到，所谓的 Headless Service，其实仍是一个标准 Service 的 YAML 文件。只不过，它的 clusterIP 字段的值是：None，即：这个 Service，没有一个 VIP 作为“头”。这也就是 Headless 的含义。所以，这个 Service 被创建后并不会被分配一个 VIP，而是会以 DNS 记录的方式暴露出它所代理的 Pod。

而它所代理的 Pod，依然是采用在前面提到的 Label Selector 机制选择出来的，即：所有携带了 app=nginx 标签的 Pod，都会被这个 Service 代理起来。

**然后关键来了。**
当你按照这样的方式创建了一个 Headless Service 之后，它所代理的所有 Pod 的 IP 地址，都会被绑定一个这样格式的 DNS 记录，如下所示：
```<pod-name>.<svc-name>.<namespace>.svc.cluster.local```

这个 DNS 记录，正是 Kubernetes 项目为 Pod 分配的唯一的“可解析身份”（Resolvable Identity）。

有了这个“可解析身份”，只要你知道了一个 Pod 的名字，以及它对应的 Service 的名字，你就可以非常确定地通过这条 DNS 记录访问到 Pod 的 IP 地址。

那么，StatefulSet 又是如何使用这个 DNS 记录来维持 Pod 的拓扑状态的呢？

#### StatefulSet
为了回答这个问题，现在我们就来编写一个 StatefulSet 的 YAML 文件，如下所示：
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
          name: web
```
**这个 YAML 文件，和前面用到的 nginx-deployment 的唯一区别，就是多了一个 serviceName=nginx 字段。**

**这个字段的作用，就是告诉 StatefulSet 控制器，在执行控制循环（Control Loop）的时候，请使用 nginx 这个 Headless Service 来保证 Pod 的“可解析身份”。**

所以，当你通过 kubectl create 创建了上面这个 Service 和 StatefulSet 之后，就会看到如下两个对象：
```sh
$ kubectl create -f svc.yaml
$ kubectl get service nginx
NAME      TYPE         CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
nginx     ClusterIP    None         <none>        80/TCP    10s

$ kubectl create -f statefulset.yaml
$ kubectl get statefulset web
NAME      DESIRED   CURRENT   AGE
web       2         1         19s
```

这时候，如果你手比较快的话，还可以通过 kubectl 的 -w 参数，即：Watch 功能，实时查看 StatefulSet 创建两个有状态实例的过程。(如果手不够快的话，Pod 很快就创建完了。不过，你依然可以通过这个 StatefulSet 的 Events 看到这些信息。)

通过 Pod 的创建过程，我们不难看到，StatefulSet 给它所管理的所有 Pod 的名字，进行了编号，编号规则是：<statefulset name>-<ordinal index>。

而且这些编号都是从 0 开始累加，与 StatefulSet 的每个 Pod 实例一一对应，绝不重复。

更重要的是，这些 Pod 的创建，也是严格按照编号顺序进行的。比如，在 web-0 进入到 Running 状态、并且细分状态（Conditions）成为 Ready 之前，web-1 会一直处于 Pending 状态。（Ready 状态再一次提醒了我们，为 Pod 设置 livenessProbe 和 readinessProbe 的重要性。）

当这两个 Pod 都进入了 Running 状态之后，你就可以查看到它们各自唯一的“网络身份”了。

我们使用 kubectl exec 命令进入到容器中查看它们的 hostname：
```sh
$ kubectl exec web-0 -- sh -c 'hostname'
web-0
$ kubectl exec web-1 -- sh -c 'hostname'
web-1
```
可以看到，这两个 Pod 的 hostname 与 Pod 名字是一致的，都被分配了对应的编号。接下来，我们再试着以 DNS 的方式，访问一下这个 Headless Service：
```sh
$ kubectl run -i --tty --image busybox:latest dns-test --restart=Never --rm /bin/sh 
```
通过这条命令，我们启动了一个一次性的 Pod，因为 --rm 意味着 Pod 退出后就会被删除掉。然后，在这个 Pod 的容器里面，我们尝试用 nslookup 命令，解析一下 Pod 对应的 Headless Service：
```sh
$ kubectl run -i --tty --image busybox:latest dns-test --restart=Never --rm /bin/sh
$ nslookup web-0.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-0.nginx
Address 1: 10.244.1.7

$ nslookup web-1.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-1.nginx
Address 1: 10.244.2.7
```
从 nslookup 命令的输出结果中，我们可以看到，在访问 web-0.nginx 的时候，最后解析到的，正是 web-0 这个 Pod 的 IP 地址；而当访问 web-1.nginx 的时候，解析到的则是 web-1 的 IP 地址。

这时候，如果你在另外一个 Terminal 里把这两个“有状态应用”的 Pod 删掉：
```sh
$ kubectl delete pod -l app=nginx
pod "web-0" deleted
pod "web-1" deleted
```
然后，再在当前 Terminal 里 Watch 一下这两个 Pod 的状态变化，就会发现一个有趣的现象：
```sh
$ kubectl get pod -w -l app=nginx
NAME      READY     STATUS              RESTARTS   AGE
web-0     0/1       ContainerCreating   0          0s
NAME      READY     STATUS    RESTARTS   AGE
web-0     1/1       Running   0          2s
web-1     0/1       Pending   0         0s
web-1     0/1       ContainerCreating   0         0s
web-1     1/1       Running   0         32s
```
可以看到，当我们把这两个 Pod 删除之后，Kubernetes 会按照原先编号的顺序，创建出了两个新的 Pod。并且，Kubernetes 依然为它们分配了与原来相同的“网络身份”：web-0.nginx 和 web-1.nginx。

通过这种严格的对应规则，StatefulSet 就保证了 Pod 网络标识的稳定性。

比如，如果 web-0 是一个需要先启动的主节点，web-1 是一个后启动的从节点，那么只要这个 StatefulSet 不被删除，你访问 web-0.nginx 时始终都会落在主节点上，访问 web-1.nginx 时，则始终都会落在从节点上，这个关系绝对不会发生任何变化。

所以，如果我们再用 nslookup 命令，查看一下这个新 Pod 对应的 Headless Service 的话:
```sh
$ kubectl run -i --tty --image busybox:latest dns-test --restart=Never --rm /bin/sh 
$ nslookup web-0.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-0.nginx
Address 1: 10.244.1.8

$ nslookup web-1.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-1.nginx
Address 1: 10.244.2.8
```
我们可以看到，在这个 StatefulSet 中，这两个新 Pod 的“网络标识”（比如：web-0.nginx 和 web-1.nginx），再次解析到了正确的 IP 地址（比如：web-0 Pod 的 IP 地址 10.244.1.8）。

通过这种方法，Kubernetes 就成功地将 Pod 的拓扑状态（比如：哪个节点先启动，哪个节点后启动），按照 Pod 的“名字 + 编号”的方式固定了下来。此外，Kubernetes 还为每一个 Pod 提供了一个固定并且唯一的访问入口，即：这个 Pod 对应的 DNS 记录。

这些状态，在 StatefulSet 的整个生命周期里都会保持不变，绝不会因为对应 Pod 的删除或者重新创建而失效。

不过，相信你也已经注意到了，尽管 web-0.nginx 这条记录本身不会变，但它解析到的 Pod 的 IP 地址，并不是固定的。这就意味着，对于“有状态应用”实例的访问，你必须使用 DNS 记录或者 hostname 的方式，而绝不应该直接访问这些 Pod 的 IP 地址。

总结一句话就是 StatefulSet 这个控制器的主要作用之一，就是使用 Pod 模板创建 Pod 的时候，对它们进行编号，并且按照编号顺序逐一完成创建工作。而当 StatefulSet 的“控制循环”发现 Pod 的“实际状态”与“期望状态”不一致，需要新建或者删除 Pod 进行“调谐”的时候，它会严格按照这些 Pod 编号的顺序，逐一完成这些操作。


## # 深入理解StatefulSet（二）：存储状态
上面分享了 StatefulSet 如何保证应用实例的拓扑状态，在 Pod 删除和再创建的过程中保持稳定。

现在了解下 StatefulSet 对存储状态的管理机制。这个机制，主要使用的是一个叫作 Persistent Volume Claim 的功能。

#### Persistent Volume Claim (pvc) 和 Persistent Volume （pv）

在前面介绍 Pod 的时候，曾提到过，要在一个 Pod 里声明 Volume，只要在 Pod 里加上 spec.volumes 字段即可。然后，你就可以在这个字段里定义一个具体类型的 Volume 了，比如：hostPath。

可是，你有没有想过这样一个场景：如果你并不知道有哪些 Volume 类型可以用，要怎么办呢？

更具体地说，作为一个应用开发者，我可能对持久化存储项目（比如 Ceph、GlusterFS 等）一窍不通，也不知道公司的 Kubernetes 集群里到底是怎么搭建出来的，我也自然不会编写它们对应的 Volume 定义文件。

所谓“术业有专攻”，这些关于 Volume 的管理和远程持久化存储的知识，不仅超越了开发者的知识储备，还会有暴露公司基础设施秘密的风险。

比如，下面这个例子，就是一个声明了 Ceph RBD 类型 Volume 的 Pod：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: rbd
spec:
  containers:
    - image: kubernetes/pause
      name: rbd-rw
      volumeMounts:
      - name: rbdpd
        mountPath: /mnt/rbd
  volumes:
    - name: rbdpd
      rbd:
        monitors:
        - '10.16.154.78:6789'
        - '10.16.154.82:6789'
        - '10.16.154.83:6789'
        pool: kube
        image: foo
        fsType: ext4
        readOnly: true
        user: admin
        keyring: /etc/ceph/keyring
        imageformat: "2"
        imagefeatures: "layering"
```
其一，如果不懂得 Ceph RBD 的使用方法，那么这个 Pod 里 Volumes 字段，你十有八九也完全看不懂。其二，这个 Ceph RBD 对应的存储服务器的地址、用户名、授权文件的位置，也都被轻易地暴露给了全公司的所有开发人员，这是一个典型的信息被“过度暴露”的例子。

这也是为什么，在后来的演化中，Kubernetes 项目引入了一组叫作 Persistent Volume Claim（PVC）和 Persistent Volume（PV）的 API 对象，大大降低了用户声明和使用持久化 Volume 的门槛。

举个例子，有了 PVC 之后，一个开发人员想要使用一个 Volume，只需要简单的两步即可。

1. 第一步：定义一个 PVC，声明想要的 Volume 的属性：
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pv-claim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```
可以看到，在这个 PVC 对象里，不需要任何关于 Volume 细节的字段，只有描述性的属性和定义。比如，storage: 1Gi，表示我想要的 Volume 大小至少是 1 GiB；accessModes: ReadWriteOnce，表示这个 Volume 的挂载方式是可读写，并且只能被挂载在一个节点上而非被多个节点共享。（关于哪种类型的 Volume 支持哪种类型的 AccessMode，你可以查看 Kubernetes 项目官方文档中的[详细列表](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes)。）

2. 第二步：在应用的 Pod 中，声明使用这个 PVC：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pv-pod
spec:
  containers:
    - name: pv-container
      image: nginx
      ports:
        - containerPort: 80
          name: "http-server"
      volumeMounts:
        - mountPath: "/usr/share/nginx/html"
          name: pv-storage
  volumes:
    - name: pv-storage
      persistentVolumeClaim:
        claimName: pv-claim
```
可以看到，在这个 Pod 的 Volumes 定义中，我们只需要声明它的类型是 persistentVolumeClaim，然后指定 PVC 的名字，而完全不必关心 Volume 本身的定义。这时候，只要我们创建这个

这时候，只要我们创建这个 PVC 对象，Kubernetes 就会自动为它绑定一个符合条件的 Volume。可是，这些符合条件的 Volume 又是从哪里来的呢？

答案是，它们来自于由运维人员维护的 PV（Persistent Volume）对象。接下来，我们一起看一个常见的 PV 对象的 YAML 文件：
```yaml
kind: PersistentVolume
apiVersion: v1
metadata:
  name: pv-volume
  labels:
    type: local
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  rbd:
    monitors:
    # 使用 kubectl get pods -n rook-ceph 查看 rook-ceph-mon- 开头的 POD IP 即可得下面的列表
    - '10.16.154.78:6789'
    - '10.16.154.82:6789'
    - '10.16.154.83:6789'
    pool: kube
    image: foo
    fsType: ext4
    readOnly: true
    user: admin
    keyring: /etc/ceph/keyring
```

可以看到，这个 PV 对象的 spec.rbd 字段，正是我们前面介绍过的 Ceph RBD Volume 的详细定义。而且，它还声明了这个 PV 的容量是 10 GiB。这样，Kubernetes 就会为我们刚刚创建的 PVC 对象绑定这个 PV。

**所以，Kubernetes 中 PVC 和 PV 的设计，实际上类似于“接口”和“实现”的思想。开发者只要知道并会使用“接口”，即：PVC；而运维人员则负责给“接口”绑定具体的实现，即：PV。**

这种解耦，就避免了因为向开发者暴露过多的存储系统细节而带来的隐患。此外，这种职责的分离，往往也意味着出现事故时可以更容易定位问题和明确责任，从而避免“扯皮”现象的出现。

而 PVC、PV 的设计，也使得 StatefulSet 对存储状态的管理成为了可能。我们还是以上面用到的 StatefulSet（拓扑状态）为例：
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
```
这次，我们为这个 StatefulSet 额外添加了一个 volumeClaimTemplates 字段。从名字就可以看出来，它跟 Deployment 里 Pod 模板（PodTemplate）的作用类似。也就是说，凡是被这个 StatefulSet 管理的 Pod，都会声明一个对应的 PVC；而这个 PVC 的定义，就来自于 volumeClaimTemplates 这个模板字段。更重要的是，这个 PVC 的名字，会被分配一个与这个 Pod 完全一致的编号。

这个自动创建的 PVC，与 PV 绑定成功后，就会进入 Bound 状态，这就意味着这个 Pod 可以挂载并使用这个 PV 了。

如果你还是不太理解 PVC 的话，可以先记住这样一个结论：PVC 其实就是一种特殊的 Volume。只不过一个 PVC 具体是什么类型的 Volume，要在跟某个 PV 绑定之后才知道。关于 PV、PVC 更详细的知识，会在下面容器存储部分做进一步解读。

当然，PVC 与 PV 的绑定得以实现的前提是，运维人员已经在系统里创建好了符合条件的 PV（比如，我们在前面用到的 pv-volume）；或者，你的 Kubernetes 集群运行在公有云上，这样 Kubernetes 就会通过 Dynamic Provisioning 的方式（在kubernetes集群内搭建一个动态数据卷，然后创建对应的StorageClass对象，就可以在Pod里直接使用PVC了，动态数据卷会自动创建匹配的PV对象），自动为你创建与 PVC 匹配的 PV。

所以，我们在使用 kubectl create 创建了 StatefulSet 之后，就会看到 Kubernetes 集群里出现了两个 PVC：
```sh
$ kubectl create -f statefulset.yaml
$ kubectl get pvc -l app=nginx
NAME        STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
www-web-0   Bound     pvc-15c268c7-b507-11e6-932f-42010a800002   1Gi        RWO           48s
www-web-1   Bound     pvc-15c79307-b507-11e6-932f-42010a800002   1Gi        RWO           48s
```
可以看到，这些 PVC，都以“--< 编号 >”的方式命名，并且处于 Bound 状态。

我们前面已经讲到过，这个 StatefulSet 创建出来的所有 Pod，都会声明使用编号的 PVC。比如，在名叫 web-0 的 Pod 的 volumes 字段，它会声明使用名叫 www-web-0 的 PVC，从而挂载到这个 PVC 所绑定的 PV。

所以，我们就可以使用如下所示的指令，在 Pod 的 Volume 目录里写入一个文件，来验证一下上述 Volume 的分配情况：
```sh
for i in 0 1; do kubectl exec web-$i -- sh -c 'echo hello $(hostname) > /usr/share/nginx/html/index.html'; done
```

如上所示，通过 kubectl exec 指令，我们在每个 Pod 的 Volume 目录里，写入了一个 index.html 文件。这个文件的内容，正是 Pod 的 hostname。比如，我们在 web-0 的 index.html 里写入的内容就是"hello web-0"。

此时，如果你在这个 Pod 容器里访问“http://localhost”，你实际访问到的就是 Pod 里 Nginx 服务器进程，而它会为你返回 /usr/share/nginx/html/index.html 里的内容。这个操作的执行方法如下所示：
```sh
$ for i in 0 1; do kubectl exec -it web-$i -- curl localhost; done
hello web-0
hello web-1
```

现在，关键来了。

如果你使用 kubectl delete 命令删除这两个 Pod，这些 Volume 里的文件会不会丢失呢？

可以看到，正如我们前面介绍过的，在被删除之后，这两个 Pod 会被按照编号的顺序被重新创建出来。而这时候，如果你在新创建的容器里通过访问“http://localhost”的方式去访问 web-0 里的 Nginx 服务。

就会发现，这个请求依然会返回：hello web-0。也就是说，原先与名叫 web-0 的 Pod 绑定的 PV，在这个 Pod 被重新创建之后，依然同新的名叫 web-0 的 Pod 绑定在了一起。对于 Pod web-1 来说，也是完全一样的情况。

这是怎么做到的呢？

其实，我和你分析一下 StatefulSet 控制器恢复这个 Pod 的过程，你就可以很容易理解了。

首先，当你把一个 Pod，比如 web-0，删除之后，这个 Pod 对应的 PVC 和 PV，并不会被删除，而这个 Volume 里已经写入的数据，也依然会保存在远程存储服务里（比如，我们在这个例子里用到的 Ceph 服务器）。

此时，StatefulSet 控制器发现，一个名叫 web-0 的 Pod 消失了。所以，控制器就会重新创建一个新的、名字还是叫作 web-0 的 Pod 来，“纠正”这个不一致的情况。

需要注意的是，在这个新的 Pod 对象的定义里，它声明使用的 PVC 的名字，还是叫作：www-web-0。这个 PVC 的定义，还是来自于 PVC 模板（volumeClaimTemplates），这是 StatefulSet 创建 Pod 的标准流程。

所以，在这个新的 web-0 Pod 被创建出来之后，Kubernetes 为它查找名叫 www-web-0 的 PVC 时，就会直接找到旧 Pod 遗留下来的同名的 PVC，进而找到跟这个 PVC 绑定在一起的 PV。

这样，新的 Pod 就可以挂载到旧 Pod 对应的那个 Volume，并且获取到保存在 Volume 里的数据。

通过这种方式，Kubernetes 的 StatefulSet 就实现了对应用存储状态的管理。

看到这里，你是不是已经大致理解了 StatefulSet 的工作原理呢？现在，我再为你详细梳理一下吧。
1. 首先，StatefulSet 的控制器直接管理的是 Pod。这是因为，StatefulSet 里的不同 Pod 实例，不再像 ReplicaSet 中那样都是完全一样的，而是有了细微区别的。比如，每个 Pod 的 hostname、名字等都是不同的、携带了编号的。而 StatefulSet 区分这些实例的方式，就是通过在 Pod 的名字里加上事先约定好的编号。
2. 其次，Kubernetes 通过 Headless Service，为这些有编号的 Pod，在 DNS 服务器中生成带有同样编号的 DNS 记录。只要 StatefulSet 能够保证这些 Pod 名字里的编号不变，那么 Service 里类似于 web-0.nginx.default.svc.cluster.local 这样的 DNS 记录也就不会变，而这条记录解析出来的 Pod 的 IP 地址，则会随着后端 Pod 的删除和再创建而自动更新。这当然是 Service 机制本身的能力，不需要 StatefulSet 操心。
3. 最后，StatefulSet 还为每一个 Pod 分配并创建一个同样编号的 PVC。这样，Kubernetes 就可以通过 Persistent Volume 机制为这个 PVC 绑定上对应的 PV，从而保证了每一个 Pod 都拥有一个独立的 Volume。

在这种情况下，即使 Pod 被删除，它所对应的 PVC 和 PV 依然会保留下来。所以当这个 Pod 被重新创建出来之后，Kubernetes 会为它找到同样编号的 PVC，挂载这个 PVC 对应的 Volume，从而获取到以前保存在 Volume 里的数据。


## # 深入理解StatefulSet（三）：有状态应用实践
上面了解了 StatefulSet 的工作原理，以及处理拓扑状态和存储状态的方法。现在通过一个实际的例子，再次深入解读一下部署一个 StatefulSet 的完整流程。

今天我选择的实例是部署一个 MySQL 集群，这也是 Kubernetes 官方文档里的一个经典案例。但是，很多工程师都曾说这个例子“完全看不懂”。

其实，这样的吐槽也可以理解：相比于 Etcd、Cassandra 等“原生”就考虑了分布式需求的项目，MySQL 以及很多其他的数据库项目，在分布式集群的搭建上并不友好，甚至有点“原始”。

首先，用自然语言来描述一下我们想要部署的“有状态应用”：
1. 是一个“主从复制”（Maser-Slave Replication）的 MySQL 集群；
2. 有 1 个主节点（Master）；
3. 有多个从节点（Slave）；
4. 从节点需要能水平扩展；
5. 所有的写操作，只能在主节点上执行；
6. 读操作可以在所有节点上执行。

在常规环境里，部署这样一个主从模式的 MySQL 集群的主要难点在于：如何让从节点能够拥有主节点的数据，即：如何配置主（Master）从（Slave）节点的复制与同步。

所以，在安装好 MySQL 的 Master 节点之后，你需要做的第一步工作，就是通过 XtraBackup 将 Master 节点的数据备份到指定目录。（XtraBackup 是业界主要使用的开源 MySQL 备份和恢复工具。）

这一步会自动在目标目录里生成一个备份信息文件，名叫：xtrabackup_binlog_info。这个文件一般会包含如下两个信息：
```sh
$ cat xtrabackup_binlog_info
TheMaster-bin.000001     481
```
这两个信息会在接下来配置 Slave 节点的时候用到。

第二步：配置 Slave 节点。Slave 节点在第一次启动前，需要先把 Master 节点的备份数据，连同备份信息文件，一起拷贝到自己的数据目录（/var/lib/mysql）下。然后，我们执行这样一句 SQL：
```sql
CHANGE MASTER TO MASTER_HOST='$masterip',
  MASTER_USER='xxx',
  MASTER_PASSWORD='xxx',
  MASTER_LOG_FILE='TheMaster-bin.000001',
  MASTER_LOG_POS=481;
```
其中，MASTER_LOG_FILE 和 MASTER_LOG_POS，就是该备份对应的二进制日志（Binary Log）文件的名称和开始的位置（偏移量），也正是 xtrabackup_binlog_info 文件里的那两部分内容（即：TheMaster-bin.000001 和 481）。

第三步，启动 Slave 节点。在这一步，我们需要执行这样一句 SQL：
```sql
START SLAVE;
```
这样，Slave 节点就启动了。它会使用备份信息文件中的二进制日志文件和偏移量，与主节点进行数据同步。

第四步，在这个集群中添加更多的 Slave 节点。

需要注意的是，新添加的 Slave 节点的备份数据，来自于已经存在的 Slave 节点。

所以，在这一步，我们需要将 Slave 节点的数据备份在指定目录。而这个备份操作会自动生成另一种备份信息文件，名叫：xtrabackup_slave_info。同样地，这个文件也包含了 MASTER_LOG_FILE 和 MASTER_LOG_POS 两个字段。

然后，我们就可以执行跟前面一样的“CHANGE MASTER TO”和“START SLAVE” 指令，来初始化并启动这个新的 Slave 节点了。

通过上面的叙述，我们不难看到，将部署 MySQL 集群的流程迁移到 Kubernetes 项目上，需要能够“容器化”地解决下面的“三座大山”：
1. Master 节点和 Slave 节点需要有不同的配置文件（即：不同的 my.cnf）；
2. Master 节点和 Slave 节点需要能够传输备份信息文件；
3. 在 Slave 节点第一次启动之前，需要执行一些初始化 SQL 操作；

而由于 MySQL 本身同时拥有拓扑状态（主从节点的区别）和存储状态（MySQL 保存在本地的数据），我们自然要通过 StatefulSet 来解决这“三座大山”的问题。

#### 第一座大山：Master 节点和 Slave 节点需要有不同的配置文件
其中，“第一座大山：Master 节点和 Slave 节点需要有不同的配置文件”，很容易处理：我们只需要给主从节点分别准备两份不同的 MySQL 配置文件，然后根据 Pod 的序号（Index）挂载进去即可。

正如前面介绍过的一样，这样的配置文件信息，应该保存在 ConfigMap 里供 Pod 使用。它的定义如下所示：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql
  labels:
    app: mysql
data:
  master.cnf: |
    # 主节点MySQL的配置文件
    [mysqld]
    log-bin
  slave.cnf: |
    # 从节点MySQL的配置文件
    [mysqld]
    super-read-only
```
在这里，我们定义了 master.cnf 和 slave.cnf 两个 MySQL 的配置文件。
- master.cnf 开启了 log-bin，即：使用二进制日志文件的方式进行主从复制，这是一个标准的设置。
- slave.cnf 的开启了 super-read-only，代表的是从节点会拒绝除了主节点的数据同步操作之外的所有写操作，即：它对用户是只读的。

而上述 ConfigMap 定义里的 data 部分，是 Key-Value 格式的。比如，master.cnf 就是这份配置数据的 Key，而“|”后面的内容，就是这份配置数据的 Value。这份数据将来挂载进 Master 节点对应的 Pod 后，就会在 Volume 目录里生成一个叫作 master.cnf 的文件。

接下来，我们需要创建两个 Service 来供 StatefulSet 以及用户使用。这两个 Service 的定义如下所示：
```yaml
# mysql
apiVersion: v1
kind: Service
metadata:
  name: mysql
  labels:
    app: mysql
spec:
  ports:
  - name: mysql
    port: 3306
  clusterIP: None
  selector:
    app: mysql

# mysql-read
apiVersion: v1
kind: Service
metadata:
  name: mysql-read
  labels:
    app: mysql
spec:
  ports:
  - name: mysql
    port: 3306
  selector:
    app: mysql
```
可以看到，这两个 Service 都代理了所有携带 app=mysql 标签的 Pod，也就是所有的 MySQL Pod。端口映射都是用 Service 的 3306 端口对应 Pod 的 3306 端口。

不同的是，第一个名叫“mysql”的 Service 是一个 Headless Service（即：clusterIP= None）。所以它的作用，是通过为 Pod 分配 DNS 记录来固定它的拓扑状态，比如“mysql-0.mysql”和“mysql-1.mysql”这样的 DNS 名字。其中，编号为 0 的节点就是我们的主节点。

而第二个名叫“mysql-read”的 Service，则是一个常规的 Service。

并且我们规定，所有用户的读请求，都必须访问第二个 Service 被自动分配的 DNS 记录，即：“mysql-read”（当然，也可以访问这个 Service 的 VIP）。这样，读请求就可以被转发到任意一个 MySQL 的主节点或者从节点上。（Kubernetes 中的所有 Service、Pod 对象，都会被自动分配同名的 DNS 记录。具体细节，参考后面的 Service 部分。）

而所有用户的写请求，则必须直接以 DNS 记录的方式访问到 MySQL 的主节点，也就是：“mysql-0.mysql“这条 DNS 记录。

#### 接下来，我们再一起解决“第二座大山：Master 节点和 Slave 节点需要能够传输备份文件”的问题。
翻越这座大山的思路，我比较推荐的做法是：先搭建框架，再完善细节。其中，Pod 部分如何定义，是完善细节时的重点。

在这一步，我们可以先为 StatefulSet 定义一些通用的字段。比如：selector 表示，这个 StatefulSet 要管理的 Pod 必须携带 app=mysql 标签；它声明要使用的 Headless Service 的名字是：mysql。

这个 StatefulSet 的 replicas 值是 3，表示它定义的 MySQL 集群有三个节点：一个 Master 节点，两个 Slave 节点。

可以看到，StatefulSet 管理的“有状态应用”的多个实例，也都是通过同一份 Pod 模板创建出来的，使用的是同一个 Docker 镜像。这也就意味着：如果你的应用要求不同节点的镜像不一样，那就不能再使用 StatefulSet 了。对于这种情况，应该考虑用下面的 Operator。

除了这些基本的字段外，作为一个有存储状态的 MySQL 集群，StatefulSet 还需要管理存储状态。所以，我们需要通过 volumeClaimTemplate（PVC 模板）来为每个 Pod 定义 PVC。比如，这个 PVC 模板的 resources.requests.strorage 指定了存储的大小为 10 GiB；ReadWriteOnce 指定了该存储的属性为可读写，并且一个 PV 只允许挂载在一个宿主机上。将来，这个 PV 对应的的 Volume 就会充当 MySQL Pod 的存储数据目录。

然后，我们来重点设计一下这个 StatefulSet 的 Pod 模板，也就是 template 字段。

由于 StatefulSet 管理的 Pod 都来自于同一个镜像，这就要求我们在编写 Pod 时，一定要保持清醒，用“人格分裂”的方式进行思考：
1. 如果这个 Pod 是 Master 节点，我们要怎么做；
2. 如果这个 Pod 是 Slave 节点，我们又要怎么做。

想清楚这两个问题，我们就可以按照 Pod 的启动过程来一步步定义它们了：
1. 第一步：从 ConfigMap 中，获取 MySQL 的 Pod 对应的配置文件。
为此，我们需要进行一个初始化操作，根据节点的角色是 Master 还是 Slave 节点，为 Pod 分配对应的配置文件。此外，MySQL 还要求集群里的每个节点都有一个唯一的 ID 文件，名叫 server-id.cnf。

而根据我们已经掌握的 Pod 知识，这些初始化操作显然适合通过 InitContainer 来完成。所以，我们首先定义了一个 InitContainer，如下所示：
```yaml
      # ...
      # template.spec
      initContainers:
      - name: init-mysql
        image: mysql:5.7
        command:
        - bash
        - "-c"
        - |
          set -ex
          # 从Pod的序号，生成server-id
          [[ `hostname` =~ -([0-9]+)$ ]] || exit 1
          ordinal=${BASH_REMATCH[1]}
          echo [mysqld] > /mnt/conf.d/server-id.cnf
          # 由于server-id=0有特殊含义，我们给ID加一个100来避开它
          echo server-id=$((100 + $ordinal)) >> /mnt/conf.d/server-id.cnf
          # 如果Pod序号是0，说明它是Master节点，从ConfigMap里把Master的配置文件拷贝到/mnt/conf.d/目录；
          # 否则，拷贝Slave的配置文件
          if [[ $ordinal -eq 0 ]]; then
            cp /mnt/config-map/master.cnf /mnt/conf.d/
          else
            cp /mnt/config-map/slave.cnf /mnt/conf.d/
          fi
        volumeMounts:
        - name: conf
          mountPath: /mnt/conf.d
        - name: config-map
          mountPath: /mnt/config-map
```
在这个名叫 init-mysql 的 InitContainer 的配置中，它从 Pod 的 hostname 里，读取到了 Pod 的序号，以此作为 MySQL 节点的 server-id。

然后，init-mysql 通过这个序号，判断当前 Pod 到底是 Master 节点（即：序号为 0）还是 Slave 节点（即：序号不为 0），从而把对应的配置文件从 /mnt/config-map 目录拷贝到 /mnt/conf.d/ 目录下。

其中，文件拷贝的源目录 /mnt/config-map，正是 ConfigMap 在这个 Pod 的 Volume，如下所示：
```yaml
      ...
      # template.spec
      volumes:
      - name: conf
        emptyDir: {}
      - name: config-map
        configMap:
          name: mysql
```
通过这个定义，init-mysql 在声明了挂载 config-map 这个 Volume 之后，ConfigMap 里保存的内容，就会以文件的方式出现在它的 /mnt/config-map 目录当中。

而文件拷贝的目标目录，即容器里的 /mnt/conf.d/ 目录，对应的则是一个名叫 conf 的、emptyDir 类型的 Volume。基于 Pod Volume 共享的原理，当 InitContainer 复制完配置文件退出后，后面启动的 MySQL 容器只需要直接声明挂载这个名叫 conf 的 Volume，它所需要的.cnf 配置文件已经出现在里面了。这跟我们之前介绍的 Tomcat 和 WAR 包的处理方法是完全一样的。

2. 第二步：在 Slave Pod 启动前，从 Master 或者其他 Slave Pod 里拷贝数据库数据到自己的目录下。
为了实现这个操作，我们就需要再定义第二个 InitContainer，如下所示：
```yaml
      ...
      # template.spec.initContainers
      - name: clone-mysql
        image: gcr.io/google-samples/xtrabackup:1.0
        command:
        - bash
        - "-c"
        - |
          set -ex
          # 拷贝操作只需要在第一次启动时进行，所以如果数据已经存在，跳过
          [[ -d /var/lib/mysql/mysql ]] && exit 0
          # Master节点(序号为0)不需要做这个操作
          [[ `hostname` =~ -([0-9]+)$ ]] || exit 1
          ordinal=${BASH_REMATCH[1]}
          [[ $ordinal -eq 0 ]] && exit 0
          # 使用ncat指令，远程地从前一个节点拷贝数据到本地
          ncat --recv-only mysql-$(($ordinal-1)).mysql 3307 | xbstream -x -C /var/lib/mysql
          # 执行--prepare，这样拷贝来的数据就可以用作恢复了
          xtrabackup --prepare --target-dir=/var/lib/mysql
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
          subPath: mysql
        - name: conf
          mountPath: /etc/mysql/conf.d
```
在这个名叫 clone-mysql 的 InitContainer 里，我们使用的是 xtrabackup 镜像（它里面安装了 xtrabackup 工具）。

而在它的启动命令里，我们首先做了一个判断。即：当初始化所需的数据（/var/lib/mysql/mysql 目录）已经存在，或者当前 Pod 是 Master 节点的时候，不需要做拷贝操作。

接下来，clone-mysql 会使用 Linux 自带的 ncat 指令，向 DNS 记录为“mysql-< 当前序号减一 >.mysql”的 Pod，也就是当前 Pod 的前一个 Pod，发起数据传输请求，并且直接用 xbstream 指令将收到的备份数据保存在 /var/lib/mysql 目录下。（3307 是一个特殊端口，运行着一个专门负责备份 MySQL 数据的辅助进程。我们后面马上会讲到它。）

当然，这一步你可以随意选择用自己喜欢的方法来传输数据。比如，用 scp 或者 rsync，都没问题。

你可能已经注意到，这个容器里的 /var/lib/mysql 目录，实际上正是一个名为 data 的 PVC，即：我们在前面声明的持久化存储。

这就可以保证，哪怕宿主机宕机了，我们数据库的数据也不会丢失。更重要的是，由于 Pod Volume 是被 Pod 里的容器共享的，所以后面启动的 MySQL 容器，就可以把这个 Volume 挂载到自己的 /var/lib/mysql 目录下，直接使用里面的备份数据进行恢复操作。

不过，clone-mysql 容器还要对 /var/lib/mysql 目录，执行一句 xtrabackup --prepare 操作，目的是让拷贝来的数据进入一致性状态，这样，这些数据才能被用作数据恢复。

至此，我们就通过 InitContainer 完成了对“主、从节点间备份文件传输”操作的处理过程，也就是翻越了“第二座大山”。

接下来，我们可以开始定义 MySQL 容器, 启动 MySQL 服务了。由于 StatefulSet 里的所有 Pod 都来自用同一个 Pod 模板，所以我们还要“人格分裂”地去思考：这个 MySQL 容器的启动命令，在 Master 和 Slave 两种情况下有什么不同。

有了 Docker 镜像，在 Pod 里声明一个 Master 角色的 MySQL 容器并不是什么困难的事情：直接执行 MySQL 启动命令即可。

但是，如果这个 Pod 是一个第一次启动的 Slave 节点，在执行 MySQL 启动命令之前，它就需要使用前面 InitContainer 拷贝来的备份数据进行初始化。

可是，别忘了，容器是一个单进程模型。

所以，一个 Slave 角色的 MySQL 容器启动之前，谁能负责给它执行初始化的 SQL 语句呢？

#### 这就是我们需要解决的“第三座大山”的问题，即：如何在 Slave 节点的 MySQL 容器第一次启动之前，执行初始化 SQL。
你可能已经想到了，我们可以为这个 MySQL 容器额外定义一个 sidecar 容器，来完成这个操作，它的定义如下所示：（之所以不用 initContainer 而是sidercar容器，是因为下文说了，需要“使用 ncat 命令启动一个工作在 3307 端口上的网络发送服务”，initContainer 是一次性的初始化动作，执行完初始化脚本直接就退出了。所以这里用sidecar是合适的，因为容器是单进程，一个常驻服务需要一个单独的容器。）
```yaml
      ...
      # template.spec.containers
      - name: xtrabackup
        image: gcr.io/google-samples/xtrabackup:1.0
        ports:
        - name: xtrabackup
          containerPort: 3307
        command:
        - bash
        - "-c"
        - |
          set -ex
          cd /var/lib/mysql
          
          # 从备份信息文件里读取MASTER_LOG_FILEM和MASTER_LOG_POS这两个字段的值，用来拼装集群初始化SQL
          if [[ -f xtrabackup_slave_info ]]; then
            # 如果xtrabackup_slave_info文件存在，说明这个备份数据来自于另一个Slave节点。这种情况下，XtraBackup工具在备份的时候，就已经在这个文件里自动生成了"CHANGE MASTER TO" SQL语句。所以，我们只需要把这个文件重命名为change_master_to.sql.in，后面直接使用即可
            mv xtrabackup_slave_info change_master_to.sql.in
            # 所以，也就用不着xtrabackup_binlog_info了
            rm -f xtrabackup_binlog_info
          elif [[ -f xtrabackup_binlog_info ]]; then
            # 如果只存在xtrabackup_binlog_inf文件，那说明备份来自于Master节点，我们就需要解析这个备份信息文件，读取所需的两个字段的值
            [[ `cat xtrabackup_binlog_info` =~ ^(.*?)[[:space:]]+(.*?)$ ]] || exit 1
            rm xtrabackup_binlog_info
            # 把两个字段的值拼装成SQL，写入change_master_to.sql.in文件
            echo "CHANGE MASTER TO MASTER_LOG_FILE='${BASH_REMATCH[1]}',\
                  MASTER_LOG_POS=${BASH_REMATCH[2]}" > change_master_to.sql.in
          fi
          
          # 如果change_master_to.sql.in，就意味着需要做集群初始化工作
          if [[ -f change_master_to.sql.in ]]; then
            # 但一定要先等MySQL容器启动之后才能进行下一步连接MySQL的操作
            echo "Waiting for mysqld to be ready (accepting connections)"
            until mysql -h 127.0.0.1 -e "SELECT 1"; do sleep 1; done
            
            echo "Initializing replication from clone position"
            # 将文件change_master_to.sql.in改个名字，防止这个Container重启的时候，因为又找到了change_master_to.sql.in，从而重复执行一遍这个初始化流程
            mv change_master_to.sql.in change_master_to.sql.orig
            # 使用change_master_to.sql.orig的内容，也是就是前面拼装的SQL，组成一个完整的初始化和启动Slave的SQL语句
            mysql -h 127.0.0.1 <<EOF
          $(<change_master_to.sql.orig),
            MASTER_HOST='mysql-0.mysql',
            MASTER_USER='root',
            MASTER_PASSWORD='',
            MASTER_CONNECT_RETRY=10;
          START SLAVE;
          EOF
          fi
          
          # 使用ncat监听3307端口。它的作用是，在收到传输请求的时候，直接执行"xtrabackup --backup"命令，备份MySQL的数据并发送给请求者
          exec ncat --listen --keep-open --send-only --max-conns=1 3307 -c \
            "xtrabackup --backup --slave-info --stream=xbstream --host=127.0.0.1 --user=root"
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
          subPath: mysql
        - name: conf
          mountPath: /etc/mysql/conf.d
```
可以看到，在这个名叫 xtrabackup 的 sidecar 容器的启动命令里，其实实现了两部分工作。
1. 第一部分工作，当然是 MySQL 节点的初始化工作。这个初始化需要使用的 SQL，是 sidecar 容器拼装出来、保存在一个名为 change_master_to.sql.in 的文件里的，具体过程如下所示：

sidecar 容器首先会判断当前 Pod 的 /var/lib/mysql 目录下，是否有 xtrabackup_slave_info 这个备份信息文件。
   - 如果有，则说明这个目录下的备份数据是由一个 Slave 节点生成的。这种情况下，XtraBackup 工具在备份的时候，就已经在这个文件里自动生成了"CHANGE MASTER TO" SQL 语句。所以，我们只需要把这个文件重命名为 change_master_to.sql.in，后面直接使用即可。
   - 如果没有 xtrabackup_slave_info 文件、但是存在 xtrabackup_binlog_info 文件，那就说明备份数据来自于 Master 节点。这种情况下，sidecar 容器就需要解析这个备份信息文件，读取 MASTER_LOG_FILE 和 MASTER_LOG_POS 这两个字段的值，用它们拼装出初始化 SQL 语句，然后把这句 SQL 写入到 change_master_to.sql.in 文件中。
接下来，sidecar 容器就可以执行初始化了。从上面的叙述中可以看到，只要这个 change_master_to.sql.in 文件存在，那就说明接下来需要进行集群初始化操作。

所以，这时候，sidecar 容器只需要读取并执行 change_master_to.sql.in 里面的“CHANGE MASTER TO”指令，再执行一句 START SLAVE 命令，一个 Slave 节点就被成功启动了。（需要注意的是：Pod 里的容器并没有先后顺序，所以在执行初始化 SQL 之前，必须先执行一句 SQL（select 1）来检查一下 MySQL 服务是否已经可用。）

当然，上述这些初始化操作完成后，我们还要删除掉前面用到的这些备份信息文件。否则，下次这个容器重启时，就会发现这些文件存在，所以又会重新执行一次数据恢复和集群初始化的操作，这是不对的。同理，change_master_to.sql.in 在使用后也要被重命名，以免容器重启时因为发现这个文件存在又执行一遍初始化。

2. 在完成 MySQL 节点的初始化后，这个 sidecar 容器的第二个工作，则是启动一个数据传输服务。
具体做法是：sidecar 容器会使用 ncat 命令启动一个工作在 3307 端口上的网络发送服务。一旦收到数据传输请求时，sidecar 容器就会调用 xtrabackup --backup 指令备份当前 MySQL 的数据，然后把这些备份数据返回给请求者。这就是为什么我们在 InitContainer 里定义数据拷贝的时候，访问的是“上一个 MySQL 节点”的 3307 端口。

值得一提的是，由于 sidecar 容器和 MySQL 容器同处于一个 Pod 里，所以它是直接通过 Localhost 来访问和备份 MySQL 容器里的数据的，非常方便。

同样地，我在这里举例用的只是一种备份方法而已，你完全可以选择其他自己喜欢的方案。比如，你可以使用 innobackupex 命令做数据备份和准备，它的使用方法几乎与这里的用的备份方法一样。

至此，我们也就翻越了“第三座大山”，完成了 Slave 节点第一次启动前的初始化工作。

扳倒了这“三座大山”后，我们终于可以定义 Pod 里的主角，MySQL 容器了。有了前面这些定义和初始化工作，MySQL 容器本身的定义就非常简单了，如下所示：
```yaml
      ...
      # template.spec
      containers:
      - name: mysql
        image: mysql:5.7
        env:
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "1"
        ports:
        - name: mysql
          containerPort: 3306
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
          subPath: mysql
        - name: conf
          mountPath: /etc/mysql/conf.d
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
        livenessProbe:
          exec:
            command: ["mysqladmin", "ping"]
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          exec:
            # 通过TCP连接的方式进行健康检查
            command: ["mysql", "-h", "127.0.0.1", "-e", "SELECT 1"]
          initialDelaySeconds: 5
          periodSeconds: 2
          timeoutSeconds: 1
```
在这个容器的定义里，我们使用了一个标准的 MySQL 5.7 的官方镜像。它的数据目录是 /var/lib/mysql，配置文件目录是 /etc/mysql/conf.d。

这时候，你应该能够明白，如果 MySQL 容器是 Slave 节点的话，它的数据目录里的数据，就来自于 InitContainer 从其他节点里拷贝而来的备份。它的配置文件目录 /etc/mysql/conf.d 里的内容，则来自于 ConfigMap 对应的 Volume。而它的初始化工作，则是由同一个 Pod 里的 sidecar 容器完成的。这些操作，正是我刚刚为你讲述的大部分内容。

另外，我们为它定义了一个 livenessProbe，通过 mysqladmin ping 命令来检查它是否健康；还定义了一个 readinessProbe，通过查询 SQL（select 1）来检查 MySQL 服务是否可用。当然，凡是 readinessProbe 检查失败的 MySQL Pod，都会从 Service 里被摘除掉。

至此，一个完整的主从复制模式的 MySQL 集群就定义完了。

现在，我们就可以使用 kubectl 命令，尝试运行一下这个 StatefulSet 了。

首先，我们需要在 Kubernetes 集群里创建满足条件的 PV。

在这里是测试项目，可以用 StorageClass 来完成这个操作。它的作用，是自动地为集群里存在的每一个 PVC，调用存储插件（Rook）创建对应的 PV，从而省去了我们手动创建 PV 的机械劳动。后续容器存储部分，会再详细介绍这个机制。（在使用 Rook 的情况下，mysql-statefulset.yaml 里的 volumeClaimTemplates 字段需要加上声明 storageClassName=rook-ceph-block，才能使用到这个 Rook 提供的持久化存储。）

然后，我们就可以创建这个 StatefulSet 了，可以看到，StatefulSet 启动成功后，会有三个 Pod 运行。

接下来，我们可以尝试向这个 MySQL 集群发起请求，执行一些 SQL 操作来验证它是否正常：
```sh
$ kubectl run mysql-client --image=mysql:5.7 -i --rm --restart=Never --\
  mysql -h mysql-0.mysql <<EOF
CREATE DATABASE test;
CREATE TABLE test.messages (message VARCHAR(250));
INSERT INTO test.messages VALUES ('hello');
EOF
```
如上所示，我们通过启动一个容器，使用 MySQL client 执行了创建数据库和表、以及插入数据的操作。需要注意的是，我们连接的 MySQL 的地址必须是 mysql-0.mysql（即：Master 节点的 DNS 记录）。因为，只有 Master 节点才能处理写操作。

而通过连接 mysql-read 这个 Service，我们就可以用 SQL 进行读操作，如下所示：
```sh
$ kubectl run mysql-client --image=mysql:5.7 -i -t --rm --restart=Never --\
 mysql -h mysql-read -e "SELECT * FROM test.messages"
Waiting for pod default/mysql-client to be running, status is Pending, pod ready: false
+---------+
| message |
+---------+
| hello   |
+---------+
pod "mysql-client" deleted
```

在有了 StatefulSet 以后，你就可以像 Deployment 那样，非常方便地扩展这个 MySQL 集群，比如：
```sh
kubectl scale statefulset mysql  --replicas=5
```
这时候，你就会发现新的 Slave Pod mysql-3 和 mysql-4 被自动创建了出来。

而如果你像如下所示的这样，直接连接 mysql-3.mysql，即 mysql-3 这个 Pod 的 DNS 名字来进行查询操作：
```sh
$ kubectl run mysql-client --image=mysql:5.7 -i -t --rm --restart=Never --\
  mysql -h mysql-3.mysql -e "SELECT * FROM test.messages"
Waiting for pod default/mysql-client to be running, status is Pending, pod ready: false
+---------+
| message |
+---------+
| hello   |
+---------+
pod "mysql-client" deleted
```

就会看到，从 StatefulSet 为我们新创建的 mysql-3 上，同样可以读取到之前插入的记录。也就是说，我们的数据备份和恢复，都是有效的。

#### StatefulSet 滚动更新
在上面，详细分享了使用 StatefulSet 编排“有状态应用”的过程。从中不难看出，StatefulSet 其实就是对现有典型运维业务的容器化抽象。也就是说，你一定有方法在不使用 Kubernetes、甚至不使用容器的情况下，自己 DIY 一个类似的方案出来。但是，一旦涉及到升级、版本管理等更工程化的能力，Kubernetes 的好处，才会更加凸现。

比如，如何对 StatefulSet 进行“滚动更新”（rolling update），很简单。你只要修改 StatefulSet 的 Pod 模板，就会自动触发“滚动更新”，如 
```sh
$ kubectl patch statefulset mysql --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"mysql:5.7.23"}]'
statefulset.apps/mysql patched
```
在这里，我使用了 kubectl patch 命令。它的意思是，以“补丁”的方式（JSON 格式的）修改一个 API 对象的指定字段，也就是我在后面指定的“spec/template/spec/containers/0/image”。

这样，StatefulSet Controller 就会按照与 Pod 编号相反的顺序，从最后一个 Pod 开始，逐一更新这个 StatefulSet 管理的每个 Pod。而如果更新发生了错误，这次“滚动更新”就会停止。**此外，StatefulSet 的“滚动更新”还允许我们进行更精细的控制，比如金丝雀发布（Canary Deploy）或者灰度发布，这意味着应用的多个实例中被指定的一部分不会被更新到最新的版本。这个字段，正是 StatefulSet 的 spec.updateStrategy.rollingUpdate 的 partition 字段。**

比如，现在我将前面这个 StatefulSet 的 partition 字段设置为 2：
```sh
$ kubectl patch statefulset mysql -p '{"spec":{"updateStrategy":{"type":"RollingUpdate","rollingUpdate":{"partition":2}}}}'
statefulset.apps/mysql patched
```
其中，kubectl patch 命令后面的参数（JSON 格式的），就是 partition 字段在 API 对象里的路径。所以，上述操作等同于直接使用 kubectl edit 命令，打开这个对象，把 partition 字段修改为 2。

**这样，我就指定了当 Pod 模板发生变化的时候，比如 MySQL 镜像更新到 5.7.23，那么只有序号大于或者等于 2 的 Pod 会被更新到这个版本。并且，如果你删除或者重启了序号小于 2 的 Pod，等它再次启动后，也会保持原先的 5.7.2 版本，绝不会被升级到 5.7.23 版本。**


**金丝雀发布（Canary Release）和灰度发布（Grayscale Release）都是为了减少新功能或更新在生产环境中引入的风险而采用的策略。金丝雀发布通常是指先在一小部分用户或服务器上推出新功能，如果没有问题，再逐步扩大到更多用户或服务器。这种策略得名于矿工在下矿之前会先放一只金丝雀进去探测是否有毒气，以此类推，金丝雀发布就是先让一小部分用户体验新功能**

**灰度发布则是一种更广泛的策略，它不仅仅局限于逐步推广新功能，还可以包括对系统进行逐步升级或修改。灰度发布可以基于多种策略，如基于用户请求的元信息（如HTTP头或Cookie）将流量路由到新版本，或者通过选择特定服务器进行测试。**

#### 总结
这里以 MySQL 集群为例，详细分享了一个实际的 StatefulSet 的编写过程。在这个过程中，有以下几个关键点（坑）特别值得你注意和体会。
1. “人格分裂”：在解决需求的过程中，一定要记得思考，该 Pod 在扮演不同角色时的不同操作。
2. “阅后即焚”：很多“有状态应用”的节点，只是在第一次启动的时候才需要做额外处理。所以，在编写 YAML 文件时，你一定要考虑“容器重启”的情况，不要让这一次的操作干扰到下一次的容器启动。
3. “容器之间平等无序”：除非是 InitContainer，否则一个 Pod 里的多个容器之间，是完全平等的。所以，你精心设计的 sidecar，绝不能对容器的顺序做出假设，否则就需要进行前置检查。

最后，StatefulSet 其实是一种特殊的 Deployment，只不过这个“Deployment”的每个 Pod 实例的名字里，都携带了一个唯一并且固定的编号。这个编号的顺序，固定了 Pod 的拓扑关系；这个编号对应的 DNS 记录，固定了 Pod 的访问方式；这个编号对应的 PV，绑定了 Pod 与持久化存储的关系。所以，当 Pod 被删除重建时，这些“状态”都会保持不变。

而一旦你的应用没办法通过上述方式进行状态的管理，那就代表了 StatefulSet 已经不能解决它的部署问题了。这时候，Operator 可能才是一个更好的选择，具体参考下面 Operator 部分。


## # 容器化守护进程的意义：DaemonSet
接下来了解下相对轻松的知识点：DaemonSet。

顾名思义，DaemonSet 的主要作用，是让你在 Kubernetes 集群里，运行一个 Daemon Pod。 所以，这个 Pod 有如下三个特征：
1. 这个 Pod 运行在 Kubernetes 集群里的每一个节点（Node）上；
2. 每个节点上只有一个这样的 Pod 实例；
3. 当有新的节点加入 Kubernetes 集群后，该 Pod 会自动地在新节点上被创建出来；而当旧节点被删除后，它上面的 Pod 也相应地会被回收掉。

这个机制听起来很简单，但 Daemon Pod 的意义确实是非常重要的。我随便给你列举几个例子：
1. 各种网络插件的 Agent 组件，都必须运行在每一个节点上，用来处理这个节点上的容器网络；
2. 各种存储插件的 Agent 组件，也必须运行在每一个节点上，用来在这个节点上挂载远程存储目录，操作容器的 Volume 目录；
3. 各种监控组件和日志组件，也必须运行在每一个节点上，负责这个节点上的监控信息和日志搜集。

更重要的是，跟其他编排对象不一样，DaemonSet 开始运行的时机，很多时候比整个 Kubernetes 集群出现的时机都要早。

这个乍一听起来可能有点儿奇怪。但其实你来想一下：如果这个 DaemonSet 正是一个网络插件的 Agent 组件呢？

这个时候，整个 Kubernetes 集群里还没有可用的容器网络，所有 Worker 节点的状态都是 NotReady（NetworkReady=false）。这种情况下，普通的 Pod 肯定不能运行在这个集群上。所以，这也就意味着 DaemonSet 的设计，必须要有某种“过人之处”才行。

为了弄清楚 DaemonSet 的工作原理，我们还是按照老规矩，先从它的 API 对象的定义说起：
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: k8s.gcr.io/fluentd-elasticsearch:1.9.0
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```
这个 DaemonSet，管理的是一个 fluentd-elasticsearch 镜像的 Pod。这个镜像的功能非常实用：通过 fluentd 将 Docker 容器里的日志转发到 ElasticSearch 中。

可以看到，DaemonSet 跟 Deployment 其实非常相似，只不过是没有 replicas 字段；它也使用 selector 选择管理所有携带了 name=fluentd-elasticsearch 标签的 Pod。

而这些 Pod 的模板，也是用 template 字段定义的。在这个字段中，我们定义了一个使用 fluentd-elasticsearch:1.9.0 镜像的容器，而且这个容器挂载了两个 hostPath 类型的 Volume，分别对应宿主机的 /var/log 目录和 /var/lib/docker/containers 目录。

显然，fluentd 启动之后，它会从这两个目录里搜集日志信息，并转发给 ElasticSearch 保存。这样，我们通过 ElasticSearch 就可以很方便地检索这些日志了。

需要注意的是，Docker 容器里应用的日志，默认会保存在宿主机的 /var/lib/docker/containers/{{. 容器 ID}}/{{. 容器 ID}}-json.log 文件里，所以这个目录正是 fluentd 的搜集目标。

那么，DaemonSet 又是如何保证每个 Node 上有且只有一个被管理的 Pod 呢？

显然，这是一个典型的“控制器模型”能够处理的问题。

DaemonSet Controller，首先从 Etcd 里获取所有的 Node 列表，然后遍历所有的 Node。这时，它就可以很容易地去检查，当前这个 Node 上是不是有一个携带了 name=fluentd-elasticsearch 标签的 Pod 在运行。

而检查的结果，可能有这么三种情况：
1. 没有这种 Pod，那么就意味着要在这个 Node 上创建这样一个 Pod；
2. 有这种 Pod，但是数量大于 1，那就说明要把多余的 Pod 从这个 Node 上删除掉；
3. 正好只有一个这种 Pod，那说明这个节点是正常的。
其中，删除节点（Node）上多余的 Pod 非常简单，直接调用 Kubernetes API 就可以了。

但是，如何在指定的 Node 上创建新 Pod 呢？

如果你已经熟悉了 Pod API 对象的话，那一定可以立刻说出答案：用 nodeSelector，选择 Node 的名字即可。
```yaml
nodeSelector:
    name: <Node名字>
```
没错。

不过，在 Kubernetes 项目里，nodeSelector 其实已经是一个将要被废弃的字段了。因为，现在有了一个新的、功能更完善的字段可以代替它，即：nodeAffinity。我来举个例子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: with-node-affinity
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: metadata.name
            operator: In
            values:
            - node-geektime
```
（注：nodeAffinity 相较于 nodeSelector 的优势在于: 1. 语言更具表现力（不仅仅是“对完全匹配规则的 AND”） 2. 你可以发现规则是“软需求”/“偏好”，而不是硬性要求，因此， 如果调度器无法满足该要求，仍然调度该 Pod 3. 你可以使用节点上（或其他拓扑域中）的 Pod 的标签来约束，而不是使用 节点本身的标签，来允许哪些 pod 可以或者不可以被放置在一起。 参考: https://kubernetes.io/zh/docs/concepts/scheduling-eviction/assign-pod-node/#%E4%BA%B2%E5%92%8C%E6%80%A7%E4%B8%8E%E5%8F%8D%E4%BA%B2%E5%92%8C%E6%80%A7）

在这个 Pod 里，声明了一个 spec.affinity 字段，然后定义了一个 nodeAffinity。其中，spec.affinity 字段，是 Pod 里跟调度相关的一个字段。关于它的完整内容，参考下面调度策略相关的部分。

而在这里，我定义的 nodeAffinity 的含义是：
1. requiredDuringSchedulingIgnoredDuringExecution：它的意思是说，这个 nodeAffinity 必须在每次调度的时候予以考虑。同时，这也意味着你可以设置在某些情况下不考虑这个 nodeAffinity；
2. 这个 Pod，将来只允许运行在“metadata.name”是“node-geektime”的节点上。

在这里，你应该注意到 nodeAffinity 的定义，可以支持更加丰富的语法，比如 operator: In（即：部分匹配；如果你定义 operator: Equal，就是完全匹配），这也正是 nodeAffinity 会取代 nodeSelector 的原因之一。

所以，我们的 DaemonSet Controller 会在创建 Pod 的时候，自动在这个 Pod 的 API 对象里，加上这样一个 nodeAffinity 定义。其中，需要绑定的节点名字，正是当前正在遍历的这个 Node。

当然，DaemonSet 并不需要修改用户提交的 YAML 文件里的 Pod 模板，而是在向 Kubernetes 发起请求之前，直接修改根据模板生成的 Pod 对象。这个思路，也正是我在前面讲解 Pod 对象时介绍过的。

此外，DaemonSet 还会给这个 Pod 自动加上另外一个与调度相关的字段，叫作 tolerations。这个字段意味着这个 Pod，会“容忍”（Toleration）某些 Node 的“污点”（Taint）。（自动添加 tolerations，定义容忍的污点，在有此污点的节点上也可以调度。）

而 DaemonSet 自动加上的 tolerations 字段，格式如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: with-toleration
spec:
  tolerations:
  - key: node.kubernetes.io/unschedulable
    operator: Exists
    effect: NoSchedule
```
这个 Toleration 的含义是：“容忍”所有被标记为 unschedulable“污点”的 Node；“容忍”的效果是允许调度。（关于如何给一个 Node 标记上“污点”，以及这里具体的语法定义，参考后面调度器部分，这里，你可以简单地把“污点”理解为一种特殊的 Label。）

而在正常情况下，被标记了 unschedulable“污点”的 Node，是不会有任何 Pod 被调度上去的（effect: NoSchedule）。可是，DaemonSet 自动地给被管理的 Pod 加上了这个特殊的 Toleration，就使得这些 Pod 可以忽略这个限制，继而保证每个节点上都会被调度一个 Pod。当然，如果这个节点有故障的话，这个 Pod 可能会启动失败，而 DaemonSet 则会始终尝试下去，直到 Pod 启动成功。

这时，你应该可以猜到，我在前面介绍到的 DaemonSet 的“过人之处”，其实就是依靠 Toleration 实现的。

假如当前 DaemonSet 管理的，是一个网络插件的 Agent Pod，那么你就必须在这个 DaemonSet 的 YAML 文件里，给它的 Pod 模板加上一个能够“容忍”node.kubernetes.io/network-unavailable“污点”的 Toleration。正如下面这个例子所示：
```yaml
...
template:
    metadata:
      labels:
        name: network-plugin-agent
    spec:
      tolerations:
      - key: node.kubernetes.io/network-unavailable
        operator: Exists
        effect: NoSchedule
```
在 Kubernetes 项目中，当一个节点的网络插件尚未安装时，这个节点就会被自动加上名为node.kubernetes.io/network-unavailable的“污点”。

而通过这样一个 Toleration，调度器在调度这个 Pod 的时候，就会忽略当前节点上的“污点”，从而成功地将网络插件的 Agent 组件调度到这台机器上启动起来。

这种机制，正是我们在部署 Kubernetes 集群的时候，能够先部署 Kubernetes 本身、再部署网络插件的根本原因：因为当时我们所创建的 Weave 的 YAML，实际上就是一个 DaemonSet。（k8s启动后，节点网络插件未安装，自动打上网络不可用的污点。daemon set通过设置，可以忽略污点，将作为网络插间agent 的pod部署上去）

至此，通过上面这些内容，你应该能够明白，DaemonSet 其实是一个非常简单的控制器。在它的控制循环中，只需要遍历所有节点，然后根据节点上是否有被管理 Pod 的情况，来决定是否要创建或者删除一个 Pod。

只不过，在创建每个 Pod 的时候，DaemonSet 会自动给这个 Pod 加上一个 nodeAffinity，从而保证这个 Pod 只会在指定节点上启动。同时，它还会自动给这个 Pod 加上一个 Toleration，从而忽略节点的 unschedulable“污点”。

当然，你也可以在 Pod 模板里加上更多种类的 Toleration，从而利用 DaemonSet 达到自己的目的。比如，在这个 fluentd-elasticsearch DaemonSet 里，我就给它加上了这样的 Toleration：
```yaml
tolerations:
- key: node-role.kubernetes.io/master
  effect: NoSchedule
```
这是因为在默认情况下，**Kubernetes 集群不允许用户在 Master 节点部署 Pod。因为，Master 节点默认携带了一个叫作node-role.kubernetes.io/master的“污点”。所以，为了能在 Master 节点上部署 DaemonSet 的 Pod，我就必须让这个 Pod“容忍”这个“污点”。**

在理解了 DaemonSet 的工作原理之后，接下来我就通过一个具体的实践来帮你更深入地掌握 DaemonSet 的使用方法。（需要注意的是，在 Kubernetes v1.11 之前，由于调度器尚不完善，DaemonSet 是由 DaemonSet Controller 自行调度的，即它会直接设置 Pod 的 spec.nodename 字段，这样就可以跳过调度器了。但是，这样的做法很快就会被废除，所以在这里我也不推荐你再花时间学习这个流程了。）

首先，创建这个 DaemonSet 对象：
```sh
kubectl create -f fluentd-elasticsearch.yaml
```
**需要注意的是，在 DaemonSet 上，我们一般都应该加上 resources 字段，来限制它的 CPU 和内存使用，防止它占用过多的宿主机资源。（DaemonSet通常是资源监控，经常会出现用死循环等来处理任务，限制资源使用度的需求比较迫切，所以这里专门针对deamonset提出。）**

而创建成功后，你就能看到，如果有 N 个节点，就会有 N 个 fluentd-elasticsearch Pod 在运行。比如在我们的例子里，会有两个 Pod，如下所示：
```sh
$ kubectl get pod -n kube-system -l name=fluentd-elasticsearch
NAME                          READY     STATUS    RESTARTS   AGE
fluentd-elasticsearch-dqfv9   1/1       Running   0          53m
fluentd-elasticsearch-pf9z5   1/1       Running   0          53m
```

而如果你此时通过 kubectl get 查看一下 Kubernetes 集群里的 DaemonSet 对象：（**Kubernetes 里比较长的 API 对象都有短名字，比如 DaemonSet 对应的是 ds，Deployment 对应的是 deploy。**）
```sh
$ kubectl get ds -n kube-system fluentd-elasticsearch
NAME                    DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
fluentd-elasticsearch   2         2         2         2            2           <none>          1h
```
就会发现 DaemonSet 和 Deployment 一样，也有 DESIRED、CURRENT 等多个状态字段。这也就意味着，DaemonSet 可以像 Deployment 那样，进行版本管理。这个版本，可以使用 kubectl rollout history 看到：
```sh
$ kubectl rollout history daemonset fluentd-elasticsearch -n kube-system
daemonsets "fluentd-elasticsearch"
REVISION  CHANGE-CAUSE
1         <none>
```
接下来，我们来把这个 DaemonSet 的容器镜像版本到 v2.2.0：
```sh
$ kubectl set image ds/fluentd-elasticsearch fluentd-elasticsearch=k8s.gcr.io/fluentd-elasticsearch:v2.2.0 --record -n=kube-system
```
这个 kubectl set image 命令里，第一个 fluentd-elasticsearch 是 DaemonSet 的名字，第二个 fluentd-elasticsearch 是容器的名字。

这时候，我们可以使用 kubectl rollout status 命令看到这个“滚动更新”的过程，如下所示：
```sh
$ kubectl rollout status ds/fluentd-elasticsearch -n kube-system
Waiting for daemon set "fluentd-elasticsearch" rollout to finish: 0 out of 2 new pods have been updated...
Waiting for daemon set "fluentd-elasticsearch" rollout to finish: 0 out of 2 new pods have been updated...
Waiting for daemon set "fluentd-elasticsearch" rollout to finish: 1 of 2 updated pods are available...
daemon set "fluentd-elasticsearch" successfully rolled out
```
注意，由于这一次我在升级命令后面加上了–record 参数，所以这次升级使用到的指令就会自动出现在 DaemonSet 的 rollout history 里面，如下所示：
```sh
$ kubectl rollout history daemonset fluentd-elasticsearch -n kube-system
daemonsets "fluentd-elasticsearch"
REVISION  CHANGE-CAUSE
1         <none>
2         kubectl set image ds/fluentd-elasticsearch fluentd-elasticsearch=k8s.gcr.io/fluentd-elasticsearch:v2.2.0 --namespace=kube-system --record=true
```
有了版本号，你也就可以像 Deployment 一样，将 DaemonSet 回滚到某个指定的历史版本了。

在前面 Deployment 对象部分，曾经提到过，Deployment 管理这些版本，靠的是“一个版本对应一个 ReplicaSet 对象”。可是，DaemonSet 控制器操作的直接就是 Pod，不可能有 ReplicaSet 这样的对象参与其中。那么，它的这些版本又是如何维护的呢？

**所谓，一切皆对象！**

在 Kubernetes 项目中，任何你觉得需要记录下来的状态，都可以被用 API 对象的方式实现。当然，“版本”也不例外。

Kubernetes v1.7 之后添加了一个 API 对象，名叫 ControllerRevision，专门用来记录某种 Controller 对象的版本。比如，你可以通过如下命令查看 fluentd-elasticsearch 对应的 ControllerRevision：
```sh
$ kubectl get controllerrevision -n kube-system -l name=fluentd-elasticsearch
NAME                               CONTROLLER                             REVISION   AGE
fluentd-elasticsearch-64dc6799c9   daemonset.apps/fluentd-elasticsearch   2          1h
```

而如果你使用 kubectl describe 查看这个 ControllerRevision 对象：
```sh
$ kubectl describe controllerrevision fluentd-elasticsearch-64dc6799c9 -n kube-system
Name:         fluentd-elasticsearch-64dc6799c9
Namespace:    kube-system
Labels:       controller-revision-hash=2087235575
              name=fluentd-elasticsearch
Annotations:  deprecated.daemonset.template.generation=2
              kubernetes.io/change-cause=kubectl set image ds/fluentd-elasticsearch fluentd-elasticsearch=k8s.gcr.io/fluentd-elasticsearch:v2.2.0 --record=true --namespace=kube-system
API Version:  apps/v1
Data:
  Spec:
    Template:
      $ Patch:  replace
      Metadata:
        Creation Timestamp:  <nil>
        Labels:
          Name:  fluentd-elasticsearch
      Spec:
        Containers:
          Image:              k8s.gcr.io/fluentd-elasticsearch:v2.2.0
          Image Pull Policy:  IfNotPresent
          Name:               fluentd-elasticsearch
...
Revision:                  2
Events:                    <none>
```

就会看到，这个 ControllerRevision 对象，实际上是在 Data 字段保存了该版本对应的完整的 DaemonSet 的 API 对象。并且，在 Annotation 字段保存了创建这个对象所使用的 kubectl 命令。

接下来，我们可以尝试将这个 DaemonSet 回滚到 Revision=1 时的状态：
```sh
$ kubectl rollout undo daemonset fluentd-elasticsearch --to-revision=1 -n kube-system
daemonset.extensions/fluentd-elasticsearch rolled back
```
这个 kubectl rollout undo 操作，实际上相当于读取到了 Revision=1 的 ControllerRevision 对象保存的 Data 字段。而这个 Data 字段里保存的信息，就是 Revision=1 时这个 DaemonSet 的完整 API 对象。

所以，现在 DaemonSet Controller 就可以使用这个历史 API 对象，对现有的 DaemonSet 做一次 PATCH 操作（等价于执行一次 kubectl apply -f “旧的 DaemonSet 对象”），从而把这个 DaemonSet“更新”到一个旧版本。

这也是为什么，在执行完这次回滚完成后，你会发现，DaemonSet 的 Revision 并不会从 Revision=2 退回到 1，而是会增加成 Revision=3。这是因为，一个新的 ControllerRevision 被创建了出来。

**相比于 Deployment，DaemonSet 只管理 Pod 对象**


## # 撬动离线业务：Job与CronJob

上面分享了 Deployment、StatefulSet，以及 DaemonSet 这三个编排概念。有没有发现它们的共同之处呢？

实际上，它们主要编排的对象，都是“在线业务”，即：Long Running Task（长作业）。比如，在前面举例时常用的 Nginx、Tomcat，以及 MySQL 等等。这些应用一旦运行起来，除非出错或者停止，它的容器进程会一直保持在 Running 状态。

但是，有一类作业显然不满足这样的条件，这就是“离线业务”，或者叫作 Batch Job（计算业务）。这种业务在计算完成后就直接退出了，而此时如果你依然用 Deployment 来管理这种业务的话，就会发现 Pod 会在计算结束后退出，然后被 Deployment Controller 不断地重启；而像“滚动更新”这样的编排功能，更无从谈起了。

所以，早在 Borg 项目中，Google 就已经对作业进行了分类处理，提出了 LRS（Long Running Service）和 Batch Jobs 两种作业形态，对它们进行“分别管理”和“混合调度”。

不过，在 2015 年 Borg 论文刚刚发布的时候，Kubernetes 项目并不支持对 Batch Job 的管理。直到 v1.4 版本之后，社区才逐步设计出了一个用来描述离线业务的 API 对象，它的名字就是：Job。

Job API 对象的定义非常简单，我来举个例子，如下所示：
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: resouer/ubuntu-bc 
        command: ["sh", "-c", "echo 'scale=10000; 4*a(1)' | bc -l "]
      restartPolicy: Never
  backoffLimit: 4
```
此时，相信你对 Kubernetes 的 API 对象已经不再陌生了。在这个 Job 的 YAML 文件里，你肯定一眼就会看到一位“老熟人”：Pod 模板，即 spec.template 字段。

在这个 Pod 模板中，我定义了一个 Ubuntu 镜像的容器（准确地说，是一个安装了 bc 命令的 Ubuntu 镜像），它运行的程序是：
```sh
echo "scale=10000; 4*a(1)" | bc -l 
```
其中，bc 命令是 Linux 里的“计算器”；-l 表示，我现在要使用标准数学库；而 a(1)，则是调用数学库中的 arctangent 函数，计算 atan(1)。这是什么意思呢？

中学知识告诉我们：tan(π/4) = 1。所以，4*atan(1)正好就是π，也就是 3.1415926…。（如果你不熟悉这个知识也不必担心，我也是在查阅资料后才知道的。kkkk）

所以，这其实就是一个计算π值的容器。而通过 scale=10000，我指定了输出的小数点后的位数是 10000。在我的计算机上，这个计算大概用时 1 分 54 秒。

但是，跟其他控制器不同的是，Job 对象并不要求你定义一个 spec.selector 来描述要控制哪些 Pod。具体原因，我马上会讲解到。

现在，我们就可以创建这个 Job 了：
```sh
kubectl create -f job.yaml
```
在成功创建后，我们来查看一下这个 Job 对象，如下所示：
```sh
$ kubectl describe jobs/pi
Name:             pi
Namespace:        default
Selector:         controller-uid=c2db599a-2c9d-11e6-b324-0209dc45a495
Labels:           controller-uid=c2db599a-2c9d-11e6-b324-0209dc45a495
                  job-name=pi
Annotations:      <none>
Parallelism:      1
Completions:      1
..
Pods Statuses:    0 Running / 1 Succeeded / 0 Failed
Pod Template:
  Labels:       controller-uid=c2db599a-2c9d-11e6-b324-0209dc45a495
                job-name=pi
  Containers:
   ...
  Volumes:              <none>
Events:
  FirstSeen    LastSeen    Count    From            SubobjectPath    Type        Reason            Message
  ---------    --------    -----    ----            -------------    --------    ------            -------
  1m           1m          1        {job-controller }                Normal      SuccessfulCreate  Created pod: pi-rq5rl
```
可以看到，这个 Job 对象在创建后，它的 Pod 模板，被自动加上了一个 controller-uid=< 一个随机字符串 > 这样的 Label。而这个 Job 对象本身，则被自动加上了这个 Label 对应的 Selector，从而 保证了 Job 与它所管理的 Pod 之间的匹配关系。

而 Job Controller 之所以要使用这种携带了 UID 的 Label，就是为了避免不同 Job 对象所管理的 Pod 发生重合。需要注意的是，这种自动生成的 Label 对用户来说并不友好，所以不太适合推广到 Deployment 等长作业编排对象上。

接下来，我们可以看到这个 Job 创建的 Pod 进入了 Running 状态，这意味着它正在计算 Pi 的值。
```sh
$ kubectl get pods
NAME                                READY     STATUS    RESTARTS   AGE
pi-rq5rl                            1/1       Running   0          10s
```

而几分钟后计算结束，这个 Pod 就会进入 Completed 状态：
```sh
$ kubectl get pods
NAME                                READY     STATUS      RESTARTS   AGE
pi-rq5rl                            0/1       Completed   0          4m
```
这也是我们需要在 Pod 模板中定义 restartPolicy=Never 的原因：离线计算的 Pod 永远都不应该被重启，否则它们会再重新计算一遍。（**事实上，restartPolicy 在 Job 对象里只允许被设置为 Never 和 OnFailure；而在 Deployment 对象里，restartPolicy 则只允许被设置为 Always。**）

此时，我们通过 kubectl logs 查看一下这个 Pod 的日志，就可以看到计算得到的 Pi 值已经被打印了出来：
```sh
$ kubectl logs pi-rq5rl
3.141592653589793238462643383279...
```
这时候，你一定会想到这样一个问题，如果这个离线作业失败了要怎么办？

比如，我们在这个例子中定义了 restartPolicy=Never，那么离线作业失败后 Job Controller 就会不断地尝试创建一个新 Pod，如下所示：
```sh
$ kubectl get pods
NAME                                READY     STATUS              RESTARTS   AGE
pi-55h89                            0/1       ContainerCreating   0          2s
pi-tqbcz                            0/1       Error               0          5s
```
可以看到，这时候会不断地有新 Pod 被创建出来。

当然，这个尝试肯定不能无限进行下去。所以，我们就在 Job 对象的 spec.backoffLimit 字段里定义了重试次数为 4（即，backoffLimit=4），而这个字段的默认值是 6。

**需要注意的是，Job Controller 重新创建 Pod 的间隔是呈指数增加的，即下一次重新创建 Pod 的动作会分别发生在 10 s、20 s、40 s …后。**

而如果你定义的 restartPolicy=OnFailure，那么离线作业失败后，Job Controller 就不会去尝试创建新的 Pod。但是，它会不断地尝试重启 Pod 里的容器。这也正好对应了 restartPolicy 的含义。

如前所述，当一个 Job 的 Pod 运行结束后，它会进入 Completed 状态。但是，如果这个 Pod 因为某种原因一直不肯结束呢？

在 Job 的 API 对象里，有一个 spec.activeDeadlineSeconds 字段可以设置最长运行时间，比如：
```yaml
spec:
 backoffLimit: 5
 activeDeadlineSeconds: 100
```
一旦运行超过了 100 s，这个 Job 的所有 Pod 都会被终止。并且，你可以在 Pod 的状态里看到终止的原因是 reason: DeadlineExceeded。

以上，就是一个 Job API 对象最主要的概念和用法了。不过，离线业务之所以被称为 Batch Job，当然是因为它们可以以“Batch”，也就是并行的方式去运行。

接下来，就了解一下 Job Controller 对并行作业的控制方法。

在 Job 对象中，负责并行控制的参数有两个：
1. spec.parallelism，它定义的是一个 Job 在任意时间最多可以启动多少个 Pod 同时运行；
2. spec.completions，它定义的是 Job 至少要完成的 Pod 数目，即 Job 的最小完成数。

这两个参数听起来有点儿抽象，所以我准备了一个例子来帮助你理解。现在，我在之前计算 Pi 值的 Job 里，添加这两个参数：
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  parallelism: 2
  completions: 4
  template:
    spec:
      containers:
      - name: pi
        image: resouer/ubuntu-bc
        command: ["sh", "-c", "echo 'scale=5000; 4*a(1)' | bc -l "]
      restartPolicy: Never
  backoffLimit: 4
```
这样，我们就指定了这个 Job 最大的并行数是 2，而最小的完成数是 4。

接下来，我们来创建这个 Job 对象：
```sh
kubectl create -f job.yaml
```
可以看到，这个 Job 其实也维护了两个状态字段，即 DESIRED 和 SUCCESSFUL，如下所示：
```sh
$ kubectl get job
NAME      DESIRED   SUCCESSFUL   AGE
pi        4         0            3s
```
其中，DESIRED 的值，正是 completions 定义的最小完成数。

然后，我们可以看到，这个 Job 首先创建了两个并行运行的 Pod 来计算 Pi：
```sh
$ kubectl get pods
NAME       READY     STATUS    RESTARTS   AGE
pi-5mt88   1/1       Running   0          6s
pi-gmcq5   1/1       Running   0          6s
```
而在 40 s 后，这两个 Pod 相继完成计算。

这时我们可以看到，每当有一个 Pod 完成计算进入 Completed 状态时，就会有一个新的 Pod 被自动创建出来，并且快速地从 Pending 状态进入到 ContainerCreating 状态：
```sh
$ kubectl get pods
NAME       READY     STATUS    RESTARTS   AGE
pi-gmcq5   0/1       Completed   0         40s
pi-84ww8   0/1       Pending   0         0s
pi-5mt88   0/1       Completed   0         41s
pi-62rbt   0/1       Pending   0         0s

$ kubectl get pods
NAME       READY     STATUS    RESTARTS   AGE
pi-gmcq5   0/1       Completed   0         40s
pi-84ww8   0/1       ContainerCreating   0         0s
pi-5mt88   0/1       Completed   0         41s
pi-62rbt   0/1       ContainerCreating   0         0s
```
紧接着，Job Controller 第二次创建出来的两个并行的 Pod 也进入了 Running 状态：
```sh
$ kubectl get pods 
NAME       READY     STATUS      RESTARTS   AGE
pi-5mt88   0/1       Completed   0          54s
pi-62rbt   1/1       Running     0          13s
pi-84ww8   1/1       Running     0          14s
pi-gmcq5   0/1       Completed   0          54s
```
最终，后面创建的这两个 Pod 也完成了计算，进入了 Completed 状态。

这时，由于所有的 Pod 均已经成功退出，这个 Job 也就执行完了，所以你会看到它的 SUCCESSFUL 字段的值变成了 4：
```sh
$ kubectl get pods 
NAME       READY     STATUS      RESTARTS   AGE
pi-5mt88   0/1       Completed   0          5m
pi-62rbt   0/1       Completed   0          4m
pi-84ww8   0/1       Completed   0          4m
pi-gmcq5   0/1       Completed   0          5m

$ kubectl get job
NAME      DESIRED   SUCCESSFUL   AGE
pi        4         4            5m
```

通过上述 Job 的 DESIRED 和 SUCCESSFUL 字段的关系，**我们就可以很容易地理解Job Controller 的工作原理了。**

**首先，Job Controller 控制的对象，直接就是 Pod。**

**其次，Job Controller 在控制循环中进行的调谐（Reconcile）操作，是根据实际在 Running 状态 Pod 的数目、已经成功退出的 Pod 的数目，以及 parallelism、completions 参数的值共同计算出在这个周期里，应该创建或者删除的 Pod 数目，然后调用 Kubernetes API 来执行这个操作。**

以创建 Pod 为例。在上面计算 Pi 值的这个例子中，当 Job 一开始创建出来时，实际处于 Running 状态的 Pod 数目 =0，已经成功退出的 Pod 数目 =0，而用户定义的 completions，也就是最终用户需要的 Pod 数目 =4。

所以，在这个时刻，需要创建的 Pod 数目 = 最终需要的 Pod 数目 - 实际在 Running 状态 Pod 数目 - 已经成功退出的 Pod 数目 = 4 - 0 - 0= 4。也就是说，Job Controller 需要创建 4 个 Pod 来纠正这个不一致状态。

可是，我们又定义了这个 Job 的 parallelism=2。也就是说，我们规定了每次并发创建的 Pod 个数不能超过 2 个。所以，Job Controller 会对前面的计算结果做一个修正，修正后的期望创建的 Pod 数目应该是：2 个。

这时候，Job Controller 就会并发地向 kube-apiserver 发起两个创建 Pod 的请求。

类似地，如果在这次调谐周期里，Job Controller 发现实际在 Running 状态的 Pod 数目，比 parallelism 还大，那么它就会删除一些 Pod，使两者相等。

综上所述，Job Controller 实际上控制了，作业执行的并行度，以及总共需要完成的任务数这两个重要参数。而在实际使用时，你需要根据作业的特性，来决定并行度（parallelism）和任务数（completions）的合理取值。

接下来，我再和你分享三种常用的、使用 Job 对象的方法。

#### 第一种用法，也是最简单粗暴的用法：外部管理器 +Job 模板。
这种模式的特定用法是：把 Job 的 YAML 文件定义为一个“模板”，然后用一个外部工具控制这些“模板”来生成 Job。这时，Job 的定义方式如下所示：
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: process-item-$ITEM
  labels:
    jobgroup: jobexample
spec:
  template:
    metadata:
      name: jobexample
      labels:
        jobgroup: jobexample
    spec:
      containers:
      - name: c
        image: busybox
        command: ["sh", "-c", "echo Processing item $ITEM && sleep 5"]
      restartPolicy: Never
```

可以看到，我们在这个 Job 的 YAML 里，定义了 $ITEM 这样的“变量”。

所以，在控制这种 Job 时，我们只要注意如下两个方面即可：
1. 创建 Job 时，替换掉 $ITEM 这样的变量；
2. 所有来自于同一个模板的 Job，都有一个 jobgroup: jobexample 标签，也就是说这一组 Job 使用这样一个相同的标识。

而做到第一点非常简单。比如，你可以通过这样一句 shell 把 $ITEM 替换掉：
```sh
$ mkdir ./jobs
$ for i in apple banana cherry
do
  cat job-tmpl.yaml | sed "s/\$ITEM/$i/" > ./jobs/job-$i.yaml
done
```
这样，一组来自于同一个模板的不同 Job 的 yaml 就生成了。接下来，你就可以通过一句 kubectl create 指令创建这些 Job 了：
```sh
$ kubectl create -f ./jobs
$ kubectl get pods -l jobgroup=jobexample
NAME                        READY     STATUS      RESTARTS   AGE
process-item-apple-kixwv    0/1       Completed   0          4m
process-item-banana-wrsf7   0/1       Completed   0          4m
process-item-cherry-dnfu9   0/1       Completed   0          4m
```
这个模式看起来虽然很“傻”，但却是 Kubernetes 社区里使用 Job 的一个很普遍的模式。

原因很简单：大多数用户在需要管理 Batch Job 的时候，都已经有了一套自己的方案，需要做的往往就是集成工作。这时候，Kubernetes 项目对这些方案来说最有价值的，就是 Job 这个 API 对象。所以，你只需要编写一个外部工具（等同于我们这里的 for 循环）来管理这些 Job 即可。

**这种模式最典型的应用，就是 TensorFlow 社区的 KubeFlow 项目。（k8s上构建的机器学习组件库）**

很容易理解，在这种模式下使用 Job 对象，completions 和 parallelism 这两个字段都应该使用默认值 1，而不应该由我们自行设置。而作业 Pod 的并行控制，应该完全交由外部工具来进行管理（比如，KubeFlow）。

#### 第二种用法：拥有固定任务数目的并行 Job。
这种模式下，我只关心最后是否有指定数目（spec.completions）个任务成功退出。至于执行时的并行度是多少，我并不关心。

比如，我们这个计算 Pi 值的例子，就是这样一个典型的、拥有固定任务数目（completions=4）的应用场景。 它的 parallelism 值是 2；或者，你可以干脆不指定 parallelism，直接使用默认的并行度（即：1）。

此外，你还可以使用一个工作队列（Work Queue）进行任务分发。这时，Job 的 YAML 文件定义如下所示：
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: job-wq-1
spec:
  completions: 8
  parallelism: 2
  template:
    metadata:
      name: job-wq-1
    spec:
      containers:
      - name: c
        image: myrepo/job-wq-1
        env:
        - name: BROKER_URL
          value: amqp://guest:guest@rabbitmq-service:5672
        - name: QUEUE
          value: job1
      restartPolicy: OnFailure
```
我们可以看到，它的 completions 的值是：8，这意味着我们总共要处理的任务数目是 8 个。也就是说，总共会有 8 个任务会被逐一放入工作队列里（你可以运行一个外部小程序作为生产者，来提交任务）。

在这个实例中，我选择充当工作队列的是一个运行在 Kubernetes 里的 RabbitMQ。所以，我们需要在 Pod 模板里定义 BROKER_URL，来作为消费者。

所以，一旦你用 kubectl create 创建了这个 Job，它就会以并发度为 2 的方式，每两个 Pod 一组，创建出 8 个 Pod。每个 Pod 都会去连接 BROKER_URL，从 RabbitMQ 里读取任务，然后各自进行处理。这个 Pod 里的执行逻辑，我们可以用这样一段伪代码来表示：
```go
/* job-wq-1的伪代码 */
queue := newQueue($BROKER_URL, $QUEUE)
task := queue.Pop()
process(task)
// exit
```
可以看到，每个 Pod 只需要将任务信息读取出来，处理完成，然后退出即可。而作为用户，我只关心最终一共有 8 个计算任务启动并且退出，只要这个目标达到，我就认为整个 Job 处理完成了。所以说，这种用法，对应的就是“任务总数固定”的场景。

#### 第三种用法，也是很常用的一个用法：指定并行度（parallelism），但不设置固定的 completions 的值。
此时，你就必须自己想办法，来决定什么时候启动新 Pod，什么时候 Job 才算执行完成。在这种情况下，任务的总数是未知的，所以你不仅需要一个工作队列来负责任务分发，还需要能够判断工作队列已经为空（即：所有的工作已经结束了）。

这时候，Job 的定义基本上没变化，只不过是不再需要定义 completions 的值了而已：
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: job-wq-2
spec:
  parallelism: 2
  template:
    metadata:
      name: job-wq-2
    spec:
      containers:
      - name: c
        image: gcr.io/myproject/job-wq-2
        env:
        - name: BROKER_URL
          value: amqp://guest:guest@rabbitmq-service:5672
        - name: QUEUE
          value: job2
      restartPolicy: OnFailure
```
而对应的 Pod 的逻辑会稍微复杂一些，我可以用这样一段伪代码来描述：
```go
/* job-wq-2的伪代码 */
for !queue.IsEmpty($BROKER_URL, $QUEUE) {
  task := queue.Pop()
  process(task)
}
print("Queue empty, exiting")
// exit
```
由于任务数目的总数不固定，所以每一个 Pod 必须能够知道，自己什么时候可以退出。比如，在这个例子中，我简单地以“队列为空”，作为任务全部完成的标志。所以说，这种用法，对应的是“任务总数不固定”的场景。

不过，在实际的应用中，你需要处理的条件往往会非常复杂。比如，任务完成后的输出、每个任务 Pod 之间是不是有资源的竞争和协同等等。

所以，这里就不再展开 Job 的用法了。因为，在实际场景里，要么干脆就用第一种用法来自己管理作业；要么，这些任务 Pod 之间的关系就不那么“单纯”，甚至还是“有状态应用”（比如，任务的输入 / 输出是在持久化数据卷里）。在这种情况下，在后面要重点了解的 Operator，加上 Job 对象一起，可能才能更好地满足实际离线任务的编排需求。

#### 最后，再来分享一个非常有用的 Job 对象，叫作：CronJob
顾名思义，CronJob 描述的，正是定时任务。它的 API 对象，如下所示：
```yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
```
**在这个 YAML 文件中，最重要的关键词就是 jobTemplate。看到它，你一定恍然大悟，原来 CronJob 是一个 Job 对象的控制器（Controller）！**

**没错，CronJob 与 Job 的关系，正如同 Deployment 与 ReplicaSet 的关系一样。CronJob 是一个专门用来管理 Job 对象的控制器。只不过，它创建和删除 Job 的依据，是 schedule 字段定义的、一个标准的Unix Cron格式的表达式。**

比如，"*/1 * * * *"。

**这个 Cron 表达式里 */1 中的 * 表示从 0 开始，/ 表示“每”，1 表示偏移量。所以，它的意思就是：从 0 开始，每 1 个时间单位执行一次。**

那么，时间单位又是什么呢？

Cron 表达式中的五个部分分别代表：分钟、小时、日、月、星期。

所以，上面这句 Cron 表达式的意思是：从当前开始，每分钟执行一次。

而这里要执行的内容，就是 jobTemplate 定义的 Job 了。

所以，这个 CronJob 对象在创建 1 分钟后，就会有一个 Job 产生了，如下所示：
```sh
$ kubectl create -f ./cronjob.yaml
cronjob "hello" created

# 一分钟后
$ kubectl get jobs
NAME               DESIRED   SUCCESSFUL   AGE
hello-4111706356   1         1         2s
```
此时，CronJob 对象会记录下这次 Job 执行的时间：
```sh
$ kubectl get cronjob hello
NAME      SCHEDULE      SUSPEND   ACTIVE    LAST-SCHEDULE
hello     */1 * * * *   False     0         Thu, 6 Sep 2018 14:34:00 -070
```
需要注意的是，由于定时任务的特殊性，很可能某个 Job 还没有执行完，另外一个新 Job 就产生了。这时候，你可以通过 spec.concurrencyPolicy 字段来定义具体的处理策略。比如：
1. concurrencyPolicy=Allow，这也是默认情况，这意味着这些 Job 可以同时存在；
2. concurrencyPolicy=Forbid，这意味着不会创建新的 Pod，该创建周期被跳过；
3. concurrencyPolicy=Replace，这意味着新产生的 Job 会替换旧的、没有执行完的 Job。

而如果某一次 Job 创建失败，这次创建就会被标记为“miss”。当在指定的时间窗口内，miss 的数目达到 100 时，那么 CronJob 会停止再创建这个 Job。

这个时间窗口，可以由 spec.startingDeadlineSeconds 字段指定。比如 startingDeadlineSeconds=200，意味着在过去 200 s 里，如果 miss 的数目达到了 100 次，那么这个 Job 就不会被创建执行了。


## # 声明式API与Kubernetes编程范式
在使用 kubectl create 创建api对象，再使用 kubectl set image 和 kubectl edit 命令更新api对象，或者 kubectl replace -f 的操作，我们称为命令式配置文件操作。
也就是说，它的处理方式，其实跟前面 Docker Swarm 的两句命令，没什么本质上的区别。只不过，它是把 Docker 命令行里的参数，写在了配置文件里而已。

那么，到底什么才是“声明式 API”呢？

答案是，kubectl apply 命令。

现在，我就使用 kubectl apply 命令来创建这个 Deployment：
```sh
$ kubectl apply -f nginx.yaml
```
这样，Nginx 的 Deployment 就被创建了出来，这看起来跟 kubectl create 的效果一样。

然后，我再修改一下 nginx.yaml 里定义的镜像，这时候，关键来了。

在修改完这个 YAML 文件之后，我不再使用 kubectl replace 命令进行更新，而是继续执行一条 kubectl apply 命令，即：
```sh
$ kubectl apply -f nginx.yaml
```
这时，Kubernetes 就会立即触发这个 Deployment 的“滚动更新”。

可是，它跟 kubectl replace 命令有什么本质区别吗？

**实际上，你可以简单地理解为，kubectl replace 的执行过程，是使用新的 YAML 文件中的 API 对象，替换原有的 API 对象；而 kubectl apply，则是执行了一个对原有 API 对象的 PATCH 操作。（类似地，kubectl set image 和 kubectl edit 也是对已有 API 对象的修改。）**

更进一步地，这意味着 kube-apiserver 在响应命令式请求（比如，kubectl replace）的时候，一次只能处理一个写请求，否则会有产生冲突的可能。而对于声明式请求（比如，kubectl apply），一次能处理多个写操作，并且具备 Merge 能力。

**kubectl create、kubectl replace命令行提交后会用本地YAML文件定义的API对象替换原有API对象，这个替换过程就是一条HTTP事务，API对象被替换后，下一条HTTP事务处理的就是新的对象，对象替换过程触发数据库写锁，所以一次只能处理一个请求。而kubectl apply则是以PATCH补丁方式在线修改API对象的多个元素，因此能并行处理多路请求，只要不同时修改同一对象的同一元素就不会产生写冲突。**

这种区别，可能乍一听起来没那么重要。而且，正是由于要照顾到这样的 API 设计，做同样一件事情，Kubernetes 需要的步骤往往要比其他项目多不少。

但是，如果你仔细思考一下 Kubernetes 项目的工作流程，就不难体会到这种声明式 API 的独到之处。

#### istio 例子

（Istio 效果：在用户提交的yaml文件中加入实现定义好的 Envoy 容器，以接管 Pod 进出流量 Istio 过程： 1.将这个 Envoy 容器本身的定义，以 ConfigMap 的方式保存在 Kubernetes 当中； 2.编写“自定义控制器”（Custom Controller）| Initializer； 3.将编写好的Initializer作为 Pod 部署； 4.编写 InitializerConfiguration ，指定 Initializer，并且在新Pod创建时添加 metadata.initializers.pending 标志； 5.Initializer 通过 Kubernetes 的“控制循环”机制，遍历获取新创建的符合筛选条件（metadata.initializers.pending）的 Pod，检查期望状态（Pod里添加了Envoy容器的定义）与实际状态 ，修改该Pod的API对象，然后清除metadata.initializers.pending 标志。 Initializer 过程： 1.从APIServer中拿到ConfigMap，将其中containers与volumes字段添加进空Pod对象； 2.调用 Kubernetes API库 strategicpatch.CreateTwoWayMergePatch(pod, newPod)，生成新旧Pod对象的TwoWayMergePatch ； 3.调用 Kubernetes 的 Client，以TwoWayMergePatch 为参数发起一个 PATCH 请求，用户提交的Pod里添加了ConfigMap中的Envoy容器相关字段。）

在 2017 年 5 月，Google、IBM 和 Lyft 公司，共同宣布了 Istio 开源项目的诞生。很快，这个项目就在技术圈儿里，掀起了一阵名叫“微服务”的热潮，把 Service Mesh 这个新的编排概念推到了风口浪尖。而 Istio 项目，实际上就是一个基于 Kubernetes 项目的微服务治理框架。主要由 Control Plane 控制层和运行在每一个应用 Pod 里的 Envoy 容器组成。

Istio 最根本的组件，是运行在每一个应用 Pod 里的 Envoy 容器。这个 Envoy 项目是 Lyft 公司推出的一个高性能 C++ 网络代理，也是 Lyft 公司对 Istio 项目的唯一贡献。

而 Istio 项目，则把这个代理服务以 sidecar 容器的方式，运行在了每一个被治理的应用 Pod 中。我们知道，Pod 里的所有容器都共享同一个 Network Namespace。所以，Envoy 容器就能够通过配置 Pod 里的 iptables 规则，把整个 Pod 的进出流量接管下来。

这时候，Istio 的控制层（Control Plane）里的 Pilot 组件，就能够通过调用每个 Envoy 容器的 API，对这个 Envoy 代理进行配置，从而实现微服务治理。

来看一个例子。

假设k8s环境中 PodA 是已经在运行的应用，而 PodB 则是我们刚刚上线的应用的新版本。这时候，Pilot 通过调节这两 Pod 里的 Envoy 容器的配置，从而将 90% 的流量分配给旧版本的应用，将 10% 的流量分配给新版本应用，并且，还可以在后续的过程中随时调整。这样，一个典型的“灰度发布”的场景就完成了。比如，Istio 可以调节这个流量从 90%-10%，改到 80%-20%，再到 50%-50%，最后到 0%-100%，就完成了这个灰度发布的过程。

更重要的是，在整个微服务治理的过程中，无论是对 Envoy 容器的部署，还是像上面这样对 Envoy 代理的配置，用户和应用都是完全“无感”的。

这时候，你可能会有所疑惑：Istio 项目明明需要在每个 Pod 里安装一个 Envoy 容器，又怎么能做到“无感”的呢？

实际上，Istio 项目使用的，是 Kubernetes 中的一个非常重要的功能，叫作 Dynamic Admission Control。

在 Kubernetes 项目中，当一个 Pod 或者任何一个 API 对象被提交给 APIServer 之后，总有一些“初始化”性质的工作需要在它们被 Kubernetes 项目正式处理之前进行。比如，自动为所有 Pod 加上某些标签（Labels）。

而这个“初始化”操作的实现，借助的是一个叫作 Admission 的功能。它其实是 Kubernetes 项目里一组被称为 Admission Controller 的代码，可以选择性地被编译进 APIServer 中，在 API 对象创建之后会被立刻调用到。

但这就意味着，如果你现在想要添加一些自己的规则到 Admission Controller，就会比较困难。因为，这要求重新编译并重启 APIServer。显然，这种使用方法对 Istio 来说，影响太大了。

所以，Kubernetes 项目为我们额外提供了一种“热插拔”式的 Admission 机制，它就是 Dynamic Admission Control，也叫作：Initializer。

比如，我有如下所示的一个应用 Pod：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox
    command: ['sh', '-c', 'echo Hello Kubernetes! && sleep 3600']
```
可以看到，这个 Pod 里面只有一个用户容器，叫作：myapp-container。

接下来，Istio 项目要做的，就是在这个 Pod YAML 被提交给 Kubernetes 之后，在它对应的 API 对象里自动加上 Envoy 容器的配置，使这个对象变成如下所示的样子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox
    command: ['sh', '-c', 'echo Hello Kubernetes! && sleep 3600']
  - name: envoy
    image: lyft/envoy:845747b88f102c0fd262ab234308e9e22f693a1
    command: ["/usr/local/bin/envoy"]
    ...
```
可以看到，被 Istio 处理后的这个 Pod 里，除了用户自己定义的 myapp-container 容器之外，多出了一个叫作 envoy 的容器，它就是 Istio 要使用的 Envoy 代理。

那么，Istio 又是如何在用户完全不知情的前提下完成这个操作的呢？

Istio 要做的，就是编写一个用来为 Pod“自动注入”Envoy 容器的 Initializer。

首先，Istio 会将这个 Envoy 容器本身的定义，以 ConfigMap 的方式保存在 Kubernetes 当中。这个 ConfigMap（名叫：envoy-initializer）的定义如下所示：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-initializer
data:
  config: |
    containers:
      - name: envoy
        image: lyft/envoy:845747db88f102c0fd262ab234308e9e22f693a1
        command: ["/usr/local/bin/envoy"]
        args:
          - "--concurrency 4"
          - "--config-path /etc/envoy/envoy.json"
          - "--mode serve"
        ports:
          - containerPort: 80
            protocol: TCP
        resources:
          limits:
            cpu: "1000m"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "64Mi"
        volumeMounts:
          - name: envoy-conf
            mountPath: /etc/envoy
    volumes:
      - name: envoy-conf
        configMap:
          name: envoy
```

相信你已经注意到了，这个 ConfigMap 的 data 部分，正是一个 Pod 对象的一部分定义。其中，我们可以看到 Envoy 容器对应的 containers 字段，以及一个用来声明 Envoy 配置文件的 volumes 字段。

不难想到，Initializer 要做的工作，就是把这部分 Envoy 相关的字段，自动添加到用户提交的 Pod 的 API 对象里。可是，用户提交的 Pod 里本来就有 containers 字段和 volumes 字段，所以 Kubernetes 在处理这样的更新请求时，就必须使用类似于 git merge 这样的操作，才能将这两部分内容合并在一起。

所以说，在 Initializer 更新用户的 Pod 对象的时候，必须使用 PATCH API 来完成。而这种 PATCH API，正是声明式 API 最主要的能力。

接下来，Istio 将一个编写好的 Initializer，作为一个 Pod 部署在 Kubernetes 中。这个 Pod 的定义非常简单，如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: envoy-initializer
  name: envoy-initializer
spec:
  containers:
    - name: envoy-initializer
      image: envoy-initializer:0.0.1
      imagePullPolicy: Always
```

我们可以看到，这个 envoy-initializer 使用的 envoy-initializer:0.0.1 镜像，就是一个事先编写好的“自定义控制器”（Custom Controller），我将会在下面描述它的编写方法。而在这里，我要先为你解释一下这个控制器的主要功能。

一个 Kubernetes 的控制器，实际上就是一个“死循环”：它不断地获取“实际状态”，然后与“期望状态”作对比，并以此为依据决定下一步的操作。

而 Initializer 的控制器，不断获取到的“实际状态”，就是用户新创建的 Pod。而它的“期望状态”，则是：这个 Pod 里被添加了 Envoy 容器的定义。

我还是用一段 Go 语言风格的伪代码，来为你描述这个控制逻辑，如下所示：
```go
for {
  // 获取新创建的Pod
  pod := client.GetLatestPod()
  // Diff一下，检查是否已经初始化过
  if !isInitialized(pod) {
    // 没有？那就来初始化一下
    doSomething(pod)
  }
}
```
- 如果这个 Pod 里面已经添加过 Envoy 容器，那么就“放过”这个 Pod，进入下一个检查周期。
- 而如果还没有添加过 Envoy 容器的话，它就要进行 Initialize 操作了，即：修改该 Pod 的 API 对象（doSomething 函数）。

这时候，你应该立刻能想到，Istio 要往这个 Pod 里合并的字段，正是我们之前保存在 envoy-initializer 这个 ConfigMap 里的数据（即：它的 data 字段的值）。

所以，在 Initializer 控制器的工作逻辑里，它首先会从 APIServer 中拿到这个 ConfigMap：
```go
func doSomething(pod) {
  cm := client.Get(ConfigMap, "envoy-initializer")
}
```
然后，把这个 ConfigMap 里存储的 containers 和 volumes 字段，直接添加进一个空的 Pod 对象里：
```go
func doSomething(pod) {
  cm := client.Get(ConfigMap, "envoy-initializer")
  
  newPod := Pod{}
  newPod.Spec.Containers = cm.Containers
  newPod.Spec.Volumes = cm.Volumes
}
```

现在，关键来了。

Kubernetes 的 API 库，为我们提供了一个方法，使得我们可以直接使用新旧两个 Pod 对象，生成一个 TwoWayMergePatch：
```go
func doSomething(pod) {
  cm := client.Get(ConfigMap, "envoy-initializer")

  newPod := Pod{}
  newPod.Spec.Containers = cm.Containers
  newPod.Spec.Volumes = cm.Volumes

  // 生成patch数据
  patchBytes := strategicpatch.CreateTwoWayMergePatch(pod, newPod)

  // 发起PATCH请求，修改这个pod对象
  client.Patch(pod.Name, patchBytes)
}
```
**有了这个 TwoWayMergePatch 之后，Initializer 的代码就可以使用这个 patch 的数据，调用 Kubernetes 的 Client，发起一个 PATCH 请求。**

这样，一个用户提交的 Pod 对象里，就会被自动加上 Envoy 容器相关的字段。

当然，Kubernetes 还允许你通过配置，来指定要对什么样的资源进行这个 Initialize 操作，比如下面这个例子：
```yaml
apiVersion: admissionregistration.k8s.io/v1alpha1
kind: InitializerConfiguration
metadata:
  name: envoy-config
initializers:
  // 这个名字必须至少包括两个 "."
  - name: envoy.initializer.kubernetes.io
    rules:
      - apiGroups:
          - "" // 前面说过， ""就是core API Group的意思
        apiVersions:
          - v1
        resources:
          - pods
```

这个配置，就意味着 Kubernetes 要对所有的 Pod 进行这个 Initialize 操作，并且，我们指定了负责这个操作的 Initializer，名叫：envoy-initializer。

而一旦这个 InitializerConfiguration 被创建，Kubernetes 就会把这个 Initializer 的名字，加在所有新创建的 Pod 的 Metadata 上，格式如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  initializers:
    pending:
      - name: envoy.initializer.kubernetes.io
  name: myapp-pod
  labels:
    app: myapp
# ...
```

这个 Metadata，正是接下来 Initializer 的控制器判断这个 Pod 有没有执行过自己所负责的初始化操作的重要依据（也就是前面伪代码中 isInitialized() 方法的含义）。

**这也就意味着，当你在 Initializer 里完成了要做的操作后，一定要记得将这个 metadata.initializers.pending 标志清除掉。这一点，你在编写 Initializer 代码的时候一定要非常注意。**

此外，除了上面的配置方法，你还可以在具体的 Pod 的 Annotation 里添加一个如下所示的字段，从而声明要使用某个 Initializer：
```yaml
apiVersion: v1
kind: Pod
metadata
  annotations:
    "initializer.kubernetes.io/envoy": "true"
    # ...
```

在这个 Pod 里，我们添加了一个 Annotation，写明： initializer.kubernetes.io/envoy=true。这样，就会使用到我们前面所定义的 envoy-initializer 了。

以上，就是关于 Initializer 最基本的工作原理和使用方法了。相信你此时已经明白，Istio 项目的核心，就是由无数个运行在应用 Pod 中的 Envoy 容器组成的服务代理网格。这也正是 Service Mesh 的含义。

而这个机制得以实现的原理，正是借助了 Kubernetes 能够对 API 对象进行在线更新的能力，这也正是 Kubernetes“声明式 API”的独特之处：
- 首先，所谓“声明式”，指的就是我只需要提交一个定义好的 API 对象来“声明”，我所期望的状态是什么样子。
- 其次，“声明式 API”允许有多个 API 写端，以 PATCH 的方式对 API 对象进行修改，而无需关心本地原始 YAML 文件的内容。
- 最后，也是最重要的，有了上述两个能力，Kubernetes 项目才可以基于对 API 对象的增、删、改、查，在完全无需外界干预的情况下，完成对“实际状态”和“期望状态”的调谐（Reconcile）过程。
- （声明我的期望状态是什么，然后以在线更新的方式达到期望）

所以说，声明式 API，才是 Kubernetes 项目编排能力“赖以生存”的核心所在，一定要认真理解。

此外，不难看到，无论是对 sidecar 容器的巧妙设计，还是对 Initializer 的合理利用，Istio 项目的设计与实现，其实都依托于 Kubernetes 的声明式 API 和它所提供的各种编排能力。可以说，Istio 是在 Kubernetes 项目使用上的一位“集大成者”。（要知道，一个 Istio 项目部署完成后，会在 Kubernetes 里创建大约 43 个 API 对象。）

所以，Kubernetes 社区也看得很明白：Istio 项目有多火热，就说明 Kubernetes 这套“声明式 API”有多成功。这，既是 Google Cloud 喜闻乐见的事情，也是 Istio 项目一推出就被 Google 公司和整个技术圈儿热捧的重要原因。

而在使用 Initializer 的流程中，最核心的步骤，莫过于 Initializer“自定义控制器”的编写过程。它遵循的，正是标准的“Kubernetes 编程范式”，即：**如何使用控制器模式，同 Kubernetes 里 API 对象的“增、删、改、查”进行协作，进而完成用户业务逻辑的编写过程。**


## # 深入解析声明式API（一）：API对象的奥秘
这个小节主要了解一下 Kubernetes 声明式 API 的工作原理，以及如何利用这套 API 机制，在 Kubernetes 里添加自定义的 API 对象。

你可能一直就很好奇：当我把一个 YAML 文件提交给 Kubernetes 之后，它究竟是如何创建出一个 API 对象的呢？

这得从声明式 API 的设计谈起了。

在 Kubernetes 项目中，一个 API 对象在 Etcd 里的完整资源路径，是由：Group（API 组）、Version（API 版本）和 Resource（API 资源类型）三个部分组成的。
如图：![](http://cdn.ipso.live/notes/k8s01.png)

在这幅图中，你可以很清楚地看到 Kubernetes 里 API 对象的组织方式，其实是层层递进的。

比如，现在我要声明要创建一个 CronJob 对象，那么我的 YAML 文件的开始部分会这么写：
```yaml
apiVersion: batch/v2alpha1
kind: CronJob
# ...
```
在这个 YAML 文件中，“CronJob”就是这个 API 对象的资源类型（Resource），“batch”就是它的组（Group），v2alpha1 就是它的版本（Version）。

当我们提交了这个 YAML 文件之后，Kubernetes 就会把这个 YAML 文件里描述的内容，转换成 Kubernetes 里的一个 CronJob 对象。

那么，Kubernetes 是如何对 Resource、Group 和 Version 进行解析，从而在 Kubernetes 项目里找到 CronJob 对象的定义呢？

1. 首先，Kubernetes 会匹配 API 对象的组。

需要明确的是，对于 Kubernetes 里的核心 API 对象，比如：Pod、Node 等，是不需要 Group 的（即：它们的 Group 是“”）。所以，对于这些 API 对象来说，Kubernetes 会直接在 /api 这个层级进行下一步的匹配过程。

而对于 CronJob 等非核心 API 对象来说，Kubernetes 就必须在 /apis 这个层级里查找它对应的 Group，进而根据“batch”这个 Group 的名字，找到 /apis/batch。

不难发现，这些 API Group 的分类是以对象功能为依据的，比如 Job 和 CronJob 就都属于“batch” （离线业务）这个 Group。

2. 然后，Kubernetes 会进一步匹配到 API 对象的版本号。

对于 CronJob 这个 API 对象来说，Kubernetes 在 batch 这个 Group 下，匹配到的版本号就是 v2alpha1。

在 Kubernetes 中，同一种 API 对象可以有多个版本，这正是 Kubernetes 进行 API 版本化管理的重要手段。这样，比如在 CronJob 的开发过程中，对于会影响到用户的变更就可以通过升级新版本来处理，从而保证了向后兼容。

3. 最后，Kubernetes 会匹配 API 对象的资源类型。

在前面匹配到正确的版本之后，Kubernetes 就知道，我要创建的原来是一个 /apis/batch/v2alpha1 下的 CronJob 对象。

**这时候，APIServer 就可以继续创建这个 CronJob 对象了。为了方便理解，我总结了如下创建过程：**
1. 首先，当我们发起了创建 CronJob 的 POST 请求之后，我们编写的 YAML 的信息就被提交给了 APIServer。而 APIServer 的第一个功能，就是过滤这个请求，并完成一些前置性的工作，比如授权、超时处理、审计等。

2. 然后，请求会进入 MUX 和 Routes 流程。如果你编写过 Web Server 的话就会知道，MUX 和 Routes 是 APIServer 完成 URL 和 Handler 绑定的场所。而 APIServer 的 Handler 要做的事情，就是按照我刚刚介绍的匹配过程，找到对应的 CronJob 类型定义。

3. 接着，APIServer 最重要的职责就来了：根据这个 CronJob 类型定义，使用用户提交的 YAML 文件里的字段，创建一个 CronJob 对象。而在这个过程中，APIServer 会进行一个 Convert 工作，即：把用户提交的 YAML 文件，转换成一个叫作 Super Version 的对象，它正是该 API 资源类型所有版本的字段全集。这样用户提交的不同版本的 YAML 文件，就都可以用这个 Super Version 对象来进行处理了。

4. 接下来，APIServer 会先后进行 Admission() 和 Validation() 操作。比如，我在上一篇文章中提到的 Admission Controller 和 Initializer，就都属于 Admission 的内容。而 Validation，则负责验证这个对象里的各个字段是否合法。这个被验证过的 API 对象，都保存在了 APIServer 里一个叫作 Registry 的数据结构中。也就是说，只要一个 API 对象的定义能在 Registry 里查到，它就是一个有效的 Kubernetes API 对象。

5. 最后，APIServer 会把验证过的 API 对象转换成用户最初提交的版本，进行序列化操作，并调用 Etcd 的 API 把它保存起来。

由此可见，声明式 API 对于 Kubernetes 来说非常重要。所以，APIServer 这样一个在其他项目里“平淡无奇”的组件，却成了 Kubernetes 项目的重中之重。它不仅是 Google Borg 设计思想的集中体现，也是 Kubernetes 项目里唯一一个被 Google 公司和 RedHat 公司双重控制、其他势力根本无法参与其中的组件。

此外，由于同时要兼顾性能、API 完备性、版本化、向后兼容等很多工程化指标，所以 Kubernetes 团队在 APIServer 项目里大量使用了 Go 语言的代码生成功能，来自动化诸如 Convert、DeepCopy 等与 API 资源相关的操作。这部分自动生成的代码，曾一度占到 Kubernetes 项目总代码的 20%~30%。

这也是为何，在过去很长一段时间里，在这样一个极其“复杂”的 APIServer 中，添加一个 Kubernetes 风格的 API 资源类型，是一个非常困难的工作。

不过，在 Kubernetes v1.7 之后，这个工作就变得轻松得多了。这，当然得益于一个全新的 API 插件机制：CRD。

#### CRD
CRD 的全称是 Custom Resource Definition。顾名思义，它指的就是，允许用户在 Kubernetes 中添加一个跟 Pod、Node 类似的、新的 API 资源类型，即：自定义 API 资源。

举个例子，我现在要为 Kubernetes 添加一个名叫 Network 的 API 资源类型。

它的作用是，一旦用户创建一个 Network 对象，那么 Kubernetes 就应该使用这个对象定义的网络参数，调用真实的网络插件，比如 Neutron 项目，为用户创建一个真正的“网络”。这样，将来用户创建的 Pod，就可以声明使用这个“网络”了。

这个 Network 对象的 YAML 文件，名叫 example-network.yaml，它的内容如下所示：
```yaml
apiVersion: samplecrd.k8s.io/v1
kind: Network
metadata:
  name: example-network
spec:
  cidr: "192.168.0.0/16"
  gateway: "192.168.0.1"
```
可以看到，我想要描述“网络”的 API 资源类型是 Network；API 组是samplecrd.k8s.io；API 版本是 v1。

那么，Kubernetes 又该如何知道这个 API（samplecrd.k8s.io/v1/network）的存在呢？

其实，上面的这个 YAML 文件，就是一个具体的“自定义 API 资源”实例，也叫 CR（Custom Resource）。而为了能够让 Kubernetes 认识这个 CR，你就需要让 Kubernetes 明白这个 CR 的宏观定义是什么，也就是 CRD（Custom Resource Definition）。

这就好比，你想让计算机认识各种兔子的照片，就得先让计算机明白，兔子的普遍定义是什么。比如，兔子“是哺乳动物”“有长耳朵，三瓣嘴”。

所以，接下来，我就先编写一个 CRD 的 YAML 文件，它的名字叫作 network.yaml，内容如下所示：
```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networks.samplecrd.k8s.io
spec:
  group: samplecrd.k8s.io
  version: v1
  names:
    kind: Network
    plural: networks
  scope: Namespaced
```
可以看到，在这个 CRD 中，我指定了“group: samplecrd.k8s.io”“version: v1”这样的 API 信息，也指定了这个 CR 的资源类型叫作 Network，复数（plural）是 networks。

然后，我还声明了它的 scope 是 Namespaced，即：我们定义的这个 Network 是一个属于 Namespace 的对象，类似于 Pod。

这就是一个 Network API 资源类型的 API 部分的宏观定义了。这就等同于告诉了计算机：“兔子是哺乳动物”。所以这时候，Kubernetes 就能够认识和处理所有声明了 API 类型是“samplecrd.k8s.io/v1/network”的 YAML 文件了。

接下来，我还需要让 Kubernetes“认识”这种 YAML 文件里描述的“网络”部分，比如“cidr”（网段），“gateway”（网关）这些字段的含义。这就相当于我要告诉计算机：“兔子有长耳朵和三瓣嘴”。

这时候呢，我就需要稍微做些代码工作了。

```sh
$ tree <your-go-proj-path>/k8s-controller-custom-resource
.
├── controller.go
├── crd
│   └── network.yaml
├── example
│   └── example-network.yaml
├── main.go
└── pkg
    └── apis
        └── samplecrd
            ├── register.go
            └── v1
                ├── doc.go
                ├── register.go
                └── types.go
```
其中，pkg/apis/samplecrd 就是 API 组的名字，v1 是版本，而 v1 下面的 types.go 文件里，则定义了 Network 对象的完整描述。

然后，我在 pkg/apis/samplecrd 目录下创建了一个 register.go 文件，用来放置后面要用到的全局变量。这个文件的内容如下所示：
```go
package samplecrd

const (
 GroupName = "samplecrd.k8s.io"
 Version   = "v1"
)
```
接着，我需要在 pkg/apis/samplecrd 目录下添加一个 doc.go 文件（Golang 的文档源文件）。这个文件里的内容如下所示：
```go
// +k8s:deepcopy-gen=package

// +groupName=samplecrd.k8s.io
package v1
```
在这个文件中，你会看到 +[=value]格式的注释，这就是 Kubernetes 进行代码生成要用的 Annotation 风格的注释。

其中，+k8s:deepcopy-gen=package 意思是，请为整个 v1 包里的所有类型定义自动生成 DeepCopy 方法；而+groupName=samplecrd.k8s.io，则定义了这个包对应的 API 组的名字。

可以看到，这些定义在 doc.go 文件的注释，起到的是全局的代码生成控制的作用，所以也被称为 Global Tags。

接下来，我需要添加 types.go 文件。顾名思义，它的作用就是定义一个 Network 类型到底有哪些字段（比如，spec 字段里的内容）。这个文件的主要内容如下所示：
```go
package v1
...
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Network describes a Network resource
type Network struct {
 // TypeMeta is the metadata for the resource, like kind and apiversion
 metav1.TypeMeta `json:",inline"`
 // ObjectMeta contains the metadata for the particular object, including
 // things like...
 //  - name
 //  - namespace
 //  - self link
 //  - labels
 //  - ... etc ...
 metav1.ObjectMeta `json:"metadata,omitempty"`
 
 Spec networkspec `json:"spec"`
}
// networkspec is the spec for a Network resource
type networkspec struct {
 Cidr    string `json:"cidr"`
 Gateway string `json:"gateway"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkList is a list of Network resources
type NetworkList struct {
 metav1.TypeMeta `json:",inline"`
 metav1.ListMeta `json:"metadata"`
 
 Items []Network `json:"items"`
}
```
在上面这部分代码里，你可以看到 Network 类型定义方法跟标准的 Kubernetes 对象一样，都包括了 TypeMeta（API 元数据）和 ObjectMeta（对象元数据）字段。

而其中的 Spec 字段，就是需要我们自己定义的部分。所以，在 networkspec 里，我定义了 Cidr 和 Gateway 两个字段。其中，每个字段最后面的部分比如json:"cidr"，指的就是这个字段被转换成 JSON 格式之后的名字，也就是 YAML 文件里的字段名字。

**此外，除了定义 Network 类型，你还需要定义一个 NetworkList 类型，用来描述一组 Network 对象应该包括哪些字段。之所以需要这样一个类型，是因为在 Kubernetes 中，获取所有 X 对象的 List() 方法，返回值都是List 类型，而不是 X 类型的数组。这是不一样的。**

同样地，在 Network 和 NetworkList 类型上，也有代码生成注释。

其中，+genclient 的意思是：请为下面这个 API 资源类型生成对应的 Client 代码（这个 Client，我马上会讲到）。而 +genclient:noStatus 的意思是：这个 API 资源类型定义里，没有 Status 字段。否则，生成的 Client 就会自动带上 UpdateStatus 方法。

如果你的类型定义包括了 Status 字段的话，就不需要这句 +genclient:noStatus 注释了。比如下面这个例子：
```go
// +genclient

// Network is a specification for a Network resource
type Network struct {
 metav1.TypeMeta   `json:",inline"`
 metav1.ObjectMeta `json:"metadata,omitempty"`
 
 Spec   NetworkSpec   `json:"spec"`
 Status NetworkStatus `json:"status"`
}
```
需要注意的是，+genclient 只需要写在 Network 类型上，而不用写在 NetworkList 上。因为 NetworkList 只是一个返回值类型，Network 才是“主类型”。

而由于我在 Global Tags 里已经定义了为所有类型生成 DeepCopy 方法，所以这里就不需要再显式地加上 +k8s:deepcopy-gen=true 了。当然，这也就意味着你可以用 +k8s:deepcopy-gen=false 来阻止为某些类型生成 DeepCopy。

你可能已经注意到，在这两个类型上面还有一句+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object的注释。它的意思是，请在生成 DeepCopy 的时候，实现 Kubernetes 提供的 runtime.Object 接口。否则，在某些版本的 Kubernetes 里，你的这个类型定义会出现编译错误。这是一个固定的操作，记住即可。

不过，你或许会有这样的顾虑：这些代码生成注释这么灵活，我该怎么掌握呢？

其实，上面我所讲述的内容，已经足以应对 99% 的场景了。当然，如果你对代码生成感兴趣的话，我推荐你阅读[这篇博客](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/)，它详细地介绍了 Kubernetes 的代码生成语法。

最后，我需要再编写一个 pkg/apis/samplecrd/v1/register.go 文件。

在前面对 APIServer 工作原理部分，我已经提到，“registry”的作用就是注册一个类型（Type）给 APIServer。其中，Network 资源类型在服务器端注册的工作，APIServer 会自动帮我们完成。但与之对应的，我们还需要让客户端也能“知道”Network 资源类型的定义。这就需要我们在项目里添加一个 register.go 文件。它最主要的功能，就是定义了如下所示的 addKnownTypes() 方法：
```go
package v1
...
// addKnownTypes adds our types to the API scheme by registering
// Network and NetworkList
func addKnownTypes(scheme *runtime.Scheme) error {
 scheme.AddKnownTypes(
  SchemeGroupVersion,
  &Network{},
  &NetworkList{},
 )
 
 // register the type in the scheme
 metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
 return nil
}
```
有了这个方法，Kubernetes 就能够在后面生成客户端的时候，“知道”Network 以及 NetworkList 类型的定义了。

**像上面这种 register.go 文件里的内容其实是非常固定的，你以后可以直接使用我提供的这部分代码做模板，然后把其中的资源类型、GroupName 和 Version 替换成你自己的定义即可。**

这样，Network 对象的定义工作就全部完成了。可以看到，它其实定义了两部分内容：
- 第一部分是，自定义资源类型的 API 描述，包括：组（Group）、版本（Version）、资源类型（Resource）等。这相当于告诉了计算机：兔子是哺乳动物。
- 第二部分是，自定义资源类型的对象描述，包括：Spec、Status 等。这相当于告诉了计算机：兔子有长耳朵和三瓣嘴。

接下来，我就要使用 Kubernetes 提供的代码生成工具，为上面定义的 Network 资源类型自动生成 clientset、informer 和 lister。其中，clientset 就是操作 Network 对象所需要使用的客户端，而 informer 和 lister 这两个包的主要功能参考下一小节的内容。

这个代码生成工具名叫k8s.io/code-generator，使用方法如下所示：
```sh
# 代码生成的工作目录，也就是我们的项目路径
$ ROOT_PACKAGE="github.com/resouer/k8s-controller-custom-resource"
# API Group
$ CUSTOM_RESOURCE_NAME="samplecrd"
# API Version
$ CUSTOM_RESOURCE_VERSION="v1"

# 安装k8s.io/code-generator
$ go get -u k8s.io/code-generator/...
$ cd $GOPATH/src/k8s.io/code-generator

# 执行代码自动生成，其中pkg/client是生成目标目录，pkg/apis是类型定义目录
$ ./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION"
```
代码生成工作完成之后，我们再查看一下这个项目的目录结构：
```sh
$ tree
.
├── controller.go
├── crd
│   └── network.yaml
├── example
│   └── example-network.yaml
├── main.go
└── pkg
    ├── apis
    │   └── samplecrd
    │       ├── constants.go
    │       └── v1
    │           ├── doc.go
    │           ├── register.go
    │           ├── types.go
    │           └── zz_generated.deepcopy.go
    └── client
        ├── clientset
        ├── informers
        └── listers
```
其中，pkg/apis/samplecrd/v1 下面的 zz_generated.deepcopy.go 文件，就是自动生成的 DeepCopy 代码文件。

而整个 client 目录，以及下面的三个包（clientset、informers、 listers），都是 Kubernetes 为 Network 类型生成的客户端库，这些库会在后面编写自定义控制器的时候用到。

可以看到，到目前为止的这些工作，其实并不要求你写多少代码，主要考验的是“复制、粘贴、替换”这样的“基本功”。

而有了这些内容，现在你就可以在 Kubernetes 集群里创建一个 Network 类型的 API 对象了。我们不妨一起来试验下。

1. 首先，使用 network.yaml 文件，在 Kubernetes 中创建 Network 对象的 CRD（Custom Resource Definition）：
```sh
$ kubectl apply -f crd/network.yaml
customresourcedefinition.apiextensions.k8s.io/networks.samplecrd.k8s.io created
```

这个操作，就告诉了 Kubernetes，我现在要添加一个自定义的 API 对象。而这个对象的 API 信息，正是 network.yaml 里定义的内容。我们可以通过 kubectl get 命令，查看这个 CRD：
```sh
$ kubectl get crd
NAME                        CREATED AT
networks.samplecrd.k8s.io   2018-09-15T10:57:12Z
```

2. 然后，我们就可以创建一个 Network 对象了，这里用到的是 example-network.yaml：
```sh
$ kubectl apply -f example/example-network.yaml 
network.samplecrd.k8s.io/example-network created
```
通过这个操作，你就在 Kubernetes 集群里创建了一个 Network 对象。它的 API 资源路径是samplecrd.k8s.io/v1/networks。

这时候，你就可以通过 kubectl get 命令，查看到新创建的 Network 对象：
```sh
$ kubectl get network
NAME              AGE
example-network   8s
```
你还可以通过 kubectl describe 命令，看到这个 Network 对象的细节：
```sh
$ kubectl describe network example-network
Name:         example-network
Namespace:    default
Labels:       <none>
...API Version:  samplecrd.k8s.io/v1
Kind:         Network
Metadata:
  ...
  Generation:          1
  Resource Version:    468239
  ...
Spec:
  Cidr:     192.168.0.0/16
  Gateway:  192.168.0.1
```
当然 ，你也可以编写更多的 YAML 文件来创建更多的 Network 对象，这和创建 Pod、Deployment 的操作，没有任何区别。


## # 深入解析声明式API（二）：编写自定义控制器
在上一小节中，详细分享了 Kubernetes 中声明式 API 的实现原理，并且通过一个添加 Network 对象的实例，讲述了在 Kubernetes 里添加 API 资源的过程。

在这小节中，继续完成剩下一半的工作，即：为 Network 这个自定义 API 对象编写一个自定义控制器（Custom Controller）。

**正如我在上一小节结尾处提到的，“声明式 API”并不像“命令式 API”那样有着明显的执行逻辑。这就使得基于声明式 API 的业务功能实现，往往需要通过控制器模式来“监视”API 对象的变化（比如，创建或者删除 Network），然后以此来决定实际要执行的具体工作。**

接下来，我就和你一起通过编写代码来实现这个过程。这个项目和上一小节里的代码是同一个项目。

**总得来说，编写自定义控制器代码的过程包括：编写 main 函数、编写自定义控制器的定义，以及编写控制器里的业务逻辑三个部分。**

#### 首先，我们来编写这个自定义控制器的 main 函数。
main 函数的主要工作就是，定义并初始化一个自定义控制器（Custom Controller），然后启动它。这部分代码的主要内容如下所示：
```go
func main() {
  // ...
  
  cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
  // ...
  kubeClient, err := kubernetes.NewForConfig(cfg)
  // ...
  networkClient, err := clientset.NewForConfig(cfg)
  // ...
  
  networkInformerFactory := informers.NewSharedInformerFactory(networkClient, ...)
  
  controller := NewController(kubeClient, networkClient,
  networkInformerFactory.Samplecrd().V1().Networks())
  
  go networkInformerFactory.Start(stopCh)
 
  if err = controller.Run(2, stopCh); err != nil {
    glog.Fatalf("Error running controller: %s", err.Error())
  }
}
```
可以看到，这个 main 函数主要通过三步完成了初始化并启动一个自定义控制器的工作。
1. 第一步：main 函数根据我提供的 Master 配置（APIServer 的地址端口和 kubeconfig 的路径），创建一个 Kubernetes 的 client（kubeClient）和 Network 对象的 client（networkClient）。

但是，如果我没有提供 Master 配置呢？

这时，main 函数会直接使用一种名叫 InClusterConfig 的方式来创建这个 client。这个方式，会假设你的自定义控制器是以 Pod 的方式运行在 Kubernetes 集群里的。

而我在 Pod 使用进阶小节中曾经提到过，Kubernetes 里所有的 Pod 都会以 Volume 的方式自动挂载 Kubernetes 的默认 ServiceAccount。所以，这个控制器就会直接使用默认 ServiceAccount 数据卷里的授权信息，来访问 APIServer。

2. 第二步：main 函数为 Network 对象创建一个叫作 InformerFactory（即：networkInformerFactory）的工厂，并使用它生成一个 Network 对象的 Informer，传递给控制器。

3. 第三步：main 函数启动上述的 Informer，然后执行 controller.Run，启动自定义控制器。

至此，main 函数就结束了。

看到这，你可能会感到非常困惑：编写自定义控制器的过程难道就这么简单吗？这个 Informer 又是个什么东西呢？

别着急。接下来，我就为你详细解释一下这个自定义控制器的工作原理。

在 Kubernetes 项目中，一个自定义控制器的工作原理，可以用下面这样一幅流程图来表示（在后面的叙述中，我会用“示意图”来指代它）：
![](http://cdn.ipso.live/notes/k8s02.png)

我们先从这幅示意图的最左边看起。

这个控制器要做的第一件事，是从 Kubernetes 的 APIServer 里获取它所关心的对象，也就是我定义的 Network 对象。

这个操作，依靠的是一个叫作 Informer（可以翻译为：通知器）的代码库完成的。Informer 与 API 对象是一一对应的，所以我传递给自定义控制器的，正是一个 Network 对象的 Informer（Network Informer）。

不知你是否已经注意到，我在创建这个 Informer 工厂的时候，需要给它传递一个 networkClient。

事实上，Network Informer 正是使用这个 networkClient，跟 APIServer 建立了连接。不过，真正负责维护这个连接的，则是 Informer 所使用的 Reflector 包。

更具体地说，Reflector 使用的是一种叫作 ListAndWatch 的方法，来“获取”并“监听”这些 Network 对象实例的变化。

在 ListAndWatch 机制下，一旦 APIServer 端有新的 Network 实例被创建、删除或者更新，Reflector 都会收到“事件通知”。这时，该事件及它对应的 API 对象这个组合，就被称为增量（Delta），它会被放进一个 Delta FIFO Queue（即：增量先进先出队列）中。

而另一方面，Informe 会不断地从这个 Delta FIFO Queue 里读取（Pop）增量。每拿到一个增量，Informer 就会判断这个增量里的事件类型，然后创建或者更新本地对象的缓存。这个缓存，在 Kubernetes 里一般被叫作 Store。

比如，如果事件类型是 Added（添加对象），那么 Informer 就会通过一个叫作 Indexer 的库把这个增量里的 API 对象保存在本地缓存中，并为它创建索引。相反，如果增量的事件类型是 Deleted（删除对象），那么 Informer 就会从本地缓存中删除这个对象。

这个同步本地缓存的工作，是 Informer 的第一个职责，也是它最重要的职责。

而 Informer 的第二个职责，则是根据这些事件的类型，触发事先注册好的 ResourceEventHandler。这些 Handler，需要在创建控制器的时候注册给它对应的 Informer。

接下来，我们就来编写这个控制器的定义，它的主要内容如下所示：
```go
func NewController(
  kubeclientset kubernetes.Interface,
  networkclientset clientset.Interface,
  networkInformer informers.NetworkInformer) *Controller {
  ...
  controller := &Controller{
    kubeclientset:    kubeclientset,
    networkclientset: networkclientset,
    networksLister:   networkInformer.Lister(),
    networksSynced:   networkInformer.Informer().HasSynced,
    workqueue:        workqueue.NewNamedRateLimitingQueue(...,  "Networks"),
    ...
  }
    networkInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: controller.enqueueNetwork,
    UpdateFunc: func(old, new interface{}) {
      oldNetwork := old.(*samplecrdv1.Network)
      newNetwork := new.(*samplecrdv1.Network)
      if oldNetwork.ResourceVersion == newNetwork.ResourceVersion {
        return
      }
      controller.enqueueNetwork(new)
    },
    DeleteFunc: controller.enqueueNetworkForDelete,
 return controller
}
```
**我前面在 main 函数里创建了两个 client（kubeclientset 和 networkclientset），然后在这段代码里，使用这两个 client 和前面创建的 Informer，初始化了自定义控制器。**

值得注意的是，在这个自定义控制器里，我还设置了一个工作队列（work queue），它正是处于示意图中间位置的 WorkQueue。这个工作队列的作用是，负责同步 Informer 和控制循环之间的数据。(实际上，Kubernetes 项目为我们提供了很多个工作队列的实现，你可以根据需要选择合适的库直接使用。)

**然后，我为 networkInformer 注册了三个 Handler（AddFunc、UpdateFunc 和 DeleteFunc），分别对应 API 对象的“添加”“更新”和“删除”事件。而具体的处理操作，都是将该事件对应的 API 对象加入到工作队列中。**

需要注意的是，实际入队的并不是 API 对象本身，而是它们的 Key，即：该 API 对象的namespace/name。

而我们后面即将编写的控制循环，则会不断地从这个工作队列里拿到这些 Key，然后开始执行真正的控制逻辑。

**综合上面的讲述，你现在应该就能明白，所谓 Informer，其实就是一个带有本地缓存和索引机制的、可以注册 EventHandler 的 client。它是自定义控制器跟 APIServer 进行数据同步的重要组件。**

更具体地说，Informer 通过一种叫作 ListAndWatch 的方法，把 APIServer 中的 API 对象缓存在了本地，并负责更新和维护这个缓存。

其中，ListAndWatch 方法的含义是：首先，通过 APIServer 的 LIST API“获取”所有最新版本的 API 对象；然后，再通过 WATCH API 来“监听”所有这些 API 对象的变化。

而通过监听到的事件变化，Informer 就可以实时地更新本地缓存，并且调用这些事件对应的 EventHandler 了。

此外，在这个过程中，每经过 resyncPeriod 指定的时间，Informer 维护的本地缓存，都会使用最近一次 LIST 返回的结果强制更新一次，从而保证缓存的有效性。在 Kubernetes 中，这个缓存强制更新的操作就叫作：resync。

需要注意的是，这个定时 resync 操作，也会触发 Informer 注册的“更新”事件。但此时，这个“更新”事件对应的 Network 对象实际上并没有发生变化，即：新、旧两个 Network 对象的 ResourceVersion 是一样的。在这种情况下，Informer 就不需要对这个更新事件再做进一步的处理了。

这也是为什么我在上面的 UpdateFunc 方法里，先判断了一下新、旧两个 Network 对象的版本（ResourceVersion）是否发生了变化，然后才开始进行的入队操作。

以上，就是 Kubernetes 中的 Informer 库的工作原理了。

接下来，我们就来到了示意图中最后面的控制循环（Control Loop）部分，也正是我在 main 函数最后调用 controller.Run() 启动的“控制循环”。它的主要内容如下所示：
```go
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
 ...
  if ok := cache.WaitForCacheSync(stopCh, c.networksSynced); !ok {
    return fmt.Errorf("failed to wait for caches to sync")
  }
  
  ...
  for i := 0; i < threadiness; i++ {
    go wait.Until(c.runWorker, time.Second, stopCh)
  }
  
  ...
  return nil
}
```
可以看到，启动控制循环的逻辑非常简单：
- 首先，等待 Informer 完成一次本地缓存的数据同步操作；
- 然后，直接通过 goroutine 启动一个（或者并发启动多个）“无限循环”的任务。

而这个“无限循环”任务的每一个循环周期，执行的正是我们真正关心的业务逻辑。

所以接下来，我们就来编写这个自定义控制器的业务逻辑，它的主要内容如下所示：
```go
func (c *Controller) runWorker() {
  for c.processNextWorkItem() {
  }
}

func (c *Controller) processNextWorkItem() bool {
  obj, shutdown := c.workqueue.Get()
  
  ...
  
  err := func(obj interface{}) error {
    ...
    if err := c.syncHandler(key); err != nil {
     return fmt.Errorf("error syncing '%s': %s", key, err.Error())
    }
    
    c.workqueue.Forget(obj)
    ...
    return nil
  }(obj)
  
  ...
  
  return true
}

func (c *Controller) syncHandler(key string) error {

  namespace, name, err := cache.SplitMetaNamespaceKey(key)
  ...
  
  network, err := c.networksLister.Networks(namespace).Get(name)
  if err != nil {
    if errors.IsNotFound(err) {
      glog.Warningf("Network does not exist in local cache: %s/%s, will delete it from Neutron ...",
      namespace, name)
      
      glog.Warningf("Network: %s/%s does not exist in local cache, will delete it from Neutron ...",
    namespace, name)
    
     // FIX ME: call Neutron API to delete this network by name.
     //
     // neutron.Delete(namespace, name)
     
     return nil
  }
    ...
    
    return err
  }
  
  glog.Infof("[Neutron] Try to process network: %#v ...", network)
  
  // FIX ME: Do diff().
  //
  // actualNetwork, exists := neutron.Get(namespace, name)
  //
  // if !exists {
  //   neutron.Create(namespace, name)
  // } else if !reflect.DeepEqual(actualNetwork, network) {
  //   neutron.Update(namespace, name)
  // }
  
  return nil
}
```
可以看到，在这个执行周期里（processNextWorkItem），我们首先从工作队列里出队（workqueue.Get）了一个成员，也就是一个 Key（Network 对象的：namespace/name）。

然后，在 syncHandler 方法中，我使用这个 Key，尝试从 Informer 维护的缓存中拿到了它所对应的 Network 对象。

可以看到，在这里，我使用了 networksLister 来尝试获取这个 Key 对应的 Network 对象。这个操作，其实就是在访问本地缓存的索引。实际上，在 Kubernetes 的源码中，你会经常看到控制器从各种 Lister 里获取对象，比如：podLister、nodeLister 等等，它们使用的都是 Informer 和缓存机制。

而如果控制循环从缓存中拿不到这个对象（即：networkLister 返回了 IsNotFound 错误），那就意味着这个 Network 对象的 Key 是通过前面的“删除”事件添加进工作队列的。所以，尽管队列里有这个 Key，但是对应的 Network 对象已经被删除了。

这时候，我就需要调用 Neutron 的 API，把这个 Key 对应的 Neutron 网络从真实的集群里删除掉。

而如果能够获取到对应的 Network 对象，我就可以执行控制器模式里的对比“期望状态”和“实际状态”的逻辑了。

其中，自定义控制器“千辛万苦”拿到的这个 Network 对象，正是 APIServer 里保存的“期望状态”，即：用户通过 YAML 文件提交到 APIServer 里的信息。当然，在我们的例子里，它已经被 Informer 缓存在了本地。

那么，“实际状态”又从哪里来呢？

当然是来自于实际的集群了。

所以，我们的控制循环需要通过 Neutron API 来查询实际的网络情况。

比如，我可以先通过 Neutron 来查询这个 Network 对象对应的真实网络是否存在。

- 如果不存在，这就是一个典型的“期望状态”与“实际状态”不一致的情形。这时，我就需要使用这个 Network 对象里的信息（比如：CIDR 和 Gateway），调用 Neutron API 来创建真实的网络。
- 如果存在，那么，我就要读取这个真实网络的信息，判断它是否跟 Network 对象里的信息一致，从而决定我是否要通过 Neutron 来更新这个已经存在的真实网络。

这样，我就通过对比“期望状态”和“实际状态”的差异，完成了一次调协（Reconcile）的过程。

至此，一个完整的自定义 API 对象和它所对应的自定义控制器，就编写完毕了。（备注：与 Neutron 相关的业务代码并不是这里的重点，所以我仅仅通过注释里的伪代码为你表述了这部分内容。如果你对这些代码感兴趣的话，可以自行完成。最简单的情况，你可以自己编写一个 Neutron Mock，然后输出对应的操作日志。）

接下来，我们就一起来把这个项目运行起来，查看一下它的工作情况。

你可以自己编译这个项目，也可以直接使用我编译好的二进制文件（samplecrd-controller）。编译并启动这个项目的具体流程如下所示：
```sh
# Clone repo
$ git clone https://github.com/resouer/k8s-controller-custom-resource$ cd k8s-controller-custom-resource

### Skip this part if you don't want to build
# Install dependency
$ go get github.com/tools/godep
$ godep restore
# Build
$ go build -o samplecrd-controller .

$ ./samplecrd-controller -kubeconfig=$HOME/.kube/config -alsologtostderr=true
I0915 12:50:29.051349   27159 controller.go:84] Setting up event handlers
I0915 12:50:29.051615   27159 controller.go:113] Starting Network control loop
I0915 12:50:29.051630   27159 controller.go:116] Waiting for informer caches to sync
E0915 12:50:29.066745   27159 reflector.go:134] github.com/resouer/k8s-controller-custom-resource/pkg/client/informers/externalversions/factory.go:117: Failed to list *v1.Network: the server could not find the requested resource (get networks.samplecrd.k8s.io)
...
```
（**注意：这里是直接运行可执行文件，也可以像 Istio 项目的 Envoy 容器组将一样使用 sider car 的方式在pod中运行**）

你可以看到，自定义控制器被启动后，一开始会报错。

这是因为，此时 Network 对象的 CRD 还没有被创建出来，所以 Informer 去 APIServer 里“获取”（List）Network 对象时，并不能找到 Network 这个 API 资源类型的定义，即：
```sh
Failed to list *v1.Network: the server could not find the requested resource (get networks.samplecrd.k8s.io)
```
所以，接下来我就需要创建 Network 对象的 CRD，这个操作在上一篇文章里已经介绍过了。

在另一个 shell 窗口里执行：
```sh
$ kubectl apply -f crd/network.yaml
```
这时候，你就会看到控制器的日志恢复了正常，控制循环启动成功：
```log
...
I0915 12:50:29.051630   27159 controller.go:116] Waiting for informer caches to sync
...
I0915 12:52:54.346854   25245 controller.go:121] Starting workers
I0915 12:52:54.346914   25245 controller.go:127] Started workers
```

接下来，我就可以进行 Network 对象的增删改查操作了。

首先，创建一个 Network 对象：
```yaml
$ cat example/example-network.yaml 
apiVersion: samplecrd.k8s.io/v1
kind: Network
metadata:
  name: example-network
spec:
  cidr: "192.168.0.0/16"
  gateway: "192.168.0.1"
  
$ kubectl apply -f example/example-network.yaml 
network.samplecrd.k8s.io/example-network created
```

这时候，查看一下控制器的输出：
```log
...
I0915 12:50:29.051349   27159 controller.go:84] Setting up event handlers
I0915 12:50:29.051615   27159 controller.go:113] Starting Network control loop
I0915 12:50:29.051630   27159 controller.go:116] Waiting for informer caches to sync
...
I0915 12:52:54.346854   25245 controller.go:121] Starting workers
I0915 12:52:54.346914   25245 controller.go:127] Started workers
I0915 12:53:18.064409   25245 controller.go:229] [Neutron] Try to process network: &v1.Network{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"example-network", GenerateName:"", Namespace:"default", ... ResourceVersion:"479015", ... Spec:v1.NetworkSpec{Cidr:"192.168.0.0/16", Gateway:"192.168.0.1"}} ...
I0915 12:53:18.064650   25245 controller.go:183] Successfully synced 'default/example-network'
...
```

可以看到，我们上面创建 example-network 的操作，触发了 EventHandler 的“添加”事件，从而被放进了工作队列。

紧接着，控制循环就从队列里拿到了这个对象，并且打印出了正在“处理”这个 Network 对象的日志。

可以看到，这个 Network 的 ResourceVersion，也就是 API 对象的版本号，是 479015，而它的 Spec 字段的内容，跟我提交的 YAML 文件一摸一样，比如，它的 CIDR 网段是：192.168.0.0/16。

这时候，我来修改一下这个 YAML 文件的内容，如下所示：
```sh
$ cat example/example-network.yaml 
apiVersion: samplecrd.k8s.io/v1
kind: Network
metadata:
  name: example-network
spec:
  cidr: "192.168.1.0/16"
  gateway: "192.168.1.1"
```
可以看到，我把这个 YAML 文件里的 CIDR 和 Gateway 字段修改成了 192.168.1.0/16 网段。

然后，我们执行了 kubectl apply 命令来提交这次更新，如下所示：
```sh
$ kubectl apply -f example/example-network.yaml 
network.samplecrd.k8s.io/example-network configured
```

这时候，我们就可以观察一下控制器的输出:
```log
...
I0915 12:53:51.126029   25245 controller.go:229] [Neutron] Try to process network: &v1.Network{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"example-network", GenerateName:"", Namespace:"default", ...  ResourceVersion:"479062", ... Spec:v1.NetworkSpec{Cidr:"192.168.1.0/16", Gateway:"192.168.1.1"}} ...
I0915 12:53:51.126348   25245 controller.go:183] Successfully synced 'default/example-network'
```

可以看到，这一次，Informer 注册的“更新”事件被触发，更新后的 Network 对象的 Key 被添加到了工作队列之中。

所以，接下来控制循环从工作队列里拿到的 Network 对象，与前一个对象是不同的：它的 ResourceVersion 的值变成了 479062；而 Spec 里的字段，则变成了 192.168.1.0/16 网段。

最后，我再把这个对象删除掉：
```sh
$ kubectl delete -f example/example-network.yaml
```

这一次，在控制器的输出里，我们就可以看到，Informer 注册的“删除”事件被触发，并且控制循环“调用”Neutron API“删除”了真实环境里的网络。这个输出如下所示：
```log
W0915 12:54:09.738464   25245 controller.go:212] Network: default/example-network does not exist in local cache, will delete it from Neutron ...
I0915 12:54:09.738832   25245 controller.go:215] [Neutron] Deleting network: default/example-network ...
I0915 12:54:09.738854   25245 controller.go:183] Successfully synced 'default/example-network'
```

以上，就是编写和使用自定义控制器的全部流程了。

**实际上，这套流程不仅可以用在自定义 API 资源上，也完全可以用在 Kubernetes 原生的默认 API 对象上。**

比如，我们在 main 函数里，除了创建一个 Network Informer 外，还可以初始化一个 Kubernetes 默认 API 对象的 Informer 工厂，比如 Deployment 对象的 Informer。这个具体做法如下所示：
```go
func main() {
  ...
  
  kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
  
  controller := NewController(kubeClient, exampleClient,
  kubeInformerFactory.Apps().V1().Deployments(),
  networkInformerFactory.Samplecrd().V1().Networks())
  
  go kubeInformerFactory.Start(stopCh)
  ...
}
```

在这段代码中，我们首先使用 Kubernetes 的 client（kubeClient）创建了一个工厂；

然后，我用跟 Network 类似的处理方法，生成了一个 Deployment Informer；

接着，我把 Deployment Informer 传递给了自定义控制器；当然，我也要调用 Start 方法来启动这个 Deployment Informer。

而有了这个 Deployment Informer 后，这个控制器也就持有了所有 Deployment 对象的信息。接下来，它既可以通过 deploymentInformer.Lister() 来获取 Etcd 里的所有 Deployment 对象，也可以为这个 Deployment Informer 注册具体的 Handler 来。

**更重要的是，这就使得在这个自定义控制器里面，我可以通过对自定义 API 对象和默认 API 对象进行协同，从而实现更加复杂的编排功能。**

比如：用户每创建一个新的 Deployment，这个自定义控制器，就可以为它创建一个对应的 Network 供它使用。

这些对 Kubernetes API 编程范式的更高级应用，我就留给你在实际的场景中去探索和实践了。

#### 总结
在今天这小节中，我剖析了 Kubernetes API 编程范式的具体原理，并编写了一个自定义控制器。这其中，有如下几个概念和机制，是你一定要理解清楚的：

所谓的 Informer，就是一个自带缓存和索引机制，可以触发 Handler 的客户端库。这个本地缓存在 Kubernetes 中一般被称为 Store，索引一般被称为 Index。

Informer 使用了 Reflector 包，它是一个可以通过 ListAndWatch 机制获取并监视 API 对象变化的客户端封装。

Reflector 和 Informer 之间，用到了一个“增量先进先出队列”进行协同。而 Informer 与你要编写的控制循环之间，则使用了一个工作队列来进行协同。

在实际应用中，除了控制循环之外的所有代码，实际上都是 Kubernetes 为你自动生成的，即：pkg/client/{informers, listers, clientset}里的内容。

而这些自动生成的代码，就为我们提供了一个可靠而高效地获取 API 对象“期望状态”的编程库。

所以，接下来，作为开发者，你就只需要关注如何拿到“实际状态”，然后如何拿它去跟“期望状态”做对比，从而决定接下来要做的业务逻辑即可。

以上内容，就是 Kubernetes API 编程范式的核心思想。

#### 请思考一下，为什么 Informer 和你编写的控制循环之间，一定要使用一个工作队列来进行协作呢？
Informer 和控制循环分开是为了解耦，防止控制循环执行过慢把Informer 拖死


## # 基于角色的权限控制：RBAC
在前面的几个小节中，已经熟悉了很多种 Kubernetes 内置的编排对象，以及对应的控制器模式的实现原理。此外，还剖析了自定义 API 资源类型和控制器的编写方式。

这时候，你可能已经冒出了这样一个想法：控制器模式看起来好像也不难嘛，我能不能自己写一个编排对象呢？

**答案当然是可以的。而且，这才是 Kubernetes 项目最具吸引力的地方。**

毕竟，在互联网级别的大规模集群里，Kubernetes 内置的编排对象，很难做到完全满足所有需求。所以，很多实际的容器化工作，都会要求你设计一个自己的编排对象，实现自己的控制器模式。

而在 Kubernetes 项目里，我们可以基于插件机制来完成这些工作，而完全不需要修改任何一行代码。

不过，你要通过一个外部插件，在 Kubernetes 里新增和操作 API 对象，那么就必须先了解一个非常重要的知识：RBAC。

我们知道，Kubernetes 中所有的 API 对象，都保存在 Etcd 里。可是，对这些 API 对象的操作，却一定都是通过访问 kube-apiserver 实现的。其中一个非常重要的原因，就是你需要 APIServer 来帮助你做授权工作。

**而在 Kubernetes 项目中，负责完成授权（Authorization）工作的机制，就是 RBAC：基于角色的访问控制（Role-Based Access Control）。**

如果你直接查看 Kubernetes 项目中关于 RBAC 的文档的话，可能会感觉非常复杂。但实际上，等到你用到这些 RBAC 的细节时，再去查阅也不迟。

**而在这里，我只希望你能明确三个最基本的概念。**
1. Role：角色，它其实是一组规则，定义了一组对 Kubernetes API 对象的操作权限。
2. Subject：被作用者，既可以是“人”，也可以是“机器”，也可以是你在 Kubernetes 里定义的“用户”。
3. RoleBinding：定义了“被作用者”和“角色”的绑定关系。

而这三个概念，其实就是整个 RBAC 体系的核心所在。

#### 先来了解一下 Role
实际上，Role 本身就是一个 Kubernetes 的 API 对象，定义如下所示：
```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: mynamespace
  name: example-role
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
```
首先，这个 Role 对象指定了它能产生作用的 Namepace 是：mynamespace。

Namespace 是 Kubernetes 项目里的一个逻辑管理单位。不同 Namespace 的 API 对象，在通过 kubectl 命令进行操作的时候，是互相隔离开的。

比如，kubectl get pods -n mynamespace。

当然，这仅限于逻辑上的“隔离”，Namespace 并不会提供任何实际的隔离或者多租户能力。而没有指定 Namespace，那就是使用的是默认 Namespace：default。

然后，这个 Role 对象的 rules 字段，就是它所定义的权限规则。在上面的例子里，这条规则的含义就是：允许“被作用者”，对 mynamespace 下面的 Pod 对象，进行 GET、WATCH 和 LIST 操作。

那么，这个具体的“被作用者”又是如何指定的呢？这就需要通过 RoleBinding 来实现了。

当然，RoleBinding 本身也是一个 Kubernetes 的 API 对象。它的定义如下所示：
```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: example-rolebinding
  namespace: mynamespace
subjects:
- kind: User
  name: example-user
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: example-role
  apiGroup: rbac.authorization.k8s.io
```
可以看到，这个 RoleBinding 对象里定义了一个 subjects 字段，即“被作用者”。它的类型是 User，即 Kubernetes 里的用户。这个用户的名字是 example-user。

可是，在 Kubernetes 中，其实并没有一个叫作“User”的 API 对象。而且，我们在前面和部署使用 Kubernetes 的流程里，既不需要 User，也没有创建过 User。

这个 User 到底是从哪里来的呢？

**实际上，Kubernetes 里的“User”，也就是“用户”，只是一个授权系统里的逻辑概念。它需要通过外部认证服务，比如 Keystone，来提供。或者，你也可以直接给 APIServer 指定一个用户名、密码文件。那么 Kubernetes 的授权系统，就能够从这个文件里找到对应的“用户”了。当然，在大多数私有的使用环境中，我们只要使用 Kubernetes 提供的内置“用户”，就足够了。这部分知识，后面马上会讲到。**

接下来，我们会看到一个 roleRef 字段。正是通过这个字段，RoleBinding 对象就可以直接通过名字，来引用我们前面定义的 Role 对象（example-role），从而定义了“被作用者（Subject）”和“角色（Role）”之间的绑定关系。

需要再次提醒的是，Role 和 RoleBinding 对象都是 Namespaced 对象（Namespaced Object），它们对权限的限制规则仅在它们自己的 Namespace 内有效，roleRef 也只能引用当前 Namespace 里的 Role 对象。

那么，对于非 Namespaced（Non-namespaced）对象（比如：Node），或者，某一个 Role 想要作用于所有的 Namespace 的时候，我们又该如何去做授权呢？

这时候，我们就必须要使用 ClusterRole 和 ClusterRoleBinding 这两个组合了。这两个 API 对象的用法跟 Role 和 RoleBinding 完全一样。只不过，它们的定义里，没有了 Namespace 字段，如下所示：
```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: example-clusterrole
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
```

```yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: example-clusterrolebinding
subjects:
- kind: User
  name: example-user
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: example-clusterrole
  apiGroup: rbac.authorization.k8s.io
```
上面的例子里的 ClusterRole 和 ClusterRoleBinding 的组合，意味着名叫 example-user 的用户，拥有对所有 Namespace 里的 Pod 进行 GET、WATCH 和 LIST 操作的权限。

更进一步地，在 Role 或者 ClusterRole 里面，如果要赋予用户 example-user 所有权限，那你就可以给它指定一个 verbs 字段的全集，如下所示：
```yaml
verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

这些就是当前 Kubernetes（v1.11）里能够对 API 对象进行的所有操作了。

类似地，Role 对象的 rules 字段也可以进一步细化。比如，你可以只针对某一个具体的对象进行权限设置，如下所示：
```yaml
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["my-config"]
  verbs: ["get"]
```
这个例子就表示，这条规则的“被作用者”，只对名叫“my-config”的 ConfigMap 对象，有进行 GET 操作的权限。

而正如我前面介绍过的，在大多数时候，我们其实都不太使用“用户”这个功能，而是直接使用 Kubernetes 里的“内置用户”。

这个由 Kubernetes 负责管理的“内置用户”，正是我们前面曾经提到过的：ServiceAccount。

接下来，我通过一个具体的实例来为你讲解一下为 ServiceAccount 分配权限的过程。

首先，我们要定义一个 ServiceAccount。它的 API 对象非常简单，如下所示：
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: mynamespace
  name: example-sa
```
可以看到，一个最简单的 ServiceAccount 对象只需要 Name 和 Namespace 这两个最基本的字段。

然后，我们通过编写 RoleBinding 的 YAML 文件，来为这个 ServiceAccount 分配权限：
```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: example-rolebinding
  namespace: mynamespace
subjects:
- kind: ServiceAccount
  name: example-sa
  namespace: mynamespace
roleRef:
  kind: Role
  name: example-role
  apiGroup: rbac.authorization.k8s.io
```
**可以看到，在这个 RoleBinding 对象里，subjects 字段的类型（kind），不再是一个 User，而是一个名叫 example-sa 的 ServiceAccount。而 roleRef 引用的 Role 对象，依然名叫 example-role，也就是我在这篇文章一开始定义的 Role 对象。**

接着，我们用 kubectl 命令创建这三个对象：
```sh
$ kubectl create -f svc-account.yaml
$ kubectl create -f role-binding.yaml
$ kubectl create -f role.yaml
```

然后，我们来查看一下这个 ServiceAccount 的详细信息：
```sh
$ kubectl get sa -n mynamespace -o yaml
- apiVersion: v1
  kind: ServiceAccount
  metadata:
    creationTimestamp: 2018-09-08T12:59:17Z
    name: example-sa
    namespace: mynamespace
    resourceVersion: "409327"
    ...
  secrets:
  - name: example-sa-token-vmfg6
```
可以看到，Kubernetes 会为一个 ServiceAccount 自动创建并分配一个 Secret 对象，即：上述 ServiceAcount 定义里最下面的 secrets 字段。

这个 Secret，就是这个 ServiceAccount 对应的、用来跟 APIServer 进行交互的授权文件，我们一般称它为：Token。Token 文件的内容一般是证书或者密码，它以一个 Secret 对象的方式保存在 Etcd 当中。

这时候，用户的 Pod，就可以声明使用这个 ServiceAccount 了，比如下面这个例子：
```yaml
apiVersion: v1
kind: Pod
metadata:
  namespace: mynamespace
  name: sa-token-test
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
  serviceAccountName: example-sa
```
在这个例子里，我定义了 Pod 要使用的 ServiceAccount 的名字是：example-sa。

等这个 Pod 运行起来之后，我们就可以看到，该 ServiceAccount 的 token，也就是一个 Secret 对象，被 Kubernetes 自动挂载到了容器的 /var/run/secrets/kubernetes.io/serviceaccount 目录下，如下所示：
```sh
$ kubectl describe pod sa-token-test -n mynamespace
Name:               sa-token-test
Namespace:          mynamespace
...
Containers:
  nginx:
    ...
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from example-sa-token-vmfg6 (ro)
```

这时候，我们可以通过 kubectl exec 查看到这个目录里的文件：
```sh
$ kubectl exec -it sa-token-test -n mynamespace -- /bin/bash
root@sa-token-test:/# ls /var/run/secrets/kubernetes.io/serviceaccount
ca.crt namespace  token
```
如上所示，容器里的应用，就可以使用这个 ca.crt 来访问 APIServer 了。更重要的是，此时它只能够做 GET、WATCH 和 LIST 操作。因为 example-sa 这个 ServiceAccount 的权限，已经被我们绑定了 Role 做了限制。

此外，我在 Pod 使用进阶小节中曾经提到过，如果一个 Pod 没有声明 serviceAccountName，Kubernetes 会自动在它的 Namespace 下创建一个名叫 default 的默认 ServiceAccount，然后分配给这个 Pod。

但在这种情况下，这个默认 ServiceAccount 并没有关联任何 Role。也就是说，此时它有访问 APIServer 的绝大多数权限。当然，这个访问所需要的 Token，还是默认 ServiceAccount 对应的 Secret 对象为它提供的，如下所示。
```sh
$kubectl describe sa default
Name:                default
Namespace:           default
Labels:              <none>
Annotations:         <none>
Image pull secrets:  <none>
Mountable secrets:   default-token-s8rbq
Tokens:              default-token-s8rbq
Events:              <none>

$ kubectl get secret
NAME                  TYPE                                  DATA      AGE
kubernetes.io/service-account-token   3         82d

$ kubectl describe secret default-token-s8rbq
Name:         default-token-s8rbq
Namespace:    default
Labels:       <none>
Annotations:  kubernetes.io/service-account.name=default
              kubernetes.io/service-account.uid=ffcb12b2-917f-11e8-abde-42010aa80002

Type:  kubernetes.io/service-account-token

Data
====
ca.crt:     1025 bytes
namespace:  7 bytes
token:      <TOKEN数据>
```
可以看到，Kubernetes 会自动为默认 ServiceAccount 创建并绑定一个特殊的 Secret：它的类型是kubernetes.io/service-account-token；它的 Annotation 字段，声明了kubernetes.io/service-account.name=default，即这个 Secret 会跟同一 Namespace 下名叫 default 的 ServiceAccount 进行绑定。

**所以，在生产环境中，我强烈建议你为所有 Namespace 下的默认 ServiceAccount，绑定一个只读权限的 Role。（相当于限制所有使用默认ServiceAccount的用户，只拥有只读权限。）**

除了前面使用的“用户”（User），Kubernetes 还拥有“用户组”（Group）的概念，也就是一组“用户”的意思。如果你为 Kubernetes 配置了外部认证服务的话，这个“用户组”的概念就会由外部认证服务提供。

而对于 Kubernetes 的内置“用户”ServiceAccount 来说，上述“用户组”的概念也同样适用。

实际上，一个 ServiceAccount，在 Kubernetes 里对应的“用户”的名字是：```system:serviceaccount:<Namespace名字>:<ServiceAccount名字>```（SA具有特定命名的用户名称，也有用户组名称，特别的同一个NameSpace下的用户同属一组）

而它对应的内置“用户组”的名字，就是：```system:serviceaccounts:<Namespace名字>```

**这两个对应关系，请你一定要牢记。**

比如，现在我们可以在 RoleBinding 里定义如下的 subjects：
```yaml
subjects:
- kind: Group
  name: system:serviceaccounts:mynamespace
  apiGroup: rbac.authorization.k8s.io
```
这就意味着这个 Role 的权限规则，作用于 mynamespace 里的所有 ServiceAccount。这就用到了“用户组”的概念。

而下面这个例子：
```yaml
subjects:
- kind: Group
  name: system:serviceaccounts
  apiGroup: rbac.authorization.k8s.io
```
就意味着这个 Role 的权限规则，作用于整个系统里的所有 ServiceAccount。

**最后，值得一提的是，在 Kubernetes 中已经内置了很多个为系统保留的 ClusterRole，它们的名字都以 system: 开头。你可以通过 kubectl get clusterroles 查看到它们。**
```sh
[root@k8s-node2 ~]# kubectl get clusterroles
NAME                                                                   CREATED AT
admin                                                                  2023-05-28T07:55:02Z
calico-cni-plugin                                                      2023-05-28T08:09:53Z
calico-crds                                                            2023-05-28T08:13:57Z
calico-extension-apiserver-auth-access                                 2023-05-28T08:13:58Z
calico-kube-controllers                                                2023-05-28T08:09:53Z
calico-node                                                            2023-05-28T08:09:53Z
calico-typha                                                           2023-05-28T08:09:52Z
calico-webhook-reader                                                  2023-05-28T08:13:58Z
cluster-admin                                                          2023-05-28T07:55:02Z
edit                                                                   2023-05-28T07:55:02Z
kubeadm:get-nodes                                                      2023-05-28T07:55:04Z
kubernetes-dashboard                                                   2023-06-29T08:18:32Z
system:aggregate-to-admin                                              2023-05-28T07:55:02Z
system:aggregate-to-edit                                               2023-05-28T07:55:02Z
system:aggregate-to-view                                               2023-05-28T07:55:02Z
system:auth-delegator                                                  2023-05-28T07:55:02Z
system:basic-user                                                      2023-05-28T07:55:02Z
system:certificates.k8s.io:certificatesigningrequests:nodeclient       2023-05-28T07:55:02Z
system:certificates.k8s.io:certificatesigningrequests:selfnodeclient   2023-05-28T07:55:02Z
system:certificates.k8s.io:kube-apiserver-client-approver              2023-05-28T07:55:02Z
system:certificates.k8s.io:kube-apiserver-client-kubelet-approver      2023-05-28T07:55:02Z
system:certificates.k8s.io:kubelet-serving-approver                    2023-05-28T07:55:02Z
system:certificates.k8s.io:legacy-unknown-approver                     2023-05-28T07:55:02Z
system:controller:attachdetach-controller                              2023-05-28T07:55:02Z
system:controller:certificate-controller                               2023-05-28T07:55:02Z
system:controller:clusterrole-aggregation-controller                   2023-05-28T07:55:02Z
system:controller:cronjob-controller                                   2023-05-28T07:55:02Z
system:controller:daemon-set-controller                                2023-05-28T07:55:02Z
system:controller:deployment-controller                                2023-05-28T07:55:02Z
system:controller:disruption-controller                                2023-05-28T07:55:02Z
system:controller:endpoint-controller                                  2023-05-28T07:55:02Z
system:controller:endpointslice-controller                             2023-05-28T07:55:02Z
system:controller:endpointslicemirroring-controller                    2023-05-28T07:55:02Z
system:controller:ephemeral-volume-controller                          2023-05-28T07:55:02Z
system:controller:expand-controller                                    2023-05-28T07:55:02Z
system:controller:generic-garbage-collector                            2023-05-28T07:55:02Z
system:controller:horizontal-pod-autoscaler                            2023-05-28T07:55:02Z
system:controller:job-controller                                       2023-05-28T07:55:02Z
system:controller:namespace-controller                                 2023-05-28T07:55:02Z
system:controller:node-controller                                      2023-05-28T07:55:02Z
system:controller:persistent-volume-binder                             2023-05-28T07:55:02Z
system:controller:pod-garbage-collector                                2023-05-28T07:55:02Z
system:controller:pv-protection-controller                             2023-05-28T07:55:02Z
system:controller:pvc-protection-controller                            2023-05-28T07:55:02Z
system:controller:replicaset-controller                                2023-05-28T07:55:02Z
system:controller:replication-controller                               2023-05-28T07:55:02Z
system:controller:resourcequota-controller                             2023-05-28T07:55:02Z
system:controller:root-ca-cert-publisher                               2023-05-28T07:55:02Z
system:controller:route-controller                                     2023-05-28T07:55:02Z
system:controller:service-account-controller                           2023-05-28T07:55:02Z
system:controller:service-controller                                   2023-05-28T07:55:02Z
system:controller:statefulset-controller                               2023-05-28T07:55:02Z
system:controller:ttl-after-finished-controller                        2023-05-28T07:55:02Z
system:controller:ttl-controller                                       2023-05-28T07:55:02Z
system:coredns                                                         2023-05-28T07:55:04Z
system:discovery                                                       2023-05-28T07:55:02Z
system:heapster                                                        2023-05-28T07:55:02Z
system:kube-aggregator                                                 2023-05-28T07:55:02Z
system:kube-controller-manager                                         2023-05-28T07:55:02Z
system:kube-dns                                                        2023-05-28T07:55:02Z
system:kube-scheduler                                                  2023-05-28T07:55:02Z
system:kubelet-api-admin                                               2023-05-28T07:55:02Z
system:monitoring                                                      2023-05-28T07:55:02Z
system:node                                                            2023-05-28T07:55:02Z
system:node-bootstrapper                                               2023-05-28T07:55:02Z
system:node-problem-detector                                           2023-05-28T07:55:02Z
system:node-proxier                                                    2023-05-28T07:55:02Z
system:persistent-volume-provisioner                                   2023-05-28T07:55:02Z
system:public-info-viewer                                              2023-05-28T07:55:02Z
system:service-account-issuer-discovery                                2023-05-28T07:55:02Z
system:volume-scheduler                                                2023-05-28T07:55:02Z
tigera-operator                                                        2023-05-28T08:07:55Z
view                                                                   2023-05-28T07:55:02Z
[root@k8s-node2 ~]#
```

一般来说，这些系统 ClusterRole，是绑定给 Kubernetes 系统组件对应的 ServiceAccount 使用的。

比如，其中一个名叫 system:kube-scheduler 的 ClusterRole，定义的权限规则是 kube-scheduler（Kubernetes 的调度器组件）运行所需要的必要权限。你可以通过如下指令查看这些权限的列表：
```sh
[root@k8s-node2 ~]# kubectl describe clusterrole system:kube-scheduler
Name:         system:kube-scheduler
Labels:       kubernetes.io/bootstrapping=rbac-defaults
Annotations:  rbac.authorization.kubernetes.io/autoupdate: true
PolicyRule:
  Resources                                  Non-Resource URLs  Resource Names    Verbs
  ---------                                  -----------------  --------------    -----
  events                                     []                 []                [create patch update]
  events.events.k8s.io                       []                 []                [create patch update]
  bindings                                   []                 []                [create]
  endpoints                                  []                 []                [create]
  pods/binding                               []                 []                [create]
  tokenreviews.authentication.k8s.io         []                 []                [create]
  subjectaccessreviews.authorization.k8s.io  []                 []                [create]
  leases.coordination.k8s.io                 []                 []                [create]
  pods                                       []                 []                [delete get list watch]
  namespaces                                 []                 []                [get list watch]
  nodes                                      []                 []                [get list watch]
  persistentvolumeclaims                     []                 []                [get list watch]
  persistentvolumes                          []                 []                [get list watch]
  replicationcontrollers                     []                 []                [get list watch]
  services                                   []                 []                [get list watch]
  replicasets.apps                           []                 []                [get list watch]
  statefulsets.apps                          []                 []                [get list watch]
  replicasets.extensions                     []                 []                [get list watch]
  poddisruptionbudgets.policy                []                 []                [get list watch]
  csidrivers.storage.k8s.io                  []                 []                [get list watch]
  csinodes.storage.k8s.io                    []                 []                [get list watch]
  csistoragecapacities.storage.k8s.io        []                 []                [get list watch]
  endpoints                                  []                 [kube-scheduler]  [get update]
  leases.coordination.k8s.io                 []                 [kube-scheduler]  [get update]
  pods/status                                []                 []                [patch update]
[root@k8s-node2 ~]#
```
这个 system:kube-scheduler 的 ClusterRole，就会被绑定给 kube-system Namesapce 下名叫 kube-scheduler 的 ServiceAccount，它正是 Kubernetes 调度器的 Pod 声明使用的 ServiceAccount。

除此之外，Kubernetes 还提供了四个预先定义好的 ClusterRole 来供用户直接使用：
1. cluster-admin；
2. admin；
3. edit；
4. view。

通过它们的名字，能大致猜出它们都定义了哪些权限。比如，这个名叫 view 的 ClusterRole，就规定了被作用者只有 Kubernetes API 的只读权限。

**还要提醒你的是，上面这个 cluster-admin 角色，对应的是整个 Kubernetes 项目中的最高权限（verbs=*），如下所示：**
```sh
[root@k8s-node2 ~]# kubectl describe clusterrole cluster-admin -n kube-system
Name:         cluster-admin
Labels:       kubernetes.io/bootstrapping=rbac-defaults        
Annotations:  rbac.authorization.kubernetes.io/autoupdate: true
PolicyRule:
  Resources  Non-Resource URLs  Resource Names  Verbs
  ---------  -----------------  --------------  -----
  *.*        []                 []              [*]
             [*]                []              [*]
[root@k8s-node2 ~]#
```

**所以，请你务必要谨慎而小心地使用 cluster-admin。**

#### 总结
主要了解了基于角色的访问控制（RBAC）。

其实，你现在已经能够理解，所谓角色（Role），其实就是一组权限规则列表。而我们分配这些权限的方式，就是通过创建 RoleBinding 对象，将被作用者（subject）和权限列表进行绑定。

另外，与之对应的 ClusterRole 和 ClusterRoleBinding，则是 Kubernetes 集群级别的 Role 和 RoleBinding，它们的作用范围不受 Namespace 限制。

而尽管权限的被作用者可以有很多种（比如，User、Group 等），但在我们平常的使用中，最普遍的用法还是 ServiceAccount。所以，Role + RoleBinding + ServiceAccount 的权限分配方式是你要重点掌握的内容。我们在后面编写和安装各种插件的时候，会经常用到这个组合。

#### 请问，如何为所有 Namespace 下的默认 ServiceAccount（default ServiceAccount），绑定一个只读权限的 Role 呢？请你提供 ClusterRoleBinding（或者 RoleBinding）的 YAML 文件。

```yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: readonly-all-default
subjects:
- kind: ServiceAccount
  name: default
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: view
  apiGroup: rbac.authorization.k8s.io
```


## # 聪明的微创新：Operator工作原理解读
在前面已经分享了 Kubernetes 项目中的大部分编排对象（比如 Deployment、StatefulSet、DaemonSet，以及 Job），也介绍了“有状态应用”的管理方法，还阐述了为 Kubernetes 添加自定义 API 对象和编写自定义控制器的原理和流程。

可能你已经感觉到，在 Kubernetes 中，管理“有状态应用”是一个比较复杂的过程，尤其是编写 Pod 模板的时候，总有一种“在 YAML 文件里编程序”的感觉，让人很不舒服。

**而在 Kubernetes 生态中，还有一个相对更加灵活和编程友好的管理“有状态应用”的解决方案，它就是：Operator。**

接下来，就以 Etcd Operator 为例，来了解一下 Operator 的工作原理和编写方法。

Etcd Operator 的使用方法非常简单，只需要两步即可完成：
1. 第一步，将这个 Operator 的代码 Clone 到本地：
```sh
$ git clone https://github.com/coreos/etcd-operator
```

2. 第二步，将这个 Etcd Operator 部署在 Kubernetes 集群里。

不过，在部署 Etcd Operator 的 Pod 之前，你需要先执行这样一个脚本：
```sh
$ example/rbac/create_role.sh
```
不用我多说你也能够明白：这个脚本的作用，就是为 Etcd Operator 创建 RBAC 规则。这是因为，Etcd Operator 需要访问 Kubernetes 的 APIServer 来创建对象。

更具体地说，上述脚本为 Etcd Operator 定义了如下所示的权限：
   1. 对 Pod、Service、PVC、Deployment、Secret 等 API 对象，有所有权限；
   2. 对 CRD 对象，有所有权限；
   3. 对属于 etcd.database.coreos.com 这个 API Group 的 CR（Custom Resource）对象，有所有权限。

而 Etcd Operator 本身，其实就是一个 Deployment，它的 YAML 文件如下所示：
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: etcd-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: etcd-operator
    spec:
      containers:
      - name: etcd-operator
        image: quay.io/coreos/etcd-operator:v0.9.2
        command:
        - etcd-operator
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
...
```

所以，我们就可以使用上述的 YAML 文件来创建 Etcd Operator，如下所示：
```sh
$ kubectl create -f example/deployment.yaml
```

而一旦 Etcd Operator 的 Pod 进入了 Running 状态，你就会发现，有一个 CRD 被自动创建了出来，如下所示：
```sh
$ kubectl get pods
NAME                              READY     STATUS      RESTARTS   AGE
etcd-operator-649dbdb5cb-bzfzp    1/1       Running     0          20s

$ kubectl get crd
NAME                                    CREATED AT
etcdclusters.etcd.database.coreos.com   2018-09-18T11:42:55Z
```

这个 CRD 名叫etcdclusters.etcd.database.coreos.com 。你可以通过 kubectl describe 命令看到它的细节，如下所示：
```sh
$ kubectl describe crd  etcdclusters.etcd.database.coreos.com
...
Group:   etcd.database.coreos.com
  Names:
    Kind:       EtcdCluster
    List Kind:  EtcdClusterList
    Plural:     etcdclusters
    Short Names:
      etcd
    Singular:  etcdcluster
  Scope:       Namespaced
  Version:     v1beta2
  
...
```
可以看到，这个 CRD 相当于告诉了 Kubernetes：接下来，如果有 API 组（Group）是etcd.database.coreos.com、API 资源类型（Kind）是“EtcdCluster”的 YAML 文件被提交上来，你可一定要认识啊。

所以说，通过上述两步操作，你实际上是在 Kubernetes 里添加了一个名叫 EtcdCluster 的自定义资源类型。而 Etcd Operator 本身，就是这个自定义资源类型对应的自定义控制器。

而当 Etcd Operator 部署好之后，接下来在这个 Kubernetes 里创建一个 Etcd 集群的工作就非常简单了。你只需要编写一个 EtcdCluster 的 YAML 文件，然后把它提交给 Kubernetes 即可，如下所示：
```sh
$ kubectl apply -f example/example-etcd-cluster.yaml
```
这个 example-etcd-cluster.yaml 文件里描述的，是一个 3 个节点的 Etcd 集群。我们可以看到它被提交给 Kubernetes 之后，就会有三个 Etcd 的 Pod 运行起来，如下所示：
```sh
$ kubectl get pods
NAME                            READY     STATUS    RESTARTS   AGE
example-etcd-cluster-dp8nqtjznc   1/1       Running     0          1m
example-etcd-cluster-mbzlg6sd56   1/1       Running     0          2m
example-etcd-cluster-v6v6s6stxd   1/1       Running     0          2m
```

那么，究竟发生了什么，让创建一个 Etcd 集群的工作如此简单呢？

我们当然还是得从这个 example-etcd-cluster.yaml 文件开始说起。

不难想到，这个文件里定义的，正是 EtcdCluster 这个 CRD 的一个具体实例，也就是一个 Custom Resource（CR）。而它的内容非常简单，如下所示：
```yaml
apiVersion: "etcd.database.coreos.com/v1beta2"
kind: "EtcdCluster"
metadata:
  name: "example-etcd-cluster"
spec:
  size: 3
  version: "3.2.13"
```

可以看到，EtcdCluster 的 spec 字段非常简单。其中，size=3 指定了它所描述的 Etcd 集群的节点个数。而 version=“3.2.13”，则指定了 Etcd 的版本，仅此而已。

而真正把这样一个 Etcd 集群创建出来的逻辑，就是 Etcd Operator 要实现的主要工作了。

看到这里，相信你应该已经对 Operator 有了一个初步的认知：

**Operator 的工作原理，实际上是利用了 Kubernetes 的自定义 API 资源（CRD），来描述我们想要部署的“有状态应用”；然后在自定义控制器里，根据自定义 API 对象的变化，来完成具体的部署和运维工作。**

