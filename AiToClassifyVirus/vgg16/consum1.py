import pika
import hashlib
import random
import mysql.connector
import torch
import torch.nn as nn
from torchvision import transforms
from PIL import Image
from torchvision import models, transforms
# 定义数据增强
transform = transforms.Compose([
    transforms.Grayscale(num_output_channels=3),
    transforms.ColorJitter(brightness=0.5, contrast=0.5, saturation=0.5, hue=0.5),
    transforms.Resize((224, 224)),
    transforms.RandomHorizontalFlip(),
    transforms.ToTensor(),
    transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225])
])


# 定义 VGG16 模型
class VGG16(nn.Module):
    def __init__(self, num_classes=9):
        super(VGG16, self).__init__()
        self.features = models.vgg16(pretrained=True).features
        self.avgpool = nn.AdaptiveAvgPool2d((7, 7))
        self.classifier = nn.Sequential(
            nn.Linear(512 * 7 * 7, 4096),
            nn.ReLU(inplace=True),
            nn.Dropout(),
            nn.Linear(4096, 4096),
            nn.ReLU(inplace=True),
            nn.Dropout(),
            nn.Linear(4096, num_classes),
        )

    def forward(self, x):
        x = self.features(x)
        x = self.avgpool(x)
        x = torch.flatten(x, 1)
        x = self.classifier(x)
        return x


# 加载 PyTorch 模型状态字典
state_dict = torch.load('model.pth')

# 创建 PyTorch 模型实例，并将状态字典加载到模型中
model = VGG16()
model.load_state_dict(state_dict)

# 将模型设置为评估模式
model.eval()




# 连接到MySQL数据库
mydb = mysql.connector.connect(
	host="127.0.0.1",
	user="root",
	password="123456",
	database="mydatabase"
  )
mycursor = mydb.cursor()
# 声明回调函数
def callback(ch, method, properties, body):
    # 解析文件名和MD5值
    filename, md5 = body.decode().split("|")

    # 处理文件
    # with open("./uploads/" + filename, "rb") as f:
    # content = f.read()
    # result = hashlib.md5(content).hexdigest()
    # process_result = random.randint(0, 9)
    # 读取灰度图像
    # try:
    print("./uploads/"+filename)
    gray_img = Image.open("./uploads/"+filename)

    # 将灰度图像转换为彩色图像
    color_img = transform(gray_img)

    # 保存彩色图像到文件中
    transforms.ToPILImage()(color_img).save('color_image.png')

    input_img = Image.open('color_image.png')
    input_tensor = transform(input_img).unsqueeze(0)
    output = model(input_tensor)
    max_value, max_index = torch.max(output, dim=1)
    # print(max_index.item())
    process_result=max_index.item()
    # 插入数据到数据库
    # mycursor = mydb.cursor()
    sql = "INSERT INTO files (filename, md5, result) VALUES (%s, %s, %s)"
    val = (filename, md5, process_result)
    mycursor.execute(sql, val)
    mydb.commit()
    # 打印处理结果
    print("Processed file: %s, MD5: %s, Result: %d" % (filename, md5, process_result))

    # 确认消息已经处理完毕
    ch.basic_ack(delivery_tag=method.delivery_tag)

# 连接到RabbitMQ服务器
connection = pika.BlockingConnection(pika.ConnectionParameters("localhost"))
channel = connection.channel()

# 声明队列
channel.queue_declare(queue="file_queue", durable=True)

# 每次只消费一个消息
channel.basic_qos(prefetch_count=1)

# 注册回调函数
channel.basic_consume(queue="file_queue", on_message_callback=callback)

# 开始消费消息
print("Waiting for messages...")
channel.start_consuming()

# CREATE TABLE files (
#     id INT(11) NOT NULL AUTO_INCREMENT,
#     filename VARCHAR(255) NOT NULL,
#     md5 VARCHAR(32) NOT NULL,
#     result INT(11) NOT NULL,
#     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
#     PRIMARY KEY (id)
# ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;