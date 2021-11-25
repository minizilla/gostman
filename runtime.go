package gostman

import (
	"errors"
	"flag"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultEnv      = "no_env"
	envFilename     = ".gostman.env.yml"
	runtimeFilename = ".gostman.runtime.yml"
)

type gostmanRuntime struct {
	once sync.Once
	g    map[string]*Gostman

	flagEnv    string
	flagSetenv string
	flagReset  bool
	flagDebug  bool

	env     map[string]map[string]string
	runtime struct {
		Env     string                       `yaml:"env"`
		Initial map[string]map[string]string `yaml:"initial"`
		Current map[string]map[string]string `yaml:"current"`
	}
}

func init() {
	flag.StringVar(&gostman.flagEnv, "env", "", "Select environment define in .gostman.env.yml")
	flag.StringVar(&gostman.flagSetenv, "setenv", "", "Select environment define in .gostman.env.yml and set it for the future request")
	flag.BoolVar(&gostman.flagReset, "reset", false, "Reset .gostman.runtime.yml")
	flag.BoolVar(&gostman.flagDebug, "debug", false, "Run gostman in debug mode")
}

func (gr *gostmanRuntime) initOnce() {
	gr.once.Do(func() {
		// check debug mode
		if gr.flagDebug {
			log.SetLevel(log.DebugLevel)
		}

		// init runtime
		if err := gostman.init(); err != nil {
			log.Fatal(err)
		}

		// check reset
		if gr.flagReset {
			gr.reset()
			gr.populate()
		}

		// check env
		env := gr.runtime.Env
		if gr.flagSetenv != "" {
			gr.runtime.Env = gr.flagSetenv
			env = gr.flagSetenv
		}
		if gr.flagEnv != "" {
			env = gr.flagEnv
		}
		log.Infof("using env %q", env)
	})
}

func (gr *gostmanRuntime) init() error {
	gostman.g = make(map[string]*Gostman)

	if err := gr.loadEnv(); err != nil {
		return err
	}
	if err := gr.loadRuntime(); err != nil {
		return err
	}

	// TODO: move to testing cleanup
	gr.cleanup()

	return nil
}

func (gr *gostmanRuntime) loadEnv() error {
	f, err := os.Open(envFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debug(err)
			gr.env = make(map[string]map[string]string)
			return nil
		}
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&gr.env); err != nil {
		if errors.Is(err, io.EOF) {
			log.Debug(err)
			gr.env = make(map[string]map[string]string)
			return nil
		}
		return err
	}

	return nil
}

func (gr *gostmanRuntime) loadRuntime() error {
	f, err := os.OpenFile(runtimeFilename, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&gr.runtime); err != nil {
		log.Debug(err)
	}

	gr.populate()
	return nil
}

// reset resets runtime initial and current value
func (gr *gostmanRuntime) reset() {
	gr.runtime.Env = ""
	gr.runtime.Initial = nil
	gr.runtime.Current = nil
}

// populate populates runtime initial and current value using env
func (gr *gostmanRuntime) populate() {
	if gr.runtime.Env == "" {
		gr.runtime.Env = defaultEnv
	}
	if gr.runtime.Initial == nil {
		gr.runtime.Initial = make(map[string]map[string]string)
	}
	if gr.runtime.Current == nil {
		gr.runtime.Current = make(map[string]map[string]string)
	}

	for env, values := range gr.env {
		if _, ok := gr.runtime.Initial[env]; !ok {
			gr.runtime.Initial[env] = make(map[string]string)
			gr.runtime.Current[env] = make(map[string]string)
		}

		for k, v := range values {
			if rv, ok := gr.runtime.Initial[env][k]; !ok || rv != v {
				gr.runtime.Initial[env][k] = v
				gr.runtime.Current[env][k] = v
			}
		}
	}

	for env, values := range gr.runtime.Initial {
		if _, ok := gr.runtime.Current[env]; !ok {
			gr.runtime.Current[env] = make(map[string]string)
		}

		for k, v := range values {
			if _, ok := gr.runtime.Current[env][k]; !ok {
				gr.runtime.Current[env][k] = v
			}
		}
	}

	for env, values := range gr.runtime.Current {
		if _, ok := gr.runtime.Initial[env]; !ok {
			delete(gr.runtime.Current, env)
		}

		for k := range values {
			if _, ok := gr.runtime.Initial[env][k]; !ok {
				delete(gr.runtime.Current[env], k)
			}
		}
	}
}

func (gr *gostmanRuntime) cleanup() {
	f, err := os.OpenFile(runtimeFilename, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	defer enc.Close()

	if err := enc.Encode(gr.runtime); err != nil {
		log.Fatal(err)
	}
}
