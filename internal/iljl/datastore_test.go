package iljl

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/jbrodriguez/mlog"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

func buildConifgTest() {
	mlog.DefaultFlags = log.Ltime | log.Lmicroseconds
	mlog.Start(mlog.LevelTrace, "")
	// path
	path, _ := ioutil.TempDir("/tmp/", "iljl")
	fmt.Println("test db folder is ", path)
	internal.Config = internal.ConfigSchema{
		Server: internal.ServerConfig{
			DbPath: path,
		},
		ShortID: internal.ShortIDConfig{
			Alphabet: "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789",
			Length:   6,
		},
	}
	internal.Config.Validate()
}

func buildConifgPanicTest() {
	mlog.Start(mlog.LevelTrace, "")
	path := " cann not exists / ssa "
	fmt.Println("test db folder is ", path)
	internal.Config = internal.ConfigSchema{
		Server: internal.ServerConfig{
			DbPath: path,
		},
		ShortID: internal.ShortIDConfig{
			Alphabet: "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789",
			Length:   6,
		},
	}
	internal.Config.Validate()
}

func buildConifgTestShortIDParams(alphabet string, length int) {
	mlog.Start(mlog.LevelTrace, "")
	path, _ := ioutil.TempDir("/tmp/", "iljl")
	fmt.Println("test db folder is ", path)
	internal.Config = internal.ConfigSchema{
		Server: internal.ServerConfig{
			DbPath: path,
		},
		ShortID: internal.ShortIDConfig{
			Alphabet:    alphabet,
			Length:      length,
			TTL:         0,
			MaxRequests: 0,
		},
		Tuning: internal.TuningConfig{
			StatsEventsWorkerNum: 20,
		},
	}
	internal.Config.Validate()
}

func TestPreprocessURL(t *testing.T) {

	type args struct {
		forceAlphabet bool
		forceLength   bool
	}
	tests := []struct {
		name    string
		args    args
		url     URLReq
		wantErr bool
	}{
		{
			name:    "all good",
			wantErr: false,
			args: args{
				forceAlphabet: true,
				forceLength:   false,
			},
			url: URLReq{
				URL: "https://ilij.li",
			},
		},
		{
			name:    "wrong target url",
			wantErr: true,
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
			url: URLReq{
				URL: "ilij.li",
			},
		},
		{
			name:    "id set",
			wantErr: false,
			args: args{
				forceAlphabet: true,
				forceLength:   false,
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "abcdef",
			},
		},
		{
			name:    "wrong alphabet",
			wantErr: true,
			args: args{
				forceAlphabet: true,
				forceLength:   false,
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "ilcdef",
			},
		},
		{
			name:    "wrong lenght",
			wantErr: true,
			args: args{
				forceAlphabet: false,
				forceLength:   true,
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "ilc",
			},
		},
		{
			name:    "wrong lenght and alphabet",
			wantErr: true,
			args: args{
				forceAlphabet: true,
				forceLength:   true,
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "abac$$ai",
			},
		},
	}

	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Log("berore:", tt.url)

			err := PreprocessURL(&tt.url, tt.args.forceAlphabet, tt.args.forceLength)

			if (err != nil) != tt.wantErr {
				t.Errorf("PreprocessURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if tt.args.forceLength && len(tt.url.ID) != internal.Config.ShortID.Length {
				t.Errorf("PreprocessURL() ID length %v, expected %v ", len(tt.url.ID), internal.Config.ShortID.Length)
				return
			}

			if tt.url.ID == "" {
				t.Errorf("PreprocessURL() empty id ")
				return
			}
		})
	}

}

func TestUpsertURL(t *testing.T) {
	type args struct {
		forceAlphabet bool
		forceLength   bool
	}
	tests := []struct {
		name    string
		url     *URLReq
		wantErr bool
		args    args
	}{
		{
			name:    "all good",
			wantErr: false,
			url: &URLReq{
				URL: "https://ilij.li",
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
		{
			name:    "id set",
			wantErr: false,
			url: &URLReq{
				URL: "https://ilij.li",
				ID:  "samesame",
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
		{
			name:    "overwrite url",
			wantErr: false,
			url: &URLReq{
				URL: "https://wikipedia.li",
				ID:  "samesame",
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
		{
			name:    "fail length",
			wantErr: true,
			url: &URLReq{
				URL: "https://ilij.li",
				ID:  "samesame",
			},
			args: args{
				forceAlphabet: false,
				forceLength:   true,
			},
		},
		{
			name:    "ttl",
			wantErr: false,
			url: &URLReq{
				URL: "https://ilij.li",
				ID:  "ttlttl",
				TTL: 30,
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
		{
			name:    "all fields",
			wantErr: false,
			url: &URLReq{
				URL:         "https://ilij.li",
				ID:          "allfields",
				TTL:         30,
				MaxRequests: 50,
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	defer CloseSession()
	ids := make(map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(tt.url, tt.args.forceAlphabet, tt.args.forceLength)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			// only records in case of success
			if err == nil {
				ids[id] = true
			}
		})
	}

	validElements := 4
	if len(ids) != validElements {
		t.Errorf("UpsertURL() length = %v, want %v", len(ids), validElements)
	}

	// test upsert
	id := "samesame"
	ur := "https://wikipedia.li"

	ui, _ := GetURLInfo(id)
	if ui.URL != ur {
		t.Errorf("UpsertURL()  %v, want %v", ui.URL, ur)
	}
}

func TestDeleteURL(t *testing.T) {
	type args struct {
		url URLReq
	}
	tests := []struct {
		name    string
		url     *URLReq
		wantErr bool
	}{
		{
			name:    "id set",
			wantErr: false,
			url: &URLReq{
				URL:         "https://ilij.li",
				MaxRequests: 10,
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)

	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		url := tt.url
		id, _ := UpsertURL(url, false, false)
		ui, _ := GetURLInfo(id)

		if id != ui.ID || url.URL != ui.URL {
			t.Errorf("DeleteURL()")
			break
		}
		t.Log(ui)

		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteURL(id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		ui, err := GetURLInfo(id)
		if err == nil {
			t.Errorf("DeleteURL() not deleted")
		}
	}

}

func TestGetURL(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantURL string
		wantErr bool
	}{
		{
			name:    "0",
			wantErr: false,
			wantURL: "https://ilij.li/?param=0",
		},
		{
			name:    "1",
			wantErr: false,
			wantURL: "https://ilij.li/?param=1",
		},
		{
			name:    "2",
			wantErr: false,
			wantURL: "https://ilij.li/?param=2",
		},
		{
			name:    "3",
			wantErr: true,
			wantURL: "https://ilij.li/?param=3",
		},
		{
			name:    "4",
			wantErr: false,
			wantURL: "https://ilij.li/?param=4",
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := "notfound"
			if !tt.wantErr {
				id, _ = UpsertURL(&URLReq{URL: tt.wantURL}, true, true)
			}
			t.Log("id:", id)
			gotURL, err := GetURLRedirect(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("GetURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestExpireUrl(t *testing.T) {
	tests := []struct {
		name    string
		numrq   int
		param   URLReq
		wantErr bool
	}{
		{
			name:    "noexpire",
			wantErr: false,
			numrq:   10,
			param: URLReq{
				URL:         "https://ilij.li/?param=noexpire",
				MaxRequests: 0,
			},
		},
		{
			name:    "expire1",
			wantErr: true,
			numrq:   10,
			param: URLReq{
				URL:         "https://ilij.li/?param=expire1",
				MaxRequests: 1,
			},
		},
		{
			name:    "expire10",
			wantErr: true,
			numrq:   11,
			param: URLReq{
				URL:         "https://ilij.li/?param=expire10",
				MaxRequests: 10,
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(&tt.param, true, true)
			mlog.Info("-- upsert %s --", id)
			// consume all the requests
			for i := 0; i < tt.numrq; i++ {
				_, err = GetURLRedirect(id)
				mlog.Info("-- get redirect %s --", id)
				url, _ := GetURLInfo(id)
				mlog.Info("-- get info %s --", id)
				mlog.Info("request n:%3d , %+v, %v", i+1, url, err)
			}
			mlog.Info("-- << end  %s --", id)
			// this should be a not found now for the expired
			hasErr := (err != nil)
			if tt.wantErr != hasErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	// get the stats
	time.Sleep(time.Duration(10) * time.Millisecond)
	s := GetStats()
	t.Log(s)
	if s.Urls != 1 {
		t.Errorf("ExpireUrl() count = %v, want %v", s.Urls, 1)
	}

}

func BenchmarkSession(b *testing.B) {
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)

	NewSession()
	defer CloseSession()

	numIds := 10000 // means 10000
	var ids []string
	//generate a bunch of urls
	ri := rand.Intn(numIds)
	for i := numIds; i < numIds+ri; i++ {
		ur := &URLReq{
			URL: fmt.Sprintf("http://ilij.li/long=%d", i),
		}
		id, err := UpsertURL(ur, true, true)
		if err != nil {
			b.Error("eror inserting url", err)
		}
		ids = append(ids, id)
	}

	numIds = len(ids)

	b.Run("thet test", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			idx := rand.Intn(numIds)
			GetURLRedirect(ids[idx])
		}
		b.Log(GetStats())
	})

	b.Run("thet test", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if 1%10 == 0 {
				idx := rand.Intn(numIds)
				DeleteURL(ids[idx])
			}

		}
		b.Log(GetStats())
	})

}
