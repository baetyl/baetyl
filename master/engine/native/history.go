package native

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/baidu/openedge/utils"
	"github.com/shirou/gopsutil/process"
)

const historyFile = "var/db/openedge/processes.history"

func (e *nativeEngine) initHistory() {
	e.cleanHistory()
	err := os.MkdirAll(path.Dir(historyFile), 0755)
	if err != nil {
		e.log.WithError(err).Warnf("failed to create path of processes.history")
	}
}

func (e *nativeEngine) cleanHistory() {
	if !utils.FileExists(historyFile) {
		return
	}
	e.mut.Lock()
	defer e.mut.Unlock()

	defer func() {
		err := os.Rename(historyFile, fmt.Sprintf("%s.%d", historyFile, time.Now().Unix()))
		if err != nil {
			e.log.WithError(err).Warnf("failed to backup processes.history")
		}
	}()

	file, err := os.Open(historyFile)
	if err != nil {
		e.log.WithError(err).Warnf("failed to open processes.history")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		e.log.Debugf("get line (%s) from processes.history", line)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			e.log.WithError(err).Warnf("get invalid line (%s) from processes.history", line)
			continue
		}
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			e.log.WithError(err).Warnf("get invalid pid (%s) from processes.history", parts[0])
			continue
		}
		p, _ := process.NewProcess(int32(pid))
		n, err := p.Exe()
		if err != nil {
			if !strings.Contains(err.Error(), "exit") {
				e.log.WithError(err).Warnf("failed to get exe of process (%d) from processes.history", pid)
			}
			continue
		}
		if parts[1] == n {
			err = p.Kill()
			if err != nil {
				e.log.WithError(err).Debugf("failed to kill process (%d) from processes.history", pid)
			} else {
				e.log.Infof("kill process (%d) from processes.history", pid)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		e.log.WithError(err).Warnf("failed to scan processes.history")
	}
	return
}

func (e *nativeEngine) appendHistory(pid int) {
	e.mut.Lock()
	defer e.mut.Unlock()

	p, _ := process.NewProcess(int32(pid))
	n, err := p.Exe()
	if err != nil {
		e.log.WithError(err).Warnf("failed to get exe of process (%d) to write into processes.history", pid)
		return
	}

	f, err := os.OpenFile(historyFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		e.log.WithError(err).Warnf("failed to open processes.history to write process (%s)", pid)
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintln(pid, n))
	if err != nil {
		e.log.WithError(err).Warnf("failed to write processes (%s) into processes.history", pid)
	}
}
