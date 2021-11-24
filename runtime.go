package gostman

import (
	"flag"
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

	flagEnv   string
	flagReset bool

	env     map[string]map[string]string
	runtime struct {
		Env     string                       `yaml:"env"`
		Initial map[string]map[string]string `yaml:"initial"`
		Current map[string]map[string]string `yaml:"current"`
	}
}

func init() {
	flag.StringVar(&gostman.flagEnv, "env", "", "Selected environment define in .gostman.env.yml")
	flag.BoolVar(&gostman.flagReset, "reset", false, "Reset .gostman.runtime.yml")
}

func (gr *gostmanRuntime) initOnce() {
	gr.once.Do(func() {
		gostman.g = make(map[string]*Gostman)
		if err := gostman.init(); err != nil {
			log.Fatal(err)
		}

		if gr.flagEnv == "" {
			gr.flagEnv = gr.runtime.Env
		}
	})
}

func (gr *gostmanRuntime) init() error {
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
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(&gr.env)
}

func (gr *gostmanRuntime) loadRuntime() error {
	f, err := os.OpenFile(runtimeFilename, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&gr.runtime); err != nil {
		gr.reset()
		return nil
	}

	gr.populate()
	return nil
}

// reset resets runtime initial and current value using env
func (gr *gostmanRuntime) reset() {
	gr.runtime.Env = defaultEnv
	gr.runtime.Initial = make(map[string]map[string]string)
	gr.runtime.Current = make(map[string]map[string]string)

	for env, values := range gr.env {
		initial := make(map[string]string)
		current := make(map[string]string)

		for k, v := range values {
			initial[k] = v
			current[k] = v
		}

		gr.runtime.Initial[env] = initial
		gr.runtime.Current[env] = current
	}
}

// populate populates runtime initial and current value using env
func (gr *gostmanRuntime) populate() {
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
