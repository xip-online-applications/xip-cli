package utils

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

func SetupSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:   "https://3443829b3dff4b068beda54dd49d0f30@o37556.ingest.sentry.io/5306873",
		Debug: false,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)
}
