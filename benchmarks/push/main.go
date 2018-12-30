package main

// Start Command eg : ./push 0 20000 localhost:7172 60

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
	httpClient *http.Client
	t          int
)

const testContent = "{\"test\":1}"

type pushBodyMsg struct {
	Msg    json.RawMessage `json:"m"`
	UserID int64           `json:"u"`
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

	num := runtime.NumCPU() * 2
	log.Printf("start routine num:%d", num)

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
	log.Printf("start Push from %d to %d", b, e)
	bodys := make([][]byte, e-b)
	for i := 0; i < e-b; i++ {
		msg := &pushBodyMsg{Msg: json.RawMessage(testContent), UserID: int64(b)}
		body, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		bodys[i] = body
	}

	for {
		for i := 0; i < len(bodys); i++ {
			resp, err := httpPost(fmt.Sprintf("http://%s/goim/push/mids?operation=1000&mids=%d", os.Args[3], b), "application/x-www-form-urlencoded", bytes.NewBuffer(bodys[i]))
			if err != nil {
				log.Printf("post error (%v)", err)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("post error (%v)", err)
				return
			}
			resp.Body.Close()

			log.Printf("response %s", string(body))
			//time.Sleep(50 * time.Millisecond)
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
