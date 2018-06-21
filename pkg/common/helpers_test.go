package common

import "testing"

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

func TestDefaultIfEmptyStr(t *testing.T) {
	type args struct {
		s        string
		defaultS string
	}
	tests := []struct {
		name        string
		args        args
		wantDefault bool
	}{
		{"no default", args{"foo", "bar"}, false},
		{"yes default", args{"", "bar"}, true},
		{"yes default", args{"\t\t", "bar"}, true},
		{"yes default", args{"\ta\t", "bar"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DefaultIfEmptyStr(&tt.args.s, tt.args.defaultS)
			if tt.wantDefault && tt.args.s != tt.args.defaultS {
				t.Errorf("DefaultIfEmptyStr expected %v got %v", tt.args.defaultS, tt.args.s)
			}
			if !tt.wantDefault && tt.args.s == tt.args.defaultS {
				t.Errorf("DefaultIfEmptyStr expected %v got %v", tt.args.s, tt.args.defaultS)
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

func TestGenerateSecret(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{"test", 50},
		{"test", 50},
		{"test", 50},
		{"test", 50},
		{"test", 50},
		{"test", 50},
		{"test", 50},
		{"test", 50},
	}
	for _, tt := range tests {
		oldOne := ""
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSecret()
			if len(got) != tt.wantLen {
				t.Errorf("GenerateSecret() = %v, want %v", len(got), tt.wantLen)
			}
			if got == oldOne {
				t.Errorf("GenerateSecret() should be fairly unique")
			}
			oldOne = got
		})
	}
}
