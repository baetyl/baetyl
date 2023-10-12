package native

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"

	"github.com/baetyl/baetyl/v2/ami"
)

func (impl *nativeImpl) FetchLog(_, pod, _ string, tailLines, _ int64) ([]byte, error) {
	logPath := impl.logHostPath
	pathArr := strings.Split(pod, ".")
	if len(pathArr) != 5 {
		return nil, errors.Trace(errors.New("log path error" + pod))
	}
	logPath = logPath + "/" + pathArr[0] + "/" + pathArr[1] + "/" + pathArr[2] + "/" + pathArr[3] + "-" + pathArr[4] + ".log"
	tailLines = 200
	if tailLines != 0 {
		tailLines = tailLines
	}
	// windows系统获取log 通过读取文件
	if runtime.GOOS == "windows" {
		return fetchLogForWindows(logPath, tailLines)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "tail", "-n", strconv.FormatInt(tailLines, 10), logPath)
	return cmd.Output()
}

func fetchLogForWindows(logPath string, tailLines int64) ([]byte, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer file.Close()
	if err = seekLastN(file, tailLines); err != nil {
		return nil, errors.Trace(err)
	}
	reader := bufio.NewReader(file)
	bytes := make([]byte, 10240)
	_, err = reader.Read(bytes)
	return bytes, err
}

// RemoteLogs use command tail -f
func (impl *nativeImpl) RemoteLogs(option *ami.LogsOptions, pipe ami.Pipe) error {
	logPath := impl.logHostPath
	pathArr := strings.Split(option.Name, ".")
	if len(pathArr) != 5 {
		return errors.Trace(errors.New("log path error"))
	}
	logPath = logPath + "/" + pathArr[0] + "/" + pathArr[1] + "/" + pathArr[2] + "/" + pathArr[3] + "-" + pathArr[4] + ".log"
	tailLines := int64(200)
	if option.TailLines != nil {
		tailLines = *option.TailLines
	}
	// windows系统获取log 通过读取文件
	if runtime.GOOS == "windows" {
		return getLogForWindows(logPath, tailLines, pipe)
	}
	cmd := exec.CommandContext(pipe.Ctx, "tail", "-n", strconv.FormatInt(tailLines, 10), "-f", logPath)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Trace(err)
	}
	defer stdoutPipe.Close()
	err = cmd.Start()
	if err != nil {
		return errors.Trace(err)
	}
	_, err = io.Copy(pipe.OutWriter, stdoutPipe)
	if err != nil {
		return errors.Trace(err)
	}
	err = cmd.Wait()
	if err != nil {
		if err.Error() == CmdKillErr {
			return nil
		}
		return errors.Trace(err)
	}
	return nil
}

func getLogForWindows(logPath string, tailLines int64, pipe ami.Pipe) error {
	file, err := os.Open(logPath)
	if err != nil {
		return errors.Trace(err)
	}
	defer file.Close()
	if err = seekLastN(file, tailLines); err != nil {
		return errors.Trace(err)
	}
	reader := bufio.NewReader(file)
	for {
		select {
		case <-pipe.Ctx.Done():
			return nil
		default:
			data, _, err := reader.ReadLine()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			data = append(data, '\n')
			_, err = pipe.OutWriter.Write(data)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
}

func seekLastN(file *os.File, tailLines int64) error {
	charBuff := make([]byte, 1)
	cursor := int64(0)
	cnt := int64(0)
	// 逆向遍历字节,寻找倒数第n行数据
	fileInfo, err := file.Stat()
	if err != nil {
		return errors.Trace(err)
	}
	size := fileInfo.Size()
	for {
		cursor += 1
		_, err = file.Seek(size-cursor, io.SeekStart)
		if err != nil {
			_, err = file.Seek(0, 0)
			if err != nil {
				return errors.Trace(err)
			}
			break
		}
		_, err = file.Read(charBuff)
		if err != nil {
			return errors.Trace(err)
		}
		if charBuff[0] == '\n' {
			cnt++
		}
		if cnt >= tailLines+1 {
			break
		}
	}
	return nil
}
