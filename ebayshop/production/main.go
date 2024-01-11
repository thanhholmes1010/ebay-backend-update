package main

import (
	"bufio"
	"ebayclone/changeset"
	"ebayclone/controller"
	"ebayclone/domain"
	"ebayclone/valueobject"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

var globalResourceServiceConfig = map[string]bool{}

func load_config_service() {
	file_name := ".config_service"
	f, err := os.Open(file_name)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, string('\n'))
		params := strings.Split(line, "=")
		resource := params[0]
		debug_param := params[1]
		if _, ok := globalResourceServiceConfig[resource]; !ok {
			globalResourceServiceConfig[resource] = debug_param == "true"
		}
	}
}

func main() {
	engine := gin.Default()
	changeset.CastValues(&domain.ProductType{}, map[string]any{
		"Attributes": &valueobject.AttributesObjectRes{},
	})
	load_config_service()
	api_group := engine.Group("/api")
	api_group.Use()
	controller.InitProductTypeController(api_group, "/product_type", globalResourceServiceConfig["ProductTypeService"])
	controller.InitProductController(api_group, "/product", globalResourceServiceConfig["ProductService"])
	engine.Run("localhost:8080")
}
