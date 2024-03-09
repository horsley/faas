package tool

import (
	"io"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	TestHTTPHost = "https://abc.com"
	TestDirKey   = "123456"
)

func TestUpload(t *testing.T) {
	type args struct {
		url     string
		key     string
		content io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test upload content",
			args: args{
				url:     TestHTTPHost + "/config/login.yaml",
				key:     TestDirKey,
				content: strings.NewReader(uuid.NewString()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Upload(tt.args.url, tt.args.key, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPoll(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	newContent := uuid.NewString()

	go Poll(TestHTTPHost+"/config/login.yaml", 1*time.Second, func(old, new []byte) {
		t.Log("changed from:", string(old), "to:", string(new))
		if string(new) != newContent {
			t.Error("content not match")
		}
		wg.Done()
	})

	time.Sleep(3 * time.Second)
	log.Println("upload new content", newContent)
	Upload(TestHTTPHost+"/config/login.yaml", TestDirKey, strings.NewReader(newContent))
	wg.Wait()
}
