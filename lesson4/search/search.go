package search

import (
	"io/fs"
	"lesson4/model"
	"os"
	"strings"
)

type SearcherInFolder struct {
	Dir string
}

func (fs SearcherInFolder) Search(extension string) ([]model.File, error) {
	dirFiles, err := os.ReadDir(fs.Dir)
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	for _, dirFile := range dirFiles {
		file, err := getFileInfo(dirFile)
		if err != nil {
			return nil, err
		}
		// Если передана пустая строка, то считаем, что нужен список всех файлов
		if file.Extension == extension || extension == "" {
			files = append(files, file)
		}
	}
	return files, nil
}

func getFileInfo(entry fs.DirEntry) (model.File, error) {
	info, err := entry.Info()
	if err != nil {
		return model.File{}, err
	}
	filename := info.Name()
	splittedFilename := strings.Split(filename, ".")
	extension := splittedFilename[len(splittedFilename)-1]
	size := info.Size()

	return model.File{
		Name:      filename,
		Extension: extension,
		Size:      size,
	}, nil
}
