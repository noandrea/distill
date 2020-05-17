package urlstore

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/noandrea/distill/pkg/common"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func _defaultConfig() (settings ConfigSchema) {
	Defaults()
	viper.Unmarshal(&settings)
	return
}

func buildConfigTest() (s ConfigSchema) {
	// path
	path, _ := ioutil.TempDir("/tmp/", "distill")
	fmt.Println("test db folder is ", path)
	s = _defaultConfig()
	s.Server.DbPath = path
	s.Server.APIKey = common.GenerateSecret()
	s.ShortID.Alphabet = "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	s.ShortID.Length = 6
	s.Validate()
	return
}

func buildConfigPanicTest() (s ConfigSchema) {
	path := " cann not exists / ssa "
	fmt.Println("test db folder is ", path)
	s = ConfigSchema{
		Server: ServerConfig{
			DbPath: path,
			APIKey: common.GenerateSecret(),
		},
		ShortID: ShortIDConfig{
			Alphabet: "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789",
			Length:   6,
		},
	}
	Defaults()
	s.Validate()
	return
}

func buildConfigTestShortIDParams(alphabet string, length int) (s ConfigSchema) {
	path, _ := ioutil.TempDir("/tmp/", "distill")
	fmt.Println("test db folder is ", path)
	s = _defaultConfig()

	s.Server.DbPath = path
	s.Server.APIKey = common.GenerateSecret()

	s.ShortID.Alphabet = alphabet
	s.ShortID.Length = length
	s.ShortID.TTL = 0
	s.ShortID.MaxRequests = 0

	s.Tuning.StatsEventsWorkerNum = 2

	return
}

func buildConfigTestExpireParams(ttl, maxr uint64, expire time.Time) (s ConfigSchema) {
	path, _ := ioutil.TempDir("/tmp/", "distill")
	fmt.Println("test db folder is ", path)
	s = _defaultConfig()

	s.Server.DbPath = path
	s.Server.APIKey = common.GenerateSecret()

	s.ShortID.Alphabet = "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	s.ShortID.Length = 6
	s.ShortID.TTL = ttl
	s.ShortID.MaxRequests = maxr
	s.ShortID.ExpireOn = expire

	s.Tuning.StatsEventsWorkerNum = 20

	settings.Validate()
	return
}

func TestGenerateID(t *testing.T) {

	tests := []struct {
		Alphabet      string
		Length        int
		InvalidRegexp string
	}{
		{"1234567890", 7, "[qwertyuiopasdfghjkl]"},
		{"1234567890", 5, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"1234567890", 4, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"1234567890", 20, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM", 6, "[1234567890]"},
		{"abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 30, "[iIl1o0O]"},
	}
	for _, tt := range tests {
		s := ConfigSchema{
			ShortID: ShortIDConfig{
				Alphabet: tt.Alphabet,
				Length:   tt.Length,
			},
		}
		t.Run(tt.Alphabet, func(t *testing.T) {
			gotShortID := generateID(&s.ShortID)
			if len(gotShortID) != tt.Length {
				t.Errorf("GenerateID() = %v, len = %v, want %v", gotShortID, len(gotShortID), tt.Length)
			}
			m, _ := regexp.MatchString(tt.InvalidRegexp, gotShortID)
			if m {
				t.Errorf("GenerateID() = %v, match = %v, expected no match", gotShortID, tt.InvalidRegexp)
			}
		})
	}
}

func TestUpsertURLQuick(t *testing.T) {
	s := buildConfigTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession(s)
	defer CloseSession()
	// test urls
	tests := []string{
		"https://battle.example.com/approval/arm.aspx",
		"http://example.com/",
		"https://www.example.com/advertisement/beginner.htm",
		"http://example.com/",
		"https://example.com/?achiever=book",
		"https://www.example.com/?act=birds&bells=believe#boat",
		"https://www.example.com/",
		"https://example.com/birthday",
		"http://www.example.com/alarm",
		"http://example.com/",
		"http://example.com/airplane",
		"http://example.com/?brick=baby",
		"https://www.example.com/boot/brick",
		"http://www.example.com/afternoon?bomb=back",
		"http://www.example.com/airport/badge",
		"http://example.net/aunt/argument.htm?apparatus=beef",
		"https://www.example.com/boy",
		"https://www.example.com/",
		"https://activity.example.org/breath.aspx",
		"https://behavior.example.com/birthday/army",
		"http://www.example.com/boy?basketball=account",
		"http://www.example.com/",
		"https://art.example.edu/authority",
		"https://belief.example.com/believe/bubble.html#acoustics",
		"https://www.example.com/animal/bike.php?bedroom=beef&authority=airplane",
		"https://www.example.com/amusement.html#bomb",
		"http://example.net/arithmetic.aspx",
		"https://example.com/bear",
		"https://example.com/",
		"http://www.example.com/",
		"https://www.example.com/action/bike.php",
		"http://www.example.com/adjustment.php",
		"https://board.example.org/?art=addition",
		"https://www.example.com/bells/bikes#believe",
		"https://www.example.net/belief.php",
		"https://example.net/behavior/basin.php",
		"https://www.example.com/baby",
		"https://back.example.com/acoustics#advertisement",
		"https://example.com/#bells",
		"http://afterthought.example.com/boot/acoustics",
		"http://example.com/bath",
		"https://example.net/",
		"http://example.edu/",
		"http://example.com/?boy=bat&advertisement=bag",
		"http://www.example.com/action/brick",
		"http://www.example.com/",
		"http://www.example.org/",
		"https://example.com/battle/brick",
		"http://www.example.net/acoustics",
		"https://www.example.com/agreement.html",
		"http://www.example.com/acoustics.html",
		"http://www.example.net/account/bait.html?believe=bag",
		"https://blow.example.net/",
		"http://example.com/?action=angle&beef=approval",
		"http://example.com/amount/airport#account",
		"https://www.example.com/basketball",
		"http://example.com/",
		"http://www.example.com/branch/bear",
		"https://bed.example.net/",
		"http://www.example.com/",
		"https://art.example.com/",
		"http://example.com/authority/books.php",
		"http://acoustics.example.org/amusement",
		"http://www.example.org/acoustics.aspx",
		"http://example.edu/?bath=bomb&basket=apparel",
		"http://www.example.com/",
		"http://example.com/behavior/arithmetic",
		"https://www.example.com/account.aspx",
		"https://bottle.example.com/bed.html",
		"http://example.net/brother.php",
		"http://www.example.com/brick/belief?angle=air&beef=birthday",
		"http://example.org/",
		"https://example.org/boot/belief",
		"https://www.example.com/",
		"http://example.com/ants/belief?boat=bike&ball=bell",
		"http://example.com/baseball.php",
		"https://alarm.example.com/brake",
		"http://example.com/believe/blade",
		"http://www.example.edu/",
		"https://example.edu/",
		"http://example.com/basin.htm?bead=afterthought",
		"http://www.example.com/animal.htm#bed",
		"https://example.com/?badge=bell&activity=believe",
		"http://www.example.com/?bedroom=bit&birth=blade",
		"https://example.com/",
		"https://www.example.com/bell/account",
		"http://www.example.com/",
		"http://bridge.example.com/attack.php?appliance=bed&aftermath=adjustment",
		"https://www.example.com/",
		"http://www.example.com/",
		"http://www.example.com/",
		"http://example.com/beginner/bridge.aspx",
		"http://bike.example.com/bait",
		"https://www.example.com/",
		"http://example.org/",
		"https://example.net/balance/action.php",
		"http://www.example.com/brother/base.aspx?agreement=bear",
		"http://bed.example.org/acoustics#ball",
		"http://example.com/",
		"http://www.example.com/",
	}
	// test random urls
	for _, u := range tests {
		urlrq := &URLReq{URL: u}
		_, err := UpsertURL(urlrq, false, false, time.Now())
		require.NoError(t, err)
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
			name:    "invalid url",
			wantErr: false,
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
	s := buildConfigTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession(s)
	defer CloseSession()
	ids := make(map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := UpsertURL(tt.url, tt.args.forceAlphabet, tt.args.forceLength, time.Now())
			log.Infof("upsert url %v, %v", id, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			// only records in case of success
			if err == nil {
				ids[id] = true
			}
		})
	}

	validElements := 5
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
		{
			name:    "does not exists",
			wantErr: true,
			url:     nil,
		},
	}
	s := buildConfigTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)

	NewSession(s)
	defer CloseSession()
	for _, tt := range tests {
		var id string
		if tt.url != nil {
			id, _ = UpsertURL(tt.url, false, false, time.Now())
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteURL(id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		if tt.url != nil {
			if _, err := GetURLInfo(id); err == nil {
				t.Errorf("DeleteURL() not deleted")
			}
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
	s := buildConfigTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)
	NewSession(s)
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
	cfg := buildConfigTestExpireParams(0, 20, zeroTime)
	NewSession(cfg)
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
	//TODO: time.Sleep(time.Duration(10) * time.Millisecond)
	s := GetStats()
	t.Log(s)
	var expected uint64 = 6
	if s.Urls != expected {
		t.Errorf("ExpireUrl() count = %v, want %v", s.Urls, expected)
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
			name:    "noexpire w1",
			wantErr: false,
			wait:    1,
			param: URLReq{
				URL: "https://ilij.li/?param=noexpireA",
				TTL: 0,
			},
		},
		{
			name:    "noexpire w2",
			wantErr: false,
			wait:    2,
			param: URLReq{
				URL:         "https://ilij.li/?param=noexpire",
				TTL:         0,
				MaxRequests: 4,
			},
		},
		{
			name:    "expire w3",
			wantErr: true,
			wait:    3,
			param: URLReq{
				URL: "https://ilij.li/?param=expire1",
				TTL: 2,
			},
		},
		{
			name:    "expire w4",
			wantErr: true,
			wait:    4,
			param: URLReq{
				URL:         "https://ilij.li/?param=expire10",
				MaxRequests: 2,
				TTL:         2,
			},
		},
	}
	var zeroTime time.Time
	c := buildConfigTestExpireParams(0, 0, zeroTime)
	NewSession(c)
	defer CloseSession()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, _ := UpsertURL(&tt.param, true, true, time.Now())
			//log.Info("-- upsert %s --", id)
			// consume all the requests
			time.Sleep(time.Duration(tt.wait) * time.Second)
			now := time.Now()
			_, err := GetURLRedirect(id)
			u, _ := GetURLInfo(id)
			fmt.Println(tt.name, tt.wantErr, "\nbat", u.BountAt, "\nexp", u.ExpireOn, "\nnow", now.UTC(), "\ndif", now.Sub(u.BountAt))
			//log.Info("-- << end  %s --", id)
			// this should be a not found now for the expired
			hasErr := (err != nil)
			if tt.wantErr != hasErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	// get the stats
	s := GetStats()
	t.Log(s)
	if s.Urls != 4 {
		t.Errorf("ExpireUrl() count = %v, want %v", s.Urls, 4)
	}
}

func BenchmarkSession(b *testing.B) {
	c := buildConfigTestShortIDParams("abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 6)

	NewSession(c)
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
		c := buildConfigTest()
		NewSession(c)
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

	tmpdir, _ := ioutil.TempDir("/tmp/", "distill-bckrestore")

	type args struct {
		bckFile string
		err     error
	}
	tests := []struct {
		name     string
		args     args
		wantRows int
		wantErr  bool
	}{
		{
			name:     "1123 without ids",
			args:     args{bckFile: filepath.Join(tmpdir, "bck.csv")},
			wantErr:  false,
			wantRows: 421,
		},
		{
			name:     "1123 with ids",
			args:     args{bckFile: filepath.Join(tmpdir, "bck.bin")},
			wantErr:  false,
			wantRows: 321,
		},
		{
			name:     "1123 with ids",
			args:     args{bckFile: filepath.Join(tmpdir, "bck.zip")},
			wantErr:  true,
			wantRows: 321,
		},
	}
	for _, tt := range tests {
		c := buildConfigTest()
		NewSession(c)
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.wantRows; i++ {
				UpsertURLSimple(&URLReq{URL: fmt.Sprintf("http://ex.com/v=%d", i)})
			}
			err := Backup(tt.args.bckFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Backup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_, err = Restore(tt.args.bckFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Restore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
		CloseSession()
	}
}
