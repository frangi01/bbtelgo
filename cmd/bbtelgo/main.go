package main

import (
	"fmt"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/logx"
)

func main() {

	logger, err := logx.New("logs/bot.log", logx.Options{
		Level:      logx.Debug, // minimum level to print
		MaxSizeMB:  10,         // rotate after ~10MB (0 to disable)
		IncludeSrc: true,       // show file:line
		// TimeFormat: time.RFC3339, // or customize
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	
	_, err = config.Load(logger)
	if err != nil {
		logger.Errorf("config - env")
		return
	}
	

	fmt.Println("hello, world go!")
}
