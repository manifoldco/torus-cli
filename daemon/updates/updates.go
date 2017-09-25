package updates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/blang/semver"
	"github.com/manifoldco/torus-cli/config"
)

const (
	url              = "https://get.torus.sh/manifest.json"
	releaseHourCheck = 6 // Check updates at 6am
)

var (
	timeLayout       = time.RFC3339
	releaseDay       = time.Wednesday
	minCheckDuration = time.Second
)

// TimeManager is a proxy interface with a Now() method which returns
// the current time with the appropriate conventions.
type TimeManager interface {
	Now() time.Time
}

type defaultTimeManager struct{}

func (m defaultTimeManager) Now() time.Time {
	return time.Now().UTC()
}

// Engine is the deamon updates check engine which fetches updates status from Torus
// repository. It can be asked for the latest version available upstream and if it's
// higher than the current one.
type Engine struct {
	config        *config.Config
	stop          chan struct{}
	lastCheck     time.Time
	targetVersion string
	timeManager   TimeManager
}

// VersionInfo maps the JSON returned from the `url` endpoint, containing the latest
// version available info.
type VersionInfo struct {
	Version  string `json:"version"`
	Released string `json:"released"`
}

// NewEngine creates a new Engine based on the provided config structure.
// It can be extended by passing specific options which alter the initialization
// of the Engine itself.
func NewEngine(cfg *config.Config, options ...func(*Engine)) *Engine {
	engine := &Engine{
		config: cfg,
		stop:   make(chan struct{}),
	}

	for _, opt := range options {
		opt(engine)
	}

	if engine.timeManager == nil {
		engine.timeManager = defaultTimeManager{}
	}

	return engine
}

// Start starts the update checking loop.
func (e *Engine) Start() error {
	// Disabled until torus dns outage is resolved
	return nil
}

// Stop stops the update checking loop.
func (e *Engine) Stop() error {
	close(e.stop)
	return nil
}

// VersionInfo returns a boolean representing if the current version
// is behind the latest one available for download and the latest version available.
func (e *Engine) VersionInfo() (bool, string) {
	return e.needsUpdate(), e.targetVersion
}

// SetTimeManager is a configuration function for `NewEngine` which sets the
// Engine's TimeManager instance to the `manager` argument.
func (e *Engine) SetTimeManager(manager TimeManager) func(*Engine) {
	return func(engine *Engine) {
		engine.timeManager = manager
	}
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
	if !e.lastCheckValid() {
		return minCheckDuration
	}
	return e.hoursToNextRelease()
}

// lastCheckValid checks if the last update check contains the most recent
// update info.
func (e *Engine) lastCheckValid() bool {
	if e.lastCheck.IsZero() {
		return false
	}
	return e.lastCheck.Unix()-e.prevReleaseDay().Unix() >= 0
}

// prevReleaseDay returns the date of the last release day, calculated based on
// the `releaseHourCheck`-th hour of the previous `releaseDay` weekday.
func (e *Engine) prevReleaseDay() time.Time {
	prevRelease := e.midnight(e.timeManager.Now())
	day := prevRelease.Weekday()
	var dayHours int
	if day < releaseDay || (day == releaseDay && prevRelease.Hour() < releaseHourCheck) {
		dayHours = 24 * (7 - int(releaseDay-day))
	} else if day > releaseDay || (day == releaseDay && prevRelease.Hour() > releaseHourCheck) {
		dayHours = 24 * int(day-releaseDay)
	}
	hours := releaseHourCheck - dayHours - prevRelease.Hour()
	prevRelease = prevRelease.Add(time.Duration(hours) * time.Hour)
	return prevRelease
}

// hoursToNextRelease returns the time delta before the next update check.
func (e *Engine) hoursToNextRelease() time.Duration {
	return e.nextReleaseDay().Sub(e.timeManager.Now())
}

// nextReleaseDay returns the date of the next release day, calculated based on
// the `releaseHourCheck`-th hour of the next `releaseDay` weekday
func (e *Engine) nextReleaseDay() time.Time {
	now := e.timeManager.Now()
	nextRelease := e.midnight(now).Add(releaseHourCheck * time.Hour)
	if nextRelease.Weekday() == releaseDay {
		if nextRelease.Unix() <= now.Unix() {
			nextRelease = nextRelease.Add(24 * 7 * time.Hour)
		}
		return nextRelease
	}

	daysDue := int(nextRelease.Weekday()) - int(releaseDay)
	if daysDue <= 0 {
		daysDue = -daysDue
	} else {
		daysDue = 7 - daysDue
	}

	return nextRelease.Add(time.Duration(daysDue*24) * time.Hour)
}

// midnight returns the date of the midnight of the provided time.
func (e Engine) midnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func (e *Engine) storeLastCheck() error {
	e.lastCheck = e.timeManager.Now()
	data := []byte(e.lastCheck.Format(timeLayout))
	return ioutil.WriteFile(e.config.LastUpdatePath, data, 0600)
}

func (e *Engine) getLastCheck() error {
	if _, err := os.Stat(e.config.LastUpdatePath); os.IsNotExist(err) {
		_, err = os.Create(e.config.LastUpdatePath)
		return err
	}

	content, err := ioutil.ReadFile(e.config.LastUpdatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
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
	// must update if there's no current version stored or it's malformed
	current, err := semver.Make(e.config.Version)
	if err != nil {
		return true
	}

	// refuse to update to an unknown/invalid target version
	target, err := semver.Make(e.targetVersion)
	if err != nil {
		return false
	}

	return target.Compare(current) > 0
}
