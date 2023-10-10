package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	filePath := "un_targz_test.tar.gz" // 要解压的文件路径

	err := extractTarGz(filePath, "./test_un_targz/") // 解压到目标文件夹路径
	if err != nil {
		fmt.Println("解压失败:", err)
	} else {
		fmt.Println("解压成功")
	}
}

func extractTarGz(filePath, destination string) error {
	// 打开要解压的 .tar.gz 文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建 gzip.Reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// 创建 tar.Reader
	tarReader := tar.NewReader(gzReader)

	// 遍历 tar 文件中的每个文件和目录
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 已达到 tar 文件末尾，结束遍历
		}
		if err != nil {
			return err
		}

		// 根据 header 中的文件信息创建或提取文件
		switch header.Typeflag {
		case tar.TypeDir: // 目录
			err = os.MkdirAll(destination+"/"+header.Name, os.ModePerm)
			if err != nil {
				return err
			}
		case tar.TypeReg: // 文件
			filePath := destination + "/" + header.Name

			// 创建目录
			err = os.MkdirAll(getDirName(filePath), os.ModePerm)
			if err != nil {
				return err
			}

			// 创建文件并写入数据
			outputFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, tarReader)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("未知文件类型: %s in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func getDirName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return ""
}
