package iljl

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/jbrodriguez/mlog"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

func setupLog() {
	mlog.DefaultFlags = log.Ltime | log.Lmicroseconds | log.Lshortfile
	//mlog.Start(mlog.LevelTrace, "")
	mlog.Start(mlog.LevelInfo, "")
}

func buildConifgTest() {
	setupLog()
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
	setupLog()
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
	setupLog()
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

func buildConifgTestExpireParams(ttl, maxr int64, expire time.Time) {
	setupLog()
	path, _ := ioutil.TempDir("/tmp/", "iljl")
	fmt.Println("test db folder is ", path)
	internal.Config = internal.ConfigSchema{
		Server: internal.ServerConfig{
			DbPath: path,
		},
		ShortID: internal.ShortIDConfig{
			Alphabet:    "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789",
			Length:      6,
			TTL:         ttl,
			MaxRequests: maxr,
			ExpireOn:    expire,
		},
		Tuning: internal.TuningConfig{
			StatsEventsWorkerNum: 20,
		},
	}
	internal.Config.Validate()
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
			name:    "invalid url",
			wantErr: true,
			url: &URLReq{
				URL: "ilij.li",
			},
			args: args{
				forceAlphabet: false,
				forceLength:   false,
			},
		},
		{
			name:    "invalid alphabet",
			wantErr: true,
			url: &URLReq{
				URL: "ilij.li",
				ID:  "abcild",
			},
			args: args{
				forceAlphabet: true,
				forceLength:   false,
			},
		},
		{
			name:    "invalid length",
			wantErr: true,
			url: &URLReq{
				URL: "ilij.li",
				ID:  "abcabcabc",
			},
			args: args{
				forceAlphabet: true,
				forceLength:   true,
			},
		},
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
			id, err := UpsertURL(tt.url, tt.args.forceAlphabet, tt.args.forceLength, time.Now())
			mlog.Info("upsert url %v, %v", id, err)
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
				URL: "https://ilij.li",
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)

	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		url := tt.url
		id, _ := UpsertURL(url, false, false, time.Now())
		for i := 0; i < 1; i++ {
			GetURLRedirect(id)
		}
		ui, _ := GetURLInfo(id)
		mlog.Trace("%#v", ui)

		if id != ui.ID || url.URL != ui.URL {
			t.Errorf("DeleteURL()")
			break
		}

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
				id, _ = UpsertURL(&URLReq{URL: tt.wantURL}, true, true, time.Now())
			}
			t.Log("id:", id)
			// a short pause to make sure the data is written
			time.Sleep(time.Duration(10) * time.Millisecond)
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

func TestExpireRequestsUrl(t *testing.T) {
	tests := []struct {
		name    string
		numrq   int
		param   URLReq
		wantErr bool
	}{
		{
			name:    "noexpire1",
			wantErr: false,
			numrq:   1,
			param: URLReq{
				URL:         "https://ilij.li/?param=noexpire",
				MaxRequests: 1,
			},
		},
		{
			name:    "expire0",
			wantErr: true,
			numrq:   2,
			param: URLReq{
				URL:         "https://ilij.li/?param=noexpire",
				MaxRequests: 1,
			},
		},
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
		{
			name:    "expire10",
			wantErr: true,
			numrq:   21,
			param: URLReq{
				URL: "https://ilij.li/?param=expire10",
			},
		},
	}
	var zeroTime time.Time
	buildConifgTestExpireParams(0, 20, zeroTime)
	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(&tt.param, true, true, time.Now())
			// consume all the requests
			for i := 0; i < tt.numrq; i++ {
				_, err = GetURLRedirect(id)
			}
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
	if s.Urls != 2 {
		t.Errorf("ExpireUrl() count = %v, want %v", s.Urls, 2)
	}
}

func TestExpireTTLUrl(t *testing.T) {
	tests := []struct {
		name    string
		wait    int
		param   URLReq
		wantErr bool
	}{
		{
			name:    "noexpire",
			wantErr: false,
			wait:    1,
			param: URLReq{
				URL: "https://ilij.li/?param=noexpire",
				TTL: 0,
			},
		},
		{
			name:    "noexpire",
			wantErr: false,
			wait:    2,
			param: URLReq{
				URL:         "https://ilij.li/?param=noexpire",
				TTL:         0,
				MaxRequests: 4,
			},
		},
		{
			name:    "expire1",
			wantErr: true,
			wait:    3,
			param: URLReq{
				URL: "https://ilij.li/?param=expire1",
				TTL: 2,
			},
		},
		{
			name:    "expire10",
			wantErr: true,
			wait:    3,
			param: URLReq{
				URL:         "https://ilij.li/?param=expire10",
				MaxRequests: 2,
				TTL:         2,
			},
		},
	}
	var zeroTime time.Time
	buildConifgTestExpireParams(0, 0, zeroTime)
	NewSession()
	defer CloseSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(&tt.param, true, true, time.Now())
			mlog.Info("-- upsert %s --", id)
			// consume all the requests
			time.Sleep(time.Duration(tt.wait) * time.Second)
			_, err = GetURLRedirect(id)
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
	if s.Urls != 2 {
		t.Errorf("ExpireUrl() count = %v, want %v", s.Urls, 2)
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
		id, err := UpsertURL(ur, true, true, time.Now())
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

func TestImportCSV(t *testing.T) {
	type args struct {
		inFile string
	}
	tests := []struct {
		name     string
		args     args
		wantRows int
		wantErr  bool
	}{
		{
			name:     "1123 without ids",
			args:     args{inFile: "testdata/urls.1123.csv"},
			wantErr:  false,
			wantRows: 1123,
		},
		{
			name:     "1123 with ids",
			args:     args{inFile: "testdata/urls.id.1123.csv"},
			wantErr:  false,
			wantRows: 1123,
		},
	}
	for _, tt := range tests {
		buildConifgTest()
		NewSession()
		t.Run(tt.name, func(t *testing.T) {
			gotRows, err := ImportCSV(tt.args.inFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRows != tt.wantRows {
				t.Errorf("ImportCSV() = %v, want %v", gotRows, tt.wantRows)
			}
		})
		CloseSession()
	}
}

func TestBackupRestore(t *testing.T) {
	type args struct {
		bckFile string
	}
	tests := []struct {
		name     string
		args     args
		wantRows int
		wantErr  bool
	}{
		{
			name:     "1123 without ids",
			args:     args{bckFile: "bck.csv"},
			wantErr:  false,
			wantRows: 421,
		},
		{
			name:     "1123 with ids",
			args:     args{bckFile: "bck.bin"},
			wantErr:  false,
			wantRows: 321,
		},
	}
	for _, tt := range tests {
		buildConifgTest()
		NewSession()
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.wantRows; i++ {
				UpsertURLSimple(&URLReq{URL: fmt.Sprintf("http://ex.com/v=%d", i)})
			}
			err := Backup(tt.args.bckFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Backup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = Restore(tt.args.bckFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Restore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
		CloseSession()
	}
}
