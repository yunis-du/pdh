package files

import (
	"github.com/duyunis/pdh/tools"
	"os"
	"path/filepath"
	"strings"
)

type Files struct {
	FilesInfo          []*FileInfo
	EmptyFolders       []*FileInfo
	TotalNumberFolders int
}

type FileInfo struct {
	Name         string `json:"Name,omitempty"`
	FolderRemote string `json:"FolderRemote,omitempty"`
	FolderSource string `json:"FolderSource,omitempty"`
	//Hash         []byte `json:"Hash,omitempty"`
	Size         int64  `json:"Size,omitempty"`
	ModTime      int64  `json:"ModTime,omitempty"`
	IsCompressed bool   `json:"IsCompressed,omitempty"`
	IsEncrypted  bool   `json:"IsEncrypted,omitempty"`
	Symlink      string `json:"Symlink,omitempty"`
	Mode         uint32 `json:"Mode,omitempty"`
	TempFile     bool   `json:"TempFile,omitempty"`
}

func GetFilesInfo(fNames []string, zipFolder bool) (*Files, error) {
	// fNames: the relative/absolute paths of files/folders that will be transfered
	filesInfo := make([]*FileInfo, 0)
	emptyFolders := make([]*FileInfo, 0)
	totalNumberFolders := 0
	var paths []string
	for _, fName := range fNames {
		// Support wildcard
		if strings.Contains(fName, "*") {
			matches, errGlob := filepath.Glob(fName)
			if errGlob != nil {
				return nil, errGlob
			}
			paths = append(paths, matches...)
			continue
		} else {
			paths = append(paths, fName)
		}
	}

	for _, path := range paths {
		stat, errStat := os.Lstat(path)

		if errStat != nil {
			return nil, errStat
		}

		absPath, errAbs := filepath.Abs(path)

		if errAbs != nil {
			return nil, errAbs
		}

		if stat.IsDir() && zipFolder {
			if path[len(path)-1:] != "/" {
				path += "/"
			}
			path := filepath.Dir(path)
			dest := filepath.Base(path) + ".zip"
			_ = tools.ZipDirectory(dest, path)
			stat, errStat = os.Lstat(dest)
			if errStat != nil {
				return nil, errStat
			}
			absPath, errAbs = filepath.Abs(dest)
			if errAbs != nil {
				return nil, errAbs
			}
			filesInfo = append(filesInfo, &FileInfo{
				Name:         stat.Name(),
				FolderRemote: "./",
				FolderSource: filepath.Dir(absPath),
				Size:         stat.Size(),
				ModTime:      stat.ModTime().UnixMilli(),
				Mode:         uint32(stat.Mode()),
				TempFile:     true,
			})
			continue
		}

		if stat.IsDir() {
			err := filepath.Walk(absPath,
				func(pathName string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					remoteFolder := strings.TrimPrefix(filepath.Dir(pathName),
						filepath.Dir(absPath)+string(os.PathSeparator))
					if !info.IsDir() {
						filesInfo = append(filesInfo, &FileInfo{
							Name:         info.Name(),
							FolderRemote: strings.Replace(remoteFolder, string(os.PathSeparator), "/", -1) + "/",
							FolderSource: filepath.Dir(pathName),
							Size:         info.Size(),
							ModTime:      info.ModTime().UnixMilli(),
							Mode:         uint32(info.Mode()),
							TempFile:     false,
						})
					} else {
						totalNumberFolders++
						isEmptyFolder, _ := tools.IsEmptyFolder(pathName)
						if isEmptyFolder {
							emptyFolders = append(emptyFolders, &FileInfo{
								// Name: info.Name(),
								FolderRemote: strings.Replace(strings.TrimPrefix(pathName,
									filepath.Dir(absPath)+string(os.PathSeparator)), string(os.PathSeparator), "/", -1) + "/",
							})
						}
					}
					return nil
				})
			if err != nil {
				return nil, err
			}

		} else {
			filesInfo = append(filesInfo, &FileInfo{
				Name:         stat.Name(),
				FolderRemote: "./",
				FolderSource: filepath.Dir(absPath),
				Size:         stat.Size(),
				ModTime:      stat.ModTime().UnixMilli(),
				Mode:         uint32(stat.Mode()),
				TempFile:     false,
			})
		}

	}
	return &Files{
		FilesInfo:          filesInfo,
		EmptyFolders:       emptyFolders,
		TotalNumberFolders: totalNumberFolders,
	}, nil
}
