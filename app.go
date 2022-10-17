package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type MetricModel struct {
	Total                 int    `json:"total"`
	ReadFileSystemEntries int    `json:"readFileSystemEntries"`
	ReadFile              int    `json:"readFile"`
	WriteFile             int    `json:"writeFile"`
	GenerateFile          int    `json:"generateFile"`
	Encryption            int    `json:"encryption"`
	Decryption            int    `json:"decryption"`
	Unit                  string `json:"unit"`
}

type GetFilesResponseModel struct {
	Files   []string    `json:"files"`
	Metrics MetricModel `json:"metrics"`
}

type PostFilesRequestModel struct {
	Key      string `json:"key"`
	Length   int    `json:"length"`
	FileName string `json:"fileName"`
}

type PostFilesResponseModel struct {
	Metrics MetricModel `json:"metrics"`
}

type GetFileRequestModel struct {
	Key string `json:"key"`
}

type GetFileResponseModel struct {
	Text    string      `json:"text"`
	Metrics MetricModel `json:"metrics"`
}

func encrypt(data string, key string) (string, error) {
	byteKey := []byte(key)
	byteData := []byte(data)
	block, err := aes.NewCipher(byteKey)
	if err != nil {
		return "", err
	}
	nonce := []byte("ICA BOSSSSSS")
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	encryptedData := aesgcm.Seal(nil, nonce, byteData, nil)
	return fmt.Sprintf("%x", encryptedData), nil
}

func decrypt(encryptedData string, key string) (string, error) {
	byteKey := []byte(key)
	byteEncryptedData, err := hex.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(byteKey)
	if err != nil {
		return "", err
	}
	nonce := []byte("ICA BOSSSSSS")
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	data, err := aesgcm.Open(nil, nonce, byteEncryptedData, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", string(data)), nil
}

func filesRoute(w http.ResponseWriter, r *http.Request) {
	totalStart := time.Now()
	fs, err := ioutil.ReadDir("store")
	if err != nil {
		fmt.Println("store folder is not created")
		fmt.Println(err.Error())
		http.Error(w, "store folder is not created: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		var getFilesResponseModel GetFilesResponseModel
		getFilesResponseModel.Metrics.Unit = "microseconds"

		start := time.Now()
		for _, f := range fs {
			getFilesResponseModel.Files = append(getFilesResponseModel.Files, f.Name())
		}
		getFilesResponseModel.Metrics.ReadFileSystemEntries = int(time.Since(start).Microseconds())

		w.WriteHeader(http.StatusOK)
		getFilesResponseModel.Metrics.Total = int(time.Since(totalStart).Microseconds())
		json.NewEncoder(w).Encode(getFilesResponseModel)
	case "POST":
		var postFilesRequestModel PostFilesRequestModel
		err := json.NewDecoder(r.Body).Decode(&postFilesRequestModel)
		if err != io.EOF && err != nil {
			fmt.Println("Failed to parse body")
			fmt.Println(err.Error())
			http.Error(w, "Failed to parse body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(postFilesRequestModel.Key) != 16 {
			fmt.Println("Key must be 16 bytes length")
			http.Error(w, "Key must be 16 bytes length", http.StatusBadRequest)
			return
		}

		var postFilesResponseModel PostFilesResponseModel
		postFilesResponseModel.Metrics.Unit = "microseconds"

		start := time.Now()
		var fileContent string
		for i := 0; i < postFilesRequestModel.Length; i++ {
			fileContent = fileContent + "a"
		}
		postFilesResponseModel.Metrics.GenerateFile = int(time.Since(start).Microseconds())

		start = time.Now()
		encriptedFileContent, err := encrypt(fileContent, postFilesRequestModel.Key)
		postFilesResponseModel.Metrics.Encryption = int(time.Since(start).Microseconds())
		if err != nil {
			fmt.Println("Failed to encrypt data")
			fmt.Println(err.Error())
			http.Error(w, "Failed to encrypt data: "+err.Error(), http.StatusBadRequest)
			return
		}

		start = time.Now()
		err = os.WriteFile("store/"+postFilesRequestModel.FileName, []byte(encriptedFileContent), 0644)
		postFilesResponseModel.Metrics.WriteFile = int(time.Since(start).Microseconds())
		if err != nil {
			fmt.Println("Failed to write to disk")
			fmt.Println(err.Error())
			http.Error(w, "Failed to write to disk: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		postFilesResponseModel.Metrics.Total = int(time.Since(totalStart).Microseconds())
		json.NewEncoder(w).Encode(postFilesResponseModel)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func fileRoute(w http.ResponseWriter, r *http.Request) {
	totalStart := time.Now()
	_, err := ioutil.ReadDir("store")
	if err != nil {
		fmt.Println("store folder is not created")
		fmt.Println(err.Error())
		http.Error(w, "store folder is not created", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	fileId, ok := vars["fileId"]
	if !ok {
		fmt.Println("id is missing in parameters")
		http.Error(w, "id is missing in parameters", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var getFileRequestModel GetFileRequestModel
		err := json.NewDecoder(r.Body).Decode(&getFileRequestModel)
		if err != io.EOF && err != nil {
			fmt.Println("Failed to parse body")
			fmt.Println(err.Error())
			http.Error(w, "Failed to parse body", http.StatusBadRequest)
			return
		}

		if len(getFileRequestModel.Key) != 16 {
			fmt.Println("Key must be 16 bytes length")
			http.Error(w, "Key must be 16 bytes length", http.StatusBadRequest)
			return
		}

		var getFileResponseModel GetFileResponseModel
		getFileResponseModel.Metrics.Unit = "microseconds"

		start := time.Now()
		encryptedFileContent, err := os.ReadFile("store/" + fileId)
		getFileResponseModel.Metrics.ReadFile = int(time.Since(start).Microseconds())
		if err != nil {
			fmt.Println("Failed to get file")
			fmt.Println(err.Error())
			http.Error(w, "Failed to get file", http.StatusInternalServerError)
			return
		}

		start = time.Now()
		fileContent, err := decrypt(string(encryptedFileContent), getFileRequestModel.Key)
		getFileResponseModel.Metrics.Decryption = int(time.Since(start).Microseconds())
		if err != nil {
			fmt.Println("Failed to decrypt data")
			fmt.Println(err.Error())
			http.Error(w, "Failed to decrypt data: "+err.Error(), http.StatusBadRequest)
			return
		}

		getFileResponseModel.Text = fileContent

		w.WriteHeader(http.StatusOK)
		getFileResponseModel.Metrics.Total = int(time.Since(totalStart).Microseconds())
		json.NewEncoder(w).Encode(getFileResponseModel)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/files", filesRoute)
	r.HandleFunc("/files/{fileId}", fileRoute)
	port := 8080
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
