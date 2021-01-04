package cache

import (
	"strconv"
	"strings"
	"time"
)

// cacheControl is the representation of the "Cache-Control"
// HTTP header.
type CacheControl struct {

	// NoStore is set when the "no-store" option is given.
	NoStore     bool

	// NoCache is set when the "no-cache" option is given.
	NoCache     bool

	// Public is set to true if the "public" option is given or
	// false if the "private" option is given.
	Public      bool

	// OnlyCached is set when the "only-if-cached" option is given.
	OnlyCached  bool

	// MaxAge represents the value for the "max-age" option.
	// This field has to be checked with the HasMaxAge field first before
	// any usage.
	MaxAge      time.Duration
	// HasMaxAge is set when the "max-age" option is given.
	HasMaxAge   bool

	// MaxStale represents the value for the "max-stale" option.
	// This field has to be checked with the HasMaxStale field first
	// before any usage.
	MaxStale    time.Duration
	// HasMaxStale is set when the "max-stale" option is given.
	HasMaxStale bool

	// MinFresh represents the value for the "min-fresh" option.
	// This field has to be checked with the HasMinFresh field first
	// before any usage.
	MinFresh    time.Duration
	// HasMinFresh is set when the "min-fresh" option is given.
	HasMinFresh bool
}

// ParseCacheControl parses the given "Cache-Control" HTTP header
// and returns its representation.
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

		if strings.HasPrefix(item, "max-stale") {
			if pos := strings.Index(item, "="); pos != -1 {
				if maxStale, err := strconv.Atoi(item[pos+1:]); err == nil {
					cc.MaxStale = time.Duration(maxStale) * time.Second
					cc.HasMaxStale = true
					continue
				}
			}
			// If no specified value, we consider a valid resource for 15 years
			// after the expiration Date.
			cc.MaxStale = 15 * (time.Hour * 24 * 365)
			cc.HasMaxStale = true
			continue
		}

		if strings.HasPrefix(item, "min-fresh") {
			minFreshStr := strings.TrimPrefix(item, "min-fresh=")
			minFresh, err := strconv.Atoi(minFreshStr)
			if err == nil {
				cc.MinFresh = time.Duration(minFresh) * time.Second
				cc.HasMinFresh = true
			}
			continue
		}
	}

	return cc
}
