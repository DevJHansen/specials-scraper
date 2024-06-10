package utils

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func DownloadFile(partialURL string) error {
	fmt.Println("Downloading file from:", partialURL)
	l := launcher.New().
		Headless(false).
		Devtools(true).
		Set("disable-notifications").
		Set("disable-popup-blocking").
		Set("disable-infobars")

	defer l.Cleanup()

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url).
		MustConnect()

	defer browser.MustClose()

	page := browser.MustPage("https://specials.com.na").MustWaitStable()

	page.MustWaitLoad()

	link := page.MustElementR("div", ".sp-form-234574")

	link.MustWaitVisible()

	// Scroll to the element (optional, improves visibility in case of dynamic content)
	link.MustScrollIntoView()

	// Click on the element
	link.MustClick()

	// Optionally, wait for a few seconds to observe the click effect
	time.Sleep(3 * time.Second)

	return nil
}
