package scrapper

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"main.go/types"
)

func containsJWPlayerSetup(scriptContent string) bool {
	return len(scriptContent) > 0 && (strings.Contains(scriptContent, "jwplayer(") || strings.Contains(scriptContent, "jwplayer.setup("))
}

func GetChannels(url string, defaultLogo string) ([]types.Channel, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", res.StatusCode, res.Status)
	}

	// Step 2: Load HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Step 3: find div with id channels
	channelsDiv := doc.Find("div#channels")
	if channelsDiv.Length() == 0 {
		log.Fatal("No div with id 'channels' found")
	}

	// step 4: find all li inside take the id and the a href from inside
	var channels []types.Channel

	channelsDiv.Find("li").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			log.Printf("No id found for li element at index %d", i)
			return
		}
		href, exists := s.Find("a").Attr("href")
		if !exists {
			log.Printf("No href found for li element with id %s", id)
			return
		}
		imgUrl := s.Find("img").AttrOr("src", "")
		if imgUrl != "" {
			imgUrl = fmt.Sprintf("%s%s", url, imgUrl)
		} else {
			imgUrl = defaultLogo
			log.Printf("No image found for li element with id %s, using default image", id)
		}
		channels = append(channels, types.Channel{
			ID:   id,
			Href: href,
			Img:  imgUrl,
		})
	})
	return channels, nil
}

func GetStreamUrlFromID(id string, base_url string) (string, error) {
	channelHref := fmt.Sprintf("%s/%s", base_url, id)
	res, err := http.Get(channelHref)
	if err != nil {
		log.Printf("Failed to get %s: %v", channelHref, err)
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("Status code error for %s: %d %s", channelHref,
			res.StatusCode, res.Status)
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Failed to parse HTML for %s: %v", channelHref, err)
		return "", err
	}
	// Find the script element with jwplayer setup
	var streamURL string
	var found bool
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent, _ := s.Html()
		if len(scriptContent) > 0 && containsJWPlayerSetup(scriptContent) {
			// Example: jwplayer("theVideoElement").setup({ file: "https://example.com/video.mp4" })
			if strings.Contains(scriptContent, "file:") {
				start := strings.Index(scriptContent, "file:") + len("file:") +
					5 // +5 to skip the quotes
				end := strings.Index(scriptContent[start:], ",") + start - 5
				// replace ampersand with &amp; if present
				if end > start {
					videoURL := scriptContent[start:end]
					streamURL = strings.Replace(videoURL, "&amp;", "&", -1)
					found = true
				} else {
					fmt.Println("No valid video URL found in script content.")
				}
			}
		}
	})
	if found {
		return streamURL, nil
	}
	return "", fmt.Errorf("no valid video URL found in script content")
}
