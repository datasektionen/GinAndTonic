package local_service

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	RootDir string
}

func NewLocalStorage(rootDir string) *LocalStorage {
	return &LocalStorage{RootDir: rootDir}
}

func (ls *LocalStorage) UploadFile(key, filePath string) error {
	destPath := filepath.Join(ls.RootDir, key)
	return copyFile(filePath, destPath)
}

func (ls *LocalStorage) DownloadFile(key, filePath string) error {
	srcPath := filepath.Join(ls.RootDir, key)
	return copyFile(srcPath, filePath)
}

func (ls *LocalStorage) GetFileURL(key string) (string, error) {
	return filepath.Join(ls.RootDir, key), nil
}

func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
