package justEmail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
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

// smtpServer := SmtpServer{host: "smtp-relay.gmail.com", port: "587", clientDomain: "autopogo.com"}

// SmtpServer is the configuration type for the email package
type SmtpServer struct {
	Host string
	Port string
	Client *smtp.Client
	ClientDomain string
}

// ServerName just concatenates the hostname and portnumber
func (s *SmtpServer) ServerName() string {
	return s.Host + ":" + s.Port
}

// StartServer just connects to an email server
func (s *SmtpServer) StartServer() {
	if s.Host == "" || s.Port == "" || s.ClientDomain == "" {
		log.Panic("StartServer called on uninitialized structure")
	}
	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Host,
	}
	var err error

	s.Client, err = smtp.Dial(s.ServerName())
	if err != nil {
		log.Panic(err)
	}

	err = s.Client.Hello(s.ClientDomain)
	if err != nil {
		log.Panic(err)
	}

	err = s.Client.StartTLS(tlsconfig)
	if err != nil {
		log.Panic(err)
	}
}

// SendMail sends an email
func (s *SmtpServer) SendMail(mail *Mail) {

	messageBody := mail.buildMessage()

	err := s.Client.Mail(mail.Sender)
	if err != nil {
		log.Panic(err)
	}

	receivers := append(mail.To, mail.Cc...)
	receivers  = append(receivers, mail.Bcc...)

	for _, k := range receivers {
		if err = s.Client.Rcpt(k); err != nil {
			log.Panic(err)
		}
	}

	// Data
	w, err := s.Client.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(messageBody))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}
}

// Quit disconnects an SmtpServer
func (s *SmtpServer) Quit() {
	s.Client.Quit()
}
