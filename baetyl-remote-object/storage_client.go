package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
	"github.com/docker/distribution/uuid"
	"github.com/panjf2000/ants"
)

// Task StorageClient
type Task struct {
	msg *EventMessage
	cb  func(msg *EventMessage, err error)
}

// FileStats upload stats
type FileStats struct {
	success uint64
	fail    uint64
	limit   uint64
	deleted uint64
}

// StorageClient StorageClient
type StorageClient struct {
	cfg    ClientInfo
	sh     IObjectStorage
	log    logger.Logger
	tomb   utils.Tomb
	pool   *ants.PoolWithFunc
	lock   sync.RWMutex
	stats  Stats
	pwd    string
	fs     *FileStats
	report report
}

// NewStorageClient creates a new newStorageClient
func NewStorageClient(cfg ClientInfo, r report) (*StorageClient, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	sh, err := NewObjectStorageHandler(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client (%s): %s", cfg.Name, err.Error())
	}
	b := &StorageClient{
		cfg:    cfg,
		report: r,
		sh:     sh,
		log:    logger.WithField("storage client", cfg.Name),
		pwd:    pwd,
		fs:     &FileStats{},
	}
	if r != nil {
		return b, b.tomb.Go(b.reporting)
	}
	return b, nil
}

// CallAsync submit task
func (cli *StorageClient) CallAsync(msg *EventMessage, cb func(msg *EventMessage, err error)) error {
	if !cli.tomb.Alive() {
		return fmt.Errorf("client (%s) closed", cli.cfg.Name)
	}
	return cli.invoke(msg, cb)
}

func (cli *StorageClient) invoke(msg *EventMessage, cb func(msg *EventMessage, err error)) error {
	if cli.pool.Running() == cli.cfg.Pool.Worker {
		cb(msg, fmt.Errorf("failed to submit task: no worker can be used"))
		return nil
	}
	task := &Task{
		msg: msg,
		cb:  cb,
	}
	if err := cli.pool.Invoke(task); err != nil {
		cb(msg, fmt.Errorf("failed to invoke pool task: %s", err.Error()))
		return nil
	}
	return nil
}

func (cli *StorageClient) call(task interface{}) {
	t, ok := task.(*Task)
	if !ok {
		return
	}
	var err error
	switch t.msg.Event.Type {
	case Upload:
		uploadEvent := t.msg.Event.Content.(*UploadEvent)
		err = cli.handleUploadEvent(uploadEvent)
	default:
		err = fmt.Errorf("EventMessage type unexpected")
	}
	if err != nil {
		cli.log.Errorf("failed to fetch: %s", err.Error())
	}
	if t.cb != nil {
		t.cb(t.msg, err)
	}
}

// upload upload object to service(BOS, CEPH or AWS S3)
func (cli *StorageClient) upload(f, remotePath string, meta map[string]string) error {
	fi, err := os.Stat(f)
	if err != nil {
		cli.log.Errorf("failed to get file info: %s", err.Error())
		return nil
	}
	fsize := fi.Size()
	md5, err := utils.CalculateFileMD5(f)
	if err != nil {
		cli.log.Errorf("failed to calculate file[%s] MD5: %s", f, err.Error())
		return nil
	}
	saved := cli.checkFile(remotePath, md5)
	if saved {
		return nil
	}
	if cli.cfg.Limit.Switch {
		month := time.Unix(0, time.Now().UnixNano()).Format("2006-01")
		err = cli.checkData(fsize, month)
		if err != nil {
			cli.log.Errorf("failed to pass data check: %s", err.Error())
			atomic.AddUint64(&cli.fs.limit, 1)
			return nil
		}
		err = cli.sh.PutObjectFromFile(cli.cfg.Bucket, remotePath, f, meta)
		if err != nil {
			atomic.AddUint64(&cli.fs.fail, 1)
			return err
		}
		atomic.AddUint64(&cli.fs.success, 1)
		return cli.increaseData(fsize, month)
	}
	err = cli.sh.PutObjectFromFile(cli.cfg.Bucket, remotePath, f, meta)
	if err != nil {
		atomic.AddUint64(&cli.fs.fail, 1)
		return err
	}
	atomic.AddUint64(&cli.fs.success, 1)
	return nil
}

func (cli *StorageClient) handleUploadEvent(e *UploadEvent) error {
	if strings.Contains(e.LocalPath, "..") {
		cli.log.Errorf("failed to pass LocalPath (%s) check: the local path can't contains ..", e.LocalPath)
		return nil
	}
	var t string
	p, err := filepath.EvalSymlinks(path.Join(cli.pwd, e.LocalPath))
	if err != nil {
		cli.log.Errorf("failed get real dir path: %s", err.Error())
		atomic.AddUint64(&cli.fs.deleted, 1)
		return nil
	}
	if ok := utils.FileExists(p); ok {
		if e.Zip {
			t = path.Join(cli.cfg.TempPath, uuid.Generate().String())
			err = utils.Zip([]string{p}, t)
			if err != nil {
				return fmt.Errorf("failed to zip dir %s: %s", t, err.Error())
			}
		} else {
			t = p
		}
	} else if ok = utils.DirExists(p); ok {
		t = path.Join(cli.cfg.TempPath, uuid.Generate().String())
		if e.Zip {
			err = utils.Zip([]string{p}, t)
			if err != nil {
				return fmt.Errorf("failed to zip dir %s: %s", t, err.Error())
			}
		} else {
			err = utils.Tar([]string{p}, t)
			if err != nil {
				return fmt.Errorf("failed to tar dir %s: %s", t, err.Error())
			}
		}
	} else {
		atomic.AddUint64(&cli.fs.deleted, 1)
		return fmt.Errorf("failed to find path: %s", p)
	}
	if t != p {
		defer os.RemoveAll(t)
	}
	return cli.upload(t, e.RemotePath, e.Meta)
}

func (cli *StorageClient) checkFile(remotePath, md5 string) bool {
	return cli.sh.FileExists(cli.cfg.Bucket, remotePath, md5)
}

func (cli *StorageClient) checkData(fsize int64, month string) error {
	if cli.cfg.Limit.Data <= 0 {
		return nil
	}
	cli.lock.RLock()
	defer cli.lock.RUnlock()
	if _, ok := cli.stats.Months[month]; ok {
		new := cli.stats.Months[month].Bytes + fsize
		if new > cli.cfg.Limit.Data {
			return fmt.Errorf("exceeds max upload data size of this month，stop to upload and will reset next month")
		}
	}
	return nil
}

func (cli *StorageClient) increaseData(fsize int64, month string) error {
	cli.lock.Lock()
	defer cli.lock.Unlock()
	if _, ok := cli.stats.Months[month]; !ok {
		cli.stats.Months[month] = &Item{}
	}
	cli.stats.Total.Bytes = cli.stats.Total.Bytes + fsize
	cli.stats.Total.Count++
	cli.stats.Months[month].Bytes = cli.stats.Months[month].Bytes + fsize
	cli.stats.Months[month].Count++
	return DumpYAML(cli.cfg.Limit.Path, &cli.stats)
}

func (cli *StorageClient) reporting() error {
	defer cli.log.Debugf("storage client reporting task stopped")
	var err error
	t := time.NewTicker(cli.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-cli.tomb.Dying():
			return nil
		case <-t.C:
			stats := map[string]interface{}{
				cli.cfg.Name: map[string]interface{}{
					"success": cli.fs.success,
					"fail":    cli.fs.fail,
					"limit":   cli.fs.limit,
					"deleted": cli.fs.deleted,
				},
			}
			err = cli.report(stats)
			if err != nil {
				cli.log.Warnf("failed to report storage client file stats")
			}
			cli.log.Debugln(stats)
		}
	}
}

// Start start all worker
func (cli *StorageClient) Start() error {
	err := os.MkdirAll(cli.cfg.TempPath, 0755)
	if err != nil {
		cli.log.Errorf("failed to make dir (%s): %s", cli.cfg.TempPath, err.Error())
		return err
	}
	if ok := utils.FileExists(cli.cfg.Limit.Path); !ok {
		basepath := path.Dir(cli.cfg.Limit.Path)
		err = os.MkdirAll(basepath, 0755)
		if err != nil {
			cli.log.Errorf("failed to make dir (%s): %s", basepath, err.Error())
			return err
		}
		f, err := os.Create(cli.cfg.Limit.Path)
		defer f.Close()
		if err != nil {
			cli.log.Errorf("failed to make file (%s): %s", cli.cfg.Limit.Path, err.Error())
			return err
		}
	}
	utils.LoadYAML(cli.cfg.Limit.Path, &cli.stats)
	p, err := ants.NewPoolWithFunc(cli.cfg.Pool.Worker, cli.call, ants.WithExpiryDuration(cli.cfg.Pool.Idletime))
	if err != nil {
		cli.log.Errorf("failed to create a pool: %s", err.Error())
		return err
	}
	cli.pool = p
	cli.log.Debugf("storage client start")
	return nil
}

// Close close client and all worker
func (cli *StorageClient) Close() error {
	cli.pool.Release()
	cli.tomb.Kill(nil)
	cli.log.Debugf("storage client closed")
	return cli.tomb.Wait()
}
