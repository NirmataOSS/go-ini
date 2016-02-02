package ini

import (
	"github.com/golang/glog"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/ini.v1"
	"sync"
)

type updateFunc func()

type IniFile interface {
	// Return value from INI file, or the supplied default
	ReadKey(section, key, defaultVal string) string
	Register(f updateFunc)
}

var updaters map[string][]updateFunc

type fileDetails struct {
	fileName string
	watch    bool      //True if file is being watched
	cfg      *ini.File //details added after file is loaded by ini pkg.
	lock     sync.RWMutex
}

var iniFiler map[string]IniFile


// Loads a new INI file and optionally watches file for changes
func NewIniFile(fileName string) (IniFile, error) {

	if f, _ := iniFiler[fileName]; f != nil {
		return f, nil
	}

	fd := &fileDetails{fileName: fileName}
	if err := fd.load(); err != nil {
		return nil, err
	}

	if updaters == nil {
		updaters = make(map[string][]updateFunc)
	}
	if iniFiler == nil {
		iniFiler =  make(map[string]IniFile)
	}
	iniFiler[fileName] = fd
	return fd, nil
}

func (fd *fileDetails) load() error {
	var err error
	fd.lock.Lock()
	defer fd.lock.Unlock()

	fd.cfg, err = ini.Load(fd.fileName)
	if err != nil {
		glog.Errorf("Unable to load the file %s", fd.fileName)
		return err
	}
	if !fd.watch {
		glog.Info("Starting watch on file: %s", fd.fileName)
		fd.keepWatch()
	}

	return nil
}

func (fd *fileDetails) ReadKey(section, key, defaultVal string) string {
	glog.V(3).Infof("Key to be read %s", key)
	fd.lock.RLock()
	defer fd.lock.RUnlock()

	if val := fd.cfg.Section(section).Key(key).String(); val != "" {
		return val
	}

	return defaultVal
}

func (fd *fileDetails) keepWatch() error {
	watchman, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	glog.V(3).Infof("Watching file %s", fd.fileName)

	go func() {
		for {
			select {
			case event := <-watchman.Events:
				glog.V(3).Infoln("Received file watch event: %s", event.String())
				if event.Op&fsnotify.Write == fsnotify.Write {
					glog.V(3).Infof("modified file: %s", event.Name)
					fd.load()
					fd.update()
				} else {
					glog.V(3).Infof("Ignoring file event: %s", event.String())
					continue
				}
			}
		}
	}()

	err = watchman.Add(fd.fileName)
	if err != nil {
		glog.Error("Failed to watch file %s: %v", fd.fileName, err)
		return err
	}
	fd.watch = true
	return nil
}

func (fd *fileDetails) update() {

	updateFunc := updaters[fd.fileName]

	if updateFunc == nil {
		glog.V(3).Infof("Update function not found for %s", fd.fileName)
		return
	}
	for _, cbUpdateFunc := range updateFunc {
		cbUpdateFunc()  /* Loop through all the registered functions */
	}
}

func (fd *fileDetails) Register(f updateFunc) {
	/* Adding all the register function in an array */
	updaters[fd.fileName] = append(updaters[fd.fileName], f)
}
