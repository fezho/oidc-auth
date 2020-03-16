package options

import (
	"github.com/spf13/pflag"
)

// Options has all the params can be passed from command line opts
type Options struct {
	PrintVersion bool

	// ConfigFile is the location of the auth server's configuration file.
	ConfigFile string
}

// NewOptions returns default scheduler app options.
func NewOptions() (*Options, error) {
	o := &Options{
		PrintVersion: false,
		ConfigFile:   "",
	}

	return o, nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.PrintVersion, "version", "v", false, "Show version and exit")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file. Flags override values in this file.")
}
