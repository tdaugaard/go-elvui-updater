package main

import (
	"fmt"
	"net/url"
	"regexp"
)

// NewElvUIUpdater returns a new updater for ElvUI
func NewElvUIUpdater() *WowAddOn {
	var checkVersion = func(addon *WowAddOn) error {
		pageURL, err := url.Parse(addon.PageURL)
		if err != nil {
			return err
		}

		doc, err := getDocument(addon.PageURL)
		if err != nil {
			return err
		}

		relLink, ok := doc.Find("a[href^='/downloads/elvui-']").Attr("href")
		if !ok {
			return fmt.Errorf("could not find the HTML element that contains the update URL")
		}

		url, err := url.Parse(relLink)
		if err != nil {
			return err
		}

		addon.ZipURL = pageURL.ResolveReference(url).String()

		rxVersion := regexp.MustCompile("-(\\d+\\.\\d+)\\.zip")
		if !rxVersion.MatchString(relLink) {
			return fmt.Errorf("cannot find a valid version in the Zip-filename (%s)", relLink)
		}

		version := rxVersion.FindStringSubmatch(relLink)[1]
		ver, err := addon.parseSemVer(version)
		if err != nil {
			return err
		}

		addon.Version.Latest = ver

		return nil
	}

	return &WowAddOn{
		Name:         "ElvUI",
		PageURL:      "https://www.tukui.org/welcome.php",
		CheckVersion: checkVersion,
	}
}
