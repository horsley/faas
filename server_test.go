package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/horsley/faas/tool"
	"github.com/horsley/svrkit"
)

func TestLegacyUpload(t *testing.T) {
	peekRootKey, _ := MetaOf("/").WriteKey()

	var buf bytes.Buffer
	m := multipart.NewWriter(&buf)
	f, err := m.CreateFormFile("file", "test_upload")
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	m.Close()

	mockReq, _ := http.NewRequest("POST", "http://abc.com/upload?k="+peekRootKey, &buf)
	mockReq.Header.Set("Content-Type", m.FormDataContentType())

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":0,"Data":null,"Message":""}` {
		t.Error("unexpected result:", resp)
	}
}

func TestLegacyUpload_BadPassword(t *testing.T) {
	var buf bytes.Buffer
	m := multipart.NewWriter(&buf)
	f, err := m.CreateFormFile("file", "test_upload")
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	m.Close()

	mockReq, _ := http.NewRequest("POST", "http://abc.com/upload?k=fake-password", &buf)
	mockReq.Header.Set("Content-Type", m.FormDataContentType())

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":401,"Data":null,"Message":"认证失败"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestLegacyUpload_BadForm(t *testing.T) {
	var buf bytes.Buffer
	m := multipart.NewWriter(&buf)
	f, err := m.CreateFormFile("fileee", "test_upload")
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	m.Close()

	mockReq, _ := http.NewRequest("POST", "http://abc.com/upload", &buf)
	mockReq.Header.Set("Content-Type", m.FormDataContentType())

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":500,"Data":null,"Message":"请选择文件上传"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestLegacyUpload_SubItemOfExistedFile(t *testing.T) {
	peekRootKey, _ := MetaOf("/").WriteKey()

	var buf bytes.Buffer
	m := multipart.NewWriter(&buf)
	f, err := m.CreateFormFile("file", "test_upload")
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	n, _ := m.CreateFormField("filename")
	n.Write([]byte("/test_upload/is/already/a/file"))
	m.Close()

	mockReq, _ := http.NewRequest("POST", "http://abc.com/upload?k="+peekRootKey, &buf)
	mockReq.Header.Set("Content-Type", m.FormDataContentType())

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":500,"Data":null,"Message":"保存失败"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestLegacyUpload_invalidPath(t *testing.T) {
	peekRootKey, _ := MetaOf("/").WriteKey()

	var buf bytes.Buffer
	m := multipart.NewWriter(&buf)
	f, err := m.CreateFormFile("file", "test_upload")
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	n, _ := m.CreateFormField("filename")
	n.Write([]byte("../../abc"))
	m.Close()

	mockReq, _ := http.NewRequest("POST", "http://abc.com/upload?k="+peekRootKey, &buf)
	mockReq.Header.Set("Content-Type", m.FormDataContentType())

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":403,"Data":null,"Message":"非法目标"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestModernUpload(t *testing.T) {
	peekRootKey, _ := MetaOf("/").WriteKey()

	var buf bytes.Buffer
	buf.WriteString("hello 2")
	mockReq, _ := http.NewRequest("PUT", "http://abc.com/test_upload2", &buf)
	tool.SignUpload(peekRootKey, mockReq)

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":0,"Data":null,"Message":""}` {
		t.Error("unexpected result:", resp)
	}
}

func TestRead(t *testing.T) {
	mockReq, _ := http.NewRequest("GET", "http://abc.com/test_upload2", nil)

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `hello 2` {
		t.Error("unexpected result:", resp)
	}
}

func TestRead304(t *testing.T) {
	mockReq, _ := http.NewRequest("GET", "http://abc.com/test_upload2", nil)
	mockReq.Header.Add("If-Modified-Since", time.Now().Add(time.Second).Format(http.TimeFormat))

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	if rec.Result().StatusCode != 304 {
		t.Error("not using 304", rec.Result().StatusCode)
	}
	resp := rec.Body.String()
	if resp != `` {
		t.Error("unexpected result:", resp)
	}
}

func TestReadBadPath(t *testing.T) {
	mockReq, _ := http.NewRequest("GET", "http://abc.com/../../test_upload3", nil)

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":403,"Data":null,"Message":"非法目标"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestRead404(t *testing.T) {
	mockReq, _ := http.NewRequest("GET", "http://abc.com/test_upload3", nil)

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != "404 page not found\n" {
		t.Error("unexpected result:", resp)
	}
}

func TestReadWithAuth(t *testing.T) {
	mockReq, _ := http.NewRequest("GET", "http://abc.com/test_upload", nil)

	MetaOf("test_upload").SetBasicAuth(map[string]string{"user": "pass"})

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	if rec.Result().StatusCode != 401 {
		t.Error("not trigger auth")
	}

	rec2 := httptest.NewRecorder()
	mockReq.SetBasicAuth("user", "bad pass")
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec2}, &svrkit.Request{Request: mockReq})
	if rec2.Result().StatusCode != 401 {
		t.Error("not reject bad auth")
	}

	rec3 := httptest.NewRecorder()
	mockReq.SetBasicAuth("user", "pass")
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec3}, &svrkit.Request{Request: mockReq})
	if rec3.Result().StatusCode != 200 {
		t.Error("not accept good auth")
	}
}

func TestDeleteFile(t *testing.T) {
	mockReq, _ := http.NewRequest("DELETE", "http://abc.com/test_upload", nil)
	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":401,"Data":null,"Message":"认证失败"}` {
		t.Error("unexpected result:", resp)
	}

	peekRootKey, _ := MetaOf("/").WriteKey()
	tool.SignUpload(peekRootKey, mockReq)
	rec2 := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec2}, &svrkit.Request{Request: mockReq})

	resp2 := rec2.Body.String()
	if resp2 != `{"Code":0,"Data":null,"Message":""}` {
		t.Error("unexpected result2:", resp2)
	}

	rec3 := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec3}, &svrkit.Request{Request: mockReq})

	resp3 := rec3.Body.String()
	if resp3 != `{"Code":0,"Data":null,"Message":""}` { //dup delete
		t.Error("unexpected result3:", resp3)
	}
}

func TestDeleteBadPath(t *testing.T) {
	mockReq, _ := http.NewRequest("DELETE", "http://abc.com/../../test_upload3", nil)

	rec := httptest.NewRecorder()
	handleRequest(&svrkit.ResponseWriter{ResponseWriter: rec}, &svrkit.Request{Request: mockReq})

	resp := rec.Body.String()
	if resp != `{"Code":403,"Data":null,"Message":"非法目标"}` {
		t.Error("unexpected result:", resp)
	}
}

func TestClean(t *testing.T) {
	MetaOf("/test_upload2").Destroy()
}
