package main

import (
	"fmt"
	"net/url"
)

// NewElvUAddOnSkinsIUpdater returns a new updater for ElvUI
func NewElvUAddOnSkinsIUpdater() *WowAddOn {
	var checkVersion = func(addon *WowAddOn) error {
		pageURL, err := url.Parse(addon.PageURL)
		if err != nil {
			return err
		}

		doc, err := getDocument(addon.PageURL)
		if err != nil {
			return err
		}

		relLink, ok := doc.Find("a[href^='addons.php?download=']").Attr("href")
		if !ok {
			return fmt.Errorf("could not find the HTML element that contains the update URL")
		}

		url, err := url.Parse(relLink)
		if err != nil {
			return err
		}

		addon.ZipURL = pageURL.ResolveReference(url).String()

		version := doc.Find("p.extras:first-child > b:first-child").Text()
		ver, err := addon.parseSemVer(version)
		if err != nil {
			return err
		}

		addon.Version.Latest = ver

		return nil
	}

	return &WowAddOn{
		Name:         "AddOnSkins",
		PageURL:      "https://www.tukui.org/addons.php?id=3",
		CheckVersion: checkVersion,
	}
}
