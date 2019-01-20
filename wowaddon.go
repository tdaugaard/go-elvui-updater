package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

// WowAddOnVersionChecker defines the function signature for a version checker
type WowAddOnVersionChecker func(addon *WowAddOn) error

// AddOnVersion defines the add-on versions
type AddOnVersion struct {
	Current semver.Version
	Latest  semver.Version
}

// WowAddOn defines a single Add-On
type WowAddOn struct {
	Name         string
	Path         string
	PageURL      string
	ZipURL       string
	Version      AddOnVersion
	CheckVersion WowAddOnVersionChecker
}

// Installed determines whether an add-on is already installed
func (c *WowAddOn) Installed() (bool, error) {
	if c.Path == "" {
		return false, fmt.Errorf("no path set for add-on")
	}

	tocFile := path.Join(c.Path, c.Name) + ".toc"
	stat, err := os.Stat(tocFile)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !stat.IsDir(), nil
}

// FindInstalledVersion finds the installed version of the add-on
func (c *WowAddOn) FindInstalledVersion() error {
	toc, err := c.readTOC()
	if err != nil {
		return err
	}

	var version string

	version, exists := toc["Version"]
	if !exists {
		return fmt.Errorf("cannot find version information in TOC")
	}

	currVer, err := c.parseSemVer(version)
	if err != nil {
		return err
	}

	c.Version.Current = currVer
	return nil
}

func (c *WowAddOn) parseSemVer(str string) (semver.Version, error) {
	var re = regexp.MustCompile(`\.0+([1-9]+)`)
	str = re.ReplaceAllString(str, ".$1")

	return semver.ParseTolerant(str)
}

func (c *WowAddOn) readTOC() (mapStringsOfStrings, error) {
	tocFile := path.Join(c.Path, c.Name) + ".toc"

	fh, err := os.Open(tocFile)
	if err != nil {
		return mapStringsOfStrings{}, err
	}

	delim := "## "
	toc := mapStringsOfStrings{}
	r := bufio.NewReader(fh)

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		index := strings.Index(line, delim)
		if index != 0 {
			continue
		}

		line = line[len(delim):]
		parts := strings.SplitN(line, ":", 2)

		toc[parts[0]] = strings.TrimSpace(parts[1])
	}

	return toc, nil
}
