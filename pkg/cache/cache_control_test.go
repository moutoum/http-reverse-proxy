package cache

import (
	"reflect"
	"testing"
	"time"
)

func TestParseCacheControl(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   *CacheControl
	}{{
		name:   "Test single no-store",
		header: "no-store",
		want:   &CacheControl{NoStore: true, Public: true},
	}, {
		name:   "Test single no-cache",
		header: "no-cache",
		want:   &CacheControl{NoCache: true, Public: true},
	}, {
		name:   "Test private",
		header: "private",
		want:   &CacheControl{Public: false},
	}, {
		name:   "Test public",
		header: "public",
		want:   &CacheControl{Public: true},
	}, {
		name:   "Test single only-if-cached",
		header: "only-if-cached",
		want:   &CacheControl{OnlyCached: true, Public: true},
	}, {
		name:   "Test invalid value",
		header: "toto",
		want:   &CacheControl{Public: true},
	}, {
		name:   "Test max-age",
		header: "max-age=20",
		want:   &CacheControl{HasMaxAge: true, MaxAge: 20 * time.Second, Public: true},
	}, {
		name:   "Test max-age with invalid format",
		header: "max-age=toto",
		want:   &CacheControl{Public: true},
	}, {
		name:   "Test simple max-stale with no value",
		header: "max-stale",
		want:   &CacheControl{HasMaxStale: true, MaxStale: 15 * time.Hour * 24 * 365, Public: true},
	}, {
		name:   "Test max-stale with value",
		header: "max-stale=20",
		want:   &CacheControl{HasMaxStale: true, MaxStale: 20 * time.Second, Public: true},
	}, {
		name:   "Test max-stale with invalid value",
		header: "max-stale=toto",
		want:   &CacheControl{Public: true},
	}, {
		name:   "Test min-fresh",
		header: "min-fresh=20",
		want:   &CacheControl{HasMinFresh: true, MinFresh: 20 * time.Second, Public: true},
	}, {
		name:   "Test max-age with invalid format",
		header: "min-fresh=toto",
		want:   &CacheControl{Public: true},
	}, {
		name: "Test with multiple values",
		header: "public, min-fresh=20, only-if-cached",
		want: &CacheControl{Public: true, HasMinFresh: true, MinFresh: 20 * time.Second, OnlyCached: true},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCacheControl(tt.header); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCacheControl() = %v, want %v", got, tt.want)
			}
		})
	}
}
