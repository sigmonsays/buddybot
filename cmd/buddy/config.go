package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	goyaml "gopkg.in/yaml.v2"
)

var defaultConfig = `
# begin default built it configuration

log_level: info
log_levels:
  git: info

verbose_log_levels:
  git: trace

buddy_server: :6060
reconnect_delay: 10s
scheme: wss
server_address: localhost
git_watch: true

# end default built it configuration
`

type BuddyConfig struct {
	Hostname         string
	LogLevel         string            `yaml:"log_level"`
	LogLevels        map[string]string `yaml:"log_levels"`
	VerboseLogLevels map[string]string `yaml:"verbose_log_levels"`
	Nick             string
	ServerAddress    string        `yaml:"server_address"`
	Scheme           string        `yaml:"scheme"`
	ReconnectDelay   time.Duration `yaml:"reconnect_delay"`
	BuddyServer      string        `yaml:"buddy_server"`

	GitWatch bool `yaml:"git_watch"`
}

func (c *BuddyConfig) LoadDefault() {
	*c = *GetDefaultConfig()
}

func (c *BuddyConfig) LoadYaml(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(nil)
	_, err = b.ReadFrom(f)
	if err != nil {
		return err
	}

	if err := c.LoadYamlBuffer(b.Bytes()); err != nil {
		return err
	}

	if err := c.FixupConfig(); err != nil {
		return err
	}

	return nil
}

func (c *BuddyConfig) LoadYamlBuffer(buf []byte) error {
	err := goyaml.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *BuddyConfig) PrintYaml() {
	PrintConfig(c)
}

func GetDefaultConfig() *BuddyConfig {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting hostname: %s\n", err)
	}

	conf := &BuddyConfig{
		Hostname: hostname,
	}

	conf.LoadYamlBuffer([]byte(defaultConfig))
	return conf
}

// after loading configuration this gives us a spot to "fix up" any configuration
// or abort the loading process
func (c *BuddyConfig) FixupConfig() error {
	// var emptyConfig BuddyConfig

	return nil
}

func PrintDefaultConfig() {
	conf := GetDefaultConfig()
	PrintConfig(conf)
}

func PrintConfig(conf *BuddyConfig) {
	d, err := goyaml.Marshal(conf)
	if err != nil {
		fmt.Println("Marshal error", err)
		return
	}
	fmt.Println("-- Configuration --")
	fmt.Println(string(d))
}
