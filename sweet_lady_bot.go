package main

import (
    "os"
    "fmt"
    "time"
    "bytes"
    "io/ioutil"
    "math/rand"
    "encoding/json"

    "github.com/mrd0ll4r/tbotapi"
    "github.com/mrd0ll4r/tbotapi/examples/boilerplate"
)

type Config struct {
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

func main() {
    if len(os.Args) != 2 {
        fmt.Printf("<config path>\n")
        return;
    }

    var config Config
    array, _ := ioutil.ReadFile(os.Args[1])
    json.Unmarshal(array, &config)

    fmt.Printf("Loading image cache...\n")
    images := getImageNames(config.Imagedir)
    cache := loadImageCache(images)
    fmt.Printf("Loaded: %d image cache.\n", len(images))

    rand.Seed(time.Now().UTC().UnixNano())
    apiToken := config.Token

    updateFunc := func(update tbotapi.Update, api *tbotapi.TelegramBotAPI) {
        switch update.Type() {
        case tbotapi.MessageUpdate:
            msg := update.Message
            typ := msg.Type()
            if typ != tbotapi.TextMessage {
                // Ignore non-text messages for now.
                fmt.Println("Ignoring non-text message")
                return
            }
            // Note: Bots cannot receive from channels, at least no text messages. So we don't have to distinguish anything here.
            // Display the incoming message.
            // msg.Chat implements fmt.Stringer, so it'll display nicely.
            // We know it's a text message, so we can safely use the Message.Text pointer.
            fmt.Printf("<-%d, From:\t%s, Text: %s \n", msg.ID, msg.Chat, *msg.Text)

            // Send a photo.
            name := getRandImageName(images)
            data := cache[name]
            reader := bytes.NewReader(data)

            // Note: Set at least a correct file extension, the API will check this.
            outMsg, err := api.NewOutgoingPhoto(tbotapi.NewRecipientFromChat(msg.Chat), "girl.jpg", reader).Send()

            if err != nil {
                fmt.Printf("Error sending: %s\n", err)
                return
            }
            fmt.Printf("->%d, To:\t%s, (Photo)\n", outMsg.Message.ID, outMsg.Message.Chat)
        case tbotapi.InlineQueryUpdate:
            fmt.Println("Ignoring received inline query: ", update.InlineQuery.Query)
        case tbotapi.ChosenInlineResultUpdate:
            fmt.Println("Ignoring chosen inline query result (ID): ", update.ChosenInlineResult.ID)
        default:
            fmt.Printf("Ignoring unknown Update type.")
        }
    }

    // Run the bot, this will block.
    boilerplate.RunBot(apiToken, updateFunc, "Photo", "Always responds to text messages with a picture")

    fmt.Printf("Staffing completion")
}
