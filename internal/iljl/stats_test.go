package iljl

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func noTest_loadGlobalStatistics(t *testing.T) {
	tests := []struct {
		name    string
		wantS   *Statistics
		wantErr bool
		match   bool
	}{
		{
			name: "1",
			wantS: &Statistics{
				Upserts: 15,
				Deletes: 5,
				Urls:    10,
				Gets:    10,
			},
			match:   true,
			wantErr: false,
		},
		{
			name: "2",
			wantS: &Statistics{
				Upserts: 1500,
				Deletes: 230,
				Urls:    1270,
				Gets:    4000,
			},
			match:   true,
			wantErr: false,
		},
		{
			name: "2",
			wantS: &Statistics{
				Upserts: 10,
				Deletes: 3,
				Urls:    10,
				Gets:    1,
			},
			match:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
			NewSession()
			ids := []string{}
			// run inserts
			for i := int64(0); i < tt.wantS.Upserts; i++ {
				id, err := UpsertURL(&URLReq{URL: fmt.Sprint("http://ilij.il/?long=", i)}, true, true, time.Now())
				if err != nil {
					t.Error(err)
				}
				ids = append(ids, id)
			}
			// run deletes
			for i := int64(0); i < tt.wantS.Deletes; i++ {
				DeleteURL(ids[i])
			}
			ids = ids[tt.wantS.Deletes:]
			// run gets
			for i := int64(0); i < tt.wantS.Gets; i++ {
				GetURLRedirect(ids[i%tt.wantS.Urls])
			}
			CloseSession()
			// Start a new session
			NewSession()
			gotS, err := loadGlobalStatistics()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadGlobalStatistics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.match == !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("loadGlobalStatistics() = %v, want %v", gotS, tt.wantS)
			}

			// alsot test reset
			gotS.Deletes = 0
			err = resetGlobalStatistics()
			if err != nil {
				t.Errorf("resetGlobalStatistics() error = %v, wantErr %v", err, false)
				return
			}

			CloseSession()
		})
	}
}

func noTest_resetGlobalStatistics(t *testing.T) {
	tests := []struct {
		name    string
		wantS   *Statistics
		wantErr bool
		match   bool
	}{
		{
			name: "1",
			wantS: &Statistics{
				Upserts: 15,
				Deletes: 5,
				Urls:    10,
				Gets:    10,
			},
			match:   true,
			wantErr: false,
		},
		{
			name: "2",
			wantS: &Statistics{
				Upserts: 1500,
				Deletes: 230,
				Urls:    1270,
				Gets:    4000,
			},
			match:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
			NewSession()
			defer CloseSession()
			ids := []string{}
			// run inserts
			for i := int64(0); i < tt.wantS.Upserts; i++ {
				id, err := UpsertURL(&URLReq{URL: fmt.Sprint("http://ilij.il/?long=", i)}, true, true, time.Now())
				if err != nil {
					t.Error(err)
				}
				ids = append(ids, id)
			}
			// run deletes
			for i := int64(0); i < tt.wantS.Deletes; i++ {
				DeleteURL(ids[i])
			}
			ids = ids[tt.wantS.Deletes:]
			// run gets
			for i := int64(0); i < tt.wantS.Gets; i++ {
				GetURLRedirect(ids[i%tt.wantS.Urls])
			}
			gotS, err := loadGlobalStatistics()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadGlobalStatistics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.match == !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("loadGlobalStatistics() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
