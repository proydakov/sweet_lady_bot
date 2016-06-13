package main

import (
    "os"
    "fmt"
    "path"
    "time"
    "strings"
    "runtime"
    "net/http"
    "io/ioutil"
    "encoding/json"
)

type Data struct {
    Id     int
    Rank   int
    Model  string
    Author string
    Image  string
}

func downloadData(urlpath string, filepath string, imagepath string) bool {
    // don't worry about errors
    response, e := http.Get(urlpath)
    if e != nil || response.StatusCode != 200 {
        return false
    }

    fmt.Printf("Download data from: %s to file: %s\n", urlpath, filepath)
    body, _ := ioutil.ReadAll(response.Body)
    response.Body.Close()

    var data []Data
    json.Unmarshal(body, &data)

    if len(data) < 1 {
        return false
    }

    data[0].Image = imagepath
    b, _ := json.Marshal(data[0])
    err  := ioutil.WriteFile(filepath, b, 0644)
    if err != nil {
        return false
    }

    return true
}

func downloadFunc(dir string, cin chan string, cout chan bool) {
    site := "http://api.oboobs.ru/boobs/get"

    for {
        imagepath, ok := <- cin
        if !ok {
            return
        }
        extension := path.Ext(imagepath)
        basename := strings.Replace(path.Base(imagepath), extension, "", 1)
        urlpath  := site + "/" + basename
        filepath := dir + "/" + basename + ".json"
        res := downloadData(urlpath, filepath, imagepath)
        cout <- res
    }
}

func getImageNames(dirpath string) []string {
    var names []string
    files, _ := ioutil.ReadDir(dirpath)
    for _, file := range files {
        names = append(names, dirpath + "/" + file.Name())
    }
    return names
}

func main() {
    if len(os.Args) != 3 {
        fmt.Printf("<image dir> <output dir>\n")
        return;
    }

    imageDir  := os.Args[1]
    outputDir := os.Args[2]

    names := getImageNames(imageDir)
    all := len(names)

    t0 := time.Now()

    cpus := runtime.GOMAXPROCS(0)
    //cpus = 1

    cin := make(chan string, all)
    cout := make(chan bool, all)

    for cpu := 0; cpu < cpus; cpu++ {
        go downloadFunc(outputDir, cin, cout)
    }

    for _, name := range names {
        cin <- name
    }

    counter := 0
    for range names {
        res, ok := <- cout
        if !ok {
            break
        }
        if res {
            counter++;
        }
    }

    close(cin)
    close(cout)

    t1 := time.Now()
    fmt.Println("###############################################################################")
    fmt.Printf("Downloaded data count: %d\n", counter)
    fmt.Printf("Total time %v.\n", t1.Sub(t0))
}
