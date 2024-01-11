package service

import (
	"context"
	"ebayclone/changeset"
	"ebayclone/domain"
	"ebayclone/dto"
	"ebayclone/dto/product"
	"ebayclone/infrastructure"
	"ebayclone/repo"
	"fmt"
	"net/http"
)

type ProductService struct {
	repo  *repo.Repo
	debug bool
}

var ProductServiceManager *ProductService

func NewProductServiceManager(debug bool) *ProductService {
	if ProductServiceManager == nil {
		ProductServiceManager = &ProductService{
			debug: debug,
			repo:  repo.NewRepo(infrastructure.MysqlConfig, debug),
		}
	}
	return ProductServiceManager
}

func (p *ProductService) CreateService(ctx context.Context, req *product.ProductCreateReq) *dto.BaseMessageResponse {
	base_response := &dto.BaseMessageResponse{
		StatusCode:    http.StatusInternalServerError,
		ErrCodeString: "",
		ReponseObject: nil,
	}
	return base_response

	// why need check product_type_id
	// fronend can get all product_type id from first on home page
	// but hacker can get api with product_type_id not exist

}

func (p *ProductService) CreateProduct(ctx context.Context, req *product.ProductCreateReq) *dto.BaseMessageResponse {
	tx := p.repo.OpenTx(ctx)
	base_response := &dto.BaseMessageResponse{
		StatusCode:    http.StatusInternalServerError,
		ErrCodeString: "",
		ReponseObject: nil,
	}
	product_type_entity_before := ProductTypeServiceManager.getProductTypeEntityExistById(req.ProductTypeId)
	if product_type_entity_before == nil {
		base_response.TransformToNotFoundEntity("ProductTypeService, sorry hacker")
		return base_response
	}

	product_entity := &domain.Product{}
	product_changeset := changeset.CastValues(product_entity, map[string]any{
		"Name": req.Name,
		"ProductTypeRel": &domain.ProductType{
			Id: req.ProductTypeId,
		},
	})

	err := p.repo.SaveTx(ctx, product_changeset, tx)
	if err != nil {
		base_response.ErrCodeString = err.Error()
		return base_response
	}

	product_type_entity_cloned_update := product_type_entity_before.CloneProductType()
	fmt.Println(product_type_entity_before.AggregateFields, product_type_entity_cloned_update.AggregateFields)
	for attributeIdCreated, optionValueIdCreated := range *req.Fields {
		if _, existAttributeId := product_type_entity_before.AggregateFields.Fields[attributeIdCreated]; !existAttributeId {
			base_response.TransformToNotFoundEntity("ProductType Not Found Attribute Id")
			tx.Rollback()
			return base_response
		}
		if _, existValueId := product_type_entity_before.AggregateFields.Fields[attributeIdCreated][optionValueIdCreated]; !existValueId {
			base_response.TransformToNotFoundEntity("ProductType Not Found OptionValue Id")
			tx.Rollback()
			return base_response
		}
		product_type_entity_cloned_update.AggregateFields.Fields[attributeIdCreated][optionValueIdCreated] += 1
	}

	fmt.Println("prepare update from product type service")
	err = ProductTypeServiceManager.UpdateAggregateFields(ctx, product_type_entity_cloned_update, tx)
	if err != nil {
		tx.Rollback()
		base_response.ErrCodeString = err.Error()
		return base_response
	}

	err = tx.Commit()
	ProductTypeServiceManager.UpdateCacheProductTypeById(product_type_entity_cloned_update.Id, product_type_entity_cloned_update)
	base_response.TransformToStatusOk(&product.ProductCreateRes{
		Id: product_entity.Id,
	})
	return base_response
}

// buy one product
// update history order, one transaction
// update product_Type count -= 1 all field have related, one transaction
//
// ship to but not get
// have one api called from shipper signal: /refund
