package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
)

const FLIP_RATIO = 0.01

var FLIP_ARRAY = []byte{1, 2, 4, 8, 16, 32, 64, 128}
var MAGIC_VALS = [][]byte{
	{0xFF},
	{0x7F},
	{0x00},
	{0xFF, 0xFF},
	{0x00, 0x00},
	{0xFF, 0xFF, 0xFF, 0xFF},
	{0x00, 0x00, 0x00, 0x00},
	{0x00, 0x00, 0x00, 0x80},
	{0x00, 0x00, 0x00, 0x40},
	{0xFF, 0xFF, 0xFF, 0x7F},
}

func main() {
	if _, err := os.Stat("crashes"); os.IsNotExist(err) {
		if err := os.Mkdir("crashes", 0755); err != nil {
			fmt.Printf("创建文件夹失败：%v\n", err)
			os.Exit(1)
		}
	}
	if len(os.Args) < 2 {
		usage()
	} else {
		filename := os.Args[2]
		// fmt.Println(filename)
		origData := getBytes(filename)
		counter := 0
		for counter < 100 {
			data := make([]byte, len(origData))
			copy(data, origData)
			mutatedData := mutate(data)
			createNew(mutatedData)
			executeFuzz(mutatedData, counter)

			if counter%10 == 0 {
				fmt.Printf("Counter: %d\n", counter)
			}

			counter++
		}
	}
}

func usage() {
	fmt.Println("Usage: go run main.go <target> <input>")
}

func getBytes(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return data
}

func createNew(data []byte) {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0755); err != nil {
			fmt.Printf("创建文件夹失败：%v\n", err)
			os.Exit(1)
		}
	}
	err := ioutil.WriteFile("data/mutated.jpg", data, 0644)
	if err != nil {
		os.Exit(1)
		return
	}

}

func mutate(data []byte) []byte {
	flips := int(float64(len(data)-4) * FLIP_RATIO)
	flipIndexes := rand.Perm(len(data) - 8)[:flips]
	// fmt.Println(len(flipIndexes))

	for _, idx := range flipIndexes {
		method := rand.Intn(2)

		if method == 0 {
			data[idx+2] = bitFlip(data[idx+2])
		} else {
			magic(data, idx+2)
		}
	}

	return data
}

func bitFlip(byteVal byte) byte {
	flipVal := FLIP_ARRAY[rand.Intn(len(FLIP_ARRAY))]
	return byteVal ^ flipVal
}

func magic(data []byte, idx int) {
	pickedMagic := MAGIC_VALS[rand.Intn(len(MAGIC_VALS))]

	for i, m := range pickedMagic {
		data[idx+i] = m
	}
}
func writeDataToFile(counter int, data []byte) error {
	filename := fmt.Sprintf("crashes/crash.%d.jpg", counter)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func executeFuzz(data []byte, counter int) {
	cmd := exec.Command("./main", "data/mutated.jpg")
	if err := cmd.Start(); err != nil {
		fmt.Printf("启动子进程失败：%v\n", err)
		os.Exit(1)
	}

	// 等待子进程退出
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// 子进程退出
	if err := <-done; err != nil {
		if err := writeDataToFile(counter, data); err != nil {
			fmt.Printf("写入文件失败：%v\n", err)
			os.Exit(1)
		}

		fmt.Println("写入文件成功！")
		fmt.Printf("子进程退出：%v\n", err)
	}
}
