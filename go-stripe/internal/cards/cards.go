package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
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
