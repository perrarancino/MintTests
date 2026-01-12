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
	GeminiKey = "AIzaSyAKz6guWs938DdF_ZZDexZ72lCDljj9zOY"
	GroqKey   = "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"
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
	json.NewDecoder(r.Body).Decode(&req)

	if req.Secret != SecretKey {
		http.Error(w, "Unauthorized", 401)
		return
	}

	// 1. ПРОБУЕМ GEMINI
	fmt.Println("Пробую Gemini...")
	ans, err := callGemini(req.Question)

	// 2. ЕСЛИ GEMINI ВЫДАЛ 429 ИЛИ ОШИБКУ — ПРОБУЕМ GROQ
	if err != nil || ans == "" {
		fmt.Println("Gemini спит (429), переключаюсь на Groq (Llama 3)...")
		ans, err = callGroq(req.Question)
	}

	if err != nil {
		http.Error(w, "Все нейронки заняты", 429)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Возвращаем ответ в формате, который ждет расширение
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidates": []map[string]interface{}{
			{"content": map[string]interface{}{
				"parts": []map[string]string{{"text": ans}},
			}},
		},
	})
}

// Функция для Gemini
func callGemini(q string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + GeminiKey
	payload := map[string]interface{}{
		"contents": []interface{}{map[string]interface{}{"parts": []interface{}{map[string]string{"text": "Give only the answer text: " + q}}}},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("error")
	}
	defer resp.Body.Close()

	var res struct {
		Candidates []struct {
			Content struct {
				Parts []struct{ Text string }
			}
		}
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if len(res.Candidates) > 0 {
		return res.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("empty")
}

// Функция для Groq (Llama 3)
func callGroq(q string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"
	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile", // Очень мощная и быстрая модель
		"messages": []interface{}{
			map[string]string{"role": "system", "content": "You are a test solver. Output ONLY the correct answer text."},
			map[string]string{"role": "user", "content": q},
		},
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+GroqKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("groq error")
	}
	defer resp.Body.Close()

	var res struct {
		Choices []struct {
			Message struct{ Content string }
		}
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
