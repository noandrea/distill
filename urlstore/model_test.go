package urlstore

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

			if gotK, _ := key(tt.prefix, tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
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
			wantK: append([]byte{keyURLPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{keyURLPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{keySysPrefix}, []byte("err")...),
			match: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if gotK, _ := keyURL(tt.id); tt.match != reflect.DeepEqual(gotK, tt.wantK) {
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
			wantK: append([]byte{keySysPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{keySysPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{keyURLPrefix}, []byte("err")...),
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
			wantK: append([]byte{keyStatPrefix}, []byte("one")...),
			match: true,
		},
		{
			id:    "two",
			wantK: append([]byte{keyStatPrefix}, []byte("two")...),
			match: true,
		},
		{
			id:    "err",
			wantK: append([]byte{keyURLPrefix}, []byte("err")...),
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

func Test_arrayconv(t *testing.T) {
	tests := []struct {
		name string
		val  uint64
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

func TestURLInfo_ExpirationDate(t *testing.T) {

	d := func(rfc3339Time string) time.Time {
		pt, _ := time.Parse(time.RFC3339, rfc3339Time)
		return pt
	}

	type fields struct {
		TTL     uint64
		BountAt time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		{
			name: "30s",
			fields: fields{
				TTL:     30,
				BountAt: d("2018-04-01T15:00:00Z"),
			},
			want: d("2018-04-01T15:00:30Z"),
		},
		{
			name: "2h",
			fields: fields{
				TTL:     7200,
				BountAt: d("2018-04-01T15:00:00Z"),
			},
			want: d("2018-04-01T17:00:00Z"),
		},
		{
			name: "1d",
			fields: fields{
				TTL:     86400,
				BountAt: d("2018-04-01T15:00:00Z"),
			},
			want: d("2018-04-02T15:00:00Z"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := URLInfo{
				TTL:     tt.fields.TTL,
				BountAt: tt.fields.BountAt,
			}
			if got := u.ExpirationDate(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("URLInfo.ExpirationDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
