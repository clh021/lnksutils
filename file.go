package lnksutils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// OpenURL open the url for reading
// It will reaturn error if open failed or the
// StatusCode is bigger than 299
// NOTE: the return reader need be closed
func OpenURL(url string, args ...string) (io.ReadCloser, error) {
	if len(args) > 0 {
		panic("Doesn't support multiple arguments")
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, fmt.Errorf("OpenURL %q failed %q", url, resp.Status)
	}
	return resp.Body, nil
}

//SaveToFile save `r` to `dst`, it will automatically create base directory.
//You can save string or bytes by
// bytes.NewBuffer([]byte) or bytes.NewBufferString(string)
func SaveToFile(r io.Reader, dst string) error {
	err := EnsureBaseDir(dst)
	if err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// FetchFile from url.
// Support http protocol and local file
func FetchFile(url string, handle func(r io.Reader) error) error {
	switch {
	case strings.HasPrefix(url, "http"):
		rc, err := OpenURL(url)
		if err != nil {
			return err
		}
		defer rc.Close()
		return handle(rc)
	default:
		f, err := os.Open(url)
		if err != nil {
			return err
		}
		defer f.Close()
		return handle(f)
	}
}

func FetchFileTmp(url string, handle func(tmpPath string) error) error {
	return DoWithTmpFile(func(tmpFile string) error {
		err := FetchFileTo(url, tmpFile)
		if err != nil {
			return err
		}
		return handle(tmpFile)
	})
}

func FetchFileTo(url string, to string) error {
	if url == to {
		return nil
	}
	return FetchFile(url, func(r io.Reader) error { return SaveToFile(r, to) })
}

// EnsureBaseDir make sure the parent directory of fpath exists
func EnsureBaseDir(fpath string) error {
	baseDir := filepath.Dir(fpath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}

func IsExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
func IsFileExist(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
func IsDirExist(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func DoWithTmpDir(fn func(string) error) error {
	tmpDir, err := ioutil.TempDir("", "lnks-dir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	return fn(tmpDir)
}
func DoWithTmpFile(fn func(string) error) error {
	f, err := ioutil.TempFile("", "lnks-file")
	if err != nil {
		return err
	}
	f.Close()
	defer os.Remove(f.Name())
	return fn(f.Name())
}
