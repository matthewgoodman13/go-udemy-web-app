package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/sub"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	// Create a Payment Intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	// Can Add Own Key Value Pairs
	// params.AddMetadata("key", "value")

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

// GetPaymentMethod gets a payment method by payment intent id
func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	pm, err := paymentmethod.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// RetrievePaymentIntent gets an existing payment intent by payment intent id
func (c *Card) RetrievePaymentIntent(s string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// Subscribe to Plan
func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan, email, last4, cardType string) (*stripe.Subscription, error) {
	stripeCustomerID := cust.ID
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
	}

	params.AddMetadata("email", email)
	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("latest_invoice.payment_intent") // To get PaymentIntent of subscription

	subscription, err := sub.New(params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// Create a customer
func (c *Card) CreateCustomer(pm string, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret

	params := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return cust, "", nil
}

// Refund a payment
func (c *Card) RefundPayment(pi string, amount int) error {
	stripe.Key = c.Secret

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(pi),
		Amount:        stripe.Int64(int64(amount)),
	}

	_, err := refund.New(params)
	if err != nil {
		return err
	}

	return nil
}

// CancelSubscription cancels a subscription
func (c *Card) CancelSubscription(subscriptionID string) error {
	stripe.Key = c.Secret

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err := sub.Update(subscriptionID, params)
	if err != nil {
		return err
	}

	return nil
}

// cardErrorMessage returns a user-friendly error message for a given Stripe error code.
func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined."
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card has expired."
	case stripe.ErrorCodeIncorrectNumber:
		msg = "Your card number is incorrect."
	case stripe.ErrorCodeIncorrectZip:
		msg = "Your card's zip code failed validation."

	case stripe.ErrorCodeInvalidExpiryMonth:
		msg = "Your card's expiration month is invalid."
	case stripe.ErrorCodeInvalidExpiryYear:
		msg = "Your card's expiration year is invalid."
	case stripe.ErrorCodeInvalidNumber:
		msg = "Your card number is invalid."
	default:
		msg = "An error occurred while processing your card."
	}
	return msg
}
