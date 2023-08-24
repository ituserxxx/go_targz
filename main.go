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

/*
	使用golang写的压缩工具，
*/
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
	err = compressTarGz(dirPath, outputPath)
	if err != nil {
		fmt.Println("Failed to create tar.gz:", err)
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
func compressTarGz(source, destination string) error {
	// 创建输出文件
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建gzip写入器
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// 创建tar写入器
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 获取源路径信息
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	// 如果是目录，则遍历目录并递归压缩
	if info.IsDir() {
		baseDir := filepath.Base(source)
		return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == destination {
				return nil
			}

			if isBeFilter(info.IsDir(), info.Name()) {
				fmt.Printf("\n已过滤--->> path=%s  info.name=%s", path, info.Name())
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// 获取相对路径
			relPath, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}

			// 创建tar头部信息
			header, err := tar.FileInfoHeader(info, relPath)
			if err != nil {
				return err
			}

			// 更新头部信息中的名称为相对路径
			header.Name = filepath.Join(baseDir, filepath.ToSlash(relPath))

			// 写入头部信息
			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			// 如果不是目录，则写入文件内容
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				// 将文件内容写入tar
				_, err = io.Copy(tw, file)
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else { // 如果是单个文件，则直接压缩
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		defer file.Close()

		// 创建tar头部信息
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// 写入头部信息
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		// 将文件内容写入tar
		_, err = io.Copy(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

type Config struct {
	SkipPrefixes []string `json:"skip_prefixes"`
}

func main2() {
	// 读取配置文件
	//configFile, err := os.Open("quickPacking.json")
	//
	//if err != nil {
	//	fmt.Println("无法打开配置文件:", err)
	//	return
	//}
	//defer configFile.Close()
	//
	//var config Config
	//decoder := json.NewDecoder(configFile)
	//if err := decoder.Decode(&config); err != nil {
	//	fmt.Println("配置文件解析错误:", err)
	//	return
	//}
	//--------------------
	// 获取当前目录名称作为压缩包名称
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录:", err)
		return
	}
	println(currentDir)
	//tarFileName := filepath.Base(currentDir) +string()+ ".tar.gz"
	tarFileName := fmt.Sprintf("%s-%d.tar.gz", filepath.Base(currentDir), time.Now().Nanosecond())

	// 创建目标 tar 文件
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		fmt.Println("无法创建 tar 文件:", err)
		return
	}
	defer tarFile.Close()
	println(tarFileName)
	//--------------------

	// 使用 gzip 创建压缩写入器
	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	// 创建 tar 写入器
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 遍历当前目录及其子目录，压缩文件和文件夹
	filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("遍历目录时出错:", err)
			return err
		}

		//// 跳过 .DS_Store 文件
		//if filepath.Base(path) == ".DS_Store" {
		//	return nil
		//}
		//
		//if info.IsDir() {
		//	// 检查是否需要跳过压缩
		//	shouldSkip := false
		//	for _, prefix := range config.SkipPrefixes {
		//		if strings.HasPrefix(filepath.Base(path), prefix) {
		//			shouldSkip = true
		//			break
		//		}
		//	}
		//	if shouldSkip {
		//		return nil // SkipDir改为返回nil
		//	}
		//}

		// 获取相对路径
		relPath, err := filepath.Rel(currentDir, path)
		if err != nil {
			fmt.Println("获取相对路径时出错:", err)
			return err
		}

		// 创建 tar 条目头部
		header := new(tar.Header)
		header.Name = relPath
		header.Mode = int64(info.Mode())
		header.Size = info.Size()
		header.ModTime = info.ModTime()

		fmt.Printf("Writing header for %s\n", relPath)

		// 写入 tar 条目头部
		if err := tarWriter.WriteHeader(header); err != nil {
			fmt.Println("写入 tar 条目头部时出错:", err)
			return err
		}

		// 如果不是目录，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				fmt.Println("无法打开文件:", err)
				return err
			}
			defer file.Close()

			fmt.Printf("Writing content for %s\n", relPath)

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				fmt.Println("复制文件内容时出错:", err)
				return err
			}
		}
		return nil
	})

}
func gentorTargzFile() {
	// 获取当前目录名称作为压缩包名称
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录:", err)
		return
	}
	println(currentDir)
	//tarFileName := filepath.Base(currentDir) +string()+ ".tar.gz"
	tarFileName := fmt.Sprintf("%s-%d.tar.gz", filepath.Base(currentDir), time.Now().Nanosecond())

	// 创建目标 tar 文件
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		fmt.Println("无法创建 tar 文件:", err)
		return
	}
	defer tarFile.Close()
	println(tarFileName)

}
func aaa() {
	// 读取配置文件
	configFile, err := os.Open("quickPacking.json")

	if err != nil {
		fmt.Println("无法打开配置文件:", err)
		return
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("配置文件解析错误:", err)
		return
	}

	// 获取当前目录名称作为压缩包名称
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录:", err)
		return
	}
	tarFileName := filepath.Base(currentDir) + ".tar.gz"

	// 创建目标 tar 文件
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		fmt.Println("无法创建 tar 文件:", err)
		return
	}
	defer tarFile.Close()

	// 使用 gzip 创建压缩写入器
	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	// 创建 tar 写入器
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 遍历当前目录及其子目录，压缩文件和文件夹
	filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("遍历目录时出错:", err)
			return err
		}

		// 跳过 .DS_Store 文件
		if filepath.Base(path) == ".DS_Store" {
			return nil
		}

		if info.IsDir() {
			// 检查是否需要跳过压缩
			shouldSkip := false
			for _, prefix := range config.SkipPrefixes {
				if strings.HasPrefix(filepath.Base(path), prefix) {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				return nil // SkipDir改为返回nil
			}
		}

		// 获取相对路径
		relPath, err := filepath.Rel(currentDir, path)
		if err != nil {
			fmt.Println("获取相对路径时出错:", err)
			return err
		}

		// 创建 tar 条目头部
		header := new(tar.Header)
		header.Name = relPath
		header.Mode = int64(info.Mode())
		header.Size = info.Size()
		header.ModTime = info.ModTime()

		fmt.Printf("Writing header for %s\n", relPath)

		// 写入 tar 条目头部
		if err := tarWriter.WriteHeader(header); err != nil {
			fmt.Println("写入 tar 条目头部时出错:", err)
			return err
		}

		// 如果不是目录，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				fmt.Println("无法打开文件:", err)
				return err
			}
			defer file.Close()

			fmt.Printf("Writing content for %s\n", relPath)

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				fmt.Println("复制文件内容时出错:", err)
				return err
			}
		}
		return nil
	})

	fmt.Println("压缩完成:", tarFileName)
}
