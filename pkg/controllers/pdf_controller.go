package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PdfController struct {
	db *gorm.DB
}

func NewPdfController(db *gorm.DB) *PdfController {
	return &PdfController{db: db}
}

func (pc *PdfController) GetSalesReportPDF(c *gin.Context) {
	pdfIdstring := c.Param("pdfID")

	pdfId, err := strconv.Atoi(pdfIdstring)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For this example, we'll assume the file path is "./pdfs/example.pdf"
	pdfPath := filepath.Join("economy", fmt.Sprintf("sales_report-%d.pdf", pdfId))

	data, err := os.ReadFile(pdfPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error reading the PDF file"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="`+pdfPath+`"`)

	c.Data(http.StatusOK, "application/pdf", data)
}
