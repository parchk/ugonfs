package util

import (
	"NfsAgent/conf"
	"NfsAgent/mlog"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func RandString(length int, c int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)
	for start := 0; start < length; start++ {

		var t int

		if c > 0 {
			t = rand.Intn(3)
		}

		if t == 0 {
			rs = append(rs, strconv.Itoa(rand.Intn(10)))
		} else if t == 1 {
			rs = append(rs, string(rand.Intn(26)+65))
		} else {
			rs = append(rs, string(rand.Intn(26)+97))
		}
	}
	return strings.Join(rs, "")
}

func GetId() (int, error) {

	id := conf.SF.Generate()

	id_str := id.String()

	l := len(id_str)
	b := l - 6
	k := id_str[b:l]

	rid := strings.TrimLeft(k, "0")

	mlog.Debug("GetId id str :", rid)

	srid, err := strconv.Atoi(rid)

	mlog.Debug("GetId id :", srid)

	return srid, err
}

func GetIdEx() (int64, error) {

	id := conf.SF.Generate()

	id_str := id.String()

	l := len(id_str)
	b := l - 11
	k := id_str[b:l]

	rid := strings.TrimLeft(k, "0")

	mlog.Debug("util GetIdEx str :", rid)

	srid, err := strconv.ParseInt(rid, 10, 0)

	mlog.Debug("util GetIdEx :", srid)

	return srid, err
}
