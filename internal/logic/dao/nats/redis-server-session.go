package nats

import (
	"context"
	"encoding/json"

	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"

	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/model"
)

const (
	_prefixMidServer    = "mid_%d" // mid -> key:server
	_prefixKeyServer    = "key_%s" // key -> server
	_prefixServerOnline = "ol_%s"  // server -> online
)

func newRedis(c *conf.Redis) *redis.Pool {
	// var c *conf.Redis
	// c.Network = "tcp"
	// c.Addr = "127.0.0.1:6379"
	// c.Active = 60000
	// c.Idle = 1024
	// c.DialTimeout = 200 * time.Second
	// c.ReadTimeout = 500 * time.Microsecond
	// c.WriteTimeout = 500 * time.Microsecond
	// c.IdleTimeout = 120 * time.Second
	// c.Expire = 30 * time.Minute

	return &redis.Pool{
		// MaxIdle:   c.Idle,
		// MaxActive: c.Active,
		// IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", c.Addr) // redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
			// redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
			// redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
			// redis.DialPassword(c.Auth),

			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get()
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, _ := json.Marshal(online)
	if err = conn.Send("HSET", key, hashKey, b); err != nil {
		log.Errorf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) serverOnline(c context.Context, key string, hashKey string) (online *model.Online, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("HGET", key, hashKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorf("conn.Do(HGET %s %s) error(%v)", key, hashKey, err)
		}
		return
	}
	online = new(model.Online)
	if err = json.Unmarshal(b, online); err != nil {
		log.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
		return
	}
	return
}

func (d *Dao) delServerOnline(c context.Context, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	key := keyServerOnline(server)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorf("conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}
