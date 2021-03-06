package worker

import (
	"fmt"
	"sync"
	"testing"

	"price_comparison/sql"
)

func Test_Crawl_Ipad(t *testing.T) {
	m := MomoQuery{keyword: "ipad"}
	page := 1
	finishQuery := make(chan bool)
	newProducts := make(chan *sql.Product)
	wgJob := &sync.WaitGroup{}
	results := []sql.Product{}
	wgJob.Add(1)
	go func() {
		for product := range newProducts {
			results = append(results, *product)
		}

	}()

	m.Crawl(page, finishQuery, newProducts, wgJob)
	fmt.Println(results)
	if len(results) == 0 {
		t.Error("error in crawl")
	}
}
func Test_FindMaxMomoPage_Ipad(t *testing.T) {
	keyword := "ipad"
	maxPage := FindMaxMomoPage(keyword)
	if maxPage < 50 {
		t.Error("error in find momopage,page=", maxPage)
	}
}
