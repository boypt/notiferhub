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

	text := `

var hq_str_sh000001="上证指数,2912.9956,2909.9746,2903.6483,2918.3030,2901.4523,0,0,42932693,45750871878,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2019-11-12,10:16:25,00,";
var hq_str_sz399001="深证成指,9675.561,9680.574,9627.075,9699.416,9614.564,0.000,0.000,7433875970,83690161241.604,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-12,10:16:30,00";
var hq_str_sz399006="创业板指,1676.427,1673.129,1667.498,1679.728,1663.957,0.000,0.000,2147767168,29643660748.580,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-12,10:16:30,00";
var hq_str_rt_hkHSI="HSI,恒生指数,27064.260,26926.551,27064.260,26942.682,26991.379,64.830,0.240,0.000,0.000,17854145.885,0,0.000,0.000,30280.119,24896.869,2019/11/12,10:16:28,,,,,,";
var hq_str_rt_hkHSCEI="HSCEI,恒生中国企业指数,10672.880,10613.631,10677.040,10623.990,10652.330,38.700,0.360,0.000,0.000,3810008.832,0,0.000,0.000,11881.680,9731.890,2019/11/12,10:16:28,,,,,,";
`
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
			got, err := StockIndexText(tt.args.stockText, true, tt.args.onlyToday)
			if (err != nil) != tt.wantErr {
				t.Errorf("StockIndexText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%v", got)
		})
	}
}

func Test_parseHKidx(t *testing.T) {
	type args struct {
		val       string
		onlyToday bool
	}
	tests := []struct {
		name    string
		args    args
		want    *Price
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{"HSI,恒生指数,27064.260,26926.551,27064.260,27038.160,27038.160,111.610,0.410,0.000,0.000,1976917.121,0,0.000,0.000,30280.119,24896.869,2019/11/12,09:30:24,,,,,,", true}, nil, false},
		{"1", args{"HSCEI,恒生中国企业指数,10672.880,10613.631,10672.880,10663.990,10666.960,53.330,0.500,0.000,0.000,457510.528,0,0.000,0.000,11881.680,9731.890,2019/11/12,09:30:24,,,,,,", true}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHKidx(tt.args.val, true, tt.args.onlyToday)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHKidx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(got.Display())
		})
	}
}

func Test_parseASidx(t *testing.T) {
	type args struct {
		val       string
		onlyToday bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{"上证指数,2912.9956,2909.9746,2917.3611,2917.6918,2910.0952,0,0,13635326,14744073631,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2019-11-12,09:40:30,00,", true}, false},
		{"2", args{"深证成指,9675.561,9680.574,9695.694,9696.392,9666.304,0.000,0.000,2463237567,27757482443.618,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-12,09:40:36,00", true}, false},
		{"3", args{"深证成指,9675.561,9680.574,9695.694,9696.392,9666.304,0.000,0.000,2463237567,27757482443.618,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,0,0.000,2019-11-11,09:40:36,00", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseASidx(tt.args.val, true, tt.args.onlyToday)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseASidx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Log(got.Display())
			}
		})
	}
}
