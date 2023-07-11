from bcc import BPF
b = BPF(text="""

struct syscall_read_logging
{
    long unsigned int buffer_addr;
    long int calling_size;
};
BPF_HASH(map_buff_addrs,  size_t,struct syscall_read_logging, 1024);
TRACEPOINT_PROBE(syscalls, sys_enter_read) {
    char comm[50];
    if(bpf_get_current_comm(&comm, 50)) {
        return 0;
    }
    const char *target_comm = "behooked";
    for (int i = 0; i < 9; i++)
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
	char *buff_addr;
        
	size_t pid_tgid = bpf_get_current_pid_tgid();
    const char *target_comm = "behooked";
    for (int i = 0; i < 9; i++)
    {
        if (comm[i] != target_comm[i])
        {
            return 0;
        }
    }
    char str[256];
   	struct syscall_read_logging *data= map_buff_addrs.lookup(&pid_tgid);
        if (data == 0) return 0;
    char hook[]="flag{true}";
    long int te=data->calling_size;
    long unsigned int tmpbuf=(long unsigned int)data->buffer_addr;
        if (te!=4096){
        return 0;
		}
    bpf_probe_write_user(tmpbuf, hook, 11); 
    return 0;
}
        """)
print("%-18s %-16s %-6s %s" % ("TIME(s)", "COMM", "PID", "message"))
while 1:
    try:
        (task, pid, cpu, flags, ts, msg) = b.trace_fields()
    except ValueError:
        continue
    print("%-18.9f %-16s %-6d %s" % (ts, task, pid, msg))
