package main

// Start Commond eg: ./push_room 1 20 localhost:3111
// first parameter: room id
// second parameter: num per seconds
// third parameter: logic server ip

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
	rountineNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	addr := os.Args[3]

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
	resp, err := http.Post("http://"+addr+"/goim/push/room?operation=1000&type=test&room="+os.Args[1], "application/json", bytes.NewBufferString(fmt.Sprintf("{\"test\":%d}", i)))
	if err != nil {
		fmt.Printf("Error: http.post() error(%v)\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: http.post() error(%v)\n", err)
		return
	}

	fmt.Printf("%s postId:%d, response:%s\n", time.Now().Format("2006-01-02 15:04:05"), i, string(body))
}
