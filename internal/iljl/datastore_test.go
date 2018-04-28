package iljl

import (
	"fmt"
	"io/ioutil"
	"testing"

	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

func buildConifgTest() {
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

func buildConifgTestShortIDParams(alphabet string, length int) {
	path, _ := ioutil.TempDir("/tmp/", "iljl")
	fmt.Println("test db folder is ", path)
	internal.Config = internal.ConfigSchema{
		Server: internal.ServerConfig{
			DbPath: path,
		},
		ShortID: internal.ShortIDConfig{
			Alphabet: alphabet,
			Length:   length,
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
				forceAlphabet: true,
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
			name:    "id set",
			wantErr: false,
			args: args{
				forceAlphabet: true,
				forceLength:   false,
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "ilcdef",
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

			if len(tt.url.ID) != internal.Config.ShortID.Length {
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

	tests := []struct {
		name    string
		url     URLReq
		wantErr bool
	}{
		{
			name:    "all good",
			wantErr: false,
			url: URLReq{
				URL: "https://ilij.li",
			},
		},
		{
			name:    "id set",
			wantErr: false,
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "samesame",
			},
		},
		{
			name:    "id set",
			wantErr: false,
			url: URLReq{
				URL: "https://wikipedia.li",
				ID:  "samesame",
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	ids := make(map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(tt.url, false, false)
			ids[id] = true
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if len(ids) != 2 {
		t.Errorf("UpsertURL() length = %v, want %v", len(ids), 2)
	}

	// test upsert
	id := "samesame"
	ur := "https://wikipedia.li"

	ui, _ := GetURL(id)
	if ui.URL != ur {
		t.Errorf("UpsertURL()  %v, want %v", ui.URL, ur)
	}

	CloseSession()
}

func TestDeleteURL(t *testing.T) {
	type args struct {
		url URLReq
	}
	tests := []struct {
		name    string
		url     URLReq
		wantErr bool
	}{
		{
			name:    "id set",
			wantErr: false,
			url: URLReq{
				URL:         "https://ilij.li",
				MaxRequests: 10,
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	for _, tt := range tests {
		url := tt.url
		id, _ := UpsertURL(url, false, false)
		ui, _ := GetURL(id)

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

		ui, err := GetURL(id)
		if err == nil {
			t.Errorf("DeleteURL() not deleted")
		}
	}
	CloseSession()
}

func TestGetURL(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantUrl URLInfo
		wantErr bool
	}{{
		name:    "0",
		wantErr: false,
		wantUrl: URLInfo{
			URL: "https://ilij.li/?param=0",
		},
	},
		{
			name:    "1",
			wantErr: false,
			wantUrl: URLInfo{
				URL: "https://ilij.li/?param=1",
			},
		},
		{
			name:    "2",
			wantErr: false,
			wantUrl: URLInfo{
				URL: "https://ilij.li/?param=2",
			},
		},
		{
			name:    "3",
			wantErr: true,
			wantUrl: URLInfo{
				URL: "https://ilij.li/?param=3",
			},
		},
		{
			name:    "4",
			wantErr: false,
			wantUrl: URLInfo{
				URL: "https://ilij.li/?param=4",
			},
		},
	}
	buildConifgTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := "notfound"
			if !tt.wantErr {
				id, _ = UpsertURL(URLReq{URL: tt.wantUrl.URL}, true, true)
			}
			t.Log("id:", id)
			gotURL, err := GetURL(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if gotURL.URL != tt.wantUrl.URL {
				t.Errorf("GetURL() = %v, want %v", gotURL, tt.wantUrl)
			}
		})
	}
	CloseSession()
}
