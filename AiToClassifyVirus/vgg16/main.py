import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, DataLoader
from torchvision import models, transforms
from PIL import Image
import os
import pandas as pd

import os
import pandas as pd
import torch
from torch.utils.data import Dataset, DataLoader
from torchvision import transforms
from PIL import Image

import os
import pandas as pd
import torch
from torch.utils.data import Dataset, DataLoader
from torchvision import transforms
from PIL import Image

class MyDataset(Dataset):
    def __init__(self, csv_file, root_dir, transform=None, train=True, test_ratio=0.2):
        self.df = pd.read_csv(csv_file)
        self.root_dir = root_dir
        self.transform = transform
        self.train = train
        self.test_ratio = test_ratio

        # 获取文件夹中的文件名
        self.filenames = os.listdir(self.root_dir)
        self.filenames = [f for f in self.filenames if f.endswith('.png')]

        # 将数据集分成训练集和测试集
        if self.train:
            self.df = self.df[self.df['Id'].isin([f[:-4] for f in self.filenames])]
            self.df = self.df.sample(frac=1).reset_index(drop=True)
            self.test_size = int(len(self.df) * self.test_ratio)
            self.train_df = self.df.iloc[self.test_size:]
            self.test_df = self.df.iloc[:self.test_size]
            print(self.train_df)
        else:
            self.train_df = self.df[self.df['Id'].isin([f[:-4] for f in self.filenames])]
            print(self.train_df)

    def __len__(self):
        return len(self.train_df)

    def __getitem__(self, idx):
        if torch.is_tensor(idx):
            idx = idx.tolist()

        # 读取图片和标签
        img_name = self.train_df.iloc[idx, 0] + '.png'
        img_path = os.path.join(self.root_dir, img_name)
        image = Image.open(img_path)
        label = self.train_df.iloc[idx, 1]

        # 数据增强
        if self.transform:
            image = self.transform(image)

        return image, label

# 定义数据增强
transform = transforms.Compose([
    transforms.Grayscale(num_output_channels=3),
    transforms.ColorJitter(brightness=0.5, contrast=0.5, saturation=0.5, hue=0.5),
    transforms.Resize((224, 224)),
    transforms.RandomHorizontalFlip(),
    transforms.ToTensor(),
    transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225])
])


# 创建训练集和测试集
train_dataset = MyDataset('trainLabels.csv', 'rtrainpng', transform=transform, train=True)
test_dataset = MyDataset('trainLabels.csv', 'rtrainpng', transform=transform, train=False)

# 创建数据加载器
train_loader = DataLoader(train_dataset, batch_size=32, shuffle=True)
test_loader = DataLoader(test_dataset, batch_size=32, shuffle=False)

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

# 创建 VGG16 模型
model = VGG16(num_classes=9)

# 定义损失函数和优化器
criterion = nn.CrossEntropyLoss()
optimizer = optim.SGD(model.parameters(), lr=0.001, momentum=0.9)

# 训练模型
for epoch in range(10):
    running_loss = 0.0
    for i, data in enumerate(train_loader, 0):
        inputs, labels = data

        optimizer.zero_grad()

        outputs = model(inputs)
        loss = criterion(outputs, labels)
        loss.backward()
        optimizer.step()

        running_loss += loss.item()

        print('[%d, %5d] loss: %.3f' % (epoch + 1, i + 1, running_loss / 100))
        running_loss = 0.0

# 测试模型
correct = 0
total = 0
with torch.no_grad():
    for data in test_loader:
        images, labels = data
        outputs = model(images)
        _, predicted = torch.max(outputs.data, 1)
        total += labels.size(0)
        correct += (predicted == labels).sum().item()

print('Accuracy of the network on the test images: %d %%' % (100 * correct / total))