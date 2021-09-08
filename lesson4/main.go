package main

import (
	"encoding/json"
	"fmt"
	"io"
	"lesson4/model"
	"lesson4/search"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 1. Возможность получить список файлов на сервере (имя, расширение, размер в байтах)
// 2. Фильтрация списка по расширению
// 4. Тесты с использованием библиотеки httptest

type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.UploadDir, header.Filename)

	err = os.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	fileLink := h.HostAddr + "/" + header.Filename
	fmt.Fprintln(w, fileLink)
}

type Searcher interface {
	Search(extension string) ([]model.File, error)
}

type GetFileListHandler struct {
	FileSearcher Searcher
}

func (h *GetFileListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ext := r.FormValue("ext")
		files, err := h.FileSearcher.Search(ext)
		if err != nil {
			http.Error(w, "Unable to get list of files", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(files)
		if err != nil {
			http.Error(w, "Error when encoding", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Unknown request", http.StatusMethodNotAllowed)
	}
}

func run(servers []*http.Server) {
	wg := sync.WaitGroup{}
	for _, srv := range servers {
		wg.Add(1)
		go func(server *http.Server) {
			_ = server.ListenAndServe()
			wg.Done()
		}(srv)
	}
	wg.Wait()

}

func main() {
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  "localhost:8080",
	}
	http.Handle("/upload", uploadHandler)

	getFileListHandler := GetFileListHandler{
		FileSearcher: search.SearcherInFolder{Dir: uploadHandler.UploadDir},
	}
	http.Handle("/files", &getFileListHandler)

	srv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fs := &http.Server{
		Addr:         uploadHandler.HostAddr,
		Handler:      http.FileServer(http.Dir(uploadHandler.UploadDir)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	servers := []*http.Server{fs, srv}
	run(servers)
}
