package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type Item struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func main() {
	// 创建 JSON 对象
	data := string("\x15")
	va := string("\x7f\x81\x81\x81\x81")

	jsonData := map[string]string{
		data: va,
	}
	// for key, value := range jsonData {
	// 	fmt.Printf("Key: %s, Value: %s\n", key, value)
	// }
	// fmt.Println(jsonData)
	// fmt.Println(jsonData)
	// 将 JSON 对象转换为 BSON 格式
	bsonData, err := bson.Marshal(jsonData)
	if err != nil {
		fmt.Println("Error marshaling JSON to BSON:", err)
		return
	}
	// 发送 POST 请求
	resp, err := http.Post("http://localhost:3000/parse-bson", "application/bson", bytes.NewBuffer(bsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var items []Item
	fmt.Println(len(bsonData))
	err1 := json.Unmarshal([]byte(body), &items)
	// fmt.Println(items)
	if err1 != nil {
		panic(err)
	}

	// 输出结果
	// for _, item := range items {
	// 	if item.Key != data || item.Value != va {
	// 		bd := []byte(data)
	// 		bv := []byte(va)
	// 		fmt.Println(bd, bv)
	// 		bk := []byte(item.Key)
	// 		bi := []byte(item.Value)
	// 		fmt.Println(bk, bi)
	// 		// fmt.Println(bd, bv)
	// 		fmt.Printf("%s: %s\n", data, va)
	// 		fmt.Printf("%s: %s\n", item.Key, item.Value)
	// 		panic("no!")
	// 	}
	// 	// fmt.Printf("%s: %s\n", item.Key, item.Value)
	// }

	// var de map[string]string
	// err = json.Unmarshal(body, &de)
	// if err != nil {
	// 	fmt.Println("wrong", err)
	// 	return
	// }
	// 打印响应体
	// fmt.Println(string(body))
}

// package main

// import (
// 	"encoding/json"
// 	"fmt"

// 	"go.mongodb.org/mongo-driver/bson"
// )

// func main() {
// 	// 创建一个Person对象
// 	p := map[string]interface{}{
// 		"111111": "\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x7f\x80",
// 	}

// 	// 将Person对象转为JSON
// 	jsonData, err := json.Marshal(p)
// 	if err != nil {
// 		fmt.Println("JSON Marshal error:", err)
// 		return
// 	}

// 	// 将JSON转为BSON
// 	var bsonData bson.M
// 	err = bson.UnmarshalExtJSON(jsonData, true, &bsonData)
// 	fmt.Println(jsonData)
// 	if err != nil {
// 		fmt.Println("BSON Unmarshal error:", err)
// 		return
// 	}

// 	// 将BSON转回JSON
// 	jsonData, err = bson.MarshalExtJSON(bsonData, true, true)
// 	if err != nil {
// 		fmt.Println("JSON Marshal error:", err)
// 		return
// 	}

// 	// 将JSON转回Person对象
// 	var p2 map[string]interface{}
// 	err = json.Unmarshal(jsonData, &p2)
// 	if err != nil {
// 		fmt.Println("JSON Unmarshal error:", err)
// 		return
// 	}

// 	fmt.Println("Name:", p2["111111"])
// }
