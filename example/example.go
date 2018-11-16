package example

import (
	"fmt"
	"github.com/febytanzil/gomailer"
	"os"
	"os/signal"
	"time"
)

func main() {
	c, _ := gomailer.NewClient(gomailer.Gomail, &gomailer.Config{
		Port:     587,
		Host:     "smtp.gmail.com",
		Email:    "user@email.com",
		Password: "user_password",
	})

	ticker := time.NewTicker(time.Minute)
	go func() {
		for t := range ticker.C {
			err := c.Send(&gomailer.Message{
				Body:   "body" + t.String(),
				Title:  "test",
				SendTo: []string{"receiver@mail.com"},
			})
			fmt.Println(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(ch, os.Interrupt)

	// Block until we receive our signal.
	<-ch
	c.Close()
	fmt.Println("shutting down")
	os.Exit(0)
}
