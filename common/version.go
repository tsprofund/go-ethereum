package common

import (
	"path/filepath"
	"os/exec"
	"regexp"
	"runtime"
	"os"
	"strings"
	"os/user"
	"github.com/denisbrodbeck/machineid"
	"strconv"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	SetClientSessionIdentity()
}

type ClientSessionIdentityT struct {
	Revision string `json:"head"`
	Version string `json:"version"`
	Hostname string `json:"host"`
	Username string `json:"user"`
	MachineID string `json:"machineid"`
	Goos string `json:"goos"`
	Goarch string `json:"goarch"`
	Pid int `json:"pid"`
	SessionID string `json:"session"`
}

func (s *ClientSessionIdentityT) String() string {
	sep := func(sep string, args ...string) string {
		var s string
		for i, v := range args {
			if i != len(args)-1 {
				s += v + sep
			} else {
				s += v
			}
		}
		return s
	}
	// Use "_" because it's unlikely to conflict with semver or other data delimiters
	return sep("_", s.Revision, s.Goos, s.Goarch, s.SessionID, s.Hostname, s.Username, s.MachineID, strconv.Itoa(s.Pid))
}

var ClientSessionIdentity *ClientSessionIdentityT

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func SetClientSessionIdentity() {
	var hostname, userName string
	var err error

	hostname, err = os.Hostname()

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}
	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)

	var mid string
	var e error
	mid, e = machineid.ID()
	if e == nil {
		mid, e = machineid.ProtectedID(mid)
	}
	if e != nil {
		mid = hostname + "." + userName
	}

	ClientSessionIdentity = &ClientSessionIdentityT{
		Revision: getGitHeadHash()[:8],
		Hostname: hostname,
		Username: userName,
		MachineID: mid[:8], // because we don't care that much
		Goos: runtime.GOOS,
		Goarch: runtime.GOARCH,
		Pid: os.Getpid(),
		SessionID: randStringBytes(8),
	}
}

func GetClientSessionIdentity() *ClientSessionIdentityT {
	return ClientSessionIdentity
}

func sourceBuildVersion() string {
	// Get the path of this file
	_, f, _, ok := runtime.Caller(1)
	if ok {
		d := filepath.Dir(f) // .../cmd/geth
		// Derive git project dir
		d = filepath.Join(d, "..", ".git")
		// Ignore error
		if o, err := exec.Command("git", "--git-dir", d, "describe", "--tags").Output(); err == nil {
			// Remove newline
			re := regexp.MustCompile(`\r?\n`) // Handle both Windows carriage returns and *nix newlines
			return re.ReplaceAllString(string(o), "")
		}
	}
	return ""
}

func SourceBuildVersionFormatted() string {
	if v := sourceBuildVersion(); v != "" {
		return "source-" + v
	} else {
		return "source-unknown"
	}
}

func getGitHeadHash() string {
	// Get the path of this file
	_, f, _, ok := runtime.Caller(1)
	if ok {
		d := filepath.Dir(f) // .../cmd/geth
		// Derive git project dir
		d = filepath.Join(d, "..", ".git")
		// Ignore error
		if o, err := exec.Command("git", "--git-dir", d, "rev-parse", "HEAD").Output(); err == nil {
			// Remove newline
			re := regexp.MustCompile(`\r?\n`) // Handle both Windows carriage returns and *nix newlines
			return re.ReplaceAllString(string(o), "")
		}
	}
	return ""
}