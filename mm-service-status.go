package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/jawher/mow.cli"
)

const version = "0.2.0"

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

func getStores() map[string]int {
	uri := "https://www.mediamarkt.se/webapp/wcs/stores/servlet/MultiChannelMARepairStatus"

	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	res := make(map[string]int)
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return res
		case tt == html.StartTagToken:
			token := z.Token()

			isAnchor := token.Data == "option"
			if isAnchor {
				// z.Next() // Go to the actual text inside the tag
				for _, a := range token.Attr {
					value, err := strconv.Atoi(a.Val)
					if err == nil && a.Key == "value" {
						// If the Key is 'value' and the Value is an Int, it's likely a store
						z.Next()
						token = z.Token()
						res[token.Data] = value
					}
				}
			}
		}
	}
}

func printStatus(status result) {
	fmt.Printf("Name: %s\n", status.name)
	fmt.Printf("Product: %s\n", status.product)
	fmt.Printf("Status: %s\n", status.status)
}

func printStoreIDs(stores map[string]int) {
	fmt.Printf("%*s: %s\n", 30, "Store Name", "Store ID")
	fmt.Printf("%s\n", strings.Repeat("=", 50))
	for store, id := range stores {
		fmt.Printf("%*s: %d\n", 30, store, id)
	}
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
			fmt.Println(path.Base(os.Args[0]), version)
		} else if *listStores {
			stores := getStores()
			if len(stores) == 0 {
				fmt.Println("Empty result")
				os.Exit(1)
			} else {
				printStoreIDs(stores)
			}
		} else {
			fmt.Printf("Checking service status for order: %s name: %s store: %s\n", *orderID, *lastName, *storeID)
			res := getURI(*orderID, *lastName, *storeID)
			if len(res) != 3 {
				fmt.Println("Empty result")
				os.Exit(1)
			} else {
				status := result{
					name:    res[0],
					product: res[1],
					status:  res[2],
				}
				printStatus(status)
			}
		}
	}

	app.Run(os.Args)
}
