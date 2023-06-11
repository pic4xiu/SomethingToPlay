def chu(fillname):
    with open(fillname, "rb") as f:
        content = f.read()
    # print(content)
    # 将每两个字符组成一个十六进制数，并用空格分隔
    hex = " ".join("{:02x}".format(c) for c in content)

    # 将空格分隔的十六进制数合并成一行
    hex = hex.replace(" ", "")

    # 将每8个字符分隔成一组，并用逗号分隔
    hex = ",".join(hex[i:i+2] for i in range(0, len(hex), 2))

    # 将结果存储到一个数组中
    arr = hex.split(",")
    ans=[]
    for i in arr:
        ans.append(int(i,16))
    return ans

chu('Ransom.WannaCryptor.exe')