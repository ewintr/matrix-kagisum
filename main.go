package main

import (
	"fmt"
	"os"
)

func main() {
	//logger := slog.New(slog.NewTextHandler(os.Stderr))
	apiKey := getParam("KAGI_API_KEY", "")
	fmt.Println(apiKey)

	k := NewKagi("https://kagi.com/api/v0", apiKey)

	res, err := k.Summarize("https://ewintr.nl/shitty-ssg/why-i-built-my-own-shitty-static-site-generator/")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
