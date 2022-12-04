package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func getUUID() uuid.UUID {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return newUUID
}

var fileRoot string = "."
var listenPort uint64 = 8000

func initializeConfig() {
	var err error
	inputFileRoot := os.Getenv("FILE_ROOT")
	if len(inputFileRoot) != 0 {
		fileRoot = inputFileRoot
	}
	inputPort := os.Getenv("PORT")
	if len(inputPort) != 0 {
		listenPort, err = strconv.ParseUint(inputPort, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getFilePath(fileId uuid.UUID) string {
	fullName := fmt.Sprintf("%s.bin", fileId.String())
	fullPath := filepath.Join(fileRoot, fullName)
	return fullPath
}

func parseFileName(r *http.Request) (uuid.UUID, error) {
	params := mux.Vars(r)
	name := params["name"]
	if len(name) == 0 {
		panic("Missing name")
	}
	parsedUUID, err := uuid.Parse(name)
	return parsedUUID, err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getFile(w http.ResponseWriter, r *http.Request) {
	fileId, err := parseFileName(r)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fp := getFilePath(fileId)
	if !fileExists(fp) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fileId := getUUID()
	fp := getFilePath(fileId)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = os.WriteFile(fp, data, 0655)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(fileId.String()))
	if err != nil {
		log.Fatal(err)
	}
}

func updateFile(w http.ResponseWriter, r *http.Request) {
	fileId, err := parseFileName(r)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fp := getFilePath(fileId)
	if !fileExists(fp) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = os.WriteFile(fp, data, 0655)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	fileId, err := parseFileName(r)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fp := getFilePath(fileId)
	if !fileExists(fp) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = os.Remove(fp)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func routePathWithId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getFile(w, r)
	case "PUT":
		updateFile(w, r)
	case "DELETE":
		deleteFile(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func initializeRouter() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", uploadFile).Methods("POST")
	rtr.HandleFunc("/{name:[a-z0-9\\-]+}", routePathWithId).Methods("GET", "PUT", "DELETE")

	http.Handle("/", rtr)
}

func main() {
	initializeConfig()
	initializeRouter()
	addr := fmt.Sprintf(":%d", listenPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
