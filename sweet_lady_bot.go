package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bot-api/telegram"
	"golang.org/x/net/context"
)

type config struct {
	Token    string
	Imagedir string
	Metadir  string
}

func getImageNames(imagedir string) []string {
	path := imagedir + "/"
	files, _ := ioutil.ReadDir(path)
	var names []string
	for _, f := range files {
		names = append(names, path+f.Name())
	}
	return names
}

func getRandImageName(images []string) string {
	size := len(images)
	index := rand.Intn(size)
	name := images[index]
	return name
}

func loadImageCache(images []string) map[string][]byte {
	cache := make(map[string][]byte)
	for _, image := range images {
		array, _ := ioutil.ReadFile(image)
		cache[image] = array
	}
	return cache
}

func extractImageName(path string) string {
	array := strings.Split(path, "/")
	size := len(array)
	if size == 0 {
		return ""
	}
	return array[len(array)-1]
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("<config path> <log path *optional*>\n")
		return
	}

	if len(os.Args) == 3 {
		f, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	var config config
	array, _ := ioutil.ReadFile(os.Args[1])
	json.Unmarshal(array, &config)

	log.Printf("Loading image cache...\n")
	images := getImageNames(config.Imagedir)
	cache := loadImageCache(images)
	log.Printf("Loaded: %d images, %d cache\n", len(images), len(cache))

	rand.Seed(time.Now().UTC().UnixNano())

	token := config.Token
	debug := false

	if token == "" {
		log.Fatal("token flag required")
	}

	api := telegram.New(token)
	api.Debug(debug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if user, err := api.GetMe(ctx); err != nil {
		log.Panic(err)
	} else {
		log.Printf("bot info: %#v", user)
	}

	updatesCh := make(chan telegram.Update)

	go telegram.GetUpdates(ctx, api, telegram.UpdateCfg{
		Timeout: 10, // Timeout in seconds for long polling.
		Offset:  0,  // Start with the oldest update
	}, updatesCh)

	for update := range updatesCh {
		rcvMsg := update.Message
		if nil == rcvMsg {
			continue
		}
		rid := rcvMsg.MessageID
		rfrom := rcvMsg.From
		rtext := rcvMsg.Text
		uname := rfrom.FirstName + " " + rfrom.LastName

		log.Printf("<-%d, From: %s, Text: %s", rid, uname, rtext)

		// Send a photo.
		name := getRandImageName(images)
		data := cache[name]

		sendFile := telegram.NewBytesFile(name, data)
		upload := telegram.NewPhotoUpload(update.Message.Chat.ID, sendFile)

		outMsg, err := api.Send(ctx, upload)
		if err != nil {
			log.Printf("send error: %v", err)
		}
		oid := outMsg.MessageID
		log.Printf("->%d, To: %s, Photo: %s", oid, uname, extractImageName(name))
	}
	log.Println("Done")
}
