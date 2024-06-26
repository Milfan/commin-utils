package utils_redis

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type redisclient struct {
	address          string
	password         string
	dbnumber         int
	defExpiryInSec   int
	defBlockingInSec int
	redisInstance    *redis.Client
	mu               sync.Mutex
	logger           *logrus.Logger
}

type IRedisClient interface {
	Connector() *redis.Client
	InitRedis() *redis.Client
	Store(key string, val interface{}, ttl ...int) error
	DoesKeyExist(key string) (status bool)
	GetContentByKey(key string) (interface{}, bool)
	DeleteKeys(keys ...string) int64
	UpdateKey(key string, newValue string) error
}

func NewRedisClient(opts ...RedisOption) (IRedisClient, error) {
	rdsCfg := &redisConfig{}
	rdsCfg.logger = logrus.StandardLogger() // default is logrus
	rdsCfg.expiredInSec = 5                 //default is 5 seconds
	rdsCfg.blockedInSec = 10                //default is localhost
	rdsCfg.host = "localhost"

	for _, opt := range opts {
		opt(rdsCfg)
	}

	r := redisclient{}
	r.logger = rdsCfg.logger
	r.address = rdsCfg.host
	r.defExpiryInSec = rdsCfg.expiredInSec
	r.defBlockingInSec = rdsCfg.blockedInSec
	r.dbnumber = rdsCfg.DBNumber
	r.password = rdsCfg.password

	client := r.Connector()
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &r, err
}

func (r *redisclient) InitRedis() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     r.address,
		Password: r.password,
		DB:       r.dbnumber,
	})
	r.redisInstance = client

	i := 1
	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		r.logger.Errorf("redis err: %s connecting to %s trial %d", err, r.address, i)
		i += 1
		time.Sleep(2000 * time.Millisecond)
	}

	return r.redisInstance
}

// Connector check redis connection
func (r *redisclient) Connector() *redis.Client {
	if r.redisInstance == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.redisInstance == nil {
			return r.InitRedis()
		}
	}

	return r.redisInstance
}

func (r *redisclient) Store(key string, val interface{}, ttl ...int) error {
	client := r.Connector()

	var expiredTimeInt int
	var err error

	if len(ttl) < 1 {
		expiredTimeInt = r.defExpiryInSec
	} else {
		expiredTimeInt = ttl[0]
	}

	durationTimeout := time.Duration(expiredTimeInt) * time.Second

	var vv []byte
	switch s := val.(type) {
	case []byte:
		if !json.Valid(s) {
			r.logger.Errorf("redis error: %+v", err)
			return err
		}
		vv = s
	case string:
		vv = []byte(s)
	default:
		vv, err = json.Marshal(val)
		if err != nil {
			r.logger.Errorf("redis error: %+v", err)
			return err
		}
	}

	err = client.Set(key, vv, durationTimeout).Err()

	return err
}

func (r *redisclient) DoesKeyExist(key string) bool {
	client := r.Connector()
	_, err := client.Get(key).Result()
	status := true
	if err == redis.Nil {
		status = false
		r.logger.Warnf("Key %s does not  exist", key)
	} else if err != nil {
		status = false
		r.logger.Errorf("DoesKeyExist err: %s ", err)
	}

	return status
}

func (r *redisclient) GetContentByKey(key string) (interface{}, bool) {
	client := r.Connector()
	val, err := client.Get(key).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		r.logger.Errorf("Error GetContentByKey err: %s ", err)
		return "", false
	}

	return val, true
}

func (r *redisclient) DeleteKeys(keys ...string) int64 {
	client := r.Connector()
	n, err := client.Del(keys...).Result()
	if err != nil {
		return 0
	}
	return n
}

// UpdateKey is updating the key value without resetting the TTL
func (r *redisclient) UpdateKey(key string, newValue string) error {

	// eval "local ttl = redis.call('ttl', ARGV[1]) if ttl > 0 then return redis.call('SETEX', ARGV[1], ttl, ARGV[2]) end" 0 key 987
	client := r.Connector()
	script := "local ttl = redis.call('ttl', KEYS[1]) if ttl > 0 then return redis.call('SETEX', KEYS[1], ttl, ARGV[1]) end"

	_, err := client.Eval(script, []string{key}, []string{newValue}).Result()

	return err
}
