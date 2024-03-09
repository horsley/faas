package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

// Poll 定时拉URL
func Poll(url string, interval time.Duration, onChange func(old, new []byte)) {

	var lastModify string
	var lastBin []byte
	for range time.NewTicker(interval).C {
		func() {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println("create poll request err:", err)
				return
			}

			if lastModify != "" {
				req.Header.Add("If-Modified-Since", lastModify)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("Poll got error:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotModified {
				return
			}

			bin, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Poll read error:", err)
				return
			}

			if resp.StatusCode == http.StatusOK {
				if bytes.Equal(lastBin, bin) {
					return
				}

				if lastBin != nil {
					onChange(lastBin, bin)
				}
				lastBin = bin
				lastModify = resp.Header.Get("Last-Modified")
				return
			}

			log.Println("Poll http error:", resp.StatusCode, string(bin[:200]))
		}()
	}
}

// Upload 上传内容
func Upload(url, key string, content io.Reader) error {
	req, err := http.NewRequest("PUT", url, content)
	if err != nil {
		return err
	}
	SignUpload(key, req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bin, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respData struct {
		Code    int
		Message string
	}

	err = json.Unmarshal(bin, &respData)
	if err != nil {
		return err
	}

	if respData.Code != 0 {
		return errors.New(respData.Message)
	}

	return nil
}
