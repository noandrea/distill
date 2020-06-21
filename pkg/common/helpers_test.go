package common

import (
	"reflect"
	"testing"
)

func TestIsEmptyStr(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"ok", "ciao", false},
		{"space", " ", true},
		{"spaces", "   ", true},
		{"one char", " 1 ", false},
		{"one char space right", "1 ", false},
		{"one char space right", " 1", false},
		{"with tabs", "\t\t    ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyStr(tt.s); got != tt.want {
				t.Errorf("IsEmptyStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultIfEmptyInt(t *testing.T) {
	type args struct {
		v        int
		defaultV int
	}
	tests := []struct {
		name        string
		args        args
		wantDefault bool
	}{
		{"no default", args{10, 1000}, false},
		{"no default", args{50000, 1011}, false},
		{"yes default", args{0, 1002}, true},
		{"yes default", args{-100, 1010}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DefaultIfEmptyInt(&tt.args.v, tt.args.defaultV)
			if tt.wantDefault && tt.args.v != tt.args.defaultV {
				t.Errorf("DefaultIfEmptyStr expected %v got %v", tt.args.defaultV, tt.args.v)
			}
			if !tt.wantDefault && tt.args.v == tt.args.defaultV {
				t.Errorf("DefaultIfEmptyStr expected %v got %v", tt.args.v, tt.args.defaultV)
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	type args struct {
		alphabet string
		length   int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{"ok 1", args{"abcdefg123", 10}, 10, false},
		{"ok 2", args{"abcdefg123", 100}, 100, false},
		{"ok 3", args{"abcdefg123", 13}, 13, false},
		{"not ok", args{"", 13}, 13, true},
		{"not ok", args{"   ", 13}, 13, true},
		{"not ok", args{"\t\t", 13}, 13, true},
		{"not ok", args{"\t\t", -13}, 13, true},
		{"not ok", args{"asdfghjk", -1}, 13, true},
		{"not ok", args{"asdfghjk", 0}, 131, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomString(tt.args.alphabet, tt.args.length)
			t.Log(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("RandomString() = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type P struct {
		a int64
		b int64
	}
	tests := []struct {
		name string
		args P
		want int64
	}{
		{"1", P{123, 321}, 123},
		{"2", P{1231, 321}, 321},
		{"1", P{100, 100}, 100},
		{"1", P{-1, 1}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtoi(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		b    []byte
		want uint64
	}{
		{"1", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Atoi(tt.b); got != tt.want {
				t.Errorf("Atoi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItoa(t *testing.T) {

	tests := []struct {
		name  string
		i     uint64
		wantB []byte
	}{
		{"1", 1, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := Itoa(tt.i); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("Itoa() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestIfEmptyThen(t *testing.T) {
	type P struct {
		s           string
		thenDefault string
	}
	tests := []struct {
		name string
		args P
		want string
	}{
		{"1", P{"this", "that"}, "this"},
		{"2", P{"", "that"}, "that"},
		{"3", P{"  ", "that"}, "that"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IfEmptyThen(tt.args.s, tt.args.thenDefault); got != tt.want {
				t.Errorf("IfEmptyThen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEqStr(t *testing.T) {
	type P struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args P
		want bool
	}{
		{"1", P{"this", "this"}, true},
		{"2", P{"This", "this"}, true},
		{"3", P{"THIS", "this"}, true},
		{"4", P{"THIS", "    this    "}, true},
		{"5", P{"this", "that"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEqStr(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsEqStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
