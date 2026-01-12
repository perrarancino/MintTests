package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	SecretKey = "iplaygodotandclaimfun"
	// Твой ключ Groq
	GroqKey = "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"
)

func solveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var req struct {
		Question string `json:"question"`
		Secret   string `json:"secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	if req.Secret != SecretKey {
		http.Error(w, "Unauthorized", 401)
		return
	}

	fmt.Println("Запрос к Groq (Llama 3.3)...")

	// Формируем строгий промпт
	systemPrompt := "Ты — решатель тестов. Твоя задача: выбрать правильный ответ из предложенных. " +
		"Выдай ТОЛЬКО текст самого ответа, без цифр в начале, без точек и без пояснений. " +
		"Если вариантов нет — ответь максимально кратко (одним словом)."

	ans, err := callGroq(systemPrompt, req.Question)
	if err != nil {
		fmt.Printf("Ошибка Groq: %v\n", err)
		http.Error(w, "Groq Error", 500)
		return
	}

	fmt.Printf("Ответ получен: %s\n", ans)

	w.Header().Set("Content-Type", "application/json")
	// Возвращаем в старом формате, чтобы не переписывать content.js
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]string{{"text": ans}},
			}},
		},
	})
}

func callGroq(system, user string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []interface{}{
			map[string]string{"role": "system", "content": system},
			map[string]string{"role": "user", "content": user},
		},
		"temperature": 0.1, // Минимум креативности, максимум точности
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+GroqKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var res struct {
		Choices []struct {
			Message struct{ Content string }
		}
	}
	json.NewDecoder(resp.Body).Decode(&res)

	if len(res.Choices) > 0 {
		return res.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from model")
}

func main() {
	http.HandleFunc("/solve", solveHandler)

	// Пустая страница для проверки работоспособности
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Groq Solver is Running")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Сервер запущен на порту %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
