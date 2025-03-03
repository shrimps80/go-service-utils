package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"sync"
)

// MailConfig 邮件客户端配置
type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// MailTemplate 邮件模板
type MailTemplate struct {
	Subject string
	Body    string
}

// MailClient 邮件客户端
type MailClient struct {
	config    *MailConfig
	templates sync.Map
}

// NewMailClient 创建邮件客户端
func NewMailClient(config *MailConfig) *MailClient {
	return &MailClient{
		config: config,
	}
}

// AddTemplate 添加邮件模板
func (c *MailClient) AddTemplate(name string, tmpl *MailTemplate) {
	c.templates.Store(name, tmpl)
}

// GetTemplate 获取邮件模板
func (c *MailClient) GetTemplate(name string) (*MailTemplate, bool) {
	tmpl, ok := c.templates.Load(name)
	if !ok {
		return nil, false
	}
	return tmpl.(*MailTemplate), true
}

// SendMail 发送邮件
func (c *MailClient) SendMail(to []string, templateName string, data interface{}) error {
	// 获取模板
	tmpl, ok := c.GetTemplate(templateName)
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	// 解析主题
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return err
	}
	var subject bytes.Buffer
	if err := subjectTmpl.Execute(&subject, data); err != nil {
		return err
	}

	// 解析正文
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	if err := bodyTmpl.Execute(&body, data); err != nil {
		return err
	}

	// 构建邮件头
	header := make(map[string]string)
	header["From"] = c.config.From
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject.String()
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body.String()

	// 配置SMTP客户端
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)

	// 发送邮件
	if c.config.UseTLS {
		// TLS配置
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.config.Host,
		}

		// 连接TLS
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, c.config.Host)
if err != nil {
	return err
}
		defer client.Close()

		// 认证
		if err = client.Auth(auth); err != nil {
			return err
		}

		// 设置发件人和收件人
		if err = client.Mail(c.config.From); err != nil {
			return err
		}
		for _, addr := range to {
			if err = client.Rcpt(addr); err != nil {
				return err
			}
		}

		// 发送邮件内容
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return client.Quit()
	} else {
		return smtp.SendMail(addr, auth, c.config.From, to, []byte(message))
	}
}