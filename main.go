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

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	// Путь для проверки доступных моделей
	http.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + GeminiKey
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	})

	// Основной путь решения
	http.HandleFunc("/solve", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			Question string `json:"question"`
			Secret   string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad JSON", 400)
			return
		}

		if req.Secret != SecretKey {
			http.Error(w, "Wrong Secret", 401)
			return
		}

		// ПРОБУЕМ САМЫЙ КЛАССИЧЕСКИЙ URL
		geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + GeminiKey

		prompt := fmt.Sprintf("Ты — профессиональный помощник по тестам. Проанализируй вопрос и варианты ответов. Выдай ТОЛЬКО текст правильного ответа. Вопрос: %s", req.Question)

		payload := map[string]interface{}{
			"contents": []interface{}{
				map[string]interface{}{
					"parts": []interface{}{
						map[string]string{"text": prompt},
					},
				},
			},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			http.Error(w, "Google Error", 500)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Заглушка для главной
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Server is UP. Use /models to check API or /solve for tasks.")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
