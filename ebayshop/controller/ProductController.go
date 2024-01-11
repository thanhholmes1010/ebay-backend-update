package controller

import (
	"ebayclone/dto/product"
	"ebayclone/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProductController struct {
	service *service.ProductService
	group   *gin.RouterGroup
}

func (c *ProductController) CreateProduct() {
	c.group.POST("/create", func(context *gin.Context) {
		var dto product.ProductCreateReq
		err := context.ShouldBindJSON(&dto)
		if err != nil {
			fmt.Println(err)
			context.JSON(http.StatusBadRequest, "wrong format")
		} else {
			base_response := c.service.CreateProduct(context, &dto)
			context.JSON(base_response.StatusCode, base_response)
		}
	})
}

func InitProductController(parentGroup *gin.RouterGroup, prefixRootApi string, debug bool) {
	p := &ProductController{
		group:   parentGroup.Group(prefixRootApi),
		service: service.NewProductServiceManager(debug),
	}

	p.CreateProduct()
}
