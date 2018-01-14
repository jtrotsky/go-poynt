package server

import (
	"fmt"
	"net/http"
	"text/template"
)

// Stock error page for failed render of HTML
var errorTemplate = `
<html>
    <body>
          <h1>Error rendering template %s</h1>
	  			<p>%s</p>
					<p>Valid templates are %s</p>
		</body>
</html>
`

// Index any .html files within a subdirectory of the templates directory.
var templates = template.Must(template.New("t").ParseGlob("server/templates/*.html"))

// RenderTemplate executes an HTML template from our templates glob.
func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(errorTemplate, name, err, *templates),
			http.StatusInternalServerError)
	}
}
