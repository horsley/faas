package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/horsley/svrkit"
)

type MetaKey string

const (
	MetaWriteKey    = MetaKey("key")
	MetaIPCheck     = MetaKey("ip_check")
	MetaReadAuth    = MetaKey("basic_auth")
	MetaContentType = MetaKey("content-type")
)

type pathMeta struct {
	root        string
	metaAbsPath string
	srcPath     string
}

const metaSubDir = "meta"
const contentSubDir = "content"

func MetaOf(path string) *pathMeta {
	metaRoot := filepath.Join(STORAGE, metaSubDir)
	absPath := filepath.Join(metaRoot, path)

	if !strings.HasPrefix(absPath, metaRoot) { // directory path traversal attack
		return nil
	}

	return &pathMeta{metaRoot, absPath, path}
}

func (p *pathMeta) Valid() bool {
	return p != nil
}

func (p *pathMeta) ContentPath() string {
	if !p.Valid() {
		return ""
	}
	return filepath.Join(STORAGE, contentSubDir, p.srcPath)
}

func (p *pathMeta) SaveContent(rd io.Reader) error {
	targetFilePath := p.ContentPath()
	if targetFilePath == "" {
		return errors.New("非法路径")
	}

	err := os.MkdirAll(filepath.Dir(targetFilePath), 0755)
	if err != nil {
		return err
	}

	targetFile, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, rd)
	return err
}

func (p *pathMeta) WriteKey() (string, bool) {
	return p.GetText(MetaWriteKey, true)
}

func (p *pathMeta) SetWriteKey(newKey string) error {
	return p.Set(MetaWriteKey, []byte(newKey))
}

func (p *pathMeta) GetBasicAuth() map[string]string {
	auth, ok := p.Get(MetaReadAuth, true)
	if ok {
		var result map[string]string
		err := json.Unmarshal(auth, &result)
		if err == nil {
			return result
		}
	}
	return nil
}

func (p *pathMeta) SetBasicAuth(in map[string]string) error {
	bin, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return err
	}
	return p.Set(MetaReadAuth, bin)
}

func (p *pathMeta) GetIPChecker() func(ip string) bool {
	auth, ok := p.Get(MetaIPCheck, true)
	if ok {
		var result []string
		err := json.Unmarshal(auth, &result)
		if err == nil {
			checker := svrkit.InSliceChecker(result)
			return func(ip string) bool {
				return checker(ip)
			}
		}
	}
	return nil
}

func (p *pathMeta) GetText(k MetaKey, inherit bool) (string, bool) {
	ret, ok := p.Get(k, inherit)
	if ok {
		return string(ret), ok
	}
	return "", false
}

func (p *pathMeta) Get(k MetaKey, inherit bool) ([]byte, bool) {
	if !p.Valid() {
		return nil, false
	}
	dir := p.metaAbsPath
	for strings.HasPrefix(dir, p.root) {
		keyPath := filepath.Join(dir, string(k))
		data, err := os.ReadFile(keyPath)
		if err == nil {
			return data, true
		}
		if !inherit {
			break
		}
		dir = filepath.Dir(dir)
	}

	return nil, false
}

func (p *pathMeta) Set(k MetaKey, content []byte) error {
	if !p.Valid() {
		return errors.New("invalid meta")
	}
	err := os.MkdirAll(p.metaAbsPath, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.metaAbsPath, string(k)), content, 0644)
}

func (p *pathMeta) Destroy() error {
	err := os.RemoveAll(p.metaAbsPath)
	if err != nil {
		return err
	}

	err = os.Remove(p.ContentPath())
	if os.IsNotExist(err) {
		return nil
	}

	return err
}
