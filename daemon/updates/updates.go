package updates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/prefs"
)

const (
	url              = "https://get.torus.sh/manifest.json"
	releaseHourCheck = 6 // Check updates at 6am
)

var (
	timeLayout = time.RFC3339
	releaseDay = time.Wednesday
)

type Engine struct {
	config        *config.Config
	stop          chan struct{}
	lastCheck     time.Time
	targetVersion string
}

type VersionInfo struct {
	Version  string `json:"version"`
	Released string `json:"released"`
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		config: cfg,
		stop:   make(chan struct{}),
	}
}

func (e *Engine) Start() error {
	pref, err := prefs.NewPreferences()
	if err != nil {
		return fmt.Errorf("cannot load preferences")
	}

	if pref.Core.EnableCheckUpdates {
		go e.start()
	}

	return nil
}

func (e *Engine) Stop() error {
	close(e.stop)
	return nil
}

func (e *Engine) start() {
	if err := e.getLastCheck(); err != nil {
		log.Printf("cannot get last update: %s", err)
	}

	log.Printf("last update check: %s", e.lastCheck)
	e.targetVersion = e.config.Version
	for {
		select {
		case <-e.stop:
			log.Printf("stopped checking for updates")
			return
		case <-time.After(e.nextCheck()):
			log.Printf("check updates")

			latest, err := e.getLatestVersion()
			if err != nil {
				log.Printf("cannot check for updates: %s", err)
				continue
			}

			e.targetVersion = latest
			if err := e.storeLastCheck(); err != nil {
				log.Printf("cannot store update check: %s", err)
			}
		}
	}
}

// nextCheck returns the time duration to wait before triggering an update check.
// It is calculated based on wether there already were a check or enough time
// has passed. By default the check is performed at the `releaseHourCheck`-th hour
// of the `releaseDay` weekday.
func (e *Engine) nextCheck() time.Duration {
	day := time.Now().Weekday()
	if e.lastCheck.IsZero() || (day >= releaseDay && time.Since(e.lastCheck).Hours() > releaseHourCheck) {
		return time.Second
	}

	daysDue := int(day) - int(releaseDay)
	if daysDue < 0 {
		daysDue = -daysDue
	} else {
		daysDue = 7 - daysDue
	}
	hoursDue := 24*(daysDue) - time.Now().Hour() + releaseHourCheck

	log.Printf("checking updates in %d hours", hoursDue)
	return time.Duration(hoursDue) * time.Hour
}

func (e *Engine) storeLastCheck() error {
	e.lastCheck = time.Now()
	data := []byte(e.lastCheck.Format(timeLayout))
	return ioutil.WriteFile(e.config.LastUpdatePath, data, 0600)
}

func (e *Engine) getLastCheck() error {
	if _, err := os.Stat(e.config.LastUpdatePath); os.IsNotExist(err) {
		if _, err := os.Create(e.config.LastUpdatePath); err != nil {
			return err
		}
		return nil
	}

	content, err := ioutil.ReadFile(e.config.LastUpdatePath)
	if err != nil {
		return err
	}

	e.lastCheck, err = time.Parse(timeLayout, string(content))
	return err
}

func (e *Engine) getLatestVersion() (string, error) {
	resp, err := http.Get(url)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", fmt.Errorf("cannot get info: %s", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unsuccessful response: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %s", err)
	}
	var info VersionInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return "", fmt.Errorf("cannot decode response body: %s", err)
	}

	return info.Version, nil
}

func (e *Engine) needsUpdate() bool {
	return e.targetVersion > e.config.Version
}

func (e *Engine) VersionInfo() (bool, string) {
	return e.needsUpdate(), e.targetVersion
}
