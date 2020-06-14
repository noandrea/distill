package common

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/labstack/gommon/log"
	gonanoid "github.com/matoous/go-nanoid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IsEqStr tells if two strings a and b are equals after trimming spaces and lowercasing
func IsEqStr(a, b string) bool {
	return strings.ToLower(strings.TrimSpace(a)) == strings.ToLower(strings.TrimSpace(b))
}

// IsEmptyStr tells if a string is empty or not
func IsEmptyStr(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// DefaultIfEmptyStr set a default for a string if it is nulled
func DefaultIfEmptyStr(s *string, defaultS string) {
	if IsEmptyStr(*s) {
		*s = defaultS
	}
}

func IfEmptyThen(s, thenDefault string) string {
	if IsEmptyStr(s) {
		return thenDefault
	}
	return s
}

// DefaultIfEmptyInt set the value of an int to a default if it is nulled (0)
func DefaultIfEmptyInt(v *int, defaultV int) {
	if *v <= 0 {
		*v = defaultV
	}
}

// DefaultIfEmptyUint64 set the value of an int to a default if it is nulled (0)
func DefaultIfEmptyUint64(v *uint64, defaultV uint64) {
	if *v <= 0 {
		*v = defaultV
	}
}

// RandomString generate a random string of required length using alphabet
func RandomString(alphabet string, length int) (s string, err error) {
	if IsEmptyStr(alphabet) {
		err = fmt.Errorf("alphabet must not be empty")
		return
	}
	if length <= 0 {
		err = fmt.Errorf("string length must be longer than 0")
		return
	}
	return gonanoid.Generate(alphabet, length)
}

// GenerateSecret generate a string that can be used as secrete api key
func GenerateSecret() string {
	a := "asdfghjklqwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890#!-"
	l := 50
	secret, _ := RandomString(a, l)
	return secret
}

// AskYes prompt a yes/no question to the prompt
func AskYes(question string, defaultYes bool) (isYes bool) {
	fmt.Print(question)
	if defaultYes {
		fmt.Print(" [yes]: ")
	} else {
		fmt.Print(" [no]: ")
	}
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	DefaultIfEmptyStr(&reply, "yes")
	if IsEqStr(reply, "yes") {
		return true
	}
	return
}

// ProtoTime convert protobuf to go time.Time
// on error returns the current time
func ProtoTime(ts *timestamppb.Timestamp) time.Time {
	goTime, err := ptypes.Timestamp(ts)
	if err != nil {
		log.Warn("cannot convert proto timestamp: ", ts)
		goTime = time.Now()
	}
	return goTime
}

// TimeProto convert go time.Time to protobuf
// on error returns the current time
func TimeProto(t time.Time) (ts *timestamppb.Timestamp) {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Warn("cannot convert time to proto: ", err)
		ts = ptypes.TimestampNow()
	}
	return
}

//   ____  ____  ________  _____ 4   _______  ________  _______     ______
//  |_   ||   _||_   __  ||_   _|   |_   __ \|_   __  ||_   __ \  .' ____ \
//    | |__| |    | |_ \_|  | |       | |__) | | |_ \_|  | |__) | | (___ \_|
//    |  __  |    |  _| _   | |   _   |  ___/  |  _| _   |  __ /   _.____`.
//   _| |  | |_  _| |__/ | _| |__/ | _| |_    _| |__/ | _| |  \ \_| \____) |
//  |____||____||________||________||_____|  |________||____| |___|\______.'
//

// FUint64 for csv printing
func FUint64(v int64) (str string) {
	if v == 0 {
		return
	}
	str = strconv.FormatInt(v, 10)
	return
}

// FTime for csv printing
func FTime(v time.Time) (str string) {
	if v.IsZero() {
		return
	}
	str = v.Format(time.RFC3339)
	return
}

// PInt64 parse an int64 from a string
func PInt64(src []string, idx, srcLen int) (v int64, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = strconv.ParseInt(src[idx], 10, 64)
	}
	return
}

// PUint64 parse an int64 from a string
func PUint64(src []string, idx, srcLen int) (v uint64, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = strconv.ParseUint(src[idx], 10, 64)
	}
	return
}

// PTime parse a time.Time from a string
func PTime(src []string, idx, srcLen int) (v time.Time, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = time.Parse(time.RFC3339, src[idx])
	}
	return
}

// Itoa byte array to uint64
func Itoa(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return
}

// Atoi byte array to uint64
func Atoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Min return the min betweek numbers
func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
