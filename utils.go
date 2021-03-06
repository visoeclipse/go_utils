package go_utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	. "fmt"
	go_deepcopy "github.com/margnus1/go-deepcopy"
	gouuid "github.com/nu7hatch/gouuid"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	ALPHABET        = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	UPPER_A_ORD     = 65
	UPPER_Z_ORD     = 90
	LOWER_A_ORD     = 97
	LOWER_Z_ORD     = 122
	NOT_FOUND_INDEX = -1
)

// Iif_string is and immediate if helper that takes a boolean expression
// and returns a string, true_val if the expression is true else false_val
func Iif_string(expr bool, true_val string, false_val string) string {
	return map[bool]string{true: true_val, false: false_val}[expr]
}

// Concat joins the strings in a slice, delimiting them with a comma, but it
// allows you to pass the delimiter string to create a single string
// Ex:  data: []string{"A", "B", "C"}; Join(data) ==> "A,B,C" ; Join(data, "|") ==> "A|B|C"
func Join(slice []string, args ...interface{}) string {
	delimiter := ","
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			delimiter = t
		default:
			panic(Sprintf("ERROR - Invalid argument (%v).  Must be a string.", arg))
		}
	}
	ret := ""
	for i, s := range slice {
		// append delimiter except at the very end
		ret += s + Iif_string((i < len(slice)-1), delimiter, "")
	}
	return ret
}

// Substr returns a portion (length characters) of string (s), beginning at a specified position (pos)
func Substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

// PadRight pads a string (s) with with a specified string (optional parameter) for padLen characters
// If no string argument is passed, then s will be padded, to the right, with a single space character
func PadRight(s string, padLen int, args ...interface{}) string {
	padStr := " "
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			padStr = t
		default:
			panic("Unknown argument")
		}
	}
	return s + strings.Repeat(padStr, padLen-len(s))
}

// PadLeft pads a string (s) with with a specified string (optional parameter) for padLen characters
// If no string argument is passed, then s will be padded, to the left, with a single space character
func PadLeft(s string, padLen int, args ...interface{}) string {
	padStr := " "
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			padStr = t
		default:
			panic("Unknown argument")
		}
	}
	return strings.Repeat(padStr, padLen-len(s)) + s
}

// reflect doesn't consider 0 or "" to be zero, so we double check those here
// Can handle a struct field (only one level)
func IsEmpty(args ...interface{}) bool {
	val := reflect.ValueOf(args[0])
	valType := val.Kind()
	switch valType {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func:
		if val.IsNil() {
			return true
		} else if valType == reflect.Slice || valType == reflect.Map {
			return val.Len() == 0
		}
	case reflect.Struct:
		return IsEmptyStruct(args[0])
	default:
		return false
	}
	return false
}

// Assumes the argument is a struct
func IsEmptyStruct(args ...interface{}) bool {
	val := reflect.ValueOf(args[0])
	valType := val.Kind()
	switch valType {
	case reflect.Struct:
		// verify that all of the struct's properties are empty
		fieldCount := val.NumField()
		for i := 0; i < fieldCount; i++ {
			field := val.Field(i)
			if field.IsValid() && !IsEmptyNonStruct(field) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Assumes the argument is not a struct
func IsEmptyNonStruct(args ...interface{}) bool {
	val := reflect.ValueOf(args[0])
	valType := val.Kind()
	switch valType {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func:
		if val.IsNil() {
			return true
		} else if valType == reflect.Slice || valType == reflect.Map {
			return val.Len() == 0
		}
	default:
		return false
	}
	return false
}

// Repeat a character (typically used for simple formatting of output)
func Dashes(repeatCount int, args ...interface{}) string {
	dashChar := "-"
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			dashChar = t
		default:
			panic("Unknown argument")
		}
	}
	return strings.Repeat(dashChar, repeatCount)
}

//	A: 65
//	B: 66
// . . .
func PrintAlphabet() {
	var thisChar string
	for _, c := range ALPHABET {
		thisChar = Sprintf("%c", c)
		Printf("%s: %d\n", thisChar, c)
	}
}

// Find index of search string in target string, starting at startPos
// Ex: domain := email[IndexOf("@", email, 0)+1:]
func IndexOf(search string, target string, startPos int) int {
	if startPos < 0 {
		startPos = 0
	}
	if len(target) < startPos {
		return NOT_FOUND_INDEX
	}
	if IsEmpty(target) || IsEmpty(search) {
		return NOT_FOUND_INDEX
	}
	foundPos := strings.Index(target[startPos:len(target)], search)
	if foundPos == -1 {
		return NOT_FOUND_INDEX
	}
	return foundPos + startPos
}

// IndexOfGeneric returns the index of an element in any type of slice
// ints := []int{1, 2, 3}
// strings := []string{"A", "B", "C"}
// IndexOfGeneric(len(ints), func(i int) bool { return ints[i] == 2 })
// IndexOfGeneric(len(strings), func(i int) bool { return strings[i] == "B" })
func IndexOfGeneric(maxLen int, findExpr func(i int) bool) int {
	for i := 0; i < maxLen; i++ {
		if findExpr(i) {
			return i
		}
	}
	return -1
}

// Printf("isLower(\"a\"): %v\n", isLower("a"))	// isLower("a"): true
// Printf("isLower(\"A\"): %v\n", isLower("A"))	// isLower("A"): false
func IsLower(letter string) bool {
	//	var thisChar string
	var ret bool
	for _, c := range letter {
		//		thisChar = Sprintf("%c", c)
		//		Printf("%s: %d\n", thisChar, c)
		ret = (c >= LOWER_A_ORD && c <= LOWER_Z_ORD)
	}
	return ret
}

// Copy makes a recursive deep copy of obj and returns the result.
// Wrap go_deepcopy.Copy (can later swap out implementation w/o breaking clients).
func DeepCopy(obj interface{}) (r interface{}) {
	return go_deepcopy.Copy(obj)
}

func NewUuid() (uuid string, err error) {
	uuidPtr, err := gouuid.NewV4()
	if err != nil {
		err = errors.New("Could not generate UUID")
	} else {
		uuid = uuidPtr.String()
	}
	return
}

// ToCurrency converts the value to a dollar and cents string
func ToCurrencyString(v interface{}) string {
	return Sprintf("%.2f", v)
}

// ToTS converts the value to a timestamp string (accepts both time.Time and *time.Time as argument)
func ToTS(v interface{}) string {
	var t time.Time
	ret := ""
	if TypeOf(v) == "time.Time" {
		t = v.(time.Time)
		ret = t.Format("2006-01-02 15:04 MST")
	} else if TypeOf(v) == "*time.Time" {
		ret = Sprintf("%s", time.Time(t).Format("2006-01-02 15:04 MST"))
	}
	return ret
}

// Prevent special CSV characters ("," and ";") from splitting a column
func CsvScrub(a interface{}) string {
	s := Sprint(a)
	commaPos := strings.Index(s, ",")
	semicolonPos := strings.Index(s, ";")
	if commaPos > -1 || semicolonPos > -1 {
		// comma or semicolon found
		s = Sprintf("\"%s\"", s) // surround with quotes per IETF RFC 2.6 guideline
	}
	return s
}

func Rand32() int32 {
	var n int32
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return n
}

// QueryString takes an http request object and returns its query string
func QueryString(req *http.Request) string {
	queryParamMap := req.URL.Query()
	i := 0
	queryParamString := ""
	for key, value := range queryParamMap {
		if len(value) > 1 {
			for _, arrayValue := range value {
				if i == 0 {
					queryParamString += "?"
				} else {
					queryParamString += "&"
				}
				i += 1
				queryParamString += key + "=" + arrayValue
			}
		} else {
			if i == 0 {
				queryParamString += "?"
			} else {
				queryParamString += "&"
			}
			i += 1
			queryParamString += key + "=" + queryParamMap.Get(key)
		}
	}
	return queryParamString
}
