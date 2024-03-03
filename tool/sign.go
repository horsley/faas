package tool

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/horsley/svrkit"
)

// TimeSpan 签名验证容忍的时间窗口
var TimeSpan = float64(10)

// SignUpload 对上传请求签名
func SignUpload(key string, req *http.Request) {
	ts := fmt.Sprint(time.Now().Unix())

	sign := svrkit.SHA1Hash(fmt.Sprint(ts, req.URL.Path, key, ts))

	req.SetBasicAuth(ts, sign)
}

// VerifySign 验证请求签名
func VerifySign(key string, req *http.Request) bool {
	ts, sign, ok := req.BasicAuth()
	if !ok {
		return false
	}

	target := svrkit.SHA1Hash(fmt.Sprint(ts, req.URL.Path, key, ts))
	if sign != target {
		return false
	}

	tsI64, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}

	diff := math.Abs(time.Since(time.Unix(tsI64, 0)).Seconds())
	return diff <= TimeSpan
}
