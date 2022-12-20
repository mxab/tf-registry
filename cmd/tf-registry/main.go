package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	modulehandler "github.com/mxab/tf-registry/internal/module/handler"

	"github.com/labstack/echo/v4"
	"github.com/mxab/tf-registry/internal/discovery"
)

func server() *echo.Echo {
	e := echo.New()

	b := discovery.DiscoveryResponse{
		ModulesV1: "/v1/modules",
	}
	discovery.NewController(b)
	modulehandler.RegisterModuleControllerGroup(e.Group(b.ModulesV1), nil)

	return e
}
func main() {

	//create a new s3 client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Printf("failed to create session, %v", err)
		return
	}

	svc := s3.New(sess)
	svc.ListBuckets(nil)

	e := server()
	e.Logger.Fatal(e.Start(":1323"))
}
