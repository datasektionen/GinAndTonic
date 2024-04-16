package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/types"
)

func ParseTemplate(templateFile string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	htmlContent := buf.String()
	// The replacement of "\x00" might not be necessary with "html/template"
	// htmlContent = strings.ReplaceAll(htmlContent, "\x00", "")

	return htmlContent, nil
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
