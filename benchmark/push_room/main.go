package main

// Start Commond eg: ./push_room 20 localhost:7172
// first parameter: post
// second parameter: amount of clients
// third parameter: comet server ip

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	rountineNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	addr := os.Args[2]

	gap := time.Second / time.Duration(rountineNum)
	delay := time.Duration(0)

	go run(addr, time.Duration(0)*time.Second)
	for i := 0; i < rountineNum-1; i++ {
		go run(addr, delay)
		delay += gap
		fmt.Println("delay:", delay)
	}
	time.Sleep(9999 * time.Hour)
}

func run(addr string, delay time.Duration) {
	time.Sleep(delay)
	i := int64(0)
	for {
		go post(addr, i)
		time.Sleep(time.Second)
		i++
	}
}

func post(addr string, i int64) {
	resp, err := http.Post("http://"+addr+"/1/push/all?rid=1", "application/json", bytes.NewBufferString(fmt.Sprintf("{\"test\":%d}", i)))
	if err != nil {
		fmt.Println("Error: http.post() error(%s)", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: http.post() error(%s)", err)
		return
	}

	fmt.Printf("%s postId:%d, response:%s\n", time.Now().Format("2006-01-02 15:04:05"), i, string(body))
}
