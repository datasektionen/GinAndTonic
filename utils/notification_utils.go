package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/types"
	"github.com/vanng822/go-premailer/premailer"
)

func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	tmpl, err := template.ParseGlob("templates/emails/*.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	// Use the file name without the directory path
	templateName := filepath.Base(templateFileName)
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		fmt.Println("Error executing template", err)
		return "", err
	}

	htmlContent := buf.String()
	// The replacement of "\x00" might not be necessary with "html/template"
	// htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	prem, err := premailer.NewPremailerFromString(htmlContent, premailer.NewOptions())
	if err != nil {
		log.Fatal(err)
	}

	html, err := prem.Transform()
	if err != nil {
		log.Fatal(err)
	}

	return html, nil
}

func GenerateEmailTable(tickets []types.EmailTicket) (string, error) {
	var emailTemplate string = `<ul class="cost-summary">`

	for _, ticket := range tickets {
		emailTemplate += fmt.Sprintf("<li><span style=\"margin: 0 10px;\">%s</span><span style=\"margin: 0 10px;\">%s SEK</span></li>", ticket.Name, ticket.Price)
	}

	emailTemplate += "</ul>"

	return emailTemplate, nil
}

func CompressHTML(html string) (string, error) {
	t := template.New("t")
	t.Funcs(template.FuncMap{"compress": func(s string) string {
		return strings.Join(strings.Fields(s), " ")
	}})
	template.Must(t.Parse(`{{compress .}}`))

	var buf bytes.Buffer
	if err := t.Execute(&buf, html); err != nil {
		return "", err
	}

	return buf.String(), nil
}
