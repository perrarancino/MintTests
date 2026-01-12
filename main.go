package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Твои данные из "Saved Information"
const (
	SecretKey = "iplaygodotandclaimfun"
	GeminiKey = "AIzaSyAKz6guWs938DdF_ZZDexZ72lCDljj9zOY"
)

type RequestBody struct {
	Question string `json:"question"`
	Secret   string `json:"secret"`
}

func solveHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем запрос от расширения
	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 2. Проверяем секрет
	if req.Secret != SecretKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Формируем запрос к Gemini API
	prompt := fmt.Sprintf("Ответь на вопрос теста. Выдай только букву правильного ответа (например, 'a' или 'b'): %s", req.Question)

	geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + GeminiKey

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Gemini Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 4. Возвращаем ответ обратно в расширение
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Для работы расширения
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/solve", solveHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server started on port", port)
	http.ListenAndServe(":"+port, nil)
}
