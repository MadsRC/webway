package koanf

import (
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	opts  *configOptions
	koanf *koanf.Koanf
}

func NewConfig(opt ...ConfigOption) (*Config, error) {
	opts := defaultConfigOptions
	for _, o := range globalConfigOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}

	cfg := &Config{
		opts: &opts,
	}

	cfg.koanf = koanf.New(cfg.opts.Delimiter)

	if len(opts.NestedMaps) > 0 {
		for _, m := range opts.NestedMaps {
			_ = cfg.koanf.Load(confmap.Provider(m, cfg.opts.Delimiter), nil)
		}
	}

	return cfg, nil
}

type configOptions struct {
	Delimiter  string
	NestedMaps []map[string]interface{}
}

var defaultConfigOptions = configOptions{
	Delimiter:  ".",
	NestedMaps: make([]map[string]interface{}, 0),
}
var globalConfigOptions []ConfigOption

type ConfigOption interface {
	apply(*configOptions)
}

// funcConfigOption wraps a function that modifies metadataStoreServerOption into an
// implementation of the ConfigOption interface.
type funcConfigOption struct {
	f func(*configOptions)
}

func (fdo *funcConfigOption) apply(do *configOptions) {
	fdo.f(do)
}

func newFuncConfigOption(f func(*configOptions)) *funcConfigOption {
	return &funcConfigOption{
		f: f,
	}
}

func WithConfigPathDelimiter(delimiter string) ConfigOption {
	return newFuncConfigOption(func(o *configOptions) {
		o.Delimiter = delimiter
	})
}

func WithConfigNestedMap(m map[string]interface{}) ConfigOption {
	return newFuncConfigOption(func(o *configOptions) {
		o.NestedMaps = append(o.NestedMaps, m)
	})
}

func (c *Config) String(key string) string {
	return c.koanf.String(key)
}

func (c *Config) Bool(key string) bool {
	return c.koanf.Bool(key)
}
