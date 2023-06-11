import pika
import re
from flask import Flask, request
import os
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LogisticRegression
import urllib.parse
import joblib
import redis
import requests
# 连接Redis
r = redis.Redis(host='localhost', port=6379, db=0)

connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()

# 声明要消费的队列
channel.queue_declare(queue='newurls', durable=True)

def load(name):
    filepath = os.path.join(str(os.getcwd()), name)
    with open(filepath,'r') as f:
        alldata = f.readlines()
    ans = []
    for i in alldata:
        i = str(urllib.parse.unquote(i))
        ans.append(i)
    return ans
badqueries = load('badqueries.txt')
goodqueries = load('goodqueries.txt')#导入两类url
vectorizer = TfidfVectorizer()#用来将url向量化
X = vectorizer.fit_transform(badqueries+goodqueries)#直接输进去
lgs = LogisticRegression(class_weight='balanced') #简单的逻辑回归二分类
lgs = joblib.load('lgs.model')
print('ready!!')
def check(url):
    X_predict = vectorizer.transform([url])
    res = lgs.predict(X_predict)
    print(res)
    return res

# 回调函数，处理消息
def callback(ch, method, properties, body):
    # 解析Session和URL字段

    session, url = None, None
    for field in body.decode().split(", "):
        if field.startswith("Session:"):
            session = field.split(":")[1].strip()
        elif field.startswith("URL:"):
            url = field.split(":")[1].strip()

    # 打印结果
    print("Session: %s, URL: %s" % (session, url))
    tmp= int(check(url))#numpy类型转int




    # Check if key exists in hash
    if r.hexists('cache', url):
        print('已经有了')#这块不用加的，我蠢了，但没删，留作教训
    else:
        # Add key-value pair to hash
        r.hset('cache', url, tmp)
        print(url+" 已加入cache数据库")


    
    if tmp:
        #注销
        # url = 'http://127.0.0.1:8080/invalidate'#你服务器地址
        # params = {'key': 'qweasd', 'session': session}#本来是想留个接口的通过一个key来请求销毁key进黑名单，但想了想干脆直接拉进黑名单不就行了，还是不是很熟悉
        #requests.get(url, params=params)
        #加拉黑
        r.sadd('blacklist',session)
        print(session+" 已拉黑")
       
# 消费消息
channel.basic_consume(queue='newurls', on_message_callback=callback, auto_ack=True)

# 启动消费者
channel.start_consuming()
