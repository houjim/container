package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

/*
启动测试程序方法
此次的环境是ubuntu18，运行程序需要sudo权限。其他环境下，直接使用root运行。
编译程序：go build -o mydocker main.go
运行程序：sudo ./docker run /bin/bash
*/
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "subcmd":
		subcmd()
	default:
		panic("help")
	}
}

func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"subcmd"}, os.Args[2:]...)...)
	/*
		完成操作系统的标准输入输出和当前环境的映射
	*/
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	/*
	Cloneflags：指定需要隔离的对象。
	syscall系统调用，建立宿主机之外的运行环境
	CLONE_NEWUTS：启动一个新的。
	CLONE_NEWPID：PID namespaces用来隔离进程的ID空间，在不同空间中可以出一样的PID。
	CLONE_NEWNS：在新的mount命名空间中创建进程
	Unshareflags：不再共享父进程中指定的命名空间
	关于更多的资源隔离，可以参考https://www.man7.org/linux/man-pages/man2/clone.2.html
	*/
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	exception(cmd.Run())
}

func subcmd() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cg()

	/*
	完成操作系统的标准输入输出和当前环境的映射
	*/
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr


	/*设置登陆docker后的主机名称*/
	exception(syscall.Sethostname([]byte("mydocker")))

	/*
	此次使用的ubuntu的基本内核作为文件挂载点。
	https://cdimage.ubuntu.com/ubuntu-base/releases/18.04/release/ubuntu-base-18.04.5-base-amd64.tar.gz
	解压下载包，并把目录修改为你需要挂载的目录名称。
	*/
	exception(syscall.Chroot("/home/master/gocode/ubuntufs"))

	/*
		Sethostname：通过系统调用更改当前环境的主机名称
		Chroot：设置docker的初始目录。
		Chdir：进入docker，将目录修改到根目录。
		Mount：在docker中，挂载相关的文件目录
	*/
	exception(os.Chdir("/"))
	exception(syscall.Mount("proc", "proc", "proc", 0, ""))

	exception(cmd.Run())

	exception(syscall.Unmount("proc", 0))
}

/*
cg（）函数，对Linux操作的Cgroup资源进行设置。
*/
func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	/*指定进程的所在目录*/
	os.Mkdir(filepath.Join(pids, "mydocker"), 0755)
	/*指定docker可以运行的进程数量*/
	exception(ioutil.WriteFile(filepath.Join(pids, "mydocker/pids.max"), []byte("20"), 0700))
	exception(ioutil.WriteFile(filepath.Join(pids, "mydocker/notify_on_release"), []byte("1"), 0700))
	exception(ioutil.WriteFile(filepath.Join(pids, "mydocker/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func exception(err error) {
	if err != nil {
		panic(err)
	}
}
