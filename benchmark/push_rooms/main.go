package main

// Start Command eg : ./push_rooms 0 20000 localhost:7172 40
// param 1 : the start of room number
// param 2 : the end of room number
// param 3 : comet server tcp address
// param 4 : push amount each goroutines per second

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

var (
	lg         *log.Logger
	httpClient *http.Client
)

const TestContent = "{\"test\":1}"

type pushBodyMsg struct {
	Msg    json.RawMessage `json:"m"`
	UserId int64           `json:"u"`
}

func init() {
	httpTransport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(30 * time.Second)
			c, err := net.DialTimeout(netw, addr, 20*time.Second)
			if err != nil {
				return nil, err
			}

			c.SetDeadline(deadline)
			return c, nil
		},
		DisableKeepAlives: false,
	}
	httpClient = &http.Client{
		Transport: httpTransport,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	infoLogfi, err := os.OpenFile("./push_rooms.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	lg = log.New(infoLogfi, "", log.LstdFlags|log.Lshortfile)

	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	length, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}
	delay := (1000 * time.Millisecond) / time.Duration(num)

	routines := runtime.NumCPU() * 2
	lg.Printf("start routine num:%d", routines)

	l := length / routines
	b, e := begin, begin+l
	for i := 0; i < routines; i++ {
		go startPush(b, e, delay)
		b += l
		e += l
	}
	if b < begin+length {
		go startPush(b, begin+length, delay)
	}

	time.Sleep(9999 * time.Hour)
}

func stop() {
	os.Exit(-1)
}

func startPush(b, e int, delay time.Duration) {
	lg.Printf("start Push from %d to %d", b, e)

	for {
		for i := b; i < e; i++ {
			resp, err := http.Post(fmt.Sprintf("http://%s/1/push/room?rid=%d", os.Args[3], i), "application/json", bytes.NewBufferString(TestContent))
			if err != nil {
				lg.Printf("post error (%v)", err)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				lg.Printf("post error (%v)", err)
				return
			}
			resp.Body.Close()

			lg.Printf("push room:%d response %s", i, string(body))
			time.Sleep(delay)
		}
	}
}

func httpPost(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
