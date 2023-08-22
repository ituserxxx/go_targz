# 使用 golang 写的 tar.gz 压缩工具

##### 实现如下功能
读取同级目录的 filter_config.json 配置文件，该 filter_config.json 是一个过滤条件，满足条件的都将过滤掉（包括子级目录下）

如没有此 json 文件，则压缩当前目录全部文件

如：okmes_，那么所有同级目录和子目录中，匹配上前缀的文件夹全部跳过压缩， 取当前所在目录的文件夹名称作为压缩包名称，将同级，子级所有文件和文件夹都压缩成tar.gz

#### 使用须知

输入  go_targz.exe help 将输出以下说明

1.将该程序放到需要打包的目录下
2.在目录创建一个 filter_config.json 文件，按需填入过滤的相关内容，
```
{
    "dir": [],
    "files": [],
    "file_prex": [],
    "file_suffix":[],
    "dir_prex": [],
    "dir_suffix":[]
}
```

字段解释：
- dir 需要过滤的目录,如: ["my_img","static"]
- files 需要过滤的文件完整名称，如：["main.go", "xxx.text"]
- file_pres 需要过滤的文件前缀，如：["img_", "aaa.fff"]
- file_suffix 需要过滤的文件后缀，如：[".go", ".txt"]
- dir_prex 需要过滤的目录前缀，如：["imgs_"]
- dir_suffix 需要过滤的目录后缀，如：["_docs"]

3.然后双击执行该程序，会在当前目录生成一个以当前目录为前缀-时间戳的新 .tar.gz 文件


#### 按需下载
windows 使用 go_targz_win.exe

mac 使用 go_targz_mac

linux 使用 go_targz_linux

