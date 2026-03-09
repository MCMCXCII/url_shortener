package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/MCMCXCII/url_shortener/internal/config"
)

func main() {

	cfg := config.NewConfig()

	endpoint := cfg.BaseURL

	fmt.Println("Введите длинный URL:")

	reader := bufio.NewReader(os.Stdin)

	longURL, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	longURL = strings.TrimSpace(longURL)

	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(longURL),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	fmt.Println("Статус-код:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Короткий URL:", string(body))
}
