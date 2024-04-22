#define BLINKER_MIOT_LIGHT
#define BLINKER_WIFI

#include <Blinker.h>
#include<Servo.h>
char auth[] = "******";//电灯科技的key
char ssid[] = "**";//wifi名称
char pswd[] = "**";//wifi密码
Servo myservo1,myservo2;//定义舵机对象
void mioPowerState(const String & state){
  BLINKER_LOG("ee",state);
  if(state==BLINKER_CMD_OFF){//如果语音接受到的是关灯就执行的动作
    myservo1.write(90);//把舵机1设置为90度
    BlinkerMIOT.powerState("off");//反馈状态
    BlinkerMIOT.print();
    delay(2000);//延迟2秒
    myservo1.write(0);//把舵机1设置为0度
 }
 else if(state==BLINKER_CMD_ON){//如果语音接受到的是开灯就执行的动作
    myservo2.write(90);//把舵机2设置为90度
    BlinkerMIOT.powerState("on");//反馈状态
    BlinkerMIOT.print();  
    delay(2000);//延迟2秒
    myservo2.write(0);//把舵机2设置为0度
 }
}

void setup() {
 //   Serial.begin(115200);// 初始化串口
    myservo1.attach(D13);//舵机信号线接口
    myservo2.attach(D12);
    myservo1.write(0);//把舵机初始角度设为0度
    myservo2.write(0);
    #if defined(BLINKER_PRINT)
        BLINKER_DEBUG.stream(BLINKER_PRINT);
    #endif

    Blinker.begin(auth, ssid, pswd);//连接wifi以及电灯科技

    BlinkerMIOT.attachPowerState(mioPowerState);
}

void loop() {
    Blinker.run();
}