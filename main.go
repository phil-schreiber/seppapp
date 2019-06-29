package main

import (
	"database/sql"
	"html/template"
	"log"
	"math"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// OilUsageItem : dataType
type OilUsageItem struct {
	UID      int
	Tstamp   int
	Crdate   int
	Year     int
	Quantity int
	Price    int
}

// OilUsage : dataType
type OilUsage struct {
	AverageUsage  int
	AverageCost   int
	OilUsageItems []OilUsageItem
}

// OilUsageItemsPerYear : dataType
type OilUsageItemsPerYear struct {
	Year          int
	OilUsageItems []OilUsageItem
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", Index)
	http.HandleFunc("/data", Data)
	http.HandleFunc("/input", Input)

	log.Println("Listening...")
	http.ListenAndServe(":8080", nil)
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "seppapp"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(localhost:3306)/"+dbName)

	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("templates/*"))

// Index : startpage
func Index(w http.ResponseWriter, r *http.Request) {
	var q struct{}
	tmpl.ExecuteTemplate(w, "Index", q)
}

// Data : get data for data view
func Data(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Data", getDataGrouped())
}

// Input : get data for data view
func Input(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Input", getData())
}

func getData() []OilUsageItemsPerYear {
	db := dbConn()
	selDb, err := db.Query("SELECT uid, tstamp, crdate, year, quantity, price FROM tx_psverbrauchsberechnung_item ORDER BY year ASC")

	if err != nil {
		panic(err.Error)
	}

	items := []OilUsageItemsPerYear{}
	OilUsageItemsPerYear := OilUsageItemsPerYear{}
	oilUsageItem := OilUsageItem{}
	oilUsageItems := []OilUsageItem{}
	lastSetYear := 0

	for selDb.Next() {
		var uid, tstamp, crdate, year, quantity, price int
		err = selDb.Scan(&uid, &tstamp, &crdate, &year, &quantity, &price)

		if err != nil {
			panic(err.Error)
		}

		if lastSetYear != year && lastSetYear != 0 {
			OilUsageItemsPerYear.Year = lastSetYear
			OilUsageItemsPerYear.OilUsageItems = oilUsageItems
			items = append(items, OilUsageItemsPerYear)
			oilUsageItems = []OilUsageItem{}
		}

		oilUsageItem.UID = uid
		oilUsageItem.Tstamp = tstamp
		oilUsageItem.Crdate = crdate
		oilUsageItem.Price = price
		oilUsageItem.Year = year
		oilUsageItem.Quantity = quantity
		lastSetYear = year

		oilUsageItems = append(oilUsageItems, oilUsageItem)
	}

	OilUsageItemsPerYear.Year = lastSetYear
	OilUsageItemsPerYear.OilUsageItems = oilUsageItems
	items = append(items, OilUsageItemsPerYear)

	defer db.Close()
	return items
}

func getDataGrouped() OilUsage {
	db := dbConn()
	selDb, err := db.Query("SELECT uid, tstamp, crdate, year, SUM(quantity), SUM(price) FROM tx_psverbrauchsberechnung_item GROUP BY year ORDER BY year ASC")

	if err != nil {
		panic(err.Error)
	}

	oilUsageItem := OilUsageItem{}
	items := []OilUsageItem{}
	oilUsage := OilUsage{}

	var count int
	var sumUsage int
	var sumPrice int
	for selDb.Next() {
		var uid, tstamp, crdate, year, quantity, price int
		err = selDb.Scan(&uid, &tstamp, &crdate, &year, &quantity, &price)

		if err != nil {
			panic(err.Error)
		}
		sumUsage = sumUsage + quantity
		sumPrice = sumPrice + price
		oilUsageItem.UID = uid
		oilUsageItem.Tstamp = tstamp
		oilUsageItem.Crdate = crdate
		oilUsageItem.Price = price
		oilUsageItem.Year = year
		oilUsageItem.Quantity = quantity

		items = append(items, oilUsageItem)
		count++
	}
	oilUsage.OilUsageItems = items
	oilUsage.AverageCost = int(math.Round(float64(sumPrice) / float64(count)))
	oilUsage.AverageUsage = int(math.Round(float64(sumUsage) / float64(count)))
	defer db.Close()
	return oilUsage
}
