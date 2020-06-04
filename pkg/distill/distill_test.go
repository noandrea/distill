package distill

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/noandrea/distill/config"
)

// func _defaultConfig() (settings config.Schema) {
// 	config.Defaults()
// 	viper.Unmarshal(&settings)
// 	return
// }

// func buildConfigTest() (s config.Schema) {
// 	// path
// 	path, _ := ioutil.TempDir("/tmp/", "distill")
// 	fmt.Println("test db folder is ", path)
// 	s = _defaultConfig()
// 	s.Datastore.URI = path
// 	s.Server.APIKey = common.GenerateSecret()
// 	s.ShortID.Alphabet = "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
// 	s.ShortID.Length = 6
// 	s.Validate()
// 	return
// }

// func buildConfigPanicTest() (s config.Schema) {
// 	path := " cann not exists / ssa "
// 	fmt.Println("test db folder is ", path)
// 	s = config.Schema{
// 		Datastore: config.DatastoreConfig{
// 			URI: path,
// 		},
// 		Server: config.ServerConfig{
// 			APIKey: common.GenerateSecret(),
// 		},
// 		ShortID: config.ShortIDConfig{
// 			Alphabet: "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789",
// 			Length:   6,
// 		},
// 	}
// 	config.Defaults()
// 	s.Validate()
// 	return
// }

// func buildConfigTestShortIDParams(alphabet string, length int) (s config.Schema) {
// 	path, _ := ioutil.TempDir("/tmp/", "distill")
// 	fmt.Println("test db folder is ", path)
// 	s = _defaultConfig()

// 	s.Datastore.URI = path
// 	s.Server.APIKey = common.GenerateSecret()

// 	s.ShortID.Alphabet = alphabet
// 	s.ShortID.Length = length
// 	s.ShortID.TTL = 0
// 	s.ShortID.MaxRequests = 0

// 	s.Tuning.StatsEventsWorkerNum = 2

// 	return
// }

// func buildConfigTestExpireParams(ttl, maxr int64, expire time.Time) (s config.Schema) {
// 	path, _ := ioutil.TempDir("/tmp/", "distill")
// 	fmt.Println("test db folder is ", path)
// 	s = _defaultConfig()

// 	s.Datastore.URI = path
// 	s.Server.APIKey = common.GenerateSecret()

// 	s.ShortID.Alphabet = "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
// 	s.ShortID.Length = 6
// 	s.ShortID.TTL = ttl
// 	s.ShortID.MaxRequests = maxr
// 	s.ShortID.ExpireOn = expire

// 	s.Tuning.StatsEventsWorkerNum = 20

// 	settings.Validate()
// 	return
// }

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
		s := config.Schema{
			ShortID: config.ShortIDConfig{
				Alphabet: tt.Alphabet,
				Length:   tt.Length,
			},
		}
		t.Run(tt.Alphabet, func(t *testing.T) {
			gotShortID := generateID(s.ShortID.Alphabet, s.ShortID.Length)
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

	// defer CloseSession()
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
		fmt.Println(u)
	}
}

func Test_calculateExpiration(t *testing.T) {

	tm := func(y, m, d, H, M, S int) (t time.Time) {
		if y+m+d+H+M+S == 0 {
			return
		}
		f := fmt.Sprintf("%d-%d-%d %d:%d:%d", d, m, y, H, M, S)
		t, err := time.Parse("1-2-2006 15:4:5", f)
		if err != nil {
			panic(err)
		}
		return
	}

	type xp struct {
		activeFrom time.Time
		ttl        int64
		expiresOn  time.Time
	}
	tests := []struct {
		name               string
		xp                 xp
		wantExpirationDate time.Time
		wantErr            bool
	}{
		{"all zero", xp{
			tm(2020, 1, 1, 10, 10, 0), // active from
			10,                        // ttl (s)
			tm(0, 0, 0, 0, 0, 0),      // expires on
		}, tm(2020, 1, 1, 10, 10, 10), false}, // results
		{"already expired", xp{
			tm(2020, 1, 1, 10, 10, 0), // active from
			0,                         // ttl (s)
			tm(2020, 1, 1, 10, 9, 59), // expires on
		}, time.Now(), true}, // results
		{"already expired/2", xp{
			tm(2020, 1, 1, 10, 10, 0), // active from
			100,                       // ttl (s)
			tm(2020, 1, 1, 10, 9, 59), // expires on
		}, time.Now(), true}, // results
		{"good to go", xp{
			tm(2020, 1, 1, 10, 10, 0), // active from
			3600,                      // ttl (s)
			tm(2020, 1, 1, 10, 10, 0), // expires on
		}, tm(2020, 1, 1, 11, 10, 0), false}, // results
		{"good2go/2", xp{
			tm(2020, 1, 1, 10, 10, 0), // active from
			3600,                      // ttl (s)
			tm(2020, 1, 1, 11, 11, 0), // expires on
		}, tm(2020, 1, 1, 11, 11, 0), false}, // results
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExpirationDate, err := calculateExpiration(tt.xp.activeFrom, tt.xp.ttl, tt.xp.expiresOn)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateExpiration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(gotExpirationDate, tt.wantExpirationDate) {
				t.Errorf("calculateExpiration() = %v, want %v", gotExpirationDate, tt.wantExpirationDate)
			}
		})
	}
}
