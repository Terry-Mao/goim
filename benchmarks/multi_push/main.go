package main

// Start Command eg : ./multi_push 0 20000 localhost:7172 60

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
	t          int
)

const testContent = "{\"test\":1}"

type pushsBodyMsg struct {
	Msg     json.RawMessage `json:"m"`
	UserIds []int64         `json:"u"`
}

func init() {
	httpTransport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(30 * time.Second)
			c, err := net.DialTimeout(netw, addr, 20*time.Second)
			if err != nil {
				return nil, err
			}

			_ = c.SetDeadline(deadline)
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
	infoLogfi, err := os.OpenFile("./multi_push.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

	t, err = strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}

	num := runtime.NumCPU() * 8

	l := length / num
	b, e := begin, begin+l
	time.AfterFunc(time.Duration(t)*time.Second, stop)
	for i := 0; i < num; i++ {
		go startPush(b, e)
		b += l
		e += l
	}
	if b < begin+length {
		go startPush(b, begin+length)
	}

	time.Sleep(9999 * time.Hour)
}

func stop() {
	os.Exit(-1)
}

func startPush(b, e int) {
	l := make([]int64, 0, e-b)
	for i := b; i < e; i++ {
		l = append(l, int64(i))
	}
	msg := &pushsBodyMsg{Msg: json.RawMessage(testContent), UserIds: l}
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	for {
		resp, err := httpPost(fmt.Sprintf("http://%s/goim/push/mids=%d", os.Args[3], b), "application/x-www-form-urlencoded", bytes.NewBuffer(body))
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

		lg.Printf("response %s", string(body))
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
