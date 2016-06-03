package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/jawher/mow.cli"
)

const version = "0.1.0"

type result struct {
	name    string
	product string
	status  string
}

func getURI(orderID string, lastName string, outletID string) []string {
	uri := "https://www.mediamarkt.se/webapp/wcs/stores/servlet/MultiChannelMARepairStatusResult"

	form := url.Values{}
	form.Set("storeId", "14401")
	form.Add("langId", "-16")
	form.Add("orderId", orderID)
	form.Add("lastname", lastName)
	form.Add("outletId", outletID)
	req, err := http.NewRequest("POST", uri, strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	var res []string
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return res
		case tt == html.StartTagToken:
			token := z.Token()

			isAnchor := token.Data == "dd"
			if isAnchor {
				z.Next() // Go to the actual text inside the tag
				f := z.Token()
				res = append(res, f.Data)
			}
		}
	}
}

func printStatus(status result) {
	fmt.Printf("Name: %s\n", status.name)
	fmt.Printf("Product: %s\n", status.product)
	fmt.Printf("Status: %s\n", status.status)
}

func main() {
	app := cli.App("mm-service-status", "Get status for MediaMarkt Service Order (Sweden)")

	app.Spec = "(--lastname=<lastname> --order-id=<order ID> --store-id=<store ID>) | (--list-stores | --version)"

	var (
		lastName     = app.StringOpt("l lastname", "", "Last Name on order")
		orderID      = app.StringOpt("o order-id", "", "Order ID, last 6 digits")
		storeID      = app.StringOpt("s store-id", "", "Store ID")
		listStores   = app.BoolOpt("list-stores", false, "List available stores")
		versionCheck = app.BoolOpt("v version", false, "Show current version")
	)

	app.Action = func() {
		if *versionCheck {
			fmt.Println(os.Args[0], version)
		} else if *listStores {
			fmt.Println("Stores:")
		} else {
			fmt.Printf("Checking service status for order: %s name: %s store: %s\n", *orderID, *lastName, *storeID)
			res := getURI(*orderID, *lastName, *storeID)
			status := result{
				name:    res[0],
				product: res[1],
				status:  res[2],
			}
			printStatus(status)
		}
	}

	app.Run(os.Args)
}
