package stock

import (
	"testing"
)

func Test_GetSinaStockTest(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{"sh000001,"}, false},
		{"1", args{"sz399001, sz399006"}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSinaStockText(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSinaStock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(got)
		})
	}
}

func TestStockIndexText(t *testing.T) {
	type args struct {
		stockText string
		onlyToday bool
	}
	text := `var hq_str_sh000001="上证指数,2993.9773,2978.7144,2964.1849,2998.7594,2962.8447,0,0,149742107,173867227160,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2019-11-08,15:01:59,00,";
var hq_str_sz399001="深证成指,9985.197,9917.487,9895.337,10008.515,9890.996,0.000,0.000,23412425421,275554236920.681,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-08,15:00:03,00";
var hq_str_sz399006="创业板指,1724.402,1715.575,1711.216,1731.003,1709.959,0.000,0.000,6583968052,97102176189.150,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-08,15:00:03,00";`
	tests := []struct {
		name string
		args args
		// want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{text, false}, false},
		{"2", args{text, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StockIndexText(tt.args.stockText, tt.args.onlyToday)
			if (err != nil) != tt.wantErr {
				t.Errorf("StockIndexText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%v", got)
		})
	}
}
