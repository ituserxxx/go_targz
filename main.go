package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)
type DiyFilter struct {
	Dir        []string `json:"dir"`
	Files      []string `json:"files"`
	FilePrex   []string `json:"file_prex"`
	FileSuffix []string `json:"file_suffix"`
	DirPrex    []string `json:"dir_prex"`
	DirSuffix  []string `json:"dir_suffix"`
}

var filterConfig *DiyFilter

var txt = `
使用须知
1.将该程序放到需要打包的目录下
2.在目录创建一个 filter_config.json 文件，按需填入过滤的相关内容，
	{
	  "dir": [],
	  "files": [],
	  "file_prex": [],
	  "file_suffix":[],
	  "dir_prex": [],
	  "dir_suffix":[]
	}
	字段解释：
	dir 需要过滤的目录,如: ["my_img","static"]
	files 需要过滤的文件完整名称，如：["main.go", "xxx.text"]
	file_pres 需要过滤的文件前缀，如：["img_", "aaa.fff"]
	file_suffix 需要过滤的文件后缀，如：[".go", ".txt"]
	dir_prex 需要过滤的目录前缀，如：["imgs_"]
	dir_suffix 需要过滤的目录后缀，如：["_docs"]
3.然后双击执行该程序，会在当前目录生成一个以当前目录为前缀-时间戳的新 .tar.gz 文件


`
func main() {
	parasm := os.Args[1:]

	if len(parasm) == 1 && parasm[0] == "help"{
		fmt.Println(txt)
		return
	}
	filename := "filter_config.json"
	_, err := os.Stat(filename)
	isExist := false
	if err == nil {
		fmt.Printf("文件 %s 存在\n", filename)
		isExist =true
	}
	if isExist {
		configFile, err := os.Open(filename)
		if err != nil {
			fmt.Println("无法打开 filter_config.json 配置文件:", err)
			return
		}
		defer configFile.Close()
		decoder := json.NewDecoder(configFile)
		if err := decoder.Decode(&filterConfig); err != nil {
			fmt.Println("配置文件 filter_config.json 解析错误:", err)
			return
		}
	}
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录:", err)
		return
	}
	fmt.Println("即将压缩目录为" + dirPath)
	outputPath := fmt.Sprintf("%s-%d.tar.gz", filepath.Base(dirPath), time.Now().Nanosecond())
	err = tarGzDirectory(dirPath, outputPath)
	if err != nil {
		fmt.Printf("Error compressing directory: %v\n", err)
		return
	}
	fmt.Println("\n压缩成功，新文件名称------" + outputPath + "-----successfully!")
}
func isBeFilter(isDir bool, objName string) bool {
	if filterConfig == nil {
		return false
	}
	if isDir {
		for _, s := range filterConfig.Dir {
			if s == objName {
				return true
			}
		}
		for _, prex := range filterConfig.DirPrex {
			if strings.HasPrefix(objName, prex) {
				return true
			}
		}
		for _, suff := range filterConfig.DirSuffix {
			if strings.HasSuffix(objName, suff) {
				return true
			}
		}
		return false
	}
	for _, f1 := range filterConfig.Files {
		if f1 == objName {
			return true
		}
	}
	for _, f2 := range filterConfig.FilePrex {
		if strings.HasPrefix(objName, f2) {
			return true
		}
	}
	for _, suff2 := range filterConfig.FileSuffix {
		if strings.HasSuffix(objName, suff2) {
			return true
		}
	}
	return false
}
func tarGzDirectory(source, tarFile string) error {
	tarFilePtr, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer tarFilePtr.Close()

	gzipWriter := gzip.NewWriter(tarFilePtr)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == tarFile {
			return nil
		}
		if isBeFilter(info.IsDir(), info.Name()) {
			fmt.Printf("\n已过滤--->> path=%s  info.name=%s", path, info.Name())
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// 将Windows风格路径转换为Linux风格路径
		header.Name = filepath.ToSlash(path[len(source):])

		err = tarWriter.WriteHeader(header)
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

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
