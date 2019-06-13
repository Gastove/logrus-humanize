package main

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/Gastove/humanize/pkg/humanize"
)

func main() {
	formatter := humanize.NewHumanizeFormatter()
	logrus.SetFormatter(formatter)
	// logrus.SetOutput(os.Stdout)

	err := errors.New("oh heavens oh no an error eep")

	logger := logrus.WithFields(logrus.Fields{
		"power_level": 9000,
		"dance":       "flhargunstow",
	})

	fmt.Println("\n// --- Long Format --- //")
	logger.Info("This is the very polite log message")

	logger.WithError(err).Error("Alas, error city!")

	formatter.Compact = true

	fmt.Println("\n\n// --- Compact Format --- //")
	logger.Info("Now, compact!")

	logger.WithError(err).Error("This is a very compact error message")
}
