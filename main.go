package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const (
	SecretKey = "iplaygodotandclaimfun"
	GroqKey   = "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"
)

// В main.go замени список моделей на этот мощный состав:
var models = []string{
	"qwen/qwen3-32b",              // Лидер по точности и RPM (60)
	"moonshotai/kimi-k2-instruct", // Лидер по RPM (60) и логике
	"gpt-oss-120b",                // Самая умная в коде и алгоритмах (120B параметров)
	"llama-3.3-70b-versatile",     // Стабильный универсал
	"llama-3.1-8b-instant",        // Страховка по дневному лимиту (14.4K)
}

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
	json.NewDecoder(r.Body).Decode(&req)

	if req.Secret != SecretKey {
		http.Error(w, "Unauthorized", 401)
		return
	}

	// Возвращаем промпт, который просит полный текст ответа
	systemPrompt := "Ты — решатель тестов. Выбери правильный ответ. Выдай ТОЛЬКО текст правильного ответа целиком, как он написан в вариантах. Никаких пояснений и цифр в начале."

	var wg sync.WaitGroup
	results := make(chan string, len(models))

	for _, model := range models {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			ans, _ := callGroq(m, systemPrompt, req.Question)
			if ans != "" {
				results <- ans
			}
		}(model)
	}

	wg.Wait()
	close(results)

	votes := make(map[string]int)
	for ans := range results {
		votes[ans]++
	}

	var finalAnswer string
	maxVotes := 0
	for ans, count := range votes {
		if count > maxVotes {
			maxVotes = count
			finalAnswer = ans
		}
	}

	fmt.Printf("Итог голосования: %s\n", finalAnswer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]string{{"text": finalAnswer}},
			}},
		},
	})
}

func callGroq(model, system, user string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"
	payload := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]string{"role": "system", "content": system},
			map[string]string{"role": "user", "content": user},
		},
		"temperature": 0.1,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+GroqKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Choices []struct{ Message struct{ Content string } }
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if len(res.Choices) > 0 {
		return res.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("empty")
}

func main() {
	http.HandleFunc("/solve", solveHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
