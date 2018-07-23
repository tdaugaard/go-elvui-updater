package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sqweek/dialog"

	"github.com/PuerkitoBio/goquery"
)

type mapStringsOfStrings map[string]string

// Updater ...
type Updater struct {
	addOnFolder   string
	forceUpdate   bool
	installAddOns bool
	updaters      []*WowAddOn
	ch            chan bool
	wg            sync.WaitGroup
}

// NewUpdater ...
func NewUpdater(appFolder string) (*Updater, error) {
	addOnFolder := path.Join(appFolder, "Interface", "AddOns")

	_, err := os.Stat(addOnFolder)
	if err != nil {
		return &Updater{}, err
	}

	return &Updater{
		addOnFolder: addOnFolder,
		ch:          make(chan bool, 1),
	}, nil
}

// SetInstallAddOns sets whether add-ons should be installed
// if they are not already.
func (c *Updater) SetInstallAddOns(flag bool) {
	c.installAddOns = flag
}

func (c *Updater) executeUpdater(addon *WowAddOn, status chan string) {
	var zipFile string

	defer c.wg.Done()

	installed, err := addon.Installed()
	if err != nil {
		log.Printf(
			"[%s] Error: %s\n",
			addon.Name,
			err.Error(),
		)

		return
	}

	if !installed {
		if !c.installAddOns {
			log.Printf(
				"[%s] Add-on not installed.\n",
				addon.Name,
			)

			return
		}

		log.Printf(
			"[%s] Installing add-on as requested.\n",
			addon.Name,
		)
	} else {
		err = addon.FindInstalledVersion()
		if err != nil {
			log.Printf(
				"[%s] Error: %s\n",
				addon.Name,
				err.Error(),
			)
		}
	}

	log.Printf(
		"[%s] Checking for latest version.\n",
		addon.Name,
	)

	err = addon.CheckVersion(addon)
	if err != nil {
		log.Printf(
			"[%s] Error: %s\n",
			addon.Name,
			err.Error(),
		)
	}

	upToDate := addon.Version.Latest.EQ(addon.Version.Current)
	if upToDate {
		if !c.forceUpdate {
			log.Printf(
				"[%s] Up to date (Installed: %s).\n",
				addon.Name,
				addon.Version.Current,
			)

			status <- fmt.Sprintf("%s v%s is up to date.", addon.Name, addon.Version.Current)

			return
		}

		log.Printf(
			"[%s] Forcing re-install of installed version: %s\n",
			addon.Name,
			addon.Version.Current,
		)
	} else {
		log.Printf(
			"[%s] Found new version: %s (Installed: %s)\n",
			addon.Name,
			addon.Version.Latest,
			addon.Version.Current,
		)
	}

	zipFile, err = c.downloadZip(addon)
	if err != nil {
		log.Printf(
			"[%s] Error: %s\n",
			addon.Name,
			err.Error(),
		)
	}

	defer func() {
		err := os.Remove(zipFile)
		if err != nil {
			log.Printf(
				"[%s] Error: %s\n",
				addon.Name,
				err.Error(),
			)
		}
	}()

	log.Printf(
		"[%s] Extracting Zip-file to %s\n",
		addon.Name,
		addon.Path,
	)

	err = c.extractZip(addon, zipFile)
	if err != nil {
		log.Printf(
			"[%s] Error: %s\n",
			addon.Name,
			err.Error(),
		)
	}

	if installed {
		if upToDate {
			status <- fmt.Sprintf("%s v%s has been re-installed.", addon.Name, addon.Version.Current)
		} else {
			status <- fmt.Sprintf("%s v%s has been updated from v%s.", addon.Name, addon.Version.Latest, addon.Version.Current)
		}
	} else {
		status <- fmt.Sprintf("%s v%s has been updated/installed.", addon.Name, addon.Version.Current)
	}
}

func (c *Updater) downloadZip(addon *WowAddOn) (string, error) {
	tmpFile, err := ioutil.TempFile("", "addon")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(addon.ZipURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func (c *Updater) extractZip(addon *WowAddOn, zipFile string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(c.addOnFolder, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close() // Close the file without defer to close before next iteration of loop

			if err != nil {
				return err
			}

		}
	}

	return nil
}

// Check if any add-ons need updating
func (c *Updater) Check() {
	status := make(chan string, 100)

	c.wg.Add(len(c.updaters))
	for _, updater := range c.updaters {
		go c.executeUpdater(updater, status)
	}
	c.wg.Wait()
	close(status)

	var statusStrings []string

	for str := range status {
		statusStrings = append(statusStrings, str)
	}

	dialog.
		Message(strings.Join(statusStrings, "\n")).
		Title("Update Status").
		Info()
}

// SetForceUpdate ....
func (c *Updater) SetForceUpdate(flag bool) {
	c.forceUpdate = flag
}

// Add ...
func (c *Updater) Add(addon *WowAddOn) {
	addon.Path = path.Join(c.addOnFolder, addon.Name)

	c.updaters = append(c.updaters, addon)
}

func getDocument(url string) (doc *goquery.Document, err error) {
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return
}
