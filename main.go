package main

import (
	"os"
	"os/signal"

	"go-mod.ewintr.nl/matrix-kagisum/bot"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	kagi := bot.NewKagi("https://kagi.com/api/v0", getParam("KAGI_API_KEY", ""))

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

	ks := bot.NewBot(mtrxConf, kagi, logger)
	if err := ks.Init(); err != nil {
		logger.Error("error running matrix ks: %v", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go ks.Run()
	defer ks.Close()
	logger.Info("matrix kagisum started")

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)
	<-done

	logger.Info("matrix kagisum stopped")

}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
