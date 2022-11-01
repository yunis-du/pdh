package tools

import (
	"archive/zip"
	"compress/flate"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/kalafut/imohash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsEmptyFolder(folderPath string) (bool, error) {
	f, err := os.Open(folderPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, nil
}

func ZipDirectory(destination string, source string) error {
	if _, err := os.Stat(destination); err == nil {
		log.Fatalf("%s file already exists!\n", destination)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Zipping %s to %s\n", source, destination)
	file, err := os.Create(destination)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	// no compression because croc does its compression on the fly
	writer.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.NoCompression)
	})
	defer writer.Close()
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalln(err)
		}
		if info.Mode().IsRegular() {
			f1, err := os.Open(path)
			if err != nil {
				log.Fatalln(err)
			}
			defer f1.Close()
			zipPath := strings.ReplaceAll(path, source, strings.TrimSuffix(destination, ".zip"))
			w1, err := writer.Create(zipPath)
			if err != nil {
				log.Fatalln(err)
			}
			if _, err := io.Copy(w1, f1); err != nil {
				log.Fatalln(err)
			}
			_, _ = fmt.Fprintf(os.Stderr, "\r\033[2K")
			_, _ = fmt.Fprintf(os.Stderr, "\rAdding %s", zipPath)
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func UnzipDirectory(destination string, source string) error {

	archive, err := zip.OpenReader(source)
	if err != nil {
		log.Fatalln(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(destination, f.Name)
		_, _ = fmt.Fprintf(os.Stderr, "\r\033[2K")
		_, _ = fmt.Fprintf(os.Stderr, "\rUnzipping file %s", filePath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Fatalln(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Fatalln(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			log.Fatalln(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			log.Fatalln(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	_, _ = fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func GetAbsolutePaths(paths []string) []string {
	absolutePaths := make([]string, 0)
	wd, _ := os.Getwd()
	lst, _ := os.Lstat(wd)
	for _, path := range paths {
		if path == lst.Name() {
			absolutePaths = append(absolutePaths, wd)
		} else {
			if IsDir(path) {
				absolutePaths = append(absolutePaths, path)
			} else if IsFile(path) {
				absolutePaths = append(absolutePaths, path)
			} else {
				p := wd + "/" + path
				if IsDir(p) {
					absolutePaths = append(absolutePaths, p)
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "cant't find this path [%s] please check it!", path)
					os.Exit(1)
				}
			}
		}
	}
	return absolutePaths
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func IsFile(path string) bool {
	_, err := os.Open(path)
	if err != nil {
		return false
	}
	return true
}

func HashFile(fileName string, algorithm string) (hash256 []byte, err error) {
	var fileStats os.FileInfo
	fileStats, err = os.Lstat(fileName)
	if err != nil {
		return nil, err
	}
	if fileStats.Mode()&os.ModeSymlink != 0 {
		var target string
		target, err = os.Readlink(fileName)
		if err != nil {
			return nil, err
		}
		return []byte(SHA256(target)), nil
	}
	switch algorithm {
	case "imohash":
		return IMOHashFile(fileName)
	case "md5":
		return MD5HashFile(fileName)
	case "xxhash":
		return XXHashFile(fileName)
	}
	err = fmt.Errorf("unspecified algorithm")
	return
}

// MD5HashFile returns MD5 hash
func MD5HashFile(fileName string) (hash256 []byte, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	hash256 = h.Sum(nil)
	return
}

// IMOHashFile returns imohash
func IMOHashFile(fileName string) (hash []byte, err error) {
	b, err := imohash.SumFile(fileName)
	hash = b[:]
	return
}

// IMOHashFileFull returns imohash of full file
func IMOHashFileFull(fileName string) (hash []byte, err error) {
	var imoFull = imohash.NewCustom(0, 0)
	b, err := imoFull.SumFile(fileName)
	hash = b[:]
	return
}

// XXHashFile returns the xxhash of a file
func XXHashFile(fileName string) (hash256 []byte, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	h := xxhash.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	hash256 = h.Sum(nil)
	return
}

// SHA256 returns sha256 sum
func SHA256(s string) string {
	sha := sha256.New()
	sha.Write([]byte(s))
	return hex.EncodeToString(sha.Sum(nil))
}
