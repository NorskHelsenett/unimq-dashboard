package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
)

//	func mustTemplate(name string) *template.Template {
//		path := "web/templates/" + name
//		return template.Must(template.New(name).Funcs(funcMap).ParseFiles(path))
//	}
//
// var (
//
//	funcMap = template.FuncMap{
//		"div": func(a, b int) int {
//			if b == 0 {
//				return 0
//			}
//			return a / b
//		},
//		"json": jsonMarshal,
//	}
//
//	indexTmpl      = mustTemplate("index.html")
//	queueTmpl      = mustTemplate("queue.html")
//	maintTmpl      = mustTemplate("maintenance.html")
//	maintAdminTmpl = mustTemplate("maintenance_admin.html")
//	notifTmpl      = mustTemplate("notifications.html")
//	notifRuleTmpl  = mustTemplate("notification_rule.html")
//	profileTmpl    = mustTemplate("profile.html")
//
// )
//
// // jsonMarshal marshals v to JSON for safe embedding in JavaScript contexts.
// // It must not be used to inject JSON directly into HTML markup; use only
// // where the template engine expects JavaScript (e.g., inside <script> tags).
// //
// // When used as an html/template function, a non-nil error will cause template
// // execution to stop and that error to be returned to the caller of Execute.
//
//	func jsonMarshal(v any) (template.JS, error) {
//		b, err := json.Marshal(v)
//		if err != nil {
//			// Return empty value along with the error; html/template will abort on error.
//			return template.JS(""), err
//		}
//		return template.JS(string(b)), nil
//	}

// This file is used to generate test data for the different stores. It is not used in the actual application.
func main() {

	logger.SetupLogger()

	config := config.NewConfig()
	if err := config.Load(); err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	err := config.CheckURLs()
	if err != nil {
		slog.Error("failed to validate URLs", "error", err)
		return
	}

	uri := database.BuildURI(config.MongoDBHost, config.MongoDBUsername, config.MongoDBPassword, config.MongoDBPort)
	db, err := database.NewDatabase(uri, config.MongoDBDatabase)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	err = db.InitCollections()
	if err != nil {
		slog.Error("failed to initialize database collections", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.Seed(ctx)
	if err != nil {
		slog.Error("failed to seed database", "error", err)
		return
	}

	slog.Info("database seeded successfully")
}
