package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Твои данные
const (
	SecretKey = "iplaygodotandclaimfun"
	// Твой ключ Gemini (рекомендуется хранить в переменных окружения Render, но пока оставляем здесь)
	GeminiKey = "AIzaSyAKz6guWs938DdF_ZZDexZ72lCDljj9zOY"
)

type RequestBody struct {
	Question string `json:"question"`
	Secret   string `json:"secret"`
}

func solveHandler(w http.ResponseWriter, r *http.Request) {
	// --- НАСТРОЙКА CORS ---
	// Разрешаем запросы с любого домена (включая Moodle)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Разрешаем типы запросов
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	// Разрешаем передачу заголовка Content-Type
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Если браузер прислал "пробный" запрос (OPTIONS), отвечаем 200 OK и выходим
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// Добавляем инструкцию, чтобы AI не писал лишнего текста
	prompt := fmt.Sprintf("Ты — помощник по тестам. Проанализируй вопрос и варианты. Выдай ТОЛЬКО текст правильного ответа или его букву. Вопрос: %s", req.Question)

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

	// Отправляем запрос в Google
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Gemini API Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 4. Возвращаем ответ обратно в расширение
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func main() {
	// Обработчик для пути /solve
	http.HandleFunc("/solve", solveHandler)

	// Стандартный обработчик для главной страницы (чтобы Render видел, что сервис жив)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
