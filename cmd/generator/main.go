package main

//
// func mustTemplate(name string) *template.Template {
// 	path := "web/templates/" + name
// 	return template.Must(template.New(name).Funcs(funcMap).ParseFiles(path))
// }
//
// var (
// 	funcMap = template.FuncMap{
// 		"div": func(a, b int) int {
// 			if b == 0 {
// 				return 0
// 			}
// 			return a / b
// 		},
// 		"json": jsonMarshal,
// 	}
//
// 	indexTmpl      = mustTemplate("index.html")
// 	queueTmpl      = mustTemplate("queue.html")
// 	maintTmpl      = mustTemplate("maintenance.html")
// 	maintAdminTmpl = mustTemplate("maintenance_admin.html")
// 	notifTmpl      = mustTemplate("notifications.html")
// 	notifRuleTmpl  = mustTemplate("notification_rule.html")
// 	profileTmpl    = mustTemplate("profile.html")
// )
//
// // jsonMarshal marshals v to JSON for safe embedding in JavaScript contexts.
// // It must not be used to inject JSON directly into HTML markup; use only
// // where the template engine expects JavaScript (e.g., inside <script> tags).
// //
// // When used as an html/template function, a non-nil error will cause template
// // execution to stop and that error to be returned to the caller of Execute.
// func jsonMarshal(v any) (template.JS, error) {
// 	b, err := json.Marshal(v)
// 	if err != nil {
// 		// Return empty value along with the error; html/template will abort on error.
// 		return template.JS(""), err
// 	}
// 	return template.JS(string(b)), nil
// }
//
// func main() {
//
// }
