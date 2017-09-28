package updates

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/manifoldco/torus-cli/config"
)

var fakeUrl = "http://hellothisdomaincantexistright:54432/test.json"

func makeEngine(uri string) *Engine {
	u, err := url.Parse(uri)
	if err != nil {
		panic("could not parse url")
	}

	return &Engine{
		config: &config.Config{
			ManifestURI: u,
		},
	}
}

type nextCheckExample struct {
	now       time.Time
	lastCheck time.Time
	expected  time.Duration
}

type mockTimeManager struct {
	now time.Time
}

func (m mockTimeManager) Now() time.Time {
	return m.now
}

func TestNextCheck(t *testing.T) {
	wednesday := time.Time{}
	daysDiff := int(releaseDay) - int(wednesday.Weekday())
	wednesday = wednesday.Add(time.Duration(daysDiff*24-wednesday.Hour()) * time.Hour)

	sunday := wednesday.Add(-24 * 3 * time.Hour)
	tuesday := wednesday.Add(-24 * time.Hour)
	thursday := wednesday.Add(24 * time.Hour)

	examples := []nextCheckExample{
		{
			lastCheck: time.Time{},
			expected:  minCheckDuration,
		},

		{
			now:       tuesday,
			lastCheck: tuesday,
			expected:  30 * time.Hour,
		},

		{
			now:       sunday,
			lastCheck: wednesday.Add(time.Duration(-24*7+releaseHourCheck) * time.Hour),
			expected:  time.Duration(24*3*time.Hour + releaseHourCheck*time.Hour),
		},

		{
			now:       tuesday,
			lastCheck: wednesday.Add(-24 * 7 * time.Hour),
			expected:  minCheckDuration,
		},

		{
			now:       thursday,
			lastCheck: wednesday.Add(releaseHourCheck * time.Hour),
			expected:  time.Duration(24*6+releaseHourCheck) * time.Hour,
		},

		{
			now:       wednesday,
			lastCheck: wednesday.Add(-24 * 7 * time.Hour),
			expected:  minCheckDuration,
		},

		{
			now:       wednesday,
			lastCheck: wednesday,
			expected:  time.Duration(releaseHourCheck * time.Hour),
		},

		{
			now:       wednesday.Add((releaseHourCheck + 1) * time.Hour),
			lastCheck: wednesday.Add(releaseHourCheck * time.Hour),
			expected:  (24*7 - 1) * time.Hour,
		},
	}

	for _, example := range examples {
		timeManager := mockTimeManager{
			now: example.now,
		}

		e := &Engine{
			timeManager: timeManager,
		}

		e.lastCheck = example.lastCheck
		if d := e.nextCheck(); d != example.expected {
			t.Fatalf("expected %s, got %s", example.expected, d)
		}
	}
}

func TestGetLatestCheck(t *testing.T) {
	t.Run("200 valid json response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"version":"0.24.1","released":"2017-09-25T04:06:19-03:00"}`))
		}))
		defer srv.Close()

		e := makeEngine(srv.URL)

		v, err := e.getLatestVersion()
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		if v != "0.24.1" {
			t.Errorf("expected `0.24.1` got %s", v)
		}
	})

	t.Run("200 invalid json response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"version":"0.24.1","released":"2017-09-25T04:06:19-03:00`))
		}))
		defer srv.Close()

		e := makeEngine(srv.URL)

		_, err := e.getLatestVersion()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "cannot decode response body") {
			t.Errorf("received different error than expected: %s", err)
		}
	})

	t.Run("200 with no body response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte{})
		}))
		defer srv.Close()

		e := makeEngine(srv.URL)

		_, err := e.getLatestVersion()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "cannot decode response body") {
			t.Errorf("received different error than expected: %s", err)
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte{})
		}))
		defer srv.Close()

		e := makeEngine(srv.URL)

		_, err := e.getLatestVersion()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "unsuccessful response") {
			t.Errorf("received different error than expected: %s", err)
		}
	})

	t.Run("cannot reach server", func(t *testing.T) {
		e := makeEngine(fakeUrl)

		_, err := e.getLatestVersion()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "cannot get info from") {
			t.Errorf("received different error than expected: %s", err)
		}
	})
}
