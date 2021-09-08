package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func main() {
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  "localhost:8080",
	}
	http.Handle("/upload", uploadHandler)

	fs := &http.Server{
		Addr:         uploadHandler.HostAddr,
		Handler:      http.FileServer(http.Dir(uploadHandler.UploadDir)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fs.ListenAndServe()
}
