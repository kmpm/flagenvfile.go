package flagenvfile

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FlagEnv struct {
	flagset    *flag.FlagSet
	envPrefix  string
	fileSuffix string
}

var fe *FlagEnv

type Option func(*FlagEnv) error

func WithFlagSet(f *flag.FlagSet) Option {
	return func(fe *FlagEnv) error {
		fe.flagset = f
		return nil
	}
}

func WithPrefix(name string) Option {
	return func(fe *FlagEnv) error {
		fe.SetEnvPrefix(name)
		return nil
	}
}

// func mustWD() string {
// 	wd, _ := os.Getwd()
// 	return wd
// }

func logFileError(err error, filename string) {
	ap, errp := filepath.Abs(filename)
	if errp != nil {
		log.Printf("error getting an absolute path for file '%s': %v\n", filename, errp)
		ap = filename
	}
	log.Printf("error dealting with file '%s' representing a flag: %v\n", ap, err)
	// log.Warn().Err(err).Str("wd", mustWD()).Str("file", filename).Msg("error dealing with flag file")
}

func readFile(filename string) string {
	//TODO: perhaps log errors
	stat, err := os.Stat(filename)
	if err != nil {
		logFileError(err, filename)
		return ""
	}
	if stat.Size() > 0 {
		data, err := os.ReadFile(filename)
		if err != nil {
			// logFileError(err, filename)
			return ""
		}
		return strings.Trim(string(data), " \t\r\n")
	}
	return ""
}

func New(options ...Option) (*FlagEnv, error) {
	fe := FlagEnv{
		fileSuffix: "_FILE",
		flagset:    flag.CommandLine,
	}
	for _, opt := range options {
		if opt != nil {
			if err := opt(&fe); err != nil {
				return nil, err
			}
		}
	}
	return &fe, nil
}

func init() {
	var err error
	fe, err = New()
	if err != nil {
		panic(err)
	}

}

func BindFlagset(flagset *flag.FlagSet) error { return fe.BindFlagset(flagset) }
func (fe *FlagEnv) BindFlagset(flagset *flag.FlagSet) error {
	fe.flagset = flagset
	return nil
}

func (fe *FlagEnv) Parse(arguments ...string) {
	fe.flagset.Parse(arguments)
}

func (fe *FlagEnv) asEnvKey(name string) string {
	return fe.envPrefix + strings.ToUpper(name)
}

func (fe *FlagEnv) asEnvFileKey(name string) string {
	return fe.asEnvKey(name) + "_FILE"
}

func SetEnvPrefix(prefix string) { fe.SetEnvPrefix(prefix) }
func (fe *FlagEnv) SetEnvPrefix(prefix string) {
	if !strings.HasSuffix(prefix, "_") {
		prefix = prefix + "_"
	}
	fe.envPrefix = strings.ToUpper(prefix)
}

// getEnvFileRaw tries environment variable then file for string value
func (fe *FlagEnv) getEnvFileRaw(name string) string {
	v := os.Getenv(fe.asEnvKey(name))
	if v == "" {
		filename := os.Getenv(fe.asEnvFileKey(name))
		if filename != "" {
			v = readFile(filename)
		}
	}
	return v
}

func (fe *FlagEnv) get(name string) (value string, defValue string) {
	if flag := fe.flagset.Lookup(name); flag != nil {
		if v := flag.Value.String(); v != "" && v != flag.DefValue {
			return v, flag.DefValue
		}
		return fe.getEnvFileRaw(name), flag.DefValue
	}
	panic(fmt.Errorf("flag named %s does not exist", name))
}

func GetString(name string) string { return fe.GetString(name) }
func (fe *FlagEnv) GetString(name string) string {
	v, dv := fe.get(name)
	if v != "" {
		return v
	}
	return dv
}

func GetBool(name string) bool { return fe.GetBool(name) }
func (fe *FlagEnv) GetBool(name string) bool {
	v, dv := fe.get(name)
	if v != "" {
		return v == "true"
	}
	return dv == "true"
}

func GetDuration(name string) time.Duration { return fe.GetDuration(name) }
func (fe *FlagEnv) GetDuration(name string) time.Duration {
	v, dv := fe.get(name)
	if v == "" {
		v = dv
	}

	dur, err := time.ParseDuration(v)
	if err != nil {
		panic(err)
	}
	return dur
}
