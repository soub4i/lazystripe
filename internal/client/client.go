package client

import (
	"context"

	stripe "github.com/stripe/stripe-go/v84"
	accountpkg "github.com/stripe/stripe-go/v84/account"
	balancepkg "github.com/stripe/stripe-go/v84/balance"
	chargepkg "github.com/stripe/stripe-go/v84/charge"
	customerpkg "github.com/stripe/stripe-go/v84/customer"
	productpkg "github.com/stripe/stripe-go/v84/product"
)

type Client struct {
	apiKey string
	sc     *stripe.Client
}

func New(apiKey string) *Client {
	stripe.Key = apiKey
	sc := stripe.NewClient(apiKey)

	return &Client{
		apiKey: apiKey,
		sc:     sc,
	}
}

func (c *Client) APIKey() string {
	return c.apiKey
}

func (c *Client) GetBalance(ctx context.Context) (*stripe.Balance, error) {
	return balancepkg.Get(&stripe.BalanceParams{})
}

func (c *Client) ListCustomers(ctx context.Context, params *stripe.CustomerListParams) *customerpkg.Iter {
	return customerpkg.List(params)
}

func (c *Client) ListCharges(ctx context.Context, params *stripe.ChargeListParams) *chargepkg.Iter {
	return chargepkg.List(params)
}

func (c *Client) ListProducts(ctx context.Context, params *stripe.ProductListParams) *productpkg.Iter {
	return productpkg.List(params)
}

func (c *Client) GetAccount() (*stripe.Account, error) {
	return accountpkg.Get()
}
