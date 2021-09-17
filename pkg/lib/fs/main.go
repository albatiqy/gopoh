package fs

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/albatiqy/gopoh/contract/log"
)

const (
	libName = "util/fs"
)

var (
	appWd string
)

func init() {
	log.Debugf("%s: %s", libName, "initialized")
}

func libError(msg string) error {
	log.Errorf("%s: %s", libName, msg)
	return fmt.Errorf("%s: %s", libName, msg)
}

/*
func GopohDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		libLog("GopohDir() error")
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "../..")
}
*/

func WorkingDir(pth ...string) string {
	if appWd == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Debugf("%s: %s", libName, err)
			return ""
		}
		appWd = wd
	}
	pth = append(pth, "")
	copy(pth[1:], pth)
	pth[0] = appWd
	return filepath.Join(pth...)
}

func MkDirIfNotExists(pth string) (bool, error) { //success, error
	fInfo, err := os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			if errDir := os.MkdirAll(pth, 0755); errDir != nil { // os.ModePerm
				return false, libError(errDir.Error())
			}
			return true, nil
		}
		return false, libError(err.Error())
	}
	if fInfo.IsDir() {
		return true, libError(fmt.Sprintf("direktori \"%s\" sudah ada", pth))
	} else {
		return false, libError(fmt.Sprintf("path ditemukan \"%s\" bukan sebuah direktori", pth))
	}
}

func FileInfo(pth string) fs.FileInfo {
	fileInfo, err := os.Stat(pth)
	if err == nil {
		return fileInfo
	}
	return nil
}

/*
func IsDir(pth string) bool {
	fileInfo := FileInfo(pth)
	if fileInfo != nil {
		return fileInfo.IsDir()
	}
	return false
}
*/

func FileCopy(srcPth string, dstPth string) error {
	src, err := os.Open(srcPth)
	if err != nil {
		return libError(err.Error())
	}
	defer src.Close()

	dst, err := os.Create(dstPth)
	if err != nil {
		return libError(err.Error())
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return libError(err.Error())
	}
	return nil
}

func BackupIfExist(pth, pathBakf string) (bool, error) { //success, error
	fInfo := FileInfo(pth)
	if fInfo != nil {
		if fInfo.IsDir() {
			return false, libError(fmt.Sprintf("hanya dapat membackup file, path \"%s\": adalah direktori", pth))
		}
		if err := FileCopy(pth, backupPath(pathBakf)); err != nil {
			return false, libError(fmt.Sprintf("tidak dapat membackup file \"%s\": %s", pth, err))
		}
		return true, nil
	}
	return true, libError("file tidak ditemukan")
}

func BackupDir(pth, pathBakf string) error {
	return Zip(pth, backupPath(pathBakf))
}

func ReadAppTextFile(pth ...string) string {
	content, err := ioutil.ReadFile(WorkingDir(pth...))
	if err != nil {
		log.Debugf("%s: %s", libName, err)
		return ""
	}
	return string(content)
}

func WriteAppTextFile(content string, pth ...string) error {
	return ioutil.WriteFile(WorkingDir(pth...), []byte(content), 0755)
}

func WriteTextFile(content, pth string) error {
	fOut, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer fOut.Close()
	if _, err := fOut.WriteString(content); err != nil {
		return err
	}
	return nil
}

func backupPath(pathBakf string) string {
	fnbakTimeStr := time.Now().Format("20060102030405")
	ext := filepath.Ext(pathBakf)
	basePath := pathBakf
	if ext != "" {
		basePath = strings.TrimSuffix(pathBakf, ext)
	}
	return basePath + "-" + fnbakTimeStr + ext
}

/*

func RemoveContents(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
    return nil
}


func testPath() {
	fmt.Println(getExecutablePath())
	fmt.Println(getWorkingDir())
	fmt.Println(filepath.Join("dir1/../dir1", "filename"))

	p := filepath.Join("dir1", "dir2", "filename.exe")
	fmt.Println("p:", p)
	fmt.Println("Dir(p):", filepath.Dir(p))
	fmt.Println("Base(p):", filepath.Base(p))

	fmt.Println(filepath.IsAbs("dir/file"))
	fmt.Println(filepath.IsAbs("/dir/file"))

	filename := "config.json"
	ext := filepath.Ext(filename)
	fmt.Println(ext)
	fmt.Println(strings.TrimSuffix(filename, ext))

	rel, err := filepath.Rel("a/b", "a/b/t/file")
    if err != nil {
        panic(err)
    }
	fmt.Println(rel)

	rel, err = filepath.Rel("a/b", "a/c/t/file")
    if err != nil {
        panic(err)
    }
	fmt.Println(rel)
	fmt.Println(isDir(getExecutablePath()))

	baseDir := filepath.Join(getWorkingDir(), "/sample")
	_ = createDirectory(baseDir)
	filePutContents(baseDir + "/oke.txt", "sfsdfsdffd")

	f, err := os.Create(baseDir + "/mbelgedes.txt")
    if err != nil {
        panic(err)
	}
	defer f.Close()

	d2 := []byte{115, 111, 109, 101, 10}
    n2, err := f.Write(d2)
    if err != nil {
        panic(err)
	}
	fmt.Printf("wrote %d bytes\n", n2)

	n3, err := f.WriteString("writes\n")
    if err != nil {
        panic(err)
	}
	fmt.Printf("wrote %d bytes\n", n3)

	f.Sync()

	w := bufio.NewWriter(f)
    n4, err := w.WriteString("buffered\n")
    if err != nil {
        panic(err)
	}
	fmt.Printf("wrote %d bytes\n", n4)
	w.Flush()
}
*/
