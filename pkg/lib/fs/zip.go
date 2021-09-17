package fs

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {

			if fileReader != nil {
				fileReader.Close()
			}

			return err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			fileReader.Close()

			if targetFile != nil {
				targetFile.Close()
			}

			return err
		}

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			fileReader.Close()
			targetFile.Close()

			return err
		}

		fileReader.Close()
		targetFile.Close()
	}

	return nil
}

func Zip(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func BackupDirFiles(pth, pathBakf string) error {
	files, err := ioutil.ReadDir(pth)
	if err != nil {
		return err
	}
	bakDest := backupPath(pathBakf)
	f, err := os.Create(bakDest)
    if err != nil {
        return err
    }
    defer f.Close()
	z := zip.NewWriter(f)
	for _, fi := range files {
		if !fi.IsDir() {
			src, err := os.Open(filepath.Join(pth, fi.Name()))
			if err != nil {
				return err
			}
			hdr, err := zip.FileInfoHeader(fi)
			if err != nil {
				return err
			}
			hdr.Name = fi.Name()
			dst, err := z.CreateHeader(hdr)
			if err != nil {
				return err
			}
			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}
			src.Close()
		}
	}
	return z.Close()
}

/*
zipit("/tmp/documents", "/tmp/backup.zip")
zipit("/tmp/report.txt", "/tmp/report-2015.zip")
unzip("/tmp/report-2015.zip", "/tmp/reports/")



func Unzip(zipFile string, destDir string) error {
    zipReader, err := zip.OpenReader(zipFile)
    if err != nil {
        return err
    }
    defer zipReader.Close()

    for _, f := range zipReader.File {
        fpath := filepath.Join(destDir, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(fpath, os.ModePerm)
        } else {
            if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
                return err
            }

            inFile, err := f.Open()
            if err != nil {
                return err
            }
            defer inFile.Close()

            outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer outFile.Close()

            _, err = io.Copy(outFile, inFile)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

*/