package utils_redis

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

type redisConfig struct {
	host         string
	password     string
	DBNumber     int
	logger       *logrus.Logger
	expiredInSec int
	blockedInSec int
	isJwt        bool
	isCookies    bool
	cookieKey    string
}

type RedisOption func(*redisConfig)

func SetHost(host string) RedisOption {
	return func(o *redisConfig) {
		o.host = host
	}
}

func SetPassword(password string) RedisOption {
	return func(o *redisConfig) {
		o.password = password
	}
}

func SetDBNumber(s string) RedisOption {
	return func(o *redisConfig) {

		dbNum, err := strconv.Atoi(s)
		if err != nil {
			if o.logger != nil {
				o.logger.Fatalf("err: %s", err)
			} else {
				logrus.Fatalf("err: %s", err)
			}
		}

		o.DBNumber = dbNum
	}
}

func WithLogger(l *logrus.Logger) RedisOption {
	return func(o *redisConfig) {
		o.logger = l
	}
}

func ExpiredTimeInSec(s string) RedisOption {
	return func(o *redisConfig) {

		defExpiredInSec, err := strconv.Atoi(s)
		if err != nil {
			if o.logger != nil {
				o.logger.Fatalf("err: %s", err)
			} else {
				logrus.Fatalf("err: %s", err)
			}
		}

		o.expiredInSec = defExpiredInSec
	}
}

func BlockedTimeInSec(s string) RedisOption {
	return func(o *redisConfig) {

		blockedInSec, err := strconv.Atoi(s)
		if err != nil {
			if o.logger != nil {
				o.logger.Fatalf("err: %s", err)
			} else {
				logrus.Fatalf("err: %s", err)
			}
		}

		o.blockedInSec = blockedInSec
	}
}

func WithPassword(s string) RedisOption {
	return func(o *redisConfig) {
		o.password = s
	}
}

func IsJWT() RedisOption {
	return func(o *redisConfig) {
		o.isJwt = true
	}
}
