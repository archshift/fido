package main

import (
	"bytes"
	"container/list"
	"encoding/json"
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
	size     int64
	entries  *list.List
}

func (ent FsEntry) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 255))

	buf.WriteString("{\"path\":")
	if b, err := json.Marshal(ent.path); err != nil {
		return nil, err
	} else {
		buf.Write(b)
	}

	buf.WriteString(",\"mod_time\":")
	if b, err := json.Marshal(ent.mod_time); err != nil {
		return nil, err
	} else {
		buf.Write(b)
	}

	buf.WriteString(",\"size\":")
	if b, err := json.Marshal(ent.size); err != nil {
		return nil, err
	} else {
		buf.Write(b)
	}

	buf.WriteString("}")

	return buf.Bytes(), nil
}

func HandleListing(writer http.ResponseWriter, request *http.Request, dir_path string) {
	info, err := ioutil.ReadDir(dir_path)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprint(writer, "{\"files\":[")
	for i, item := range info {
		if i != 0 {
			fmt.Fprint(writer, ",")
		}

		b, err := json.Marshal(FsEntry{
			path:     dir_path + item.Name(),
			mod_time: item.ModTime(),
			size:     item.Size(),
			entries:  nil,
		})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if err := json.Compact(&buf, b); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		buf.WriteTo(writer)
	}
	fmt.Fprint(writer, "]}")
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
