package main

import (
    "container/list"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"
)

type FsEntriesByDate list.List

type FsEntry struct {
    path     string
    mod_time time.Time
    entries  *list.List
}

func HandleListing(writer http.ResponseWriter, request *http.Request, dir_path string) {
    info, err := ioutil.ReadDir(dir_path)
    if err != nil {
        writer.WriteHeader(http.StatusNotFound)
        return
    }
    for _, i := range info {
        fmt.Printf("%s/%s\n", dir_path, i.Name())
    }
}

func HandleHttp(writer http.ResponseWriter, request *http.Request) {
    path := request.URL.Path
    fmt.Println(path)

    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            writer.WriteHeader(http.StatusNotFound)
        } else {
            writer.WriteHeader(http.StatusInternalServerError)
        }
        return
    }

    if info.IsDir() {
        HandleListing(writer, request, path)
    } else {
        http.ServeFile(writer, request, path)
    }
}

func main() {
    fmt.Println("Starting...")

    mux := http.NewServeMux()
    mux.HandleFunc("/", HandleHttp)
    server := http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
