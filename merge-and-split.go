package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	version          = "1.0.0-nightly"
	defaultChunkSizeMB = 24
)

func main() {
	if len(os.Args) < 2 {
	printUsage()
	return
    }

	mode := os.Args[1]

    if mode == "version" || mode == "-v" || mode == "--version" {
	fmt.Printf("merge-and-split %s\n", version)
	return
    }


	switch mode {
	case "split":
		if len(os.Args) < 3 {
			fmt.Println("错误：缺少要分割的文件名！")
			fmt.Println("用法: merge-and-split split 文件名.pdf [每片大小MB]")
			os.Exit(1)
		}
		sourceFile := os.Args[2]
		chunkSizeMB := defaultChunkSizeMB
		if len(os.Args) >= 4 {
			size, err := strconv.Atoi(os.Args[3])
			if err == nil && size > 0 {
				chunkSizeMB = size
			}
		}
		doSplit(sourceFile, chunkSizeMB)

	case "merge":
		if len(os.Args) < 3 {
			fmt.Println("错误：缺少分片所在的文件夹路径！")
			fmt.Println("用法: merge-and-split merge 分片所在文件夹路径")
			os.Exit(1)
		}
		dirPath := os.Args[2]
		doMerge(dirPath)

	default:
		fmt.Println("未知模式，请使用 'split' 或 'merge'")
		os.Exit(1)
	}
}

func doSplit(sourceFile string, chunkSizeMB int) {
	source, err := os.Open(sourceFile)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		os.Exit(1)
	}
	defer source.Close()

	fileInfo, _ := source.Stat()
	totalSize := fileInfo.Size()
	chunkSize := int64(chunkSizeMB) * 1024 * 1024
	
	if totalSize < chunkSize {
		fmt.Println("文件大小不足一个分片，无需分割。")
		return
	}

	dir := filepath.Dir(sourceFile)
	ext := filepath.Ext(sourceFile)
	baseName := strings.TrimSuffix(filepath.Base(sourceFile), ext)

	chunkCount := 0
	offset := int64(0)
	buf := make([]byte, chunkSize)

	fmt.Printf("正在分割 %s (共 %d MB，每片 %d MB)...\n", filepath.Base(sourceFile), totalSize/(1024*1024), chunkSizeMB)

	for {
		n, err := source.ReadAt(buf, offset)
		if n > 0 {
			chunkCount++
			chunkFilePath := filepath.Join(dir, fmt.Sprintf("%s.part%d%s", baseName, chunkCount, ext))
			partFile, err := os.Create(chunkFilePath)
			if err != nil {
				fmt.Printf("创建分片文件失败: %v\n", err)
				os.Exit(1)
			}
			partFile.Write(buf[:n])
			partFile.Close()
			fmt.Printf(" -> 生成: %s\n", filepath.Base(chunkFilePath))
		}
		if err != nil {
			break
		}
		offset += chunkSize
	}
	fmt.Printf("分割完成，共生成 %d 个分片。\n", chunkCount)
}

func doMerge(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("读取文件夹失败: %v\n", err)
		os.Exit(1)
	}

	var parts []string
	var ext string
	for _, f := range files {
		if !f.IsDir() && strings.Contains(f.Name(), ".part") {
			parts = append(parts, f.Name())
			parts[len(parts)-1] = filepath.Join(dirPath, f.Name())
			ext = f.Name()
			ext = filepath.Ext(ext)
		}
	}

	if len(parts) == 0 {
		fmt.Println("未找到任何 .part 分片文件，无需合并。")
		return
	}

	sort.Strings(parts)

	baseName := filepath.Base(parts[0])
	prefix := strings.Split(baseName, ".part")[0]
	outFile := filepath.Join(dirPath, prefix+ext)

	output, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	fmt.Printf("正在合并生成 %s ...\n", filepath.Base(outFile))
	buf := make([]byte, 1024*1024)

	for i, partFile := range parts {
		fmt.Printf(" -> 读取分片: %s\n", filepath.Base(partFile))
		f, err := os.Open(partFile)
		if err != nil {
			fmt.Printf("打开分片失败: %v\n", err)
			os.Exit(1)
		}
		for {
			n, err := f.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		f.Close()
		
		if i == len(parts)-1 {
			f.Close()
			os.Remove(partFile)
			fmt.Println("  [已删除分片]")
		}
	}
	fmt.Printf("合并完成，文件已保存为 %s\n", filepath.Base(outFile))
}

func printUsage() {
	fmt.Println("merge-and-split — 文件分割与合并工具")
	fmt.Printf("版本: %s\n\n", version)
	fmt.Println("用法:")
	fmt.Println("  merge-and-split split 文件名 [每片大小MB]")
	fmt.Println("  merge-and-split merge 分片所在文件夹")
	fmt.Println("  merge-and-split version          查看版本号")
}
