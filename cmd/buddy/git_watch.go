package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sigmonsays/git-watch/watch/git"
)

type upgrader struct {
	gopath      string
	upgrade_dir string
	go_pkg      string
}

func GitWatch(conf *BuddyConfig) {
	code_directory := "."
	branch := "master"
	interval := 30
	gopath := os.Getenv("GOPATH")
	upgrade_dir := filepath.Join(gopath, "src/github.com/sigmonsays/buddybot")

	gw := git.NewGitWatch(code_directory, branch)
	gw.Interval = interval
	gw.Dir = upgrade_dir

	err := gw.Start()
	if err != nil {
		log.Warnf("GitWatch: %s", err)
		return
	}

	upgrader := &upgrader{
		gopath:      gopath,
		upgrade_dir: upgrade_dir,
		go_pkg:      "github.com/sigmonsays/buddybot/...",
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
	var err error

	// do a git pull
	cmdline = []string{"git", "pull"}
	err = me.doCommand(cmdline)
	if err != nil {
		log.Warnf("%s", err)
		return nil
	}

	// do the upgrade (this will pull it any dependencies)
	cmdline = []string{"go", "get", "-u", me.go_pkg}
	err = me.doCommand(cmdline)
	if err != nil {
		log.Warnf("%s", err)
		return nil
	}

	// do the install
	cmdline = []string{"go", "install", me.go_pkg}
	err = me.doCommand(cmdline)
	if err != nil {
		log.Warnf("%s", err)
		return nil
	}

	new_version, _ := me.GetVersion()
	log.Infof("Upgraded version from %s to %s", old_version, new_version)

	return nil
}

func (me *upgrader) doCommand(cmdline []string) error {
	// do the install
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = me.upgrade_dir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("upgrade error: %s: %s", cmdline, err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("upgrade error: %s: %s", cmdline, err)
	}
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
