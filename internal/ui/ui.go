package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	stripeclient "github.ibm.com/soub4i/lazystripe/internal/client"

	"github.ibm.com/soub4i/lazystripe/internal/screens"
)

type App struct {
	app    *tview.Application
	pages  *tview.Pages
	menu   *tview.List
	footer *tview.TextView
	stripe *stripeclient.Client
}

func Run(apiKey string) error {
	stripe := stripeclient.New(apiKey)
	app := &App{
		app:    tview.NewApplication(),
		pages:  tview.NewPages(),
		menu:   createMenu(),
		footer: nil,
		stripe: stripe,
	}

	balanceView := screens.NewBalanceView(app.app)
	customersTable := screens.NewCustomersTable(app.app)
	transactionsTable := screens.NewTransactionsTable(app.app)
	donateView := screens.NewDonateView()
	productsTable := screens.NewProductTable(app.app)

	accountName := "Account: Unknown"
	if acct, err := stripe.GetAccount(); err == nil && acct != nil {
		id := acct.ID
		if len(id) > 8 {
			id = id[:4] + "..." + id[len(id)-4:]
		}
		accountName = fmt.Sprintf("Account: %s (%s)", id, acct.BusinessType)
	}

	app.footer = createFooter(accountName)

	app.pages.AddPage("balance", balanceView, true, true)
	app.pages.AddPage("customers", customersTable, true, false)
	app.pages.AddPage("transactions", transactionsTable, true, false)
	app.pages.AddPage("products", productsTable, true, false)
	app.pages.AddPage("donate", donateView, true, false)

	app.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'm':
			app.app.SetFocus(app.menu)
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(app.menu, 30, 1, true).
		AddItem(app.pages, 0, 4, false)

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flex, 0, 1, true). // main content
		AddItem(app.footer, 3, 0, false)

	app.menu.SetSelectedFunc(func(index int, mainText, _ string, _ rune) {
		switch mainText {
		case "Balance":
			app.pages.SwitchToPage("balance")
			screens.LoadBalance(balanceView, app.stripe)
		case "Customers":
			app.pages.SwitchToPage("customers")
			screens.LoadCustomers(customersTable, app.stripe, main, true)
		case "Transactions":
			app.pages.SwitchToPage("transactions")
			screens.LoadTransactions(transactionsTable, app.stripe, main, true)
		case "Products":
			app.pages.SwitchToPage("products")
			screens.LoadProducts(productsTable, app.stripe, main, true)
		case "Donate":
			app.pages.SwitchToPage("donate")
		case "Quit":
			app.app.Stop()
		}
		app.app.SetFocus(app.pages)
	})

	screens.LoadBalance(balanceView, app.stripe)
	app.pages.SwitchToPage("balance")

	if err := app.app.SetRoot(main, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("UI run error: %w", err)
	}

	return nil
}

func createMenu() *tview.List {
	menu := tview.NewList().
		AddItem("Balance", "View current Stripe balance", 'b', nil).
		AddItem("Customers", "List customers", 'c', nil).
		AddItem("Transactions", "List charges/transactions", 't', nil).
		AddItem("Products", "List products", 'r', nil).
		AddItem("Donate", "Buy me a coffee link", 'd', nil).
		AddItem("Quit", "Exit the program", 'q', nil)
	menu.SetBorder(true).SetTitle("Lazystripe")

	return menu
}

func createFooter(acc string) *tview.TextView {
	footer := tview.NewTextView().
		SetText(fmt.Sprintf("Lazystripe - %s | (m)enu | (q)uit", acc)).
		SetTextAlign(tview.AlignCenter)
	footer.SetBorder(true)

	return footer
}
