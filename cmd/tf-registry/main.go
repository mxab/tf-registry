package main

import (
	"github.com/mxab/tf-registry/internal/module/handler"

	"github.com/labstack/echo/v4"
)

func server() *echo.Echo {
	e := echo.New()
	module_handler := &handler.Controller{}
	e.GET("/v1/modules", module_handler.ListModules)
	return e
}
func main() {

	// create a new s3 client
	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String("us-east-1")},
	// )
	// if err != nil {
	// 	e.StdLogger.Fatalf("failed to create session, %v", err)
	// }

	// svc := s3.New(sess)
	// svc.ListBuckets(nil)

	e := server()
	e.Logger.Fatal(e.Start(":1323"))
}
