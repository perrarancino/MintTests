package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
)

const SecretKey = "iplaygodotandclaimfun"

// ТВОИ КЛЮЧИ (Вставь сюда 3 разных ключа из Google AI Studio)
var keys = []string{
	"AIzaSyAKz6guWs938DdF_ZZDexZ72lCDljj9zOY", // Твой первый ключ
	"AIzaSyDdgyihZD8DJIMfXcl6zxriHbSS5NVyVow",
	"AIzaSyBkxw_ZDhW8GTERy6uipCCZK39fxjoH790",
}

// Счетчик для переключения ключей
var currentKeyIdx uint32

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
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

		// ВЫБИРАЕМ КЛЮЧ (циклически переключаемся 0 -> 1 -> 2 -> 0)
		idx := atomic.AddUint32(&currentKeyIdx, 1) % uint32(len(keys))
		selectedKey := keys[idx]

		// Используем 2.5-flash (или 2.0-flash-lite для еще большей стабильности)
		// В main.go замени строку с URL на эту:
		geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + selectedKey

		fmt.Printf("Использую ключ №%d\n", idx+1)

		// В файле main.go замени строку prompt на эту:

		prompt := fmt.Sprintf(`ИНСТРУКЦИЯ: Ты — автоматический решатель тестов. 
		Твоя задача: прочитать вопрос и варианты, затем выбрать ОДИН правильный.
		ВЫХОДНЫЕ ДАННЫЕ: Выдай ТОЛЬКО текст правильного ответа. 
		ЗАПРЕЩЕНО: Писать пояснения, вступления или использовать знаки препинания в конце, если их нет в ответе.
		ВОПРОС: %s`, req.Question)
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is UP. Using %d keys.", len(keys))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
