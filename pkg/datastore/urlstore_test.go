package datastore

import (
	"fmt"
	"testing"
	"time"

	"github.com/noandrea/distill/pkg/model"
)

func dt(y, m, d, H, M, S int) (t time.Time) {
	if y+m+d+H+M+S == 0 {
		return
	}
	f := fmt.Sprintf("%d-%d-%d %d:%d:%d", d, m, y, H, M, S)
	t, err := time.Parse("1-2-2006 15:4:5", f)
	if err != nil {
		panic(err)
	}
	return
}

func TestUpdateCounters(t *testing.T) {

	tests := []struct {
		name             string
		u                *model.URLInfo
		ExpectedHits     int64
		ExpectedResolves int64
	}{
		{
			"all good",
			&model.URLInfo{
				ActiveFrom:   time.Now().Add(time.Duration(-1) * time.Hour),
				ExpiresOn:    time.Now().Add(time.Duration(1) * time.Hour),
				Hits:         100,
				ResolveLimit: 0,
				ResolveCount: 50,
			},
			101, // hits
			51,  // resolves
		},
		{
			"inactive",
			&model.URLInfo{
				ActiveFrom:   time.Now().Add(time.Duration(-2) * time.Hour),
				ExpiresOn:    time.Now().Add(time.Duration(-1) * time.Hour),
				Hits:         100,
				ResolveLimit: 0,
				ResolveCount: 50,
			},
			101, // hits
			50,  // resolves
		},
		{
			"expired",
			&model.URLInfo{
				ActiveFrom:   time.Now().Add(time.Duration(1) * time.Hour),
				ExpiresOn:    time.Now().Add(time.Duration(2) * time.Hour),
				Hits:         100,
				ResolveLimit: 0,
				ResolveCount: 50,
			},
			101, // hits
			50,  // resolves
		},
		{
			"exhausted",
			&model.URLInfo{
				ActiveFrom:   time.Now().Add(time.Duration(-1) * time.Hour),
				ExpiresOn:    time.Now().Add(time.Duration(2) * time.Hour),
				Hits:         100,
				ResolveLimit: 10,
				ResolveCount: 50, // this should not be possible
			},
			101, // hits
			11,  // resolves
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateCounters(tt.u)
		})
		if tt.ExpectedHits != tt.u.Hits {
			t.Errorf("UpdateCounters() expected hits = %v, got %v", tt.ExpectedHits, tt.u.Hits)
			return
		}
		if tt.ExpectedResolves != tt.u.ResolveCount {
			t.Errorf("UpdateCounters() expected Resolves = %v, got %v", tt.ExpectedResolves, tt.u.ResolveCount)
			return
		}
	}
}
