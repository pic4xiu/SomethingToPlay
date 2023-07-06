#!/bin/bash

# 设置文件夹路径
folder_path="/home/pic/下载/openjpeg/build/bin/tga/slimit1/hangs"

# 遍历文件夹中的所有文件
for file in "$folder_path"/*
do
    # 检查是否为文件
    if [ -f "$file" ]; then
        # 执行命令
        st=$(date +%s)
        echo "$file"
        opj_decompress -i "$file" -o te  > /dev/null 2>&1
        end=$(date +%s)
        ex=$((end - st))
        echo "执行了 $ex 秒"
    fi
done

