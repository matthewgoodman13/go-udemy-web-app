package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	CreatedAt time.Time `json:"created_at"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// CreateAndSendInvoice creates and sends an email with an invoice
func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {

	// Receive JSON from request
	var order Order

	err := app.readJSON(w, r, &order)
	if err != nil {
		app.badRequest(w, r, err)
		app.errorLog.Println("Error reading JSON")
		return
	}

	// Generate PDF invoice
	err = app.createInvoicePDF(order)
	if err != nil {
		app.badRequest(w, r, err)
		app.errorLog.Println("Error creating PDF invoice")
		return
	}

	// Create Mail
	attachments := []string{fmt.Sprintf("./invoices/%d.pdf", order.ID)}

	// Send email with PDF invoice
	err = app.SendMail("info@widgets.com", order.Email, "Your Invoice", "invoice", attachments, nil)
	if err != nil {
		app.badRequest(w, r, err)
		app.errorLog.Println("Error sending email")
		return
	}

	// Send Response
	var res struct {
		Message string `json:"message"`
		Error   bool   `json:"error"`
	}
	res.Error = false
	res.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)

	err = app.writeJSON(w, http.StatusOK, res)
	if err != nil {
		app.badRequest(w, r, err)
	}
}

// createInvoicePDF
func (app *application) createInvoicePDF(order Order) error {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	pdf.SetAutoPageBreak(true, 0)

	importer := gofpdi.NewImporter()

	t := importer.ImportPage(pdf, "./pdf-templates/invoice.pdf", 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	// Write Bill-To Information
	pdf.SetX(10)
	pdf.SetY(50)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(97, 8, fmt.Sprintf("%s %s", order.FirstName, order.LastName), "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.Email, "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.CreatedAt.Format("2006-01-02"), "", 0, "L", false, 0, "")

	// Add Products to Invoice Table
	pdf.SetX(58)
	pdf.SetY(93)
	pdf.CellFormat(155, 8, order.Product, "", 0, "L", false, 0, "")
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", order.Quantity), "", 0, "C", false, 0, "")
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$%.2f", float32(order.Amount/100.0)), "", 0, "R", false, 0, "")

	// Save PDF
	invoicePath := fmt.Sprintf("./invoices/%d.pdf", order.ID)
	err := pdf.OutputFileAndClose(invoicePath)
	if err != nil {
		app.errorLog.Println("Error saving PDF")
		return err
	}
	return nil
}
