import os

# 定义一个函数，将 bytes 文件转换成图片
def bytes_to_image(file_path, output_folder):
    # 读取文件内容
    with open(file_path, 'r') as f:
        content = f.read()

    # 将文件内容转换成像素点的亮度
    pixels = []
    for line in content.split('\n'):
        for byte in line.split()[1:]:
            if byte=='??':continue
            pixels.append(int(byte, 16))

    # 计算图片高度
    print(len(pixels)//1024)
    width = next(w for s, w in widths if len(pixels) < s)
    height = len(pixels) // width
    if len(pixels) % width != 0:
        height += 1

    # 创建图片
    from PIL import Image
    img = Image.new('L', (width, height), 0)
    img.putdata(pixels)

    # 保存图片
    file_name = os.path.splitext(os.path.basename(file_path))[0] + '.png'
    output_path = os.path.join(output_folder, file_name)
    img.save(output_path)

# 定义一个列表，用于存储不同文件大小对应的图片宽度
widths = [(10 * 1024, 32), (30 * 1024, 64), (60 * 1024, 128),(100 * 1024, 256),(200 * 1024, 384),(500 * 1024, 512),(1000 * 1024, 768),(float('inf'), 1024)]

# 遍历文件夹，将所有 .bytes 文件转换成图片
input_folder = 'train'
output_folder = 'rtrainpng'
for file_name in os.listdir(input_folder):
    if file_name.endswith('.bytes'):
        file_path = os.path.join(input_folder, file_name)

        # 读取文件内容，计算像素点数量
        with open(file_path, 'r') as f:
            content = f.read()
        pixels_count = sum(1 for line in content.split('\n') if line.startswith('004010'))  

        # 将文件转换成图片
        bytes_to_image(file_path, output_folder)