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