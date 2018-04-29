package iljl

import (
	"reflect"
	"testing"
	"time"
)

func Test_key(t *testing.T) {
	type args struct {
		prefix byte
		id     string
	}
	tests := []struct {
		prefix byte
		id     string
		wantK  []byte
		match  bool
	}{
		{
			prefix: 0x00,
			id:     "one",
			wantK:  append([]byte{0x00}, []byte("one")...),
			match:  true,
		},
		{
			prefix: 0x02,
			id:     "two",
			wantK:  append([]byte{0x02}, []byte("two")...),
			match:  true,
		},
		{
			prefix: 0x06,
			id:     "err",
			wantK:  append([]byte{0x00}, []byte("err")...),
			match:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {

			if gotK := key(tt.prefix, tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
				t.Errorf("key() = %v, want %v", gotK, tt.wantK)
			}
		})
	}
}

func Test_keyUrl(t *testing.T) {
	type args struct {
		prefix byte
		id     string
	}
	tests := []struct {
		id    string
		wantK []byte
		match bool
	}{
		{
			id:    "one",
			wantK: append([]byte{KeyURLPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{KeyURLPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{KeySysPrefix}, []byte("err")...),
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if gotK := keyURL(tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
				t.Errorf("key() = %v, want %v", gotK, tt.wantK)
			}
		})
	}
}

func Test_keySys(t *testing.T) {
	type args struct {
		prefix byte
		id     string
	}
	tests := []struct {
		id    string
		wantK []byte
		match bool
	}{
		{
			id:    "one",
			wantK: append([]byte{KeySysPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{KeySysPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{KeyURLPrefix}, []byte("err")...),
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if gotK := keySys(tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
				t.Errorf("key() = %v, want %v", gotK, tt.wantK)
			}
		})
	}
}

func Test_keyGlobalStat(t *testing.T) {
	type args struct {
		prefix byte
		id     string
	}
	tests := []struct {
		id    string
		wantK []byte
		match bool
	}{
		{
			id:    "one",
			wantK: append([]byte{KeyStatPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{KeyStatPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{KeyURLPrefix}, []byte("err")...),
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if gotK := keyGlobalStat(tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
				t.Errorf("key() = %v, want %v", gotK, tt.wantK)
			}
		})
	}
}

func Test_keyURLStatCount(t *testing.T) {
	type args struct {
		prefix byte
		id     string
	}
	tests := []struct {
		id    string
		wantK []byte
		match bool
	}{
		{
			id:    "one",
			wantK: append([]byte{KeyURLStatCountPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{KeyURLStatCountPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{KeyURLPrefix}, []byte("err")...),
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if gotK := keyURLStatCount(tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
				t.Errorf("key() = %v, want %v", gotK, tt.wantK)
			}
		})
	}
}

func Test_ttl(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		wantD   time.Duration
		match   bool
	}{
		{name: "10s", seconds: 10, match: true, wantD: time.Duration(10) * time.Second},
		{name: "20s", seconds: 20, match: true, wantD: time.Duration(20) * time.Second},
		{name: "1s", seconds: 1, match: true, wantD: time.Duration(1) * time.Second},
		{name: "10s", seconds: 10, match: false, wantD: time.Duration(5) * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotD := ttl(tt.seconds); tt.match == (gotD != tt.wantD) {
				t.Errorf("ttl() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func Test_arrayconv(t *testing.T) {
	tests := []struct {
		name string
		val  int64
	}{
		{name: "1", val: 1},
		{name: "2", val: 2},
		{name: "3213211", val: 3213211},
		{name: "83913127831", val: 83913127831},
		{name: "10000000000", val: 10000000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := atoi(itoa(tt.val)); atoi(itoa(tt.val)) != tt.val {
				t.Errorf("itoa() = %v, want %v", gotB, tt.val)
			}
		})
	}
}
