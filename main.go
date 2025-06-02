package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var (
	maxResults = flag.Int64("max-results", 50, "Max YouTube results")
)

const developerKey = "" //YOUR DEVELOPER KEY

// use curl call directly
func main() {
	ParsingYear := 2024
	filepathPrivate := ""
	filepath := filepathPrivate + strconv.Itoa(ParsingYear) + "_using"
	//fmt.Println(filepath)
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("can't open youtube channel list file")
	}
	defer file.Close()

	outPrivatePath := ""
	outFilepath := outPrivatePath + strconv.Itoa(ParsingYear)
	outputFile, err := os.OpenFile(outFilepath, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		fmt.Printf("fail to open or create outputfile %v\n", outFilepath)
		os.Exit(1)
	}
	defer outputFile.Close()
	w := bufio.NewWriter(outputFile)

	scanner := bufio.NewScanner(file)
	cnt, cant := 0, 0
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		//fmt.Printf("%v: %v\n", lineNum, line)

		pair := strings.Split(line, ", ")
		link := pair[0]
		author := pair[1]
		note := pair[2:]
		//fmt.Printf("link %v, author %v, note %v\n", link, author, note)

		if link[:4] == "null" {
			fmt.Fprintf(w, "%v$ $ $%v$ 無youtube頻道$ $ %v$ %v$\n", lineNum, author, ParsingYear, note)
		} else {
			pre := len("https://www.youtube.com/")
			left := link[pre:]
			slashLoca := strings.Index(left, string('/'))
			//fmt.Printf("slashLoca: %v, link[pre:]: %v\n", slashLoca, link[pre:])
			kind := ""
			if slashLoca == -1 {
				kind = "author"
			} else {
				kind = link[pre : pre+slashLoca]
				if left[0] == '@' {
					kind = "author"
				}
			}

			id := link[pre+slashLoca+1:]

			if kind == "channel" {
				GetCols(id, kind, ParsingYear, w, lineNum, author)
			} else if kind == "user" {
				GetCols(id, kind, ParsingYear, w, lineNum, author)
			} else if kind == "c" {
				cnt++
				/*} else if kind == "author" {
				//Not really need to call. Can add "view-source:" in sublime directly
				printViewSourceForChannels(link)*/
			} else {
				fmt.Fprintf(w, "can't handle")
				cant++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln("scanner err for readline from youtube channel list file: ", err)
	}

	fmt.Fprintf(w, "num of links as /c/ : %v, num can't parsed: %v\n", cnt, cant)
}

func GetCols(channelId string, IdCategory string, year int, w *bufio.Writer, lineNum int, author string) {
	url := ""
	previousY := strconv.Itoa(year - 1)
	nextY := strconv.Itoa(year + 1)

	if IdCategory == "channel" {
		url = "https://www.googleapis.com/youtube/v3/channels?part=snippet,contentDetails&key=" + developerKey + "&" + "id=" + channelId + "&publishedAfter=" + previousY + "-12-31T23%3A59%3A59Z&publishedBefore=" + nextY + "-01-01T00%3A00%3A00Z&order=date&maxResults=50"

	} else if IdCategory == "user" {
		url = "https://www.googleapis.com/youtube/v3/channels?part=snippet,contentDetails&key=" + developerKey + "&" + "forUsername=" + channelId + "&publishedAfter=" + previousY + "-12-31T23%3A59%3A59Z&publishedBefore=" + nextY + "-01-01T00%3A00%3A00Z&order=date&maxResults=50"
	} else {
		fmt.Printf("Can't parse: %v with category %v\n", channelId, IdCategory)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("io read err")
		fmt.Println(err)
	}

	var data ChannelList
	json.Unmarshal(body, &data)

	playlistId := data.Items[0].ContentDetails.RelatedPlaylists.Uploads
	nextPageToken := ""
	for {
		url2 := "https://youtube.googleapis.com/youtube/v3/playlistItems?part=snippet&part=contentDetails&maxResults=50&playlistId=" + playlistId + "&pageToken=" + nextPageToken + "&key=" + developerKey
		resp2, err := http.Get(url2)
		if err != nil {
			fmt.Println("err")
			fmt.Println(err)
		}

		defer resp2.Body.Close()

		body2, err := ioutil.ReadAll(resp2.Body)
		if err != nil {
			fmt.Println("io read err")
			fmt.Println(err)
		}

		var data2 PlaylistItem
		json.Unmarshal(body2, &data2)

		if data2.PageInfo.TotalResults != 0 {
			for _, item := range data2.Items {
				if item.Snippet.PublishedAt.Before(time.Date(year+1, time.January, 1, 00, 00, 00, 00, time.Local)) && item.Snippet.PublishedAt.After(time.Date(year-1, time.December, 31, 23, 59, 59, 59, time.Local)) {
					link := "https://www.youtube.com/watch?v=" + item.Snippet.ResourceID.VideoID
					title := item.Snippet.Title
					date := item.Snippet.PublishedAt
					description := item.Snippet.Description
					des := strings.Replace(description, "\n", "__", -1)

					urlVideo := "https://www.googleapis.com/youtube/v3/videos?id=" + item.Snippet.ResourceID.VideoID + "&part=contentDetails&key=" + developerKey
					respVideo, err := http.Get(urlVideo)
					if err != nil {
						fmt.Println("err")
						fmt.Println(err)
					}
					defer respVideo.Body.Close()

					bodyVideo, err := ioutil.ReadAll(respVideo.Body)
					var dataVideo Videos
					json.Unmarshal(bodyVideo, &dataVideo)
					duration := dataVideo.Items[0].ContentDetails.Duration
					fmt.Fprintf(w, "%v$ $ $%v$ %v$ $%v$ %v$ %v$ 2024$ %v$  %v$ $%v\n", lineNum, author, title, date, link, CalSeconds(duration), duration[2:], CalSeconds(duration), des)
				}
			}
			nextPageToken = data2.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
		if nextPageToken == "" {
			break
		}
	}
}

func CalSeconds(input string) int {

	if input[len(input)-1] == 'M' {
		input = input[:len(input)-1]
		input = input[2:]
		m, _ := strconv.Atoi(input)
		return 60 * m
	}

	input = input[:len(input)-1]
	input = input[2:]

	total := 0
	loca := -1

	for idx, c := range input {
		if c == 'M' {
			loca = idx
		}
	}

	if loca != -1 {
		s, _ := strconv.Atoi(input[loca+1:])
		m, _ := strconv.Atoi(input[:loca])

		total = m*60 + s
		return total
	} else {
		s, _ := strconv.Atoi(input)
		return s
	}

}

// just for trying youtube API before, not using
func YoutubeQuery() {
	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	var part []string
	call := service.Search.List(part).ChannelId("UCbijK1E1aI7b_GMOBvNDKcQ").Order("date")
	response, err := call.Do()
	if err != nil {
		fmt.Println(err)
	}

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)

	fmt.Println("response:")
	fmt.Println(response)

	printIDs("Videos", videos)
	//printIDs("Channels", channels)
	//printIDs("Playlists", playlists)
}

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
// Retrieve resource for the authenticated user's channel

// just for trying youtube API before, not using at the end
func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[%v] %v\n", "https://www.youtube.com/watch?v="+id, title)
	}
	fmt.Printf("\n\n")
}

// Add "view-source:" directly in sublime in front of YoutubeChannelLsit might be easier.
/*func printViewSourceForChannels(link string) {
	fmt.Println("view-source:" + link)
}*/
