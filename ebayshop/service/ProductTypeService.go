package service

import (
	"context"
	"database/sql"
	"ebayclone/changeset"
	"ebayclone/domain"
	dto2 "ebayclone/dto"
	"ebayclone/dto/product_type_dto"
	"ebayclone/infrastructure"
	"ebayclone/log_util"
	"ebayclone/repo"
	"ebayclone/valueobject"
	"fmt"
	"net/http"
)

type ProductTypeService struct {
	repo                *repo.Repo
	cacheAllProductType map[uint32]*domain.ProductType
	debug               bool
	serviceName         string
}

func (s *ProductTypeService) CreateProductType(ctx context.Context, req *product_type_dto.ProductTypeCreateReq) *dto2.BaseMessageResponse {
	product_type_entity := &domain.ProductType{}
	base_message_response := &dto2.BaseMessageResponse{
		StatusCode:    http.StatusInternalServerError,
		ErrCodeString: "Internal Server Error",
		ReponseObject: nil,
	}

	attributeObjectRes := &valueobject.AttributesObjectRes{
		Attributes: make([]*valueobject.OneAttributeObjectRes, 0),
	}
	var globalAttributeId uint32 = 0
	var aggregateFieldsJSON = map[valueobject.AttributeId]map[valueobject.OptionValueId]int{}
	for attributeNameReq, optionValues := range req.Attributes {
		oneAttributeObjectRes := &valueobject.OneAttributeObjectRes{
			Id:           valueobject.AttributeId(globalAttributeId + 1),
			Name:         attributeNameReq,
			OptionValues: make([]*valueobject.OptionValueRes, 0),
		}

		if _, existMapOptionValueJsonFieldAggregate := aggregateFieldsJSON[oneAttributeObjectRes.Id]; !existMapOptionValueJsonFieldAggregate {
			aggregateFieldsJSON[oneAttributeObjectRes.Id] = make(map[valueobject.OptionValueId]int)
		}

		var globalOptionValueId uint32 = 0
		for _, optionValueReq := range optionValues {
			oneOptionValueRes := &valueobject.OptionValueRes{
				Id:    valueobject.OptionValueId(globalOptionValueId + 1),
				Value: optionValueReq,
			}
			oneAttributeObjectRes.OptionValues = append(
				oneAttributeObjectRes.OptionValues, oneOptionValueRes)
			globalOptionValueId++
			aggregateFieldsJSON[oneAttributeObjectRes.Id][oneOptionValueRes.Id] = 0
		}
		attributeObjectRes.Attributes = append(
			attributeObjectRes.Attributes, oneAttributeObjectRes)
		globalAttributeId++
	}
	product_type_changeset := changeset.CastValues(product_type_entity, map[string]any{
		"Name":       req.Name,
		"Attributes": attributeObjectRes,
		"AggregateFields": &valueobject.AggregateFieldJSON{
			Fields: aggregateFieldsJSON,
		},
	})

	err := s.repo.Save(ctx, product_type_changeset)
	if err != nil {
		base_message_response.ErrCodeString = err.Error()
		return base_message_response
	}

	// write into cache new product_type
	fmt.Println("product type entity add: ", product_type_entity.AggregateFields)
	s.addProductTypeEntityIntoCache(product_type_entity)
	base_message_response.TransformToStatusOk(&product_type_dto.ProductTypeCreateRes{
		Id:         product_type_entity.Id,
		Attributes: product_type_entity.Attributes,
	})

	return base_message_response
}

var ProductTypeServiceManager *ProductTypeService

func NewProductTypeService(debug bool) *ProductTypeService {
	if ProductTypeServiceManager == nil {
		ProductTypeServiceManager = &ProductTypeService{
			repo:                repo.NewRepo(infrastructure.MysqlConfig, debug),
			cacheAllProductType: make(map[uint32]*domain.ProductType),
			debug:               debug,
			serviceName:         "ProductTypeService",
		}

		// load all product type here
		ProductTypeServiceManager.FetchAllProductTypesInMemoryFromDatabase()
	}
	return ProductTypeServiceManager
}

func (p *ProductTypeService) FetchAllProductTypesInMemoryFromDatabase() {
	// create builder query
	builder := p.repo.GetById(&domain.ProductType{})
	table_name := "producttypes"
	builder.
		Select(repo.Col("Id", table_name)).
		Select(repo.Col("Name", table_name)).
		Select(repo.Col("Attributes", table_name)).
		Select(repo.Col("AggregateFields", table_name))

	query, args := builder.Query()
	entities, _ := p.repo.RawQuery(query, args, &domain.ProductType{})
	if len(entities) > 0 {
		for _, entity := range entities {
			productEntity := entity.(*domain.ProductType)
			Id := productEntity.Id
			if _, ok := p.cacheAllProductType[Id]; !ok {
				p.cacheAllProductType[Id] = productEntity
			}
			for _, oneAttribute := range productEntity.Attributes.Attributes {
				log_util.PrintFlag(p.serviceName, p.debug, fmt.Sprintf("type_name [%v], attribute_name [%v]",
					productEntity.Name, oneAttribute.Name, productEntity.AggregateFields))
			}
		}
	}
}

func (p *ProductTypeService) addProductTypeEntityIntoCache(entity *domain.ProductType) {
	if _, exist := p.cacheAllProductType[entity.Id]; !exist {
		p.cacheAllProductType[entity.Id] = entity
	}
}

func (p *ProductTypeService) getProductTypeEntityExistById(productTypeId uint32) *domain.ProductType {
	fmt.Println("global cache: ", p.cacheAllProductType)
	return p.cacheAllProductType[productTypeId]
}

func (s *ProductTypeService) GetAllProductType(ctx context.Context) *dto2.BaseMessageResponse {
	base_message := &dto2.BaseMessageResponse{
		StatusCode:    http.StatusInternalServerError,
		ErrCodeString: "",
		ReponseObject: nil,
	}
	if len(s.cacheAllProductType) == 0 {
		base_message.TransformToNotFoundEntity("ProductType")
		return base_message
	}
	product_type_res := product_type_dto.ProductTypeGetAllRes{
		ProductTypes: make([]*domain.ProductType, 0),
	}
	for _, productEntity := range s.cacheAllProductType {
		product_type_res.ProductTypes = append(product_type_res.ProductTypes, productEntity)
	}

	base_message.TransformToStatusOk(product_type_res)
	return base_message
}

func (p *ProductTypeService) UpdateAggregateFields(ctx context.Context, entity *domain.ProductType, tx *sql.Tx) error {
	product_entity := &domain.ProductType{Id: entity.Id} // must add id here , to it get reflection id where id update
	product_update_changeset := changeset.CastValues(product_entity, map[string]any{

		"AggregateFields": entity.AggregateFields,
	})

	fmt.Println("before update repo save")
	err := p.repo.UpdateTxById(ctx, product_update_changeset, tx)
	if err != nil {
		fmt.Println("error: ", err)
		log_util.PrintFlag("ProductService", p.debug, fmt.Sprintf("error: ", err))
		return err
	}
	fmt.Println("update success")
	return nil
}

func (p *ProductTypeService) UpdateCacheProductTypeById(id uint32, new_product_type *domain.ProductType) {
	if _, ok := p.cacheAllProductType[id]; ok {
		p.cacheAllProductType[id] = new_product_type
	}
}
