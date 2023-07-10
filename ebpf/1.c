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