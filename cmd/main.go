package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/ekediala/fem-htmx-proj/views"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type headers struct {
	key   string
	value string
}

type Data struct {
	contacts []views.Contact
	count    views.Count
}

type Page struct {
	Data     Data
	FormData views.FormData
}

func newData() Data {
	return Data{
		contacts: []views.Contact{
			createContact("Alice", "wetana3835@dabeixin.com"),
			createContact("Bob", "hogomag615@abudat.com"),
		},
		count: views.Count{Count: 0},
	}
}

func (d Data) hasEmail(email string) bool {
	for _, contact := range d.contacts {
		if strings.ToLower(contact.Email) == strings.ToLower(email) {
			return true
		}
	}
	return false
}

func newPage() Page {
	return Page{
		Data:     newData(),
		FormData: views.NewFormData(),
	}
}

func render(c *fiber.Ctx, component templ.Component, status int, h ...headers) error {
	componentHandler := templ.Handler(component)
	componentHandler.Status = status
	if len(h) > 0 {
		for _, header := range h {
			c.Set(header.key, header.value)
		}
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

func createContact(name, email string) views.Contact {
	contact := views.Contact{
		Name:  name,
		Email: email,
	}

	return contact
}

func main() {
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "${pid} ${locals:requestid} [${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	page := newPage()
	data := page.Data
	fd := page.FormData

	app.Post("/count", func(c *fiber.Ctx) error {
		data.count.Count++
		return render(c, views.Counter(data.count), http.StatusOK)
	})

	app.Get("/count", func(c *fiber.Ctx) error {
		return render(c, views.Index(data.count), http.StatusOK)
	})

	app.Post("/contacts", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		name := c.FormValue("name")
		if data.hasEmail(email) {
			fd := views.NewFormData()
			fd.Values["name"] = name
			fd.Values["email"] = email

			fd.Errors["email"] = "Email already exists"

			// h := []headers{
			// 	{key: "HX-Retarget", value: "#form"},
			// 	{key: "HX-Reswap", value: "outerHTML"},
			// }

			// I personally prefer retargetting and reswapping because it doesn't require any further scripting
			// from me but to us the beforeSwap event on the browser for 422 errors we use
			return render(c, views.ContactForm(fd), http.StatusUnprocessableEntity)
			// return render(c, views.ContactForm(fd), http.StatusOK, h...)
		}
		contact := createContact(name, email)
		data.contacts = append(data.contacts, contact)
		// we can just do this but we want to do out of band updates
		// return render(c, views.SingleContact(contact), http.StatusOK)
		return render(c, views.OObContact(contact), http.StatusOK)
	})

	app.Get("/contacts", func(c *fiber.Ctx) error {
		return render(c, views.Contacts(data.contacts, fd), http.StatusOK)
	})

	app.Static("/", "./public")

	log.Fatal(app.Listen(":4222"))
}
