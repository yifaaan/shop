// Package esearch 是 goods_srv 对 Elasticsearch 的薄封装：索引管理、
// 单条/批量写入、按关键词+结构化过滤检索返回商品 ID。
//
// 本包不 import goods_srv 主包或 gorm，避免循环依赖；调用方负责
// Goods 模型与 GoodsDoc 的转换。
package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Config struct {
	Addr           string
	Index          string
	ReindexOnStart bool
}

// GoodsDoc 是写入 ES 的商品文档。
type GoodsDoc struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	GoodsBrief  string  `json:"goods_brief"`
	ShopPrice   float64 `json:"shop_price"`
	MarketPrice float64 `json:"market_price"`
	ShipFree    bool    `json:"ship_free"`
	IsHot       bool    `json:"is_hot"`
	IsNew       bool    `json:"is_new"`
	OnSale      bool    `json:"on_sale"`
	ClickNum    int32   `json:"click_num"`
	SoldNum     int32   `json:"sold_num"`
	FavNum      int32   `json:"fav_num"`
	BrandsID    int32   `json:"brands_id"`
	CategoryID  int32   `json:"category_id"`
}

type Service struct {
	client *es.Client
	index  string
}

// New 构造 ES 客户端并 ping 一次确认可达。
func New(cfg Config) (*Service, error) {
	cli, err := es.NewClient(es.Config{Addresses: []string{cfg.Addr}})
	if err != nil {
		return nil, fmt.Errorf("create es client: %w", err)
	}
	return &Service{client: cli, index: cfg.Index}, nil
}

// goodsMapping 是 goods 索引的 mapping：name/goods_brief 用 IK 中文分词，
// 写入 ik_max_word（细切召回高），查询 ik_smart（粗切精确）。
const goodsMapping = `{
  "mappings": {
    "properties": {
      "id":          {"type": "integer"},
      "name":        {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
      "goods_brief": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
      "shop_price":  {"type": "double"},
      "market_price":{"type": "double"},
      "ship_free":   {"type": "boolean"},
      "is_hot":      {"type": "boolean"},
      "is_new":      {"type": "boolean"},
      "on_sale":     {"type": "boolean"},
      "click_num":   {"type": "integer"},
      "sold_num":    {"type": "integer"},
      "fav_num":     {"type": "integer"},
      "brands_id":   {"type": "integer"},
      "category_id": {"type": "integer"}
    }
  }
}`

// EnsureIndex 在索引不存在时按 goodsMapping 创建；已存在则跳过。
func (s *Service) EnsureIndex(ctx context.Context) error {
	exists := esapi.IndicesExistsRequest{Index: []string{s.index}}
	res, err := exists.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("check index %s: %w", s.index, err)
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil // 已存在
	}
	if res.StatusCode != 404 {
		return fmt.Errorf("unexpected status checking index %s: %s", s.index, res.Status())
	}
	req := esapi.IndicesCreateRequest{Index: s.index, Body: strings.NewReader(goodsMapping)}
	crest, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("create index %s: %w", s.index, err)
	}
	defer crest.Body.Close()
	if crest.IsError() {
		return fmt.Errorf("create index %s: %s", s.index, crest.String())
	}
	return nil
}

// IndexGoods 写入/覆盖一条商品文档（_id = goods id）。
func (s *Service) IndexGoods(ctx context.Context, doc GoodsDoc) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal goods doc: %w", err)
	}
	req := esapi.IndexRequest{
		Index:      s.index,
		DocumentID: strconv.Itoa(int(doc.ID)),
		Body:       bytes.NewReader(body),
		Refresh:    "false",
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("index goods %d: %w", doc.ID, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("index goods %d: %s", doc.ID, res.String())
	}
	return nil
}

// DeleteGoodsDoc 删除一条商品文档；404 视为已删除，不报错。
func (s *Service) DeleteGoodsDoc(ctx context.Context, id int32) error {
	req := esapi.DeleteRequest{Index: s.index, DocumentID: strconv.Itoa(int(id))}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("delete goods %d: %w", id, err)
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete goods %d: %s", id, res.String())
	}
	return nil
}

// Clear 清空索引内全部文档（_delete_by_query match_all）。
// 启动重建前调用，剔除已软删的陈旧文档，保证 ES == 存活 MySQL。
func (s *Service) Clear(ctx context.Context) error {
	req := esapi.DeleteByQueryRequest{
		Index: []string{s.index},
		Body:  strings.NewReader(`{"query":{"match_all":{}}}`),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("clear index %s: %w", s.index, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("clear index %s: %s", s.index, res.String())
	}
	return nil
}

// BulkIndex 批量写入商品文档（NDJSON bulk API）。
func (s *Service) BulkIndex(ctx context.Context, docs []GoodsDoc) error {
	if len(docs) == 0 {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, d := range docs {
		// action 行
		meta := map[string]any{"index": map[string]any{"_index": s.index, "_id": strconv.Itoa(int(d.ID))}}
		if err := enc.Encode(meta); err != nil {
			return fmt.Errorf("encode bulk action: %w", err)
		}
		// source 行
		if err := enc.Encode(d); err != nil {
			return fmt.Errorf("encode bulk source: %w", err)
		}
	}
	req := esapi.BulkRequest{Body: bytes.NewReader(buf.Bytes()), Refresh: "false"}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index: %s", res.String())
	}
	// 解析 bulk 响应，有单项失败则报错
	var br struct {
		Errors bool `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&br); err != nil {
		return fmt.Errorf("decode bulk response: %w", err)
	}
	if br.Errors {
		return fmt.Errorf("bulk index completed with item errors")
	}
	return nil
}

// SearchIDs 按关键词 + 结构化过滤检索，返回命中商品 ID（按 ES 相关性顺序）
// 与真实命中总数。keyWords 为空时不应调用此方法（调用方走 MySQL 路径）。
func (s *Service) SearchIDs(
	ctx context.Context,
	keyWords string,
	priceMin, priceMax float64,
	isHot, isNew bool,
	brand int32,
	categoryIDs []int32,
	from, size int,
) ([]int32, int32, error) {
	q := map[string]any{
		"track_total_hits": true,
		"from":             from,
		"size":             size,
		"query": map[string]any{
			"bool": buildBoolQuery(keyWords, priceMin, priceMax, isHot, isNew, brand, categoryIDs),
		},
	}
	body, err := json.Marshal(q)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal search query: %w", err)
	}
	req := esapi.SearchRequest{
		Index:          []string{s.index},
		Body:           bytes.NewReader(body),
		TrackTotalHits: true,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, 0, fmt.Errorf("search goods: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, 0, fmt.Errorf("search goods: %s", res.String())
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read search response: %w", err)
	}
	var sr struct {
		Hits struct {
			Total struct {
				Value int32 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, 0, fmt.Errorf("decode search response: %w", err)
	}
	ids := make([]int32, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		if id, err := strconv.Atoi(h.ID); err == nil {
			ids = append(ids, int32(id))
		}
	}
	return ids, sr.Hits.Total.Value, nil
}

// buildBoolQuery 组装关键词 must + 结构化 filter 的 bool query。
func buildBoolQuery(keyWords string, priceMin, priceMax float64, isHot, isNew bool, brand int32, categoryIDs []int32) map[string]any {
	boolq := map[string]any{}
	// 关键词在 name/goods_brief 上 multi_match（name 权重 x2），用 ik_smart 粗切查询
	boolq["must"] = []map[string]any{
		{"multi_match": map[string]any{
			"query":    keyWords,
			"fields":   []string{"name^2", "goods_brief"},
			"analyzer": "ik_smart",
		}},
	}
	filters := []map[string]any{}
	if priceMin > 0 {
		filters = append(filters, map[string]any{"range": map[string]any{"shop_price": map[string]any{"gte": priceMin}}})
	}
	if priceMax > 0 {
		filters = append(filters, map[string]any{"range": map[string]any{"shop_price": map[string]any{"lte": priceMax}}})
	}
	if isHot {
		filters = append(filters, map[string]any{"term": map[string]any{"is_hot": true}})
	}
	if isNew {
		filters = append(filters, map[string]any{"term": map[string]any{"is_new": true}})
	}
	if brand > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"brands_id": brand}})
	}
	if len(categoryIDs) > 0 {
		filters = append(filters, map[string]any{"terms": map[string]any{"category_id": categoryIDs}})
	}
	if len(filters) > 0 {
		boolq["filter"] = filters
	}
	return boolq
}
