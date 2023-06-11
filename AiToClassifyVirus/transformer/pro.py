# 读取数据文件
with open('all_analysis_data.txt', 'r') as f:
    data = f.readlines()
# 将单词转换为独一无二的数字
word_to_id = {}
id_to_word = {}
id = 0
for line in data:
    words = line.strip().split()
    for word in words:
        if word not in word_to_id:
            word_to_id[word] = id
            id_to_word[id] = word
            id += 1
# 输出单词和对应的数字
for word, id in word_to_id.items():
    print(f'{word}: {id}')
# ...从零开始的
# getusernameexa: 269
# netusergetlocalgroups: 270
# findwindowexw: 271
# deleteurlcacheentryw: 272
# rtlcreateuserthread: 273
# setinformationjobobject: 274
# cryptprotectmemory: 275
# cryptunprotectmemory: 276
# findfirstfileexa: 277
# 将数据转换为数字序列
data_ids = []
for line in data:
    words = line.strip().split()
    ids = [word_to_id[word] for word in words]
    data_ids.append(ids)


# 将数字序列保存到文件中
with open('data_ids.txt', 'w') as f:
    for ids in data_ids:
        f.write(' '.join(str(id) for id in ids))
        f.write('\n')