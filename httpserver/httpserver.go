package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"main.go/config"
	"main.go/scrapper"
)

func Start(port string) {
	http.HandleFunc("/manifest.json", manifestHandler)
	http.HandleFunc("/catalog/tv/channels.json", catalogHandler)
	http.HandleFunc("/stream/tv/", streamHandler)
	http.HandleFunc("/catalog/tv/channels", catalogHandler)
	http.HandleFunc("/meta/tv/", metaHandler)
	http.HandleFunc("/", redirectToManifest)

	log.Println("Stremio Addon listening on :" + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func redirectToManifest(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/manifest.json", http.StatusFound)
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	conf, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config: "+err.Error(), 500)
		return
	}

	manifest := map[string]interface{}{
		"id":          "stremio-bgtv",
		"version":     "1.0.0",
		"name":        "Live Bulgarian TV",
		"description": "Live Bulgarian TV channels",
		"logo":        conf.DefaultLogo,
		"resources": []interface{}{
			"catalog",
			map[string]interface{}{
				"name":  "stream",
				"types": []string{"tv"},
			},
			map[string]interface{}{
				"name":  "meta",
				"types": []string{"tv"},
			},
		},

		"types": []string{"tv"},
		"catalogs": []map[string]interface{}{
			{
				"type": "tv",
				"id":   "channels",
				"name": "Channels",
				"extra": []map[string]interface{}{
					{
						"name":       "search",
						"isRequired": false,
					},
				},
			},
		},
	}

	json.NewEncoder(w).Encode(manifest)
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	conf, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config: "+err.Error(), 500)
		return
	}
	channels, err := scrapper.GetChannels(conf.Base, conf.DefaultLogo)
	if err != nil {
		http.Error(w, "Failed to fetch channels: "+err.Error(), 500)

		return
	}

	metas := []map[string]string{}
	for _, ch := range channels {
		metas = append(metas, map[string]string{
			"id":     ch.ID,
			"name":   ch.ID,
			"type":   "tv",
			"poster": ch.Img,
		})
	}

	response := map[string]interface{}{"metas": metas}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id := strings.TrimPrefix(r.URL.Path, "/stream/tv/")
	id = strings.TrimSuffix(id, ".json")

	log.Println("Requested stream ID:", id)
	conf, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config: "+err.Error(), 500)
		return
	}
	streamURL, err := scrapper.GetStreamUrlFromID(id, conf.Base)
	if err != nil {
		http.Error(w, "Failed to fetch streams: "+err.Error(), 500)
		return
	}

	if streamURL == "" {
		log.Println("Stream URL not found for ID:", id)
		http.Error(w, "Stream URL not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"url":  streamURL,
				"name": id,
			},
		},
	})

	http.Error(w, "Stream not found", 404)
}

func metaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id := strings.TrimPrefix(r.URL.Path, "/meta/tv/")
	id = strings.TrimSuffix(id, ".json")

	conf, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config: "+err.Error(), 500)
		return
	}

	channels, err := scrapper.GetChannels(conf.Base, conf.DefaultLogo)
	if err != nil {
		http.Error(w, "Failed to fetch metadata: "+err.Error(), 500)
		return
	}

	for _, ch := range channels {
		if ch.ID == id {
			meta := map[string]interface{}{
				"id":     ch.ID,
				"name":   ch.ID,
				"type":   "tv",
				"poster": ch.Img,
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"meta": meta,
			})
			return
		}
	}
	http.Error(w, "Meta not found", 404)
}
