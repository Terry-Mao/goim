package main

import (
	"encoding/json"
	inet "goim/libs/net"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/thinkboy/log4go"
)

func InitHTTP() (err error) {
	// http listen
	var network, addr string
	for i := 0; i < len(Conf.HTTPAddrs); i++ {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/push", Push)
		httpServeMux.HandleFunc("/1/pushs", Pushs)
		httpServeMux.HandleFunc("/1/push/all", PushAll)
		httpServeMux.HandleFunc("/1/push/room", PushRoom)
		httpServeMux.HandleFunc("/1/server/del", DelServer)
		httpServeMux.HandleFunc("/1/count", Count)
		log.Info("start http listen:\"%s\"", Conf.HTTPAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.HTTPAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go httpListen(httpServeMux, network, addr)
	}
	return
}

func httpListen(mux *http.ServeMux, network, addr string) {
	httpServer := &http.Server{Handler: mux, ReadTimeout: Conf.HTTPReadTimeout, WriteTimeout: Conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	if err := httpServer.Serve(l); err != nil {
		log.Error("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retWrite marshal the result and write to client(get).
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Info("req: \"%s\", get: res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// retPWrite marshal the result and write to client(post).
func retPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Info("req: \"%s\", post: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

func Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body      string
		serverId  int32
		keys      []string
		subKeys   map[int32][]string
		bodyBytes []byte
		userId    int64
		err       error
		uidStr    = r.URL.Query().Get("uid")
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if userId, err = strconv.ParseInt(uidStr, 10, 64); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", uidStr, err)
		res["ret"] = InternalErr
		return
	}
	subKeys = genSubKey(userId)
	for serverId, keys = range subKeys {
		if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
}

type pushsBodyMsg struct {
	Msg     json.RawMessage `json:"m"`
	UserIds []int64         `json:"u"`
}

func parsePushsBody(body []byte) (msg []byte, userIds []int64, err error) {
	tmp := pushsBodyMsg{}
	if err = json.Unmarshal(body, &tmp); err != nil {
		return
	}
	msg = tmp.Msg
	userIds = tmp.UserIds
	return
}

// {"m":{"test":1},"u":"1,2,3"}
func Pushs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body      string
		bodyBytes []byte
		serverId  int32
		userIds   []int64
		err       error
		res       = map[string]interface{}{"ret": OK}
		subKeys   map[int32][]string
		keys      []string
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if bodyBytes, userIds, err = parsePushsBody(bodyBytes); err != nil {
		log.Error("parsePushsBody(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	subKeys = genSubKeys(userIds)
	for serverId, keys = range subKeys {
		if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
}

func PushRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		bodyBytes []byte
		body      string
		rid       int
		err       error
		param     = r.URL.Query()
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%v)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	ridStr := param.Get("rid")
	enable, _ := strconv.ParseBool(param.Get("ensure"))
	// push room
	if rid, err = strconv.Atoi(ridStr); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", ridStr, err)
		res["ret"] = InternalErr
		return
	}
	if err = broadcastRoomKafka(int32(rid), bodyBytes, enable); err != nil {
		log.Error("broadcastRoomKafka(\"%s\",\"%s\",\"%d\") error(%s)", rid, body, enable, err)
		res["ret"] = InternalErr
		return
	}
	res["ret"] = OK
	return
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		bodyBytes []byte
		body      string
		err       error
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%v)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	// push all
	if err := broadcastKafka(bodyBytes); err != nil {
		log.Error("broadcastKafka(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	res["ret"] = OK
	return
}

type RoomCounter struct {
	RoomId int32
	Count  int32
}

type ServerCounter struct {
	Server int32
	Count  int32
}

func Count(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		typeStr = r.URL.Query().Get("type")
		res     = map[string]interface{}{"ret": OK}
	)
	defer retWrite(w, r, res, time.Now())
	if typeStr == "room" {
		d := make([]*RoomCounter, 0, len(RoomCountMap))
		for roomId, count := range RoomCountMap {
			d = append(d, &RoomCounter{RoomId: roomId, Count: count})
		}
		res["data"] = d
	} else if typeStr == "server" {
		d := make([]*ServerCounter, 0, len(ServerCountMap))
		for server, count := range ServerCountMap {
			d = append(d, &ServerCounter{Server: server, Count: count})
		}
		res["data"] = d
	}
	return
}

func DelServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		err       error
		serverStr = r.URL.Query().Get("server")
		server    int64
		res       = map[string]interface{}{"ret": OK}
	)
	if server, err = strconv.ParseInt(serverStr, 10, 32); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", serverStr, err)
		res["ret"] = InternalErr
		return
	}
	defer retWrite(w, r, res, time.Now())
	if err = delServer(int32(server)); err != nil {
		res["ret"] = InternalErr
		return
	}
	return
}
