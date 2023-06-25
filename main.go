package main

import (
	"ewintr.nl/matrix-kagisum/bot"
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"os/signal"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	apiKey := getParam("KAGI_API_KEY", "")
	fmt.Println(apiKey)

	kagi := bot.NewKagi("https://kagi.com/api/v0", apiKey)

	//res, err := kagi.Summarize("https://ewintr.nl/shitty-ssg/why-i-built-my-own-shitty-static-site-generator/")
	//if err != nil {
	//	fmt.Println(err)
	//}

	mtrxConf := bot.MatrixConfig{
		Homeserver:    getParam("MATRIX_HOMESERVER", "http://localhost/"),
		UserID:        getParam("MATRIX_USER_ID", "@user:localhost"),
		UserAccessKey: getParam("MATRIX_USER_ACCESS_KEY", "secret"),
		UserPassword:  getParam("MATRIX_USER_PASSWORD", "secret"),
		RoomID:        getParam("MATRIX_ROOM_ID", "!room:localhost"),
		DBPath:        getParam("MATRIX_DB_PATH", "matrix.db"),
		Pickle:        getParam("MATRIX_PICKLE", "matrix.pickle"),
		AcceptInvites: getParam("MATRIX_ACCEPT_INVITES", "false") == "true",
	}

	bot := bot.NewBot(mtrxConf, kagi, logger)
	if err := bot.Init(); err != nil {
		logger.Error("error running matrix bot: %v", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go bot.Run()
	defer bot.Close()
	logger.Info("matrix bot started")

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)
	<-done

	logger.Info("matrix bot stopped")

}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
