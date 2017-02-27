package main

import (
    "os"
    "log"
    "time"
    "bytes"
    "strings"
    "io/ioutil"
    "math/rand"
    "encoding/json"

    "github.com/mrd0ll4r/tbotapi"
    "github.com/mrd0ll4r/tbotapi/examples/boilerplate"
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
        names = append(names, path + f.Name())
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
    size  := len(array)
    if size == 0 {
        return ""
    }
    return array[len(array) - 1]
}

func main() {
    if len(os.Args) < 2 {
        log.Printf("<config path> <log path *optional*>\n")
        return;
    }

    if len(os.Args) == 3 {
        f, err := os.OpenFile(os.Args[3], os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
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
    log.Printf("Loaded: %d image cache.\n", len(images))

    rand.Seed(time.Now().UTC().UnixNano())
    apiToken := config.Token

    updateFunc := func(update tbotapi.Update, api *tbotapi.TelegramBotAPI) {
        switch update.Type() {
        case tbotapi.MessageUpdate:
            msg := update.Message
            typ := msg.Type()
            if typ != tbotapi.TextMessage {
                // Ignore non-text messages for now.
                log.Println("Ignoring non-text message")
                return
            }
            // Note: Bots cannot receive from channels, at least no text messages. So we don't have to distinguish anything here.
            // Display the incoming message.
            // msg.Chat implements log.Stringer, so it'll display nicely.
            // We know it's a text message, so we can safely use the Message.Text pointer.
            log.Printf("<-%d, From:\t%s, Text: %s \n", msg.ID, msg.Chat, *msg.Text)

            // Send a photo.
            name := getRandImageName(images)
            data := cache[name]
            reader := bytes.NewReader(data)

            // Note: Set at least a correct file extension, the API will check this.
            outMsg, err := api.NewOutgoingPhoto(tbotapi.NewRecipientFromChat(msg.Chat), "girl.jpg", reader).Send()

            if err != nil {
                log.Printf("Error sending: %s\n", err)
                return
            }
            log.Printf("->%d, To:\t%s, Photo: %s\n", outMsg.Message.ID, outMsg.Message.Chat, extractImageName(name))
        case tbotapi.InlineQueryUpdate:
            log.Println("Ignoring received inline query: ", update.InlineQuery.Query)
        case tbotapi.ChosenInlineResultUpdate:
            log.Println("Ignoring chosen inline query result (ID): ", update.ChosenInlineResult.ID)
        default:
            log.Printf("Ignoring unknown Update type.")
        }
    }

    // Run the bot, this will block.
    boilerplate.RunBot(apiToken, updateFunc, "Photo", "Always responds to text messages with a picture")

    log.Printf("Staffing completion")
}
