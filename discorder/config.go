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
)

type Config struct {
	Email             string   `json:"email"`
	AuthToken         string   `json:"authToken"` // Not used currently, but planned
	Theme             string   `json:"theme"`
	AllPrivateMode    bool     `json:"allPrivateMode"`
	LastChannel       string   `json:"lastChannel"`
	ListeningChannels []string `json:"listeningChannels"`
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
