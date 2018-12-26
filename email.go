/* package justEmail provides basic functions to use SMTP in a way that most people would.  As HTML, over TLS, and also retrying if it sends on a disconnected client */
package justEmail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
  log "github.com/autopogo/justLogging"
	"errors"
)

var (
 ErrServerUnavailable = errors.New("justEmail: Couldn't connect to the relay server")
 ErrBadConfig = errors.New("justEmail: Tried to initialize SMTP w/o required parameters set")
)

// The Mail struct allows you to build emails easily
type Mail struct {
	Sender	 string
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	Body     string
}

// buildMessage turns a Mail struct into an email. ... does text/plain cover html? i don't know... tutorials ugh
func (mail *Mail) buildMessage() string {
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	header := ""
	header += fmt.Sprintf("From: %s\r\n", mail.Sender)
	if len(mail.To) > 0 {
		header += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	}
	if len(mail.Cc) > 0 {
		header += fmt.Sprintf("Cc: %s\r\n", strings.Join(mail.Cc, ";"))
	}

	header += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	header += fmt.Sprintf("%s\r\n", mime)
	header += "\r\n" + mail.Body

	return header
}

// SmtpServer is the configuration type for the email package
type SmtpServer struct {
	Host string
	Port string
	Client *smtp.Client
	ClientDomain string
}

// ServerName just concatenates the hostname and portnumber
func (s *SmtpServer) serverName() string {
	return s.Host + ":" + s.Port
}

// StartServer just connects to an email server
func (s *SmtpServer) StartServer() error {
	if s.Host == "" || s.Port == "" || s.ClientDomain == "" {
		log.Errorf("Trying to start a server with an empty initialization string")
		return ErrBadConfig
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Host,
	}
	var err error

	s.Client, err = smtp.Dial(s.serverName())
	if err != nil {
		log.Errorf("justSQL, .StartServer(): Error on dial: %v", err)
		return err
	}

	err = s.Client.Hello(s.ClientDomain)
	if err != nil {
		log.Errorf("justSQL, .StartServer(): Error on hello: %v", err)
		return err
	}

	err = s.Client.StartTLS(tlsconfig)
	if err != nil {
		log.Errorf("justSQL, .StartServer(): Error on startTLS: %v", err)
		return err
	}
	return nil
}

// SendMail sends an email TODO: add return error
func (s *SmtpServer) SendMail(mail *Mail) (err error) {
	messageBody := mail.buildMessage()
	var messageSent bool = false
	var messageRetry uint8 = 3
	defer func() {
		if (messageRetry == 0) {
			log.Errorf("We failed to send an email: %v!", mail)
			err = ErrServerUnavailable
		}
	}()
	handleError := func() {
		log.Warningf("justSql, .SendMail(): Trying to restart mail server, %v-rsth time", 3-messageRetry)
		if err := s.StartServer(); err != nil {
			log.Errorf("justSql, .SendMail(): ServerStart error on restarted: %v", err)
		}
		messageRetry -= 1
		if (messageRetry == 0) {
			err = ErrServerUnavailable
		}
	}

	for ((!messageSent) && (messageRetry != 0)){
		err := s.Client.Mail(mail.Sender)
		if err != nil {
			log.Errorf("justSQL, .SendMail(): Error on Mail", err)
			handleError()
			continue
		}

		receivers := append(mail.To, mail.Cc...) // NOTE: receivers may or may not be pointing to mail.To
		receivers  = append(receivers, mail.Bcc...)

		for _, k := range receivers {
			if err = s.Client.Rcpt(k); err != nil {
				log.Errorf("justSQL, .SendMail: Error on Client.Rcpt", err)
				handleError()
				continue
			}
		}

		// Data
		w, err := s.Client.Data()
		if err != nil {
			log.Errorf("justSQL, .SendMail: Error on Data", err)
			handleError()
			continue
		}

		_, err = w.Write([]byte(messageBody))
		if err != nil {
			log.Errorf("justSQL, .SendMail: Error on Write", err)
			handleError()
			continue
		}

		err = w.Close()
		if err != nil {
			log.Errorf("justSQL, .SendMail: Error on Close", err)
		}
		messageSent = true
	}
	return nil
}

// Quit disconnects an SmtpServer
func (s *SmtpServer) Quit() {
	s.Client.Quit()
}
