//go:generate goversioninfo

package main

// https://learnku.com/articles/71107
// Go语言给编译出来的程序添加图标和版本信息

// go mod init example.com/myproject  // 初始化 Go 模块
// go mod tidy //扫描所有我们 import 到的包，并生成对应的记录到 gomod 文件里。
// go build -ldflags="-H windowsgui" -o jh.exe main.go //编译成可执行文件，-H windowsgui 选项可以隐藏控制台窗口。
// go build -ldflags="-H windowsgui" -o jh.exe main.go -trimpath //编译成可执行文件，-H windowsgui 选项可以隐藏控制台窗口，-trimpath 选项可以去掉编译时的路径信息。

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

const (
	IDYES = 6 // 用户点击了 "是"
	IDNO  = 7 // 用户点击了 "否"
)

func main() {
	var folderA, folderB string
	var updateFlag bool

	// 长度为3的字符串数组
	// 根据约定俗成的惯例，Shoes目录都是放在根目录下
	arr := []string{`c:\shoes\jh\`, `d:\shoes\jh\`, `e:\shoes\jh\`}

	folderA = `\\10.10.1.60\shoes\jh\`

	// 遍历数组，打印每个元素的索引和值
	// for key, value := range arr {
	// 	fmt.Println("数组遍历，Key = ", key, " value = ", value)  // 数组遍历，Key =  0  value =  c:\shoes\jh\
	// }

	for _, item := range arr {
		// 调用os.Stat 函数，它会返回文件或目录的信息（FileInfo）以及一个错误（error）
		// 如果文件或目录存在，err 为 nil，并返回文件的元信息
		// 如果文件或目录不存在，err 会包含一个特定的错误值（ErrNotExist）

		// os.IsNotExist(err)
		// 这是一个辅助函数，用来判断 err 是否表示文件或目录不存在。
		// 它的实现是通过检查 err 是否与ErrNotExist 匹配
		// !os.IsNotExist(err)的意思是：如果 err不是表示文件或目录不存在（即文件存在或发生了其他错误），则进入if`语句块。

		//在if表达式之前添加一个执行语句，再根据变量值进行判断.由于是在 if 语句中声明的变量，因此变量的作用域也只在 if 语句中，外部无法访问这些变量。
		if _, err := os.Stat(item); !os.IsNotExist(err) {
			folderB = item
			break
		}
	}

	/*
		// 获取当前可执行文件所在路径
		ex, err := os.Executable()
		if err != nil {
			// fmt.Println("获取可执行文件路径失败:", err)
			return
		}

		// 获取可执行文件所在目录
		exPath := filepath.Dir(ex)

		// 打印可执行文件路径和目录
		// fmt.Println("可执行文件路径:", ex)
		// fmt.Println("可执行文件所在目录:", exPath)

		folderB = exPath
	*/

	if !isServerOS() {
		serverFile := filepath.Join(folderB, "server.txt")
		if _, err := os.Stat(serverFile); err == nil {
			os.Remove(serverFile)
		}
	} else {
		os.Exit(0) //远程服务器上的程序不能由ERP用户有来自动更新
	}

	if fileModifiedTime(filepath.Join(folderB, "jh.exe")).After(fileModifiedTime(filepath.Join(folderA, "jh.exe"))) {
		updateFlag = false // 1不需要更新

		// result := showMessageBox("提示", "jh.exe正在运行中，是否继续运行？")
		// // 根据用户选择执行逻辑
		// if result == IDYES {
		// 	fmt.Println("用户选择继续运行")
		// 	// 继续程序逻辑
		// } else if result == IDNO {
		// 	fmt.Println("用户选择退出程序")
		// 	// 退出程序
		// 	return
		// }
	} else {
		if isProcessRunning("jh.exe") {
			// fmt.Println("请关掉所有ERP窗口再点击本程序更新！")
			// os.Exit(0)
			// 弹出对话框
			// result := showMessageBox("提示", "jh.exe正在运行中，下次运行时自动更新程序？")
			showMessageBox("提示", "服务器上的ERP程序已更新，如果需要更新请关掉所有ERP窗口再次重新运行！")
			os.Exit(0)
			// 根据用户选择执行逻辑
			// if result == IDYES {
			// 	fmt.Println("用户选择继续运行")
			// 	// 继续程序逻辑
			// } else if result == IDNO {
			// 	fmt.Println("用户选择退出程序")
			// 	// 退出程序
			// return
			// }
		}

		copyFiles(folderA, folderB)
		updateFlag = true // 2需要更新
	}

	//程序有更新
	if updateFlag {
		runCommand(filepath.Join(folderB, "jhshoes.exe"))
	}

	runCommand(filepath.Join(folderB, "jh.exe"))

	// c := exec.Command(filepath.Join(folderB, "jh.exe"))
	// 设置进程属性，使其独立运行
	// 设置进程属性，使其独立运行（Windows 版本）
	// c.SysProcAttr = &syscall.SysProcAttr{
	// 	CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	// }
	// if err := c.Start(); err != nil {
	// 	log.Printf("启动命令失败: %v", err)
	// }
	// fmt.Println(filepath.Join(folderB, "jh.exe"))
	os.Exit(0)

}

func isServerOS() bool {
	out, err := exec.Command("cmd", "ver").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "Server")
}

func fileModifiedTime(filePath string) time.Time {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}
	}
	return fileInfo.ModTime()
}

func isProcessRunning(processName string) bool {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), processName)
}

func copyFiles(srcDir, dstDir string) {
	srcFiles, err := os.ReadDir(srcDir)
	if err != nil {
		return
	}
	for _, file := range srcFiles {
		// 跳过目录和以 .pbd 结尾的文件。pbd包含在jhshoes.exe中
		if strings.ToLower(filepath.Ext(file.Name())) != ".pbd" {
			srcFile := filepath.Join(srcDir, file.Name())
			dstFile := filepath.Join(dstDir, file.Name())
			if _, err := os.Stat(dstFile); err == nil {
				if fileModifiedTime(srcFile).After(fileModifiedTime(dstFile)) {
					copyFile(srcFile, dstDir)
				}
				// } else {
				// 	copyFile(srcFile, dstDir)
			}
		}
	}
}

func copyFile(srcFile, dstDir string) {
	input, err := os.ReadFile(srcFile)
	if err != nil {
		return
	}
	dstFile := filepath.Join(dstDir, filepath.Base(srcFile))
	os.WriteFile(dstFile, input, 0644)
}

func runCommand(command string) {
	cmd := exec.Command(command)
	cmd.Dir = filepath.Dir(command)
	// cmd.Run() //会阻塞当前进程，直到命令执行完成
	cmd.Start() //不会阻塞当前进程，命令在后台执行
}

// showMessageBox 使用 windows 包调用 MessageBoxW
func showMessageBox(title, text string) int {
	// 调用 MessageBoxW
	ret, _ := windows.MessageBox(
		0,                               // HWND，0 表示没有父窗口
		windows.StringToUTF16Ptr(text),  // 消息内容
		windows.StringToUTF16Ptr(title), // 标题
		// windows.MB_YESNO|windows.MB_ICONQUESTION, // 弹出一个对话框，用户可以选择 "是" 或 "否"。
		windows.MB_OK|windows.MB_ICONINFORMATION, // 显示一个对话框，只有一个“确定”按钮
	)
	return int(ret)
}
