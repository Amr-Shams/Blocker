package main

import (
	//"os"
    "html/template"
    "io"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
  //  "github.com/Amr-Shams/Blocker/cmd"
)

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate() *Template {
    return &Template{
        templates: template.Must(template.ParseGlob("views/*.html")),
    }
}
type User struct {
    Name  string
    Email string
}
func main() {
    e:= echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Renderer = NewTemplate()
    e.GET("/", func(c echo.Context) error {
        return c.Render(200, "index", nil)
    })
    e.Logger.Fatal(e.Start(":8080"))
}
