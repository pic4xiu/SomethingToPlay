#!/usr/bin/env python3

import mysql.connector
# import smtplib
import os
# 恶意域名库文件路径
MALWARE_DOMAINS_FILE = "domains.list"

# MySQL数据库连接信息
MYSQL_HOST = "localhost"
MYSQL_PORT = 3306
MYSQL_USER = "root"
MYSQL_PASSWORD = "123456"
MYSQL_DATABASE = "database"

cnx = mysql.connector.connect(user=MYSQL_USER, password=MYSQL_PASSWORD,
                              host=MYSQL_HOST, port=MYSQL_PORT,
                              database=MYSQL_DATABASE)
cursor = cnx.cursor()

# 检查恶意域名库中的域名是否存在于数据库中
with open(MALWARE_DOMAINS_FILE) as f:
    for domain in f:
        domain = domain.strip()
        query = "SELECT COUNT(*) FROM dns_packets WHERE queries = %s"
        
        cursor.execute(query, (domain,))
        result = cursor.fetchone()[0]
        print(result)
        if result >= 1:
            # 触发告警
            subject = "Malware Alert"
            body = f"Alert: {domain} found in malware domain list"
            print('echo '+body+'| mail -s "恶意域名！" pic4xiu@qq.com')
            os.system('echo '+body+'| mail -s "bad query!" pic4xiu@qq.com')

cursor.close()
cnx.close()
