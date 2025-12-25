package screens

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/stripe/stripe-go/v84"
	customerpkg "github.com/stripe/stripe-go/v84/customer"
	stripeclient "github.ibm.com/soub4i/lazystripe/internal/client"
)

var Transactions map[string]interface{}
var Customers map[string]interface{}
var Products map[string]interface{}

type BalanceView struct {
	*tview.TextView
	app *tview.Application
}

func NewBalanceView(app *tview.Application) *BalanceView {
	view := &BalanceView{
		TextView: tview.NewTextView().SetDynamicColors(true).SetWrap(true),
		app:      app,
	}
	view.SetBorder(true).SetTitle("Balance")
	return view
}

func LoadBalance(view *BalanceView, client *stripeclient.Client) {
	showLoading(view, "Balance")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		bal, err := client.GetBalance(ctx)
		view.app.QueueUpdateDraw(func() {
			view.Clear()
			if err != nil {
				fmt.Fprintf(view, "[red]Error: %v", err)
				return
			}
			fmt.Fprint(view, "[green]Available:\n")
			for _, a := range bal.Available {
				fmt.Fprintf(view, "  %s %d (currency: %s)\n", a.Currency, a.Amount, a.Currency)
			}
			fmt.Fprint(view, "\n[purple]Pending:\n")
			for _, p := range bal.Pending {
				fmt.Fprintf(view, "  %s %d (currency: %s)\n", p.Currency, p.Amount, p.Currency)
			}
			fmt.Fprintf(view, "\n[white]Retrieved at: %s", time.Now().Format(time.RFC1123))
		})
	}()
}

type CustomersTable struct {
	*tview.Table
	app *tview.Application
}

type ProductsTable struct {
	*tview.Table
	app *tview.Application
}

func NewCustomersTable(app *tview.Application) *CustomersTable {
	table := &CustomersTable{
		Table: tview.NewTable().SetSelectable(true, false),
		app:   app,
	}
	table.SetBorder(true).SetTitle("Customers - (p)age")
	return table
}

func LoadCustomers(table *CustomersTable, client *stripeclient.Client, root *tview.Flex, reset bool) {
	if reset {
		table.Clear()
		Customers = make(map[string]interface{})
	}
	table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("Email").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("Name").SetSelectable(false))
	row := 1

	go func() {
		params := &stripe.CustomerListParams{}
		iter := customerpkg.List(params)

		for iter.Next() {
			cust := iter.Customer()
			Customers[cust.ID] = cust
			table.app.QueueUpdateDraw(func() {
				table.SetCell(row, 0, tview.NewTableCell(cust.ID))
				table.SetCell(row, 1, tview.NewTableCell(cust.Email))
				table.SetCell(row, 2, tview.NewTableCell(cust.Name))
			})

			table.SetSelectedFunc(func(row, column int) {
				custID := table.GetCell(row, 0).Text
				if custData, ok := Customers[custID]; ok {
					if custDetails, ok := custData.(*stripe.Customer); ok {
						modal := tview.NewModal().
							SetText(fmt.Sprintf("Customer ID: %s\nEmail: %s\nName: %s\nCreated: %s",
								custDetails.ID,
								custDetails.Email,
								custDetails.Name,
								time.Unix(custDetails.Created, 0).Format(time.RFC1123))).
							AddButtons([]string{"Close"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								table.app.SetRoot(root, true).SetFocus(table)
							})
						table.app.SetRoot(modal, false).SetFocus(modal)
					}

				}
			})

			row++
		}
		if err := iter.Err(); err != nil {
			log.Printf("Customers error: %v", err)
			return
		}
	}()
}

type TransactionsTable struct {
	*tview.Table
	app *tview.Application
}

func NewTransactionsTable(app *tview.Application) *TransactionsTable {
	table := &TransactionsTable{
		Table: tview.NewTable().SetSelectable(true, false),
		app:   app,
	}
	table.SetBorder(true).SetTitle("Transactions - (p)age")
	return table
}

func LoadTransactions(table *TransactionsTable, client *stripeclient.Client, root *tview.Flex, reset bool) {
	{
		if reset {
			table.Clear()
			Transactions = make(map[string]interface{})
		}
		table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
		table.SetCell(0, 1, tview.NewTableCell("Amount").SetSelectable(false))
		table.SetCell(0, 2, tview.NewTableCell("Currency").SetSelectable(false))
		table.SetCell(0, 3, tview.NewTableCell("Status").SetSelectable(false))

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			iter := client.ListCharges(ctx, &stripe.ChargeListParams{})
			row := 1

			for iter.Next() {
				charge := iter.Charge()
				Transactions[charge.ID] = charge
				amt := float64(charge.Amount) / 100.0
				table.app.QueueUpdateDraw(func() {
					table.SetCell(row, 0, tview.NewTableCell(charge.ID))
					table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%.2f", amt)))
					table.SetCell(row, 2, tview.NewTableCell(string(charge.Currency)))
					table.SetCell(row, 3, tview.NewTableCell(string(charge.Status)))
					row++
				})
				table.SetSelectedFunc(func(row, column int) {
					chargeID := table.GetCell(row, 0).Text
					if chargeData, ok := Transactions[chargeID]; ok {
						if chargeDetails, ok := chargeData.(*stripe.Charge); ok {
							modal := tview.NewModal().
								SetText(fmt.Sprintf("Charge ID: %s\nAmount: %.2f\nCurrency: %s\nStatus: %s\nDescription: %s\nCreated: %s",
									chargeDetails.ID,
									float64(chargeDetails.Amount)/100.0,
									chargeDetails.Currency,
									chargeDetails.Status,
									chargeDetails.Description,
									time.Unix(chargeDetails.Created, 0).Format(time.RFC1123))).
								AddButtons([]string{"Close"}).
								SetDoneFunc(func(buttonIndex int, buttonLabel string) {
									table.app.SetRoot(root, true).SetFocus(table)
								})
							table.app.SetRoot(modal, false).SetFocus(modal)
						}
					}
				})
			}

			if err := iter.Err(); err != nil {
				log.Printf("Transactions error: %v", err)
				return
			}
		}()
	}
}

type DonateView struct {
	*tview.TextView
}

func NewDonateView() *DonateView {
	view := &DonateView{TextView: tview.NewTextView().SetDynamicColors(true).SetWrap(true)}

	txt := colorate("If you like this tool, buy me a coffee!", "yellow") + "\n\n" +
		colorate("Your support helps me keep improving it.", "white") + "\n\n" +
		colorate("Press Enter to open the donation link in your browser. (https://buymeacoffee.com/soubai)", "green")

	view.SetText(txt)
	view.SetBorder(true).SetTitle("Donate - (p)age")
	view.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			err := exec.Command("open", "https://buymeacoffee.com/soubai").Start()
			if err != nil {
				log.Printf("Failed to open browser: %v", err)
			}
		}
	})

	return view
}

func showLoading(view tview.Primitive, msg string) {
	if tv, ok := view.(*tview.TextView); ok {
		tv.Clear()
		fmt.Fprintf(tv, "[yellow]%s\n\n[white]Fetching...\n", msg)
	}
}

func NewProductTable(app *tview.Application) *ProductsTable {
	table := &ProductsTable{
		Table: tview.NewTable().SetSelectable(true, false),
		app:   app,
	}
	table.SetBorder(true).SetTitle("Products - (p)age")
	return table
}

func LoadProducts(table *ProductsTable, client *stripeclient.Client, root *tview.Flex, reset bool) {

	if reset {
		table.Clear()
		Products = make(map[string]interface{})
	}
	table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("Name").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("Active").SetSelectable(false))
	row := 1

	go func() {
		params := &stripe.ProductListParams{}
		cxt, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		iter := client.ListProducts(cxt, params)

		for iter.Next() {
			product := iter.Product()
			Products[product.ID] = product
			table.app.QueueUpdateDraw(func() {
				table.SetCell(row, 0, tview.NewTableCell(product.ID))
				table.SetCell(row, 1, tview.NewTableCell(product.Name))
				table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%v", product.Active)))
			})

			table.SetSelectedFunc(func(row, column int) {
				productID := table.GetCell(row, 0).Text
				if productData, ok := Products[productID]; ok {
					if productDetails, ok := productData.(*stripe.Product); ok {
						modal := tview.NewModal().
							SetText(fmt.Sprintf("Product ID: %s\nName: %s\nActive: %v\nDescription: %s\nCreated: %s",
								productDetails.ID,
								productDetails.Name,
								productDetails.Active,
								productDetails.Description,
								time.Unix(productDetails.Created, 0).Format(time.RFC1123))).
							AddButtons([]string{"Close"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								table.app.SetRoot(root, true).SetFocus(table)
							})
						table.app.SetRoot(modal, false).SetFocus(modal)
					}

				}
			})

			row++
		}
		if err := iter.Err(); err != nil {
			log.Printf("Products error: %v", err)
			return
		}
	}()
}

func colorate(text string, color string) string {
	return fmt.Sprintf("[%s]%s[white]", color, text)
}
