package cache

import (
	"strconv"
	"strings"
	"time"
)

type CacheControl struct {
	NoStore                   bool
	NoCache                   bool
	Public                    bool
	OnlyCached, HasOnlyCached bool
	MaxAge                    time.Duration
	HasMaxAge                 bool
}

func ParseCacheControl(cacheControlHeader string) *CacheControl {
	cc := &CacheControl{Public: true}

	items := strings.Split(cacheControlHeader, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		switch item {
		case "no-store":
			cc.NoStore = true
			continue
		case "no-cache":
			cc.NoCache = true
			continue
		case "public":
			cc.Public = true
			continue
		case "private":
			cc.Public = false
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
				cc.HasMaxAge = true
			}
			continue
		}
	}

	return cc
}
