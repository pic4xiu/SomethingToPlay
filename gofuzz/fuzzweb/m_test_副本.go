package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

type Item struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func FuzzReverse(f *testing.F) {
	f.Fuzz(func(a *testing.T, data string, va string) { //接收参数
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
			// fmt.Println("Error marshaling JSON to BSON:", err)
			return
		}

		// 发送 POST 请求
		resp, err := http.Post("http://localhost:3000/parse-bson", "application/bson", bytes.NewBuffer(bsonData))
		if err != nil {
			// fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()
		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// fmt.Println("Error reading response body:", err)
			return
		}

		var items []Item
		err1 := json.Unmarshal([]byte(body), &items)
		if err1 != nil {
			return
		}

		// 输出结果
		// for _, item := range items {
		// 	if item.Key != data && item.Value != va {
		// 		panic("no!")
		// 	}
		// 	// fmt.Printf("%s: %s\n", item.Key, item.Value)
		// }
	})
}
