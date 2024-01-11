package controller

import (
	dto2 "ebayclone/dto"
	"ebayclone/dto/product_type_dto"
	"ebayclone/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProductTypeController struct {
	service *service.ProductTypeService
	group   *gin.RouterGroup
}

func (c *ProductTypeController) CreateProductType() {
	c.group.POST("/create", func(context *gin.Context) {
		var dto product_type_dto.ProductTypeCreateReq
		err := context.ShouldBindJSON(&dto)
		if err != nil {
			context.JSON(http.StatusBadRequest, nil)
		} else {
			var response_message *dto2.BaseMessageResponse
			response_message = c.service.CreateProductType(context, &dto)
			context.JSON(response_message.StatusCode, response_message)
		}
	})
}

func (c *ProductTypeController) UpdateProductType() {

}

func (c *ProductTypeController) GetAllProductType() {
	c.group.GET("/get_all", func(context *gin.Context) {
		base_response := c.service.GetAllProductType(context)
		context.JSON(base_response.StatusCode, base_response)
	})
}

func InitProductTypeController(parentGroup *gin.RouterGroup, rootApiPathResource string, debug bool) {
	productTypeObjectController := &ProductTypeController{
		service: service.NewProductTypeService(debug),
		group:   parentGroup.Group(rootApiPathResource),
	}
	productTypeObjectController.CreateProductType()
	productTypeObjectController.UpdateProductType()
	productTypeObjectController.GetAllProductType()
}
