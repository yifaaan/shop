package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"shop/pkg/proto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Health is the liveness/readiness endpoint used by Consul.
func (s *Server) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleGrpcErrorToHttp maps a gRPC status error onto an HTTP response.
func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err == nil {
		return
	}
	if e, ok := status.FromError(err); ok {
		switch e.Code() {
		case codes.NotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"msg": e.Message()})
		case codes.Internal:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		case codes.InvalidArgument:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		case codes.AlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{"msg": e.Message()})
		case codes.PermissionDenied:
			ctx.JSON(http.StatusForbidden, gin.H{"msg": e.Message()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		}
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
}

func handleValidatorError(err error, ctx *gin.Context) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": removeTopStruct(errs.Translate(translator))})
}

// ==================== 商品 Handlers ====================

// GoodsList 商品列表（分页 + 过滤）
func (s *Server) GoodsList(ctx *gin.Context) {
	req := &proto.GoodsFilterRequest{}

	priceMin, _ := strconv.Atoi(ctx.DefaultQuery("pmin", "0"))
	priceMax, _ := strconv.Atoi(ctx.DefaultQuery("pmax", "0"))
	req.PriceMin = int32(priceMin)
	req.PriceMax = int32(priceMax)

	isHot := ctx.DefaultQuery("ih", "0")
	if isHot == "1" {
		req.IsHot = true
	}

	isNew := ctx.DefaultQuery("in", "0")
	if isNew == "1" {
		req.IsNew = true
	}

	isTab := ctx.DefaultQuery("it", "0")
	if isTab == "1" {
		req.IsTab = true
	}

	categoryId, _ := strconv.Atoi(ctx.DefaultQuery("c", "0"))
	req.TopCategory = int32(categoryId)

	pages, _ := strconv.Atoi(ctx.DefaultQuery("p", "1"))
	pagePerNums, _ := strconv.Atoi(ctx.DefaultQuery("pnum", "10"))
	req.Pages = int32(pages)
	req.PagePerNums = int32(pagePerNums)

	keywords := ctx.DefaultQuery("q", "")
	req.KeyWords = keywords

	brandId, _ := strconv.Atoi(ctx.DefaultQuery("b", "0"))
	req.Brand = int32(brandId)

	rsp, err := s.goodsSrv.GoodsList(ctx.Request.Context(), req)
	if err != nil {
		s.log.Error("GoodsList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	data := make([]*goodsResponse, 0, len(rsp.Data))
	for i := range rsp.Data {
		data = append(data, GoodsModelToResponse(rsp.Data[i]))
	}
	ctx.JSON(http.StatusOK, goodsListResponse{
		Total: rsp.Total,
		Data:  data,
	})
}

// GetGoodsDetail 通过 id 查询商品详情
func (s *Server) GetGoodsDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	rsp, err := s.goodsSrv.GetGoodsDetail(ctx.Request.Context(), &proto.GoodInfoRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("GetGoodsDetail grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, GoodsModelToResponse(rsp))
}

// CreateGoods 新建商品
func (s *Server) CreateGoods(ctx *gin.Context) {
	var form CreateGoodsForm
	if err := ctx.ShouldBind(&form); err != nil {
		s.log.Error("CreateGoods form binding error: ", err)
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.goodsSrv.CreateGoods(ctx.Request.Context(), &proto.CreateGoodsInfo{
		Name:            form.Name,
		GoodsSn:         form.GoodsSn,
		Stocks:          form.Stocks,
		MarketPrice:     form.MarketPrice,
		ShopPrice:       form.ShopPrice,
		GoodsBrief:      form.GoodsBrief,
		ShipFree:        form.ShipFree,
		Images:          form.Images,
		DescImages:      form.DescImages,
		GoodsFrontImage: form.FrontImage,
		CategoryId:      form.Category,
		BrandId:         form.Brand,
	})
	if err != nil {
		s.log.Error("CreateGoods grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// UpdateGoods 更新商品
func (s *Server) UpdateGoods(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateGoodsForm
	if err := ctx.ShouldBind(&form); err != nil {
		s.log.Error("UpdateGoods form binding error: ", err)
		handleValidatorError(err, ctx)
		return
	}
	form.Id = int32(id)
	_, err = s.goodsSrv.UpdateGoods(ctx.Request.Context(), &proto.CreateGoodsInfo{
		Id:              form.Id,
		Name:            form.Name,
		GoodsSn:         form.GoodsSn,
		Stocks:          form.Stocks,
		MarketPrice:     form.MarketPrice,
		ShopPrice:       form.ShopPrice,
		GoodsBrief:      form.GoodsBrief,
		ShipFree:        form.ShipFree,
		Images:          form.Images,
		DescImages:      form.DescImages,
		GoodsFrontImage: form.FrontImage,
		IsNew:           form.IsNew,
		IsHot:           form.IsHot,
		OnSale:          form.OnSale,
		CategoryId:      form.Category,
		BrandId:         form.Brand,
	})
	if err != nil {
		s.log.Error("UpdateGoods grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteGoods 删除商品
func (s *Server) DeleteGoods(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.goodsSrv.DeleteGoods(ctx.Request.Context(), &proto.DeleteGoodsInfo{Id: int32(id)})
	if err != nil {
		s.log.Error("DeleteGoods grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 分类 Handlers ====================

// GetAllCategorysList 获取全部分类列表（含树形 JSON）
func (s *Server) GetAllCategorysList(ctx *gin.Context) {
	rsp, err := s.goodsSrv.GetAllCategorysList(ctx.Request.Context(), &emptypb.Empty{})
	if err != nil {
		s.log.Error("GetAllCategorysList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"total": rsp.Total,
		"data":  json.RawMessage(rsp.JsonData),
	})
}

// GetSubCategory 获取子分类
func (s *Server) GetSubCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	rsp, err := s.goodsSrv.GetSubCategory(ctx.Request.Context(), &proto.CategoryListRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("GetSubCategory grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, SubCategoryListToResponse(rsp))
}

// CreateCategory 新建分类
func (s *Server) CreateCategory(ctx *gin.Context) {
	var form CreateCategoryForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.goodsSrv.CreateCategory(ctx.Request.Context(), &proto.CategoryInfoRequest{
		Name:           form.Name,
		ParentCategory: form.ParentCategory,
		Level:          form.Level,
		IsTab:          form.IsTab,
	})
	if err != nil {
		s.log.Error("CreateCategory grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, CategoryModelToResponse(rsp))
}

// UpdateCategory 更新分类
func (s *Server) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateCategoryForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	form.Id = int32(id)
	_, err = s.goodsSrv.UpdateCategory(ctx.Request.Context(), &proto.CategoryInfoRequest{
		Id:    form.Id,
		Name:  form.Name,
		IsTab: form.IsTab,
	})
	if err != nil {
		s.log.Error("UpdateCategory grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteCategory 删除分类
func (s *Server) DeleteCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.goodsSrv.DeleteCategory(ctx.Request.Context(), &proto.DeleteCategoryRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("DeleteCategory grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 品牌 Handlers ====================

// BrandList 品牌列表（分页）
func (s *Server) BrandList(ctx *gin.Context) {
	pages := int32(ctx.DefaultQuery("pages", "1")[0] - '0')
	pagePerNumsStr := ctx.DefaultQuery("page_per_nums", "10")
	pagePerNums, _ := strconv.Atoi(pagePerNumsStr)
	rsp, err := s.goodsSrv.BrandList(ctx.Request.Context(), &proto.BrandFilterRequest{
		Pages:       pages,
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		s.log.Error("BrandList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// GetBrandDetail 通过 id 查询品牌详情（通过 BrandList 转发，proto 无单查接口）
func (s *Server) GetBrandDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	// proto 无 GetBrandDetail 接口，用 BrandList 拉取后过滤
	rsp, err := s.goodsSrv.BrandList(ctx.Request.Context(), &proto.BrandFilterRequest{
		Pages:       1,
		PagePerNums: 100,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	for _, b := range rsp.Data {
		if b.Id == int32(id) {
			ctx.JSON(http.StatusOK, b)
			return
		}
	}
	ctx.JSON(http.StatusNotFound, gin.H{"msg": "品牌不存在"})
}

// CreateBrand 新建品牌
func (s *Server) CreateBrand(ctx *gin.Context) {
	var form CreateBrandForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.goodsSrv.CreateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Name: form.Name,
		Logo: form.Logo,
	})
	if err != nil {
		s.log.Error("CreateBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// UpdateBrand 更新品牌
func (s *Server) UpdateBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateBrandForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	form.Id = int32(id)
	_, err = s.goodsSrv.UpdateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Id:   form.Id,
		Name: form.Name,
		Logo: form.Logo,
	})
	if err != nil {
		s.log.Error("UpdateBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteBrand 删除品牌
func (s *Server) DeleteBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.goodsSrv.DeleteBrand(ctx.Request.Context(), &proto.BrandRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("DeleteBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 轮播图 Handlers ====================

// BannerList 轮播图列表
func (s *Server) BannerList(ctx *gin.Context) {
	rsp, err := s.goodsSrv.BannerList(ctx.Request.Context(), &emptypb.Empty{})
	if err != nil {
		s.log.Error("BannerList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// CreateBanner 新建轮播图
func (s *Server) CreateBanner(ctx *gin.Context) {
	var form CreateBannerForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.goodsSrv.CreateBanner(ctx.Request.Context(), &proto.BannerRequest{
		Image: form.Image,
		Url:   form.Url,
		Index: form.Index,
	})
	if err != nil {
		s.log.Error("CreateBanner grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// UpdateBanner 更新轮播图
func (s *Server) UpdateBanner(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateBannerForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	form.Id = int32(id)
	_, err = s.goodsSrv.UpdateBanner(ctx.Request.Context(), &proto.BannerRequest{
		Id:    form.Id,
		Image: form.Image,
		Url:   form.Url,
		Index: form.Index,
	})
	if err != nil {
		s.log.Error("UpdateBanner grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteBanner 删除轮播图
func (s *Server) DeleteBanner(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.goodsSrv.DeleteBanner(ctx.Request.Context(), &proto.BannerRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("DeleteBanner grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 分类品牌关系 Handlers ====================

// CategoryBrandList 分类品牌关系列表（分页）
func (s *Server) CategoryBrandList(ctx *gin.Context) {
	pagesStr := ctx.DefaultQuery("pages", "1")
	pagePerNumsStr := ctx.DefaultQuery("page_per_nums", "10")
	pages, _ := strconv.Atoi(pagesStr)
	pagePerNums, _ := strconv.Atoi(pagePerNumsStr)
	rsp, err := s.goodsSrv.CategoryBrandList(ctx.Request.Context(), &proto.CategoryBrandFilterRequest{
		Pages:       int32(pages),
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		s.log.Error("CategoryBrandList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// GetCategoryBrandList 通过分类获取其关联的品牌列表
func (s *Server) GetCategoryBrandList(ctx *gin.Context) {
	categoryId, err := strconv.Atoi(ctx.Param("category_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid category_id"})
		return
	}
	rsp, err := s.goodsSrv.GetCategoryBrandList(ctx.Request.Context(), &proto.CategoryInfoRequest{Id: int32(categoryId)})
	if err != nil {
		s.log.Error("GetCategoryBrandList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// CreateCategoryBrand 新建分类品牌关系
func (s *Server) CreateCategoryBrand(ctx *gin.Context) {
	var form CreateCategoryBrandForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.goodsSrv.CreateCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{
		CategoryId: form.Category,
		BrandId:    form.Brand,
	})
	if err != nil {
		s.log.Error("CreateCategoryBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, rsp)
}

// UpdateCategoryBrand 更新分类品牌关系
func (s *Server) UpdateCategoryBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateCategoryBrandForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	form.Id = int32(id)
	_, err = s.goodsSrv.UpdateCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{
		Id:         form.Id,
		CategoryId: form.Category,
		BrandId:    form.Brand,
	})
	if err != nil {
		s.log.Error("UpdateCategoryBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteCategoryBrand 删除分类品牌关系
func (s *Server) DeleteCategoryBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.goodsSrv.DeleteCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{Id: int32(id)})
	if err != nil {
		s.log.Error("DeleteCategoryBrand grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// suppress unused import (strings used in translator.go)
var _ = strings.HasPrefix
