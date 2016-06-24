package main

import (
    "os"
    "io"
    "fmt"
    "time"
    "runtime"
    "strconv"
    "net/http"
)

func formatID(id int, size int) string {
    s := strconv.Itoa(id)
    ssize := len(s)
    for i := 0; i < (size - ssize); i++ {
        s = "0" + s
    }
    return s
}

func downloadImage(urlpath string, filepath string) bool {
    // don't worry about errors
    response, e := http.Get(urlpath)
    if e != nil || response.StatusCode != 200 {
        return false
    }
    defer response.Body.Close()

    fmt.Printf("Download image from: %s to file: %s\n", urlpath, filepath)

    //open a file for writing
    file, err := os.Create(filepath)
    if err != nil {
        return false
    }
    // Use io.Copy to just dump the response body to the file. This supports huge files
    _, err = io.Copy(file, response.Body)
    if err != nil {
        return false
    }
    file.Close()

    return true
}

func downloadFunc(dir string, cin chan int, cout chan bool) {
    site := "http://media.oboobs.ru"

    for {
        id, ok := <- cin
        if !ok {
            return
        }
        imagename := formatID(id, 5)
        urlpath  := site + "/boobs/" + imagename + ".jpg"
        filepath := dir + "/" + imagename + ".jpg"
        res := downloadImage(urlpath, filepath)
        cout <- res
    }
}

func main() {
    if len(os.Args) != 4 {
        fmt.Printf("<from> <to> <output directory>\n")
        return;
    }

    t0 := time.Now()

    from, _ := strconv.Atoi(os.Args[1])
    to, _   := strconv.Atoi(os.Args[2])
    all     := to - from + 1
    dir     := os.Args[3]

    cpus := runtime.GOMAXPROCS(0)
    cpus = 1

    cin := make(chan int, all)
    cout := make(chan bool, all)

    for cpu := 0; cpu < cpus; cpu++ {
        go downloadFunc(dir, cin, cout)
    }

    for id := from; id <= to; id++ {
        cin <- id
    }

    counter := 0
    for id := from; id <= to; id++ {
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
    fmt.Printf("Downloaded image count: %d\n", counter)
    fmt.Printf("Total time %v.\n", t1.Sub(t0))
}
