package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const (
	templateLocation  string = "templates"
	templateExtension string = ".html"
	errorTemplateName string = "error"
)

var (
	FuncMap template.FuncMap
)

type ErrorPageData struct {
	Title        string
	Error        error
	TemplateName string
	Data         interface{}
	// TODO: Figure out how to make this generic
	LoggedInPlayer struct {
		Name  string
		Email string
		ID    int
	}
}

func init() {
	FuncMap = template.FuncMap{}
}

func Page(rw http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	log.Printf("Rendering: %s\n", templateName)
	FuncMap["toggle"] = toggle(true)

	tmpl, err := template.New("base").Funcs(FuncMap).ParseFiles(
		filepath.Join(templateLocation, "layouts", "base"+templateExtension),
		filepath.Join(templateLocation, templateName+templateExtension),
	)

	if err != nil {
		Error(rw, r, err, templateName, data)
		return
	}
	err = tmpl.ExecuteTemplate(rw, "base", data)
	if err != nil {
		Error(rw, r, err, templateName, data)
		return
	}
}

func Error(rw http.ResponseWriter, r *http.Request, err error, templateName string, data interface{}) {
	log.Printf("Rendering Error: %s\n", err)
	if templateName == errorTemplateName {
		rw.Write([]byte(fmt.Sprintf("Error rendering error. Oops. %s", err)))
		return
	}
	Page(rw, r, errorTemplateName, ErrorPageData{
		Title:        "Error!",
		Error:        err,
		TemplateName: templateName,
		Data:         data,
	})
}

func JSON(rw http.ResponseWriter, r *http.Request, responseData interface{}) {
	log.Printf("Rendering JSON")

	rw.Header().Set("Content-Type", "application/json")

	jsonOutput, err := json.Marshal(responseData)
	if err != nil {
		InternalErrorJSON(rw, r, responseData, err)
		return
	}
	rw.Write(jsonOutput)
}

func InternalErrorJSON(rw http.ResponseWriter, r *http.Request, responseData interface{}, err error) {
	log.Printf("Rendering internal error JSON: %s", err)

	http.Error(rw, err.Error(), http.StatusInternalServerError)

	data := make(map[string]interface{})
	data["success"] = false
	data["reason"] = err.Error()
	data["data"] = responseData
	rw.Header().Set("Content-Type", "application/json")

	jsonOutput, err := json.Marshal(data)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf("Error writing json error: %s", err.Error())))
		return
	}
	rw.Write(jsonOutput)
}

func UserErrorJSON(rw http.ResponseWriter, r *http.Request, responseData map[string]interface{}, err error) {
	if responseData == nil {
		responseData = make(map[string]interface{})
	}
	log.Printf("Rendering internal error JSON: %s", err)

	http.Error(rw, err.Error(), http.StatusBadRequest)

	responseData["success"] = false
	responseData["reason"] = err.Error()
	rw.Header().Set("Content-Type", "application/json")

	jsonOutput, err := json.Marshal(responseData)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf("Error writing json error: %s", err.Error())))
		return
	}
	rw.Write(jsonOutput)
}

func toggle(first bool) func() bool {
	return func() bool {
		first = !first
		return first == false
	}
}
