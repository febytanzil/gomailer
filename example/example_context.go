package example

import (
	"context"
	"fmt"
	"github.com/febytanzil/gomailer"
	"log"
	"os"
	"os/signal"
	"time"
)

func main_ctx() {
	c, _ := gomailer.NewClient(gomailer.Gomail, &gomailer.Config{
		Port:     587,
		Host:     "smtp.gmail.com",
		Email:    "user@email.com",
		Password: "user_password",
	})

	ticker := time.NewTicker(time.Second)
	go func() {
		for t := range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*50)
			ftr := make(chan struct{})
			go func(ctx context.Context, ch chan struct{}) {
				select {
				case <-ctx.Done():
					log.Println("context timeout")
				default:
					err := c.SendContext(ctx, &gomailer.Message{
						Body:   "body" + t.String(),
						Title:  "test",
						SendTo: []string{"receiver@mail.com"},
					})
					if nil != err {
						fmt.Println("err from client sendContext", err)
						close(ch)
						return
					}
					ch <- struct{}{}
				}
				close(ch)
			}(ctx, ftr)
			select {
			case <-ctx.Done():
				fmt.Println("timeout", t.String(), ctx.Err())
			case <-ftr:
				fmt.Println("success")
			}
			cancel()
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
