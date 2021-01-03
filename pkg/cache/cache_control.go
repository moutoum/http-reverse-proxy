package cache

import (
	"strconv"
	"strings"
	"time"
)

type CacheControl struct {
	NoStore, HasNotStore      bool
	NoCache, HasNoCache       bool
	Public, HasPublic         bool
	OnlyCached, HasOnlyCached bool
	MaxAge                    time.Duration
	HasMaxAge                 bool
	Expiration                time.Time
}

func ParseCacheControl(cacheControlHeader string) *CacheControl {
	cc := &CacheControl{}

	items := strings.Split(cacheControlHeader, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		switch item {
		case "no-store":
			cc.NoStore = true
			cc.HasNotStore = true
			continue
		case "no-cache":
			cc.NoCache = true
			cc.HasNoCache = true
			continue
		case "public":
			cc.Public = true
			cc.HasPublic = true
			continue
		case "only-if-cached":
			cc.OnlyCached = true
			cc.HasOnlyCached = true
		}

		if strings.HasPrefix(item, "max-age") {
			ageStr := strings.TrimPrefix(item, "max-age=")
			age, err := strconv.Atoi(ageStr)
			if err == nil {
				cc.MaxAge = time.Duration(age) * time.Second
				cc.Expiration = time.Now().Add(cc.MaxAge)
				cc.HasMaxAge = true
			}
			continue
		}
	}

	return cc
}
