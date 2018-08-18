/*
 * Author: gao88<fbg@live.com>
 * Created Time: 2016-08-17 9:30
 * Last Modified: 2016-08-18 18:02:36
 * File Name: updatesvr.go
 * Description:当统计在线人数为0时，而且自已的MD5和上次不一样，则自动重启并更新自已，监控程序可以自己写或用supervisor等工具，让监控程序拉起服务。
 */

package autoupdate

import (
	"os"
	"time"
	"strings"
	"runtime"
	"errors"
	"log"
	"io"
	"fmt"
	"os/exec"
	"path/filepath"
	"crypto/md5"
	"encoding/hex"
)

var updateSvrInstance = &UpdateSvr{}

type UpdateSvr struct {
	svrExeDir    string
	svrExeName   string
	svrExeMd5    string
	count        int32
	logger       *log.Logger
}

func init() {
	initUpdateSvr()
}

func PushCount(count int32) {
	if updateSvrInstance == nil {
		return
	}

	updateSvrInstance.count = count
}

func initUpdateSvr() {
	svrExePath := getCurExePath()

	var substr string
	if "windows" == runtime.GOOS {
		substr = "\\"
	} else {
		substr = "/"
	}

	updateSvrInstance = &UpdateSvr{}
	updateSvrInstance.svrExeDir = getFilePath(svrExePath) + substr
	updateSvrInstance.svrExeName = getFileName(svrExePath)
	initLogger()

	md5, err := getFileMd5(svrExePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	updateSvrInstance.svrExeMd5 = md5
	fmt.Println("svrExePath:", updateSvrInstance.svrExeDir, ", svrExeName:", updateSvrInstance.svrExeName, ", svrExeMd5:", updateSvrInstance.svrExeMd5)

	go updateSvrInstance.onTimer()
}

func getCurExePath() string {
	return os.Args[0]
}

func getFilePath(filePath string) string {
	file, _ := exec.LookPath(filePath)
	path, _ := filepath.Abs(file)

	var substr string

	if "windows" == runtime.GOOS {
		substr = "\\"
	} else {
		substr = "/"
	}

	path = string(path[0:strings.LastIndex(path, substr)])

	return path
}

func getFileName(filePath string) string {
	var substr string

	if "windows" == runtime.GOOS {
		substr = "\\"
	} else {
		substr = "/"
	}

	return string(filePath[strings.LastIndex(filePath, substr)+1:])
}

func getFileMd5(filePath string) (str string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	md5h := md5.New()
	io.Copy(md5h, file)

	md5Str := md5h.Sum([]byte(""))
	if len(md5Str) == 0 {
		return "", errors.New("len(md5Str) == 0")
	}

	return hex.EncodeToString(md5Str), nil
}

func killAll(processName string) error {
	cmd := exec.Command("killall", processName)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("kill:", processName)

	return nil
}

func startProcess(fileDir string, fileName string) *os.Process {
	if fileDir == "" || fileName == "" {
		return nil
	}

	cmd := exec.Command("./"+fileName, "-d=true")
	cmd.Dir = fileDir
	err := cmd.Run()
	if err == nil {
		fmt.Println("Start, fileDir:", fileDir, ", fileName:", fileName, ", OK!")
	} else {
		fmt.Println("Error:", err)
	}

	return cmd.Process
}

func initLogger() {
	cdir, err := os.Getwd()
	if err != nil {
		fmt.Println("err:", err.Error())
	}

	path := cdir

	if "windows" == runtime.GOOS {
		path = path + "\\..\\log\\"
	} else {
		path = path + "/../log/"
	}

	err = os.MkdirAll(path, 0777)

	logFile, err := os.OpenFile(path+"update_svr.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("open file error=%s\r\n", err.Error())
		os.Exit(-1)
	}

	writers := []io.Writer{
		logFile,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)

	updateSvrInstance.logger = log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime)
}

func updateLog(log string) {
	if updateSvrInstance != nil && updateSvrInstance.logger != nil {
		updateSvrInstance.logger.Println(log)
	}
}

func (this *UpdateSvr) onTimer() {
	fmt.Println("onTimer...")

	for {
		newSvrMd5, err := getFileMd5(this.svrExeDir + this.svrExeName)
		if err != nil {
			fmt.Println(err)
			time.Sleep(3 * time.Second)
			continue
		}

		if newSvrMd5 == "" {
			fmt.Println("newSvrMd5 == \"\"")
			time.Sleep(3 * time.Second)
			continue
		}

		//fmt.Println("svrExeFileMd5:", this.svrExeMd5, ", newSvrMd5:", newSvrMd5, ", robotMd5:", this.robotMd5, ", newRobotMd5:", newRobotMd5, ", count:", this.count)

		if this.svrExeMd5 != newSvrMd5 && this.count <= 0 {
			fmt.Println("Server exit, svrExeFileMd5:", this.svrExeMd5, ", newSvrMd5:", newSvrMd5, ", count:", this.count)

			updateLog("update: " + updateSvrInstance.svrExeName)

			os.Exit(-1)
			return
		}

		time.Sleep(10 * time.Second)
	}
}
