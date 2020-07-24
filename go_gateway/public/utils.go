package public

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

func GenSaltPassword(salt, password string) string {
	s1 := sha256.New()
	s1.Write([]byte(password))
	str1 := fmt.Sprintf("%x", s1.Sum(nil))
	s2 := sha256.New()
	s2.Write([]byte(str1 + salt))
	return fmt.Sprintf("%x", s2.Sum(nil))
}

func MD5(s string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, s)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Obj2Json(s interface{}) string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func InStringSlice(slice []string, str string) bool {
	for _, item := range slice {
		if str == item {
			return true
		}
	}
	return false
}
