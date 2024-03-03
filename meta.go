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
	return p.GetText("key", true)
}

func (p *pathMeta) SetWriteKey(newKey string) error {
	return p.Set("key", []byte(newKey))
}

func (p *pathMeta) GetBasicAuth() map[string]string {
	auth, ok := p.Get("basic_auth", true)
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
	return p.Set("basic_auth", bin)
}

func (p *pathMeta) GetIPChecker() func(ip string) bool {
	auth, ok := p.Get("ip_check", true)
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

func (p *pathMeta) GetText(k string, inherit bool) (string, bool) {
	ret, ok := p.Get(k, inherit)
	if ok {
		return string(ret), ok
	}
	return "", false
}

func (p *pathMeta) Get(k string, inherit bool) ([]byte, bool) {
	if !p.Valid() {
		return nil, false
	}
	dir := p.metaAbsPath
	for strings.HasPrefix(dir, p.root) {
		keyPath := filepath.Join(dir, k)
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

func (p *pathMeta) Set(k string, content []byte) error {
	if !p.Valid() {
		return errors.New("invalid meta")
	}
	err := os.MkdirAll(p.metaAbsPath, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.metaAbsPath, k), content, 0644)
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
