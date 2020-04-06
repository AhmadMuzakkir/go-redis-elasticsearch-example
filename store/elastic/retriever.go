package elastic

import (
	"context"
	"strconv"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var _ store.ViewRetriever = (*viewRetriever)(nil)

type viewRetriever struct {
	client *elastic.Client
}

func (v *viewRetriever) Retrieve(ctx context.Context, id string, ranges ...store.Range) ([]store.ViewCount, error) {
	if len(ranges) == 0 {
		return []store.ViewCount{}, nil
	}

	aggs := elastic.NewRangeAggregation().Field("timestamp")

	// Convert each range into Range Aggregation and add it into the Aggregation
	for _, rang := range ranges {
		unit := rangeUnit(rang)
		if unit == "" {
			return nil, errors.Errorf("unimplemented range unit %v", rang)
		}

		// Use range constant as the bucket key.
		key := strconv.Itoa(int(rang))

		aggs.AddRangeWithKey(key, "now-"+unit, "now")
	}

	res, err := v.client.Search(indexName).
		Query(elastic.NewTermQuery("id", id)).
		Aggregation("views", aggs).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	rangeRes, _ := res.Aggregations.Range("views")
	// This should never happens
	if rangeRes == nil || len(rangeRes.Buckets) == 0 {
		return nil, errors.New("elastic response empty")
	}

	viewCounts := make([]store.ViewCount, 0, len(ranges))

	for _, bucket := range rangeRes.Buckets {
		// The bucket key is the ranges constant.
		rang, err := strconv.Atoi(bucket.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "bucket key %d", rang)
		}

		desc := store.RangeDescription(store.Range(rang))
		if desc == "" {
			return nil, errors.Errorf("unimplemented range description %v", rang)
		}

		viewCounts = append(viewCounts, store.ViewCount{
			Description: desc,
			Count:       bucket.DocCount,
		})
	}

	return viewCounts, nil
}

// TODO add unit test.
// Get ElasticSearch's time unit for the range constant
func rangeUnit(rang store.Range) string {
	switch rang {
	case store.OneMinute:
		return "1m"
	case store.FiveMinute:
		return "5m"
	case store.OneHour:
		return "1h"
	case store.OneDay:
		return "1d"
	case store.OneWeek:
		return "7d"
	case store.OneMonth:
		return "30d"
	default:
		return ""
	}
}
