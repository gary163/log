package log

import (
	"sync"
	"os"
	"time"
	"encoding/json"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
	"strconv"
	"path"
	"io"
	"bytes"
	"fmt"
)

type file struct {
	sync.RWMutex
	Filename      string `json:"filename"`
	fileNameOnly  string
	suffix        string
	writer        *os.File
	MaxSize       int `json:"maxsize"`
	sizeNow       int
	Daily         bool  `json:"daily"`
	MaxDays       int64 `json:"maxdays"`
	dailyOpenTime time.Time
	Rotate        bool `json:"rotate"`
	Level         int `json:"level"`
	Perm          string `json:"perm"`
}

func newFile() Logger {
	w := &file{
		Daily:      true,
		MaxDays:    7,
		Rotate:     true,
		Level:      LevelDebug,
		Perm:       "0660",
		MaxSize:    1 << 28,
	}
	return w
}

func init() {
	Register(AdapterFile, newFile)
}

func (f *file) Init(config string) error {
	err := json.Unmarshal([]byte(config),f)
	if err != nil {
		return err
	}

	if len(f.Filename) == 0 {
		return errors.New("Filename must be setting")
	}

	f.suffix = filepath.Ext(f.Filename)
	f.fileNameOnly = strings.TrimSuffix(f.Filename, f.suffix)
	if f.suffix == "" {
		f.suffix = ".log"
	}
	go f.monitorFile()
	return f.startLogger()
}

func (f *file) WriteMsg(when time.Time, msg string, level int) error {
	if f.Level < level {
		fmt.Println(f.Level,level)
		return nil
	}

	t := when.Format("2006-01-02 15:04:05")
	msg = t + " " + msg + "\n"

	f.Lock()
	_,err := f.writer.Write([]byte(msg))
	if err == nil {
		f.sizeNow += len(msg)
	}
	f.Unlock()

	return nil
}

func (f *file) monitorFile() {
	ticker := time.NewTicker(time.Second)
	fmt.Println("monitorFile")
	for{
		select {
		case <-ticker.C:
			go func(){
				f.Lock()
				if err := f.doRotate(time.Now()); err != nil {
					fmt.Fprintf(os.Stderr, "MonitourFile:FileLogWriter(%q): %s\n", f.Filename, err)
				}
				f.Unlock()
			}()
		}
	}
}

func (f *file) doRotate(t time.Time) error {
	if !f.Rotate {
		return nil
	}
	if ( f.MaxSize > 0 && f.sizeNow > f.MaxSize )|| f.dailyOpenTime.Day() != t.Day() {
		_,err := os.Lstat(f.Filename)
		if err != nil {
			f.restartLogger()
		}

		format := "2006-01-02"
		newFileName := fmt.Sprintf("%s%s_%d%02d%02d%s",f.fileNameOnly,t.Format(format),t.Hour(),t.Minute(),t.Second(),f.suffix)

		if _,err = os.Lstat(newFileName); err == nil {
			return fmt.Errorf("Rotate:Trying make a filename,but the file %s is existed",newFileName)
		}

		f.writer.Close()
		if err := os.Rename(f.Filename,newFileName); err != nil {
			f.restartLogger()
		}
		if err = os.Chmod(newFileName,os.FileMode(0440)); err != nil {
			return fmt.Errorf("Rotate: chomd err: %s",err)
		}

		return f.restartLogger()
	}
	return nil
}

func (f *file) restartLogger() error {
	f.startLogger()
	go f.removeOldFile()
	return nil
}

func (f *file) removeOldFile() {
	dir := filepath.Dir(f.Filename)
	filepath.Walk(dir, func(path string,info os.FileInfo,err error) error {
		defer func(){
			if err := recover(); err != nil {
				fmt.Fprintf(os.Stderr,"Unable to delete old log '%s', error: %v\n",f.Filename,err)
			}
		}()
		if f.Daily {
			if !info.IsDir() && info.ModTime().Add(24 * time.Hour * time.Duration(f.MaxDays)).Before(time.Now()) {
				if strings.HasPrefix(filepath.Base(path), filepath.Base(f.fileNameOnly)) &&
					strings.HasSuffix(filepath.Base(path), f.suffix) {
					os.Remove(path)
				}
			}
		}
		return nil
	})
}

func (f *file) Destroy() {
	f.writer.Close()
}

func (f *file) Flush() {
	f.writer.Sync()
}

func (f *file) createFile() (*os.File,error) {
	perm,err := strconv.ParseInt(f.Perm,8,64)
	if err != nil {
		return nil,err
	}
	filePerm := os.FileMode(perm)
	dir := path.Dir(f.Filename)
	err = os.MkdirAll(dir,filePerm)
	if err != nil {
		return nil,err
	}

	fd,err := os.OpenFile(f.Filename,os.O_CREATE|os.O_APPEND|os.O_WRONLY,filePerm)
	if err != nil {
		return nil,err
	}

	return fd,nil
}

func (f *file) startLogger() error {
	fd,err :=f.createFile()
	if err != nil {
		return err
	}

	if f.writer != nil {
		f.writer.Close()
	}
	f.writer = fd
	return f.initFd()
}

func (f *file) initFd() error {
	f.dailyOpenTime = time.Now()
	fileInfo,err := f.writer.Stat()
	if err != nil {
		return err
	}
	f.sizeNow = int(fileInfo.Size())
	return nil
}

func (f *file) getFileLine() (int,error) {
	fd,err := os.Open(f.Filename)
	if err != nil {
		return 0,err
	}
	defer fd.Close()

	buf := make([]byte,32768)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c,err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count,err
		}
		count += bytes.Count(buf[:c],lineSep)

		if err == io.EOF {
			break
		}
	}
	return count,nil
}


