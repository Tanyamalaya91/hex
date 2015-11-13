package connectors

import (
	"github.com/mmcquillan/jane/models"
	"log"
	"net/smtp"
)

type Email struct {
	Connector models.Connector
}

func (x Email) Run(config *models.Config, connector models.Connector) {
	defer Recovery(config, connector)
	return
}

func (x Email) Send(config *models.Config, connector models.Connector, message models.Message, target string) {
	if target == "" {
		log.Print("No email provided to the email connector")
	} else {
		auth := smtp.PlainAuth("", connector.Login, connector.Pass, connector.Server)
		to := []string{target}
		msg := []byte("To: " + target + "\r\n" +
			"Subject: " + message.Title + "\r\n" +
			"\r\n" + message.Description + "\r\n\r\n" + message.Link + "\r\n")
		err := smtp.SendMail(connector.Server+":25", auth, connector.From, to, msg)
		if err != nil {
			log.Print(err)
		}
	}
}
