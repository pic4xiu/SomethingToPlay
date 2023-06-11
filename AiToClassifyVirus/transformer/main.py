# 导入必要的库
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import Dataset, DataLoader

# 定义数据预处理函数
def preprocess_data(data_file, label_file, max_seq_len):
    # 读取数据文件和标签文件
    with open(data_file, 'r') as f:
        data = f.readlines()
    with open(label_file, 'r') as f:
        labels = f.readlines()
    # 将数据和标签转换为数字序列
    data = [[int(x) for x in row.strip().split()] for row in data]
    labels = [int(x) for x in labels]
    # 将数据分成子序列并补齐
    sub_data = []
    sub_labels = []
    for i in range(len(data)):
        row = data[i]
        label = labels[i]
        for j in range(0, len(row), max_seq_len):
            sub_row = row[j:j+max_seq_len] + [278]*(max_seq_len-len(row[j:j+max_seq_len]))
            sub_data.append(sub_row)
            sub_labels.append(label)
    # 将标签转换为one-hot编码
    sub_labels = [torch.eye(8)[label] for label in sub_labels]
    # 返回数据和标签
    return sub_data, sub_labels

# 定义数据集类
class MyDataset(Dataset):
    def __init__(self, data, labels):
        self.data = data
        self.labels = labels
    def __len__(self):
        return len(self.data)
    def __getitem__(self, idx):
        return torch.tensor(self.data[idx]), self.labels[idx]
    # 定义Transformer模型类
class TransformerModel(nn.Module):
    def __init__(self, input_size, output_size, max_seq_len, num_layers=3, hidden_size=128, num_heads=8, dropout=0.1):
        super(TransformerModel, self).__init__()
        self.embedding = nn.Embedding(input_size, hidden_size)
        self.positional_encoding = nn.Parameter(torch.zeros(max_seq_len, hidden_size))
        self.encoder_layers = nn.TransformerEncoderLayer(hidden_size, num_heads, hidden_size*4, dropout)
        self.encoder = nn.TransformerEncoder(self.encoder_layers, num_layers=num_layers)
        self.fc = nn.Linear(hidden_size, output_size)
    def forward(self, x):
        x = self.embedding(x) + self.positional_encoding[:x.size(1), :]
        x = self.encoder(x.transpose(0, 1)).transpose(0, 1)
        x = self.fc(x[:, -1, :])
        return x


# 定义训练函数
def train(model, train_loader, optimizer, criterion):
    model.train()
    train_loss = 0
    i=0
    for data, labels in train_loader:
        data=data.cuda()
        labels=labels.cuda()
        print(i,len(train_loader))
        i+=1
        optimizer.zero_grad()
        outputs = model(data)
        loss = criterion(outputs, labels)
        loss.backward()
        optimizer.step()
        train_loss += loss.item() * data.size(0)
    train_loss /= len(train_loader.dataset)
    torch.save(model.state_dict(), f'model.pth')
    return train_loss

# 定义测试函数
def test(model, test_loader, criterion):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, labels in test_loader:
            data=data.cuda()
            labels=labels.cuda()
            outputs = model(data)
            test_loss += criterion(outputs, labels).item() * data.size(0)
            pred = outputs.argmax(dim=1, keepdim=True)
            correct += pred.eq(labels.argmax(dim=1, keepdim=True)).sum().item()
    test_loss /= len(test_loader.dataset)
    accuracy = correct / len(test_loader.dataset)
    return test_loss, accuracy

# 加载数据并进行预处理
data_file = "data_ids.txt"
label_file = "label_ids.txt"
max_seq_len = 100
data, labels = preprocess_data(data_file, label_file, max_seq_len)

# 划分训练集和测试集
train_size = int(0.8 * len(data))
test_size = len(data) - train_size
train_data, test_data = data[:train_size], data[train_size:]
train_labels, test_labels = labels[:train_size], labels[train_size:]

# 创建数据集和数据加载器
train_dataset = MyDataset(train_data, train_labels)
test_dataset = MyDataset(test_data, test_labels)
batch_size = 32
train_loader = DataLoader(train_dataset, batch_size=batch_size, shuffle=True)
test_loader = DataLoader(test_dataset, batch_size=batch_size, shuffle=False)

# 创建模型并定义优化器和损失函数
input_size = 300
output_size = 8
model = TransformerModel(input_size, output_size, max_seq_len).cuda()
optimizer = optim.Adam(model.parameters(), lr=0.001)
criterion = nn.BCEWithLogitsLoss()

# 训练模型并测试
num_epochs = 10
for epoch in range(num_epochs):
    print(epoch)
    train_loss = train(model, train_loader, optimizer, criterion)
    test_loss, accuracy = test(model, test_loader, criterion)
    print("Epoch [{}/{}], Train Loss: {:.4f}, Test Loss: {:.4f}, Accuracy: {:.2f}%".format(epoch+1, num_epochs, train_loss, test_loss, accuracy*100))