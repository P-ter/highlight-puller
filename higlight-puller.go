package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type HightlightResponse struct {
	Results []struct {
		Options struct {
			SerializedData string `json:"serializedData"`
		} `json:"options"`
	} `json:"results"`
}

type Highlight = struct {
	PageTitle string `json:"pageTitle"`
	Color     string `json:"color"`
	Text      string `json:"text"`
}

func main() {
	var apToken string
	var sharingId string
	var appId string
	var size string
	flag.StringVar(&apToken, "apToken", "", "AP Token")
	flag.StringVar(&sharingId, "sharingId", "", "Sharing Id")
	flag.StringVar(&appId, "appId", "", "App Id")
	flag.StringVar(&size, "size", "360", "Size")
	flag.Parse()
	//fetch data from a url
	linkToHighlight := "https://hapicen.com/ap/usersInputs/advanced?aptoken=" + apToken
	body := strings.NewReader(`{
		"filters": {
			"sharingId": "` + sharingId + `",
			"appId": "` + appId + `"
		},
		"pagination": {
			"page": 1,
			"size": ` + size + `,
			"sort": {
				"field": "$created_at",
				"direction": "asc"
			},
			"type": "highlight"
		},
		"count": true,
		"distinctCount": true
	}`)
	resp, err := http.Post(linkToHighlight, "application/json", body)
	if err != nil {
		log.Println("Failed to pull data")
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("Pulled data")
	var highlightResponse HightlightResponse
	err = json.NewDecoder(resp.Body).Decode(&highlightResponse)
	if err != nil {
		panic(err)
	}

	var serializedDataList []string
	for _, results := range highlightResponse.Results {
		serializedDataList = append(serializedDataList, results.Options.SerializedData)
	}
	var highlightList []Highlight
	for _, serializedData := range serializedDataList {
		var highlight Highlight
		var jsonObject interface{}
		err := json.Unmarshal([]byte(serializedData), &jsonObject)
		if err != nil {
			panic(err)
		}
		highlightJson, err := json.Marshal(jsonObject)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(highlightJson, &highlight)
		if err != nil {
			panic(err)
		}

		highlightList = append(highlightList, highlight)
	}

	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")

	pageMap := make(map[string][]string)
	var orderedPages []string
	for _, highlight := range highlightList {
		decodedHightlight, err := url.QueryUnescape(highlight.Text)
		if err != nil {
			panic(err)
		}
		if pageMap[highlight.PageTitle] == nil {
			orderedPages = append(orderedPages, highlight.PageTitle)
			pageMap[highlight.PageTitle] = []string{decodedHightlight}
		} else if decodedHightlight != highlight.PageTitle {
			pageMap[highlight.PageTitle] = append(pageMap[highlight.PageTitle], decodedHightlight)
		}
	}

	highlightFile, err := os.Create("highlight-" + formattedDate + ".txt")
	if err != nil {
		println("Failed to create file")
		panic(err)
	}
	defer highlightFile.Close()

	for _, pageTitle := range orderedPages {
		log.Println("Extracting highlights for page: ", pageTitle)
		highlightFile.WriteString("## " + pageTitle + "\n")
		for _, highlight := range pageMap[pageTitle] {
			highlightFile.WriteString(highlight + "\n")
		}
	}
	// END OMIT
}
