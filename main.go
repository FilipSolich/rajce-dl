package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	url := os.Args[1]
	scrollingIterations := 100
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &scrollingIterations)
	}

	dirParts := strings.Split(url, "/")
	dir := dirParts[len(dirParts)-1]
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("Failed to create output directory:", err)
		return
	}

	u := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(url)
	page.MustWaitLoad()
	page.MustWaitIdle()

	for range scrollingIterations {
		page.Mouse.MustScroll(0, 100)
		page.MustWaitIdle()
	}

	images := page.MustElements("img")
	var total int
	for _, img := range images {
		src := img.MustAttribute("src")
		if src == nil || *src == "" {
			continue
		}
		if !strings.Contains(*src, "thumb") {
			continue
		}

		url := strings.ReplaceAll(*src, "thumb", "images")
		fmt.Println("Found:", url)
		total++

		err := downloadImage(url, dir)
		if err != nil {
			fmt.Println("Failed:", err)
		}
	}
	fmt.Println("Total images found:", total)
}

func downloadImage(url, dir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	nameParts := strings.Split(url, "/")
	name := nameParts[len(nameParts)-1]
	filename := filepath.Join(dir, name)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
