package metrics

import (
	"testing"
)

func Test(t *testing.T) {

	response := &struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metrics struct {
					Stage string `json:"instance"`
				} `json:"metric"`
				Values []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}{}

	re := &TikvWritingByte{}
	//response := ""
	err := re.queryPrometheusQuery("localhost:9090", "sum(rate(tikv_scheduler_stage_total{}[2m]))by(instance)", &response)

	if err != nil {
		t.Error(err)
	}
}
