package discorder

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Email     string `json:"email"`
	AuthToken string `json:"auth_token"` // Not used currently, but planned
	Theme     string `json:"theme"`

	Tabs []*TabConfig `json:"tabs"`

	TimeFormatSameDay string `json:"time_format_same_day"`
	TimeFormatFull    string `json:"time_format_full"`

	// General settings
	ShortGuilds   bool `json:"short_guilds"`
	HideNicknames bool `json:"hide_nicknames"`
	// The blow gives guilds, channels and users dtereministic "random" colors from the
	// Active theme's discrim table
	ColoredGuilds   bool `json:"colored_guilds"`
	ColoredChannels bool `json:"colored_channels"`
	ColoredUsers    bool `json:"colored_users"`
}

type TabConfig struct {
	Name              string   `json:"name"`
	AllPrivateMode    bool     `json:"all_private_mode"`
	Index             int      `json:"index"`
	SendChannel       string   `json:"send_channel"`
	ListeningChannels []string `json:"listening_cannels"`
}

const (
	DefaultTimeFormatSameDay = "15:04:05"
	DefaultTimeFormatFull    = time.Stamp
)

func (c *Config) GetTimeFormatSameDay() string {
	if c.TimeFormatSameDay == "" {
		return DefaultTimeFormatSameDay
	}
	return c.TimeFormatSameDay
}

func (c *Config) GetTimeFormatFull() string {
	if c.TimeFormatFull == "" {
		return DefaultTimeFormatFull
	}
	return c.TimeFormatFull
}

func LoadOrCreateConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err == nil {
		var decoded Config
		err = json.Unmarshal(file, &decoded)
		return &decoded, err
	}

	log.Println("Failed loading config, creating new one")
	c := &Config{}
	err = c.Save(path)
	return c, err
}

func (c *Config) Save(path string) error {
	eencoded, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, eencoded, os.FileMode(0755))
}

func LoadTheme(themePath string) *Theme {
	if themePath == "" {
		return nil
	}

	file, err := ioutil.ReadFile(themePath)
	if err != nil {
		log.Println("Failed loading theme", themePath, ":", err)
		return nil
	}

	var theme Theme
	err = json.Unmarshal(file, &theme)
	if err != nil {
		log.Println("Failed loading theme", themePath, ":", err)
		return nil
	}
	log.Println("Loaded theme", theme.Name, "By", theme.Author)
	return &theme
}

func GetCreateConfigDir() (result string, err error) {
	if runtime.GOOS == "windows" {
		result, err = dirWindows()
	} else {
		// Unix-like system, so just assume Unix
		result, err = dirUnix()
	}

	if err != nil {
		return
	}

	_, err = os.Stat(result)
	if os.IsNotExist(err) {
		err = os.MkdirAll(result, 0755)
		log.Println("Couldn't find config dir, creating:", result)
	} else {
		return
	}

	return result, nil
}

// Expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func ExpandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}

	dir, err := GetCreateConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, path[1:]), nil
}

func dirUnix() (string, error) {
	home, err := homeDirUnix()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, ".config", "discorder")
	return configPath, nil
}

func homeDirUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try getent
	var stdout bytes.Buffer
	cmd := exec.Command("getent", "passwd", strconv.Itoa(os.Getuid()))
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		// If "getent" is missing, ignore it
		if err == exec.ErrNotFound {
			return "", err
		}
	} else {
		if passwd := strings.TrimSpace(stdout.String()); passwd != "" {
			// username:password:uid:gid:gecos:home:shell
			passwdParts := strings.SplitN(passwd, ":", 7)
			if len(passwdParts) > 5 {
				return passwdParts[5], nil
			}
		}
	}

	// If all else fails, try the shell
	stdout.Reset()
	cmd = exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func dirWindows() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return "", errors.New("No appdata in path")
	}
	return filepath.Join(appdata, "discorder"), nil
}
