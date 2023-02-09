package main

import (
	"fmt"
	"myapp/internal/cards"
	"myapp/internal/encryption"
	"myapp/internal/models"
	"myapp/internal/urlsigner"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Displays the virtual terminal page
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

type TransactionData struct {
	FirstName       string
	LastName        string
	Cardholder      string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// GetTransactionData reads the posted data and stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// Read Posted Data
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	cardholder := r.Form.Get("cardholder_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastfour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// Can add validation here if needed!

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Cardholder:      cardholder,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastfour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return txnData, nil
}

// VirtualTerminalPaymentSucceeded displays receipt page for virtual terminal transactions
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create New Transaction
	transaction := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}
	_, err = app.SaveTransaction(transaction)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Write data to session and redirect user to new page
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// PaymentSucceeded displays receipt page for store checkout transactions
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.errorLog.Println(err)
		return
	}

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	widgetID, _ := strconv.Atoi(r.Form.Get("product_id")) // Ignoring error!

	// Create New Customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create New Transaction

	transaction := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}
	transactionID, err := app.SaveTransaction(transaction)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create New Order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: transactionID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Write data to session and redirect user to new page
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

// Displays the receipt page for virtual terminal transactions
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Pop(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}
}

// Displays the receipt page for store checkout transactions
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Pop(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	if err := app.renderTemplate(w, r, "receipt", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a customer to the database and returns the ID
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveTransaction saves a transaction to the database and returns the ID
func (app *application) SaveTransaction(transaction models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(transaction)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveOrder saves an order to the database and returns the ID
func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Displays page to buy one widget
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{Data: data}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// Displays page to buy Bronze plan subscription
func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {

	widget, err := app.DB.GetWidget(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{Data: data}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// Displays the receipt page for subscription transactions
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Authentication Handlers
// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// PostLoginPage handles the login form submission
func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, err := app.DB.Authenticate(email, password)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles the logout request
func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ForgotPassword displays the forgot password page
func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// ShowResetPassword displays the reset password page
func (app *application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, theURL)
	email := r.URL.Query().Get("email")

	// Check if the URL is valid and not expired
	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}
	if !signer.VerifyToken(testURL) {
		app.errorLog.Println("Invalid URL - tampering detected")
		return
	}
	if signer.Expired(testURL, 60) {
		app.errorLog.Println("URL expired")
		return
	}

	// Encrypt the email
	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretkey),
	}
	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["email"] = encryptedEmail

	if err := app.renderTemplate(w, r, "reset-password", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}
}

// AllSales displays all sales
func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-sales", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// AllSubscriptions displays all subscriptions
func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-subscriptions", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// ShowSale displays a sale
func (app *application) ShowSale(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Sale"
	stringMap["return_url"] = "/admin/all-sales"

	if err := app.renderTemplate(w, r, "sale", &templateData{StringMap: stringMap}); err != nil {
		app.errorLog.Println(err)
	}
}

// ShowSubscription displays a subscription
func (app *application) ShowSubscription(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Subscription"
	stringMap["return_url"] = "/admin/all-subscriptions"

	if err := app.renderTemplate(w, r, "sale", &templateData{StringMap: stringMap}); err != nil {
		app.errorLog.Println(err)
	}
}
