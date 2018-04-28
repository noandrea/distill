package iljl

import (
	"fmt"
	"io/ioutil"
	"reflect"
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

func TestPreprocessURL(t *testing.T) {
	buildConifgTest()

	u := &URLReq{
		URL: "http://ai.ai",
	}
	t.Log("before", u)
	PreprocessURL(u, true)
	t.Log("after", u)

	type args struct {
		forceAlphabet bool
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
			},
			url: URLReq{
				URL: "https://ilij.li",
				ID:  "abcdef",
			},
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Log("berore:", tt.url)

			err := PreprocessURL(&tt.url, tt.args.forceAlphabet)

			if (err != nil) != tt.wantErr {
				t.Errorf("PreprocessURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Log("after :", tt.url)
			if (err == nil) && tt.url.ID == "" {
				t.Errorf("PreprocessURL() empty id ")
				return
			}
		})
	}

}

func TestUpsertUrl(t *testing.T) {
	type args struct {
		url URLReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpsertUrl(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("UpsertUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteUrl(t *testing.T) {
	type args struct {
		url URLReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteUrl(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("DeleteUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUrlByID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantUrl URLInfo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUrl := GetUrlByID(tt.args.id); !reflect.DeepEqual(gotUrl, tt.wantUrl) {
				t.Errorf("GetUrlByID() = %v, want %v", gotUrl, tt.wantUrl)
			}
		})
	}
}
