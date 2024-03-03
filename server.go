package main

import (
	"log"
	"net/http"

	"github.com/horsley/faas/tool"
	"github.com/horsley/svrkit"
)

func newServer() http.Handler {
	mux := svrkit.NewRouter()

	mux.HandleFuncEx("/", handleRequest)
	return mux
}

func handleRequest(rw *svrkit.ResponseWriter, r *svrkit.Request) {
	if r.Method == "POST" || r.Method == "PUT" {
		uploadHandler(rw, r)
		return
	}

	if r.Method == "DELETE" {
		deleteHandler(rw, r)
		return
	}

	readFileHandler(rw, r)
}

func uploadHandler(rw *svrkit.ResponseWriter, r *svrkit.Request) {
	targetPath := r.URL.Path
	contentReader := r.Body
	legacyAuthCheck := false

	if r.Method == "POST" && r.URL.Path == "/upload" { //legacy upload support
		legacyAuthCheck = true
		f, info, err := r.FormFile("file")
		if err != nil {
			rw.WriteCommonResponse(500, "请选择文件上传", nil)
			return
		}
		defer f.Close()

		targetPath = info.Filename
		contentReader = f
		if r.FormValue("filename") != "" {
			targetPath = r.FormValue("filename")
		}
	}

	targetMeta := MetaOf(targetPath)
	writeKey, ok := targetMeta.WriteKey()
	if !ok {
		rw.WriteCommonResponse(403, "非法目标", nil)
		return
	}

	var authPass bool
	if legacyAuthCheck {
		authPass = r.URL.Query().Get("k") == writeKey
	} else {
		authPass = tool.VerifySign(writeKey, r.Request)
	}
	if !authPass {
		rw.WriteCommonResponse(401, "认证失败", nil)
		return
	}

	err := targetMeta.SaveContent(contentReader)
	if err != nil {
		log.Println("SaveContent err:", err, targetPath)
		rw.WriteCommonResponse(500, "保存失败", nil)
		return
	}

	rw.WriteCommonResponse(0, "", nil)
}

func deleteHandler(rw *svrkit.ResponseWriter, r *svrkit.Request) {
	targetPath := r.URL.Path
	targetMeta := MetaOf(targetPath)
	writeKey, ok := targetMeta.WriteKey()
	if !ok {
		rw.WriteCommonResponse(403, "非法目标", nil)
		return
	}

	if !tool.VerifySign(writeKey, r.Request) {
		rw.WriteCommonResponse(401, "认证失败", nil)
		return
	}

	err := targetMeta.Destroy()
	if err != nil {
		log.Println("Destroy err:", err, targetPath)
		rw.WriteCommonResponse(500, "删除失败", nil)
		return
	}

	rw.WriteCommonResponse(0, "", nil)

}

func readFileHandler(rw *svrkit.ResponseWriter, r *svrkit.Request) {
	targetPath := r.URL.Path
	targetMeta := MetaOf(targetPath)
	if targetMeta == nil {
		rw.WriteCommonResponse(403, "非法目标", nil)
		return
	}

	if validUserPass := targetMeta.GetBasicAuth(); validUserPass != nil {
		user, pass, ok := r.BasicAuth()
		if !ok {
			rw.Header().Add("WWW-Authenticate", `Basic realm="Give me username and password"`)
			rw.HTTPError(http.StatusUnauthorized, "need auth")
			return
		}

		if validUserPass[user] != pass {
			rw.HTTPError(http.StatusUnauthorized, "auth fail")
			return

		}
	}

	if ipChecker := targetMeta.GetIPChecker(); ipChecker != nil && !ipChecker(r.ClientIP()) {
		rw.HTTPError(http.StatusForbidden, "bad ip:"+r.ClientIP())
		return
	}

	if contentType, ok := targetMeta.GetText(MetaContentType, false); ok {
		rw.Header().Set("Content-Type", contentType)
	}
	http.ServeFile(rw, r.Request, targetMeta.ContentPath())
}
