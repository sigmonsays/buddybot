package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sigmonsays/git-watch/watch/git"
)

type upgrader struct {
	gopath string
	go_pkg string
}

func GitWatch(conf *BuddyConfig) {
	code_directory := "."
	branch := "master"
	interval := 60
	gopath := os.Getenv("GOPATH")

	gw := git.NewGitWatch(code_directory, branch)
	gw.Interval = interval
	gw.Dir = gopath

	err := gw.Start()
	if err != nil {
		log.Warnf("GitWatch: %s", err)
		return
	}

	upgrader := &upgrader{
		gopath: gopath,
		go_pkg: "github.com/sigmonsays/buddybot/...",
	}

	gw.OnChange = upgrader.OnChange
	gw.OnCheck = upgrader.OnCheck

}
func (me *upgrader) OnCheck(dir, branch, lhash, rhash string) error {
	return nil
}

func (me *upgrader) OnChange(dir, branch, lhash, rhash string) error {
	old_version, _ := me.GetVersion()

	log.Debugf("dir=%s branch=%s local-hash=%s remote-hash=%s", dir, branch, lhash, rhash)
	if lhash == "" || rhash == "" {
		log.Debugf("Update aborted due to empty rhash or lhash")
		return nil
	}
	var cmdline []string

	// do a git pull
	cmdline = []string{"git", "pull"}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = me.gopath
	err := cmd.Start()
	if err != nil {
		log.Warnf("upgarde error: git pull: %s", err)
		return nil
	}

	// do the upgrade (this will pull it any dependencies)
	cmdline = []string{"go", "get", "-u", me.go_pkg}
	cmd = exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = me.gopath
	err = cmd.Start()
	if err != nil {
		log.Warnf("upgrade error: go get %s", err)
		return nil
	}

	// do the install
	cmdline = []string{"go", "install", me.go_pkg}
	cmd = exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = me.gopath
	err = cmd.Start()
	if err != nil {
		log.Warnf("upgrade error: go install %s", err)
		return nil
	}

	new_version, _ := me.GetVersion()
	log.Infof("Upgraded version from %s to %s", old_version, new_version)

	return nil
}

func (me *upgrader) GetVersion() (string, error) {
	cmdline := []string{"git", "describe", "--tags"}
	out, err := exec.Command(cmdline[0], cmdline[1:]...).Output()
	if err != nil {
		return "", err
	}
	version := strings.Trim(string(out), "\n")
	return version, nil
}
