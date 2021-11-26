package gostman

import (
	"errors"
	"flag"
	"io"
	"os"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func init() {
	flag.StringVar(&runtime.flagEnv, "env", "", "Select environment define in .gostman.env.yml")
	flag.StringVar(&runtime.flagSetenv, "setenv", "", "Select environment define in .gostman.env.yml and set it for the future request")
	flag.BoolVar(&runtime.flagReset, "reset", false, "Reset .gostman.runtime.yml")
	flag.BoolVar(&runtime.flagDebug, "debug", false, "Run gostman in debug mode")
}

const (
	defaultEnv      = "no_env"
	envFilename     = ".gostman.env.yml"
	runtimeFilename = ".gostman.runtime.yml"
)

type gmRuntime struct {
	once sync.Once
	m    *testing.M

	flagEnv    string
	flagSetenv string
	flagReset  bool
	flagDebug  bool

	cfgRuntimeFile *os.File

	cfgEnv     map[string]map[string]string
	cfgRuntime struct {
		Env     string                       `yaml:"env"`
		Initial map[string]map[string]string `yaml:"initial"`
		Current map[string]map[string]string `yaml:"current"`
	}

	env     string
	mu      sync.RWMutex
	initial map[string]string // initial variable for the selected environment
	current map[string]string // current variable for the selected environment
}

var runtime = new(gmRuntime)

// Run run gostman runtime and the test. It returns an exit code to pass to os.Exit.
// The runtime should be run in the TestMain before using Gostman.
//
//  func TestMain(m *testing.M) {
// 	 os.Exit(gostman.Run(m))
//  }
func Run(m *testing.M) (code int) {
	runtime.once.Do(func() {
		if !flag.Parsed() {
			flag.Parse()
		}

		if err := runtime.init(m); err != nil {
			log.Fatal(err)
		}
	})
	defer runtime.close()

	return runtime.m.Run()
}

func (gmr *gmRuntime) init(m *testing.M) error {
	// register *testing.M
	gmr.m = m

	// check debug mode
	if gmr.flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	// load config env and runtime
	if err := gmr.loadCfgEnv(); err != nil {
		return err
	}
	if err := gmr.loadCfgRuntime(); err != nil {
		return err
	}

	// check reset
	if gmr.flagReset {
		gmr.reset()
		gmr.populate()
	}

	// check env
	gmr.env = gmr.cfgRuntime.Env
	if gmr.flagSetenv != "" {
		gmr.cfgRuntime.Env = gmr.flagSetenv
		gmr.env = gmr.flagSetenv
	}
	if gmr.flagEnv != "" {
		gmr.env = gmr.flagEnv
	}
	log.Infof("using env %q", gmr.env)

	// set variable for the selected environment
	gmr.initial = gmr.cfgRuntime.Initial[gmr.env]
	gmr.current = gmr.cfgRuntime.Current[gmr.env]

	return nil
}

func (gmr *gmRuntime) loadCfgEnv() error {
	// init first
	gmr.cfgEnv = make(map[string]map[string]string)

	f, err := os.Open(envFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debug(err)
			return nil
		}
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&gmr.cfgEnv); err != nil {
		if errors.Is(err, io.EOF) {
			log.Debug(err)
			return nil
		}
		return err
	}

	return nil
}

func (gmr *gmRuntime) loadCfgRuntime() error {
	// init first
	gmr.reset()

	f, err := os.OpenFile(runtimeFilename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	gmr.cfgRuntimeFile = f

	if err := yaml.NewDecoder(f).Decode(&gmr.cfgRuntime); err != nil {
		// format error, debug it and continue populate the runtime
		log.Debug(err)
	}
	gmr.populate()

	return nil
}

func (gmr *gmRuntime) reset() {
	gmr.cfgRuntime.Env = defaultEnv
	gmr.cfgRuntime.Initial = make(map[string]map[string]string)
	gmr.cfgRuntime.Current = make(map[string]map[string]string)
}

func (gmr *gmRuntime) populate() {
	for env, values := range gmr.cfgEnv {
		if _, ok := gmr.cfgRuntime.Initial[env]; !ok {
			gmr.cfgRuntime.Initial[env] = make(map[string]string)
			gmr.cfgRuntime.Current[env] = make(map[string]string)
		}

		for k, v := range values {
			if rv, ok := gmr.cfgRuntime.Initial[env][k]; !ok || rv != v {
				gmr.cfgRuntime.Initial[env][k] = v
				gmr.cfgRuntime.Current[env][k] = v
			}
		}
	}

	for env, values := range gmr.cfgRuntime.Initial {
		if _, ok := gmr.cfgRuntime.Current[env]; !ok {
			gmr.cfgRuntime.Current[env] = make(map[string]string)
		}

		for k, v := range values {
			if _, ok := gmr.cfgRuntime.Current[env][k]; !ok {
				gmr.cfgRuntime.Current[env][k] = v
			}
		}
	}

	for env, values := range gmr.cfgRuntime.Current {
		if _, ok := gmr.cfgRuntime.Initial[env]; !ok {
			delete(gmr.cfgRuntime.Current, env)
		}

		for k := range values {
			if _, ok := gmr.cfgRuntime.Initial[env][k]; !ok {
				delete(gmr.cfgRuntime.Current[env], k)
			}
		}
	}
}

func (gmr *gmRuntime) close() {
	f := gmr.cfgRuntimeFile
	if f == nil {
		return
	}
	defer f.Close()

	f.Truncate(0)
	f.Seek(0, 0)

	enc := yaml.NewEncoder(f)
	defer enc.Close()

	if err := enc.Encode(gmr.cfgRuntime); err != nil {
		log.Fatal(err)
	}
}

func (gmr *gmRuntime) setEnvVar(name, val string) {
	if runtime.initial == nil {
		return
	}

	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	if _, ok := runtime.initial[name]; !ok {
		runtime.initial[name] = ""
	}
	runtime.current[name] = val
}

func (gmr *gmRuntime) envVar(name string) string {
	if runtime.initial == nil {
		return ""
	}

	runtime.mu.RLock()
	defer runtime.mu.RUnlock()

	return runtime.current[name]
}
