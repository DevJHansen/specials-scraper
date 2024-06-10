package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DevJHansen/specials/internal"
	"github.com/DevJHansen/specials/pkg/firebase"
	"github.com/DevJHansen/specials/pkg/scraper"
	"github.com/gocolly/colly"
)

func main() {
	ctx := context.Background()
	app, _ := firebase.NewFirebaseApp(ctx)
	specialsScrapingChan := make(chan internal.Special, 100)
	scrapedSpecials := make([]internal.Special, 0)
	var wg sync.WaitGroup

	go func() {
		for special := range specialsScrapingChan {
			scrapedSpecials = append(scrapedSpecials, special)
		}
	}()

	for _, category := range internal.WebsiteCategories {
		wg.Add(1)
		collector := colly.NewCollector()
		go scraper.ScrapeSpecialsNamibiaCategory(collector, category, specialsScrapingChan, &wg, app, ctx)
	}

	wg.Add(1)
	go scraper.ScrapePickAndPay(colly.NewCollector(), internal.WebsiteCategories[0], specialsScrapingChan, &wg, app, ctx)

	wg.Add(1)
	go scraper.ScrapeShoprite(colly.NewCollector(), internal.WebsiteCategories[0], specialsScrapingChan, &wg, app, ctx)

	wg.Add(1)
	go scraper.ScrapeSparMarua(colly.NewCollector(), internal.WebsiteCategories[0], specialsScrapingChan, &wg, app, ctx)

	wg.Add(1)
	go scraper.ScrapeCheckers(colly.NewCollector(), internal.WebsiteCategories[0], specialsScrapingChan, &wg, app, ctx)

	wg.Add(1)
	go scraper.ScrapeOkFoods(colly.NewCollector(), internal.WebsiteCategories[0], specialsScrapingChan, &wg, app, ctx)

	wg.Wait()
	close(specialsScrapingChan)

	jsonItems, _ := json.MarshalIndent(scrapedSpecials, "", "    ")
	fmt.Println("Collected Items:")
	fmt.Println(string(jsonItems))

	if app != nil {
		for _, special := range scrapedSpecials {
			wg.Add(1)

			go func(sp internal.Special) {
				defer wg.Done()

				err := firebase.AddSpecial(app, context.Background(), sp)
				if err != nil {
					fmt.Println(err)
				}
			}(special)
		}
	}

	wg.Wait()
}
