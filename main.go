package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	SecretKey = "iplaygodotandclaimfun"
	GeminiKey = "AIzaSyAKz6guWs938DdF_ZZDexZ72lCDljj9zOY"
)

type RequestBody struct {
	Question string `json:"question"`
	Secret   string `json:"secret"`
}

func solveHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Настройка CORS (чтобы браузер не блокировал запрос)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Проверка пути
	if r.URL.Path != "/solve" {
		// Для главной страницы просто выводим статус
		if r.URL.Path == "/" {
			fmt.Fprint(w, "Server is live! Send POST to /solve")
			return
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// 3. Чтение данных
	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Empty or bad body", http.StatusBadRequest)
		return
	}

	if req.Secret != SecretKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 4. Запрос к Gemini
	geminiURL := "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-flash:generateContent?key=" + GeminiKey

	prompt := fmt.Sprintf("Ты — помощник по тестам. Проанализируй вопрос и варианты. Выдай ТОЛЬКО текст правильного ответа или его букву. Вопрос: %s", req.Question)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Gemini connection error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 5. Отправка ответа клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/", solveHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	http.ListenAndServe(":"+port, nil)
}
