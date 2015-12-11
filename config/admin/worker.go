package admin

import (
	"fmt"
	"path"
	"time"

	"github.com/qor/exchange/backends/csv"
	"github.com/qor/media_library"
	"github.com/qor/qor"
	"github.com/qor/qor-example/db"
	"github.com/qor/worker"
)

func getWorker() *worker.Worker {
	Worker := worker.New()

	type sendNewsletterArgument struct {
		Subject      string
		Content      string `sql:"size:65532"`
		SendPassword string
	}

	Worker.RegisterJob(worker.Job{
		Name: "send_newsletter",
		Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Started sending newsletters...")
			qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))
			for i := 1; i <= 100; i++ {
				time.Sleep(100 * time.Millisecond)
				qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
				qorJob.SetProgress(uint(i))
			}
			qorJob.AddLog("Finished send newsletters")
			return nil
		},
		Resource: Admin.NewResource(&sendNewsletterArgument{}),
	})

	type importProductArgument struct {
		File media_library.FileSystem
	}

	Worker.RegisterJob(worker.Job{
		Name: "import_products",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			argument := arg.(*importProductArgument)

			context := &qor.Context{DB: db.DB}

			ProductExchange.Import(csv.New(path.Join("public", argument.File.URL())), context)
			return nil
		},
		Resource: Admin.NewResource(&importProductArgument{}),
	})

	Worker.RegisterJob(worker.Job{
		Name: "export_products",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Exporting products...")

			context := &qor.Context{DB: db.DB}
			fileName := fmt.Sprintf("/downloads/products.%v.csv", time.Now().UnixNano())
			ProductExchange.Export(csv.New(path.Join("public", fileName)), context)

			qorJob.SetProgressText(fmt.Sprintf("Download it from <a href='%v'>Download exported products</a>", fileName))
			return nil
		},
	})
	return Worker
}