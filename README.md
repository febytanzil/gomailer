# gomailer
wrapper golang email with automatic reconnection

## Intentions & Features
- Based on https://gopkg.in/gomail.v2
- Auto reconnection broken pipe from SMTP server
- Avoid too many login attempts because of redial & close
- Unit-testable/ mock-able with gomock provided

## Usage
Example is provided in `example/` folder/ package
```
c, _ := gomailer.NewClient(gomailer.Gomail, &gomailer.Config{
  Port:     587,
  Host:     "smtp.gmail.com",
  Email:    "user@email.com",
  Password: "user_password",
})
  
err := c.Send(&gomailer.Message{
  Body:   "body" + t.String(),
  Title:  "test",
  SendTo: []string{"receiver@mail.com"},
})
```

## Offline Contributors
[erosaru](https://github.com/erosaru)
