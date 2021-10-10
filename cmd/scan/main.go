package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	sepStr   = "    " // 开始换行符
	startStr = "- [ ] "
)

/**
生成形如
- [ ] dirName1
    - [ ] subDirName1
        - [ ] file1.go
    - [ ] subDirName2
    - [ ] file2.go
- [ ] dirName2
    - [ ] file3.go
*/
func main() {
	dirPath := "../../"
	excludeDir := []string{".github", "benchmarks", "cmd", "docs", "examples", "scripts", "target", ".git", ".idea"}

	d, err := os.Open(dirPath)
	if err != nil {
		log.Printf("open dirPath %v fail, err: %v", dirPath, err)
		os.Exit(1)
	}
	fAbsPath, err := filepath.Abs(filepath.Dir(d.Name()))
	if err != nil {
		log.Printf("Abs dirPath %v fail, err: %v", dirPath, err)
		os.Exit(1)
	}

	dirs, err := d.Readdir(0)
	if err != nil {
		log.Printf("read dir %v fail, err: %v", dirPath, err)
		os.Exit(1)
	}
	var res []string
	for _, dir := range dirs {
		if dir.IsDir() && !in(dir.Name(), excludeDir) {
			//log.Printf("start scan dir %v", dir.Name())
			res = append(res, scanDir(dir.Name(), fAbsPath, 0)...)
		}
	}
	//log.Printf("scan res:\n%v", strings.Join(res, "\n"))
	fmt.Println(strings.Join(res, "\n"))
}

// 扫描目录
func scanDir(d string, pPath string /*父文件绝对路径*/, dep int /*表示树的层级*/) (res []string) {
	// log.Printf("d %v parentPath %v dep %v", d, pPath, dep)

	curAbsPath := strings.Join([]string{pPath, d}, "/")
	res = append(res, format(d, dep))

	f, _ := os.Open(curAbsPath)
	subDirs, _ := f.Readdir(0)
	for _, subDir := range subDirs {
		if subDir.IsDir() {
			res = append(res, scanDir(subDir.Name(), curAbsPath, dep+1)...)
		} else if !strings.HasSuffix(subDir.Name(), "_test.go") { // 去除掉测试文件
			res = append(res, format(subDir.Name(), dep+1))
		}
	}
	return
}

func in(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func format(fPath string, dep int) string {
	if dep == 0 {
		return startStr + fPath
	}
	var str string
	for dep > 0 {
		dep--
		str += sepStr
	}
	str += startStr + fPath
	return str
}
