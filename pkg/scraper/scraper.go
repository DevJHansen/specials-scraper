package scraper

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	firebaseSDK "firebase.google.com/go"
	"github.com/DevJHansen/specials/internal"
	firebaseUtils "github.com/DevJHansen/specials/pkg/firebase"
	"github.com/gocolly/colly"
)

func ScrapeSpecialsNamibiaCategory(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()
	c.OnHTML(category.TabNumber, func(e *colly.HTMLElement) {
		e.ForEach(".col-lg-6.col-md-6.col-12", func(_ int, el *colly.HTMLElement) {
			websiteLink := ""
			el.ForEach("a", func(_ int, a *colly.HTMLElement) {
				if a.Text == "Store Info" {
					websiteLink = a.Attr("href")
				}
			})
			specialDates := el.ChildText("h4")
			title := el.ChildText("h3")
			scrapingID := title + specialDates

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if title != "" && specialDates != "" {
				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				specialsChan <- internal.Special{IsActive: true, Title: title, Category: category.Title, BeenSent: false, ScrapingID: &scrapingID, WebsiteLink: websiteLink, DateAdded: unixTimestamp}
			}
		})
	})
	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", category.Title)
	})
	c.Visit("https://specials.com.na/")
}

func ScrapePickAndPay(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()

	c.OnHTML("div.elementor-posts-container", func(e *colly.HTMLElement) {
		e.ForEach("article.elementor-post", func(_ int, el *colly.HTMLElement) {
			websiteLink := el.ChildAttr("a", "href")
			title := el.ChildText("h2.elementor-heading-title")
			specialDates := el.ChildText("div[data-widget_type='text-editor.default'] > div")
			scrapingID := title + " " + specialDates

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if title != "" && specialDates != "" {
				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				uploadedFile, err := firebaseUtils.UploadFileToFirebase(app, ctx, websiteLink, scrapingID)

				if err != nil {
					fmt.Println(err)
				}

				specialsChan <- internal.Special{
					IsActive:     true,
					Title:        title,
					Category:     "Groceries",
					BeenSent:     false,
					ScrapingID:   &scrapingID,
					WebsiteLink:  websiteLink,
					DateAdded:    unixTimestamp,
					DownloadLink: uploadedFile,
				}
			}
		})
	})

	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", "Pick n Pay")
	})

	c.Visit("https://pnp.na/specials-2/")
}

func ScrapeShoprite(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()

	c.OnHTML("div.cmp-leaflets-specials__leaflets", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, el *colly.HTMLElement) {

			websiteLink := el.ChildAttr("button.cmp-leaflets-specials__leaflets__item__catalogue", "data-leaflet-external-url")
			title := "Shoprite Specials "
			specialDates := el.ChildText("p.cmp-leaflets-specials__leaflets__item__validity")
			previewLink := el.ChildAttr("img.cmp-leaflets-specials__leaflets__item__image", "src")
			scrapingID := title + " " + specialDates + " " + previewLink

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if title != "" && specialDates != "" {
				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				specialsChan <- internal.Special{
					IsActive:    true,
					Title:       title,
					Category:    "Groceries",
					BeenSent:    false,
					ScrapingID:  &scrapingID,
					WebsiteLink: websiteLink,
					DateAdded:   unixTimestamp,
				}
			}
		})
	})

	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", "Shoprite")
	})

	c.Visit("https://www.shoprite.com.na/specials.html?storeId=42&provinceId=175")
}

func ScrapeSparMarua(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()

	c.OnHTML("ul.dropdown-menu", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {

			websiteLink := el.ChildAttr("a", "href")
			title := "Spar Marua "
			specialDates := el.ChildText("a")
			scrapingID := title + " " + specialDates + " " + websiteLink

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if title != "" && specialDates != "" {
				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				specialsChan <- internal.Special{
					IsActive:    true,
					Title:       title,
					Category:    "Groceries",
					BeenSent:    false,
					ScrapingID:  &scrapingID,
					WebsiteLink: websiteLink,
					DateAdded:   unixTimestamp,
				}
			}
		})
	})

	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", "Spar Marua")
	})

	c.Visit("https://www.spar.co.za/Home/Store-View/SUPERSPAR-Maerua-Namibia")
}

func ScrapeCheckers(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()

	c.OnHTML("div.aem-Grid.aem-Grid--12.aem-Grid--default--12.aem-Grid--phone--12 ", func(e *colly.HTMLElement) {
		e.ForEach("div.responsivegrid.aem-GridColumn--default--none.aem-GridColumn--phone--none.aem-GridColumn--phone--12.aem-GridColumn.aem-GridColumn--offset--phone--0.aem-GridColumn--offset--default--1.aem-GridColumn--default--2", func(_ int, el *colly.HTMLElement) {

			websiteLink := el.ChildAttr("a", "href")
			title := "Checkers "
			scrapingID := title + websiteLink

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if websiteLink != "" {
				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				specialsChan <- internal.Special{
					IsActive:    true,
					Title:       title,
					Category:    "Groceries",
					BeenSent:    false,
					ScrapingID:  &scrapingID,
					WebsiteLink: websiteLink,
					DateAdded:   unixTimestamp,
				}
			}
		})
	})

	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", "Checkers")
	})

	c.Visit("https://www.checkers.com.na/promotions.html")
}

func ScrapeOkFoods(c *colly.Collector, category internal.WebsiteCategory, specialsChan chan<- internal.Special, wg *sync.WaitGroup, app *firebaseSDK.App, ctx context.Context) {
	defer wg.Done()

	c.OnHTML("ul.search_results__content.search_results__multi", func(e *colly.HTMLElement) {
		e.ForEach("div.specials-actions-cell", func(_ int, el *colly.HTMLElement) {

			baseURL := "https://www.okfoods.co.za"

			linkElement := el.ChildAttr("button.viewButton.cmp-button", "onclick")
			re := regexp.MustCompile(`window.open\('(.*?)','_blank'\)`)
			matches := re.FindStringSubmatch(linkElement)

			title := "OK Foods"

			currentTime := time.Now()
			unixTimestamp := currentTime.Unix()

			if len(matches) > 1 {
				relativeURL := matches[1]
				fullURL := baseURL + relativeURL
				scrapingID := title + fullURL

				fbSpecial, _ := firebaseUtils.GetSpecialByField(app, ctx, "ScrapingID", scrapingID)

				if fbSpecial.Title == title {
					return
				}

				uploadedFile, err := firebaseUtils.UploadFileToFirebase(app, ctx, fullURL, scrapingID)

				if err != nil {
					fmt.Println(err)
				}

				specialsChan <- internal.Special{
					IsActive:     true,
					Title:        title,
					Category:     "Groceries",
					BeenSent:     false,
					ScrapingID:   &scrapingID,
					WebsiteLink:  fullURL,
					DateAdded:    unixTimestamp,
					DownloadLink: uploadedFile,
				}
			}
		})
	})

	c.OnScraped(func(_ *colly.Response) {
		fmt.Println("Finished scraping:", "OK Foods")
	})

	c.Visit("https://www.okfoods.co.za/content/okfoods/na/en_NA/specials.html")
}
