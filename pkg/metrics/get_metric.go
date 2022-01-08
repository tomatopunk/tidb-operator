package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap/tidb-operator/pkg/monitor/monitor"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type TikvWritingByteRate struct {
	Instance string
	Value    float64
}

type TikvWritingByte struct {
	cache     []TikvWritingByteRate
	Namespace string
	Name      string
	Shard     int32
	context   context.Context
	done      chan struct{}
}

func (b *TikvWritingByte) Get() (error, []TikvWritingByteRate) {
	return nil, b.cache
}

func (b *TikvWritingByte) StartFlush(tickets time.Duration) {
	go func() {
		for {
			select {
			case <-time.After(tickets):
				err, currentData := b.getTikvWritingByte()
				if err != nil {
					b.cache = currentData
				}
			case <-b.done:
				return
			}
		}
	}()

}

func (b *TikvWritingByte) getTikvWritingByte() (error, []TikvWritingByteRate) {
	prometheusAddr := fmt.Sprintf("%s.%s:9090", monitor.PrometheusName(b.Name, b.Shard), b.Namespace)
	response := &struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metrics struct {
					Instance string `json:"instance"`
				} `json:"metric"`
				Values []interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}{}
	err := b.queryPrometheusQuery(prometheusAddr, "sum(tikv_scheduler_writing_bytes) by (instance)", response)
	if response.Status != "success" {
		return err, nil
	}
	var result []TikvWritingByteRate
	for _, data := range response.Data.Result {
		vStr := data.Values[1].(string)
		v, err := strconv.ParseFloat(vStr, 64)
		if err != nil {
			return err, nil
		}
		result = append(result, TikvWritingByteRate{
			Instance: data.Metrics.Instance,
			Value:    v,
		})
	}
	return nil, result
}

func (b *TikvWritingByte) queryPrometheusQuery(prometheusAddr string, query string, result interface{}) error {
	prometheusSvc := fmt.Sprintf("http://%s/api/v1/query?query=%s", prometheusAddr, query)
	resp, err := http.Get(prometheusSvc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("resp is nil")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("resp is failed")
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return err
	}

	return nil
}
