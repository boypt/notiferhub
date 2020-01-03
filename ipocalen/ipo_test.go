package ipocalen

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

var (
	days = []string{
		`<div class="cal_date"><span></span>3日</div><div class="cal_content"><div class="cal_item"><b>申 购</b><ul><li><a href='/xg/detail/300813.html' title='微生物检测与控制技术系统产品、有机物分析仪器等制药装备的研发、制造和销售'>泰林生物</a></li><li class='zq'><a href='/kzz/detail/128093.html'>百川转债(债)</a></li><li class='zq'><a href='/kzz/detail/127015.html'>希望转债(债)</a></li><li class='zq'><a href='/kzz/detail/123040.html'>乐普转债(债)</a></li></ul></div><div class="cal_item"><b>上 市</b><ul><li class='zq'><a href='/kzz/detail/110063.html'>N鹰19转(债)</a></li></ul></div><div class="cal_item"><b>中签号</b><ul><li class='zq'><a href='/kzz/detail/113561.html'>正裕转债(债)</a></li></ul></div><div class="cal_item"><b>中签率</b><ul><li><a href='/xg/detail/002971.html' title='各类气体产品的研发、生产、销售、服务以及工业尾气回收循环利用'>和远气体</a></li><li><a href='/xg/detail/688178.html' title='公司专业提供先进环保技术装备开发、系统集成与环境问题整体解决方案,主营业务聚焦垃圾污染削减及修复业务、高难度废水处理业务等'>万德斯</a></li><li><a href='/xg/detail/603551.html' title='从事浴霸、集成吊顶等家居产品的研发、生产、销售及相关服务的提供'>奥普家居</a></li><li class='zq'><a href='/kzz/detail/128093.html'>百川转债(债)</a></li><li class='zq'><a href='/kzz/detail/127015.html'>希望转债(债)</a></li><li class='zq'><a href='/kzz/detail/123040.html'>乐普转债(债)</a></li></ul></div><div class="cal_item"><b >缴款日</b><ul><li class='zq'><a href='/kzz/detail/113561.html'>正裕转债(债)</a></li></ul></div></div>`,
		`<td valign="top"><div class="cal_date"><span></span>6日</div><div class="cal_content"><div class="cal_item"><b>申 购</b><ul><li><a href="/xg/detail/601816.html" title="京沪高速铁路及沿线车站的投资,建设,运营">京沪高铁</a></li></ul></div><div class="cal_item"><b>上 市</b><ul><li><a href="/xg/detail/688081.html" title="网络通信的军队专用视频指挥控制系统提供商,专注于视音频领域的技术创新和产品创新">兴图新科</a></li><li><a href="/xg/detail/002973.html" title="城乡环卫保洁、生活垃圾处置、市政环卫工程及其他环卫服务">侨银环保</a></li><li><a href="/xg/detail/688181.html" title="液晶显示材料的研发、生产和销售">八亿时空</a></li><li class="zq"><a href="/kzz/detail/128082.html">华锋转债(债)</a></li></ul></div><div class="cal_item"><b>中签号</b><ul><li><a href="/xg/detail/002971.html" title="各类气体产品的研发、生产、销售、服务以及工业尾气回收循环利用">和远气体</a></li><li><a href="/xg/detail/688178.html" title="公司专业提供先进环保技术装备开发、系统集成与环境问题整体解决方案,主营业务聚焦垃圾污染削减及修复业务、高难度废水处理业务等">万德斯</a></li><li><a href="/xg/detail/603551.html" title="从事浴霸、集成吊顶等家居产品的研发、生产、销售及相关服务的提供">奥普家居</a></li><li class="zq"><a href="/kzz/detail/113562.html">璞泰转债(债)</a></li></ul></div><div class="cal_item"><b>中签率</b><ul><li><a href="/xg/detail/300813.html" title="微生物检测与控制技术系统产品、有机物分析仪器等制药装备的研发、制造和销售">泰林生物</a></li></ul></div><div class="cal_item"><b>缴款日</b><ul><li><a href="/xg/detail/002971.html" title="各类气体产品的研发、生产、销售、服务以及工业尾气回收循环利用">和远气体</a></li><li><a href="/xg/detail/688178.html" title="公司专业提供先进环保技术装备开发、系统集成与环境问题整体解决方案,主营业务聚焦垃圾污染削减及修复业务、高难度废水处理业务等">万德斯</a></li><li><a href="/xg/detail/603551.html" title="从事浴霸、集成吊顶等家居产品的研发、生产、销售及相关服务的提供">奥普家居</a></li><li class="zq"><a href="/kzz/detail/113562.html">璞泰转债(债)</a></li></ul></div></div></td>`,
		`<td valign="top"><div class="cal_date"><span></span>8日</div><div class="cal_content"><div class="cal_item"><b>申 购</b><ul><li><a href="/xg/detail/688158.html" title="自主研发并提供计算,网络,存储等基础资源和构建在这些基础资源之上的基础IT架构产品,以及大数据,人工智能等产品,通过公有云,私有云,混合云三种模式为用户提供服务">优刻得</a></li></ul></div><div class="cal_item"><b>上 市</b><ul><li>无</li></ul></div><div class="cal_item"><b>中签号</b><ul><li><a href="/xg/detail/601816.html" title="京沪高速铁路及沿线车站的投资,建设,运营">京沪高铁</a></li></ul></div><div class="cal_item"><b>中签率</b><ul><li><a href="/xg/detail/688278.html" title="从事重组蛋白质及其长效修饰药物研发,生产及销售的创新型生物医药企业">特宝生物</a></li><li><a href="/xg/detail/688100.html" title="公司为聚焦于智慧公用事业领域的物联网综合应用解决方案提供商,致力于以物联网技术重塑电、水、气、热等能源的管理方式,以提供智慧能源管理完整解决方案为核心,并逐步向智慧消防、智慧路灯等领域拓展,是国内最早专业从事智慧公用事业的厂商之一">威胜信息</a></li></ul></div><div class="cal_item"><b>缴款日</b><ul><li><a href="/xg/detail/601816.html" title="京沪高速铁路及沿线车站的投资,建设,运营">京沪高铁</a></li></ul></div></div></td>`,
		`<td valign="top"><div class="cal_date"><span></span>9日</div><div class="cal_content"><div class="cal_item"><b>申 购</b><ul><li>无</li></ul></div><div class="cal_item"><b>上 市</b><ul><li>无</li></ul></div><div class="cal_item"><b>中签号</b><ul><li><a href="/xg/detail/688278.html" title="从事重组蛋白质及其长效修饰药物研发,生产及销售的创新型生物医药企业">特宝生物</a></li><li><a href="/xg/detail/688100.html" title="公司为聚焦于智慧公用事业领域的物联网综合应用解决方案提供商,致力于以物联网技术重塑电、水、气、热等能源的管理方式,以提供智慧能源管理完整解决方案为核心,并逐步向智慧消防、智慧路灯等领域拓展,是国内最早专业从事智慧公用事业的厂商之一">威胜信息</a></li></ul></div><div class="cal_item"><b>中签率</b><ul><li><a href="/xg/detail/688158.html" title="自主研发并提供计算,网络,存储等基础资源和构建在这些基础资源之上的基础IT架构产品,以及大数据,人工智能等产品,通过公有云,私有云,混合云三种模式为用户提供服务">优刻得</a></li></ul></div><div class="cal_item"><b>缴款日</b><ul><li><a href="/xg/detail/688278.html" title="从事重组蛋白质及其长效修饰药物研发,生产及销售的创新型生物医药企业">特宝生物</a></li><li><a href="/xg/detail/688100.html" title="公司为聚焦于智慧公用事业领域的物联网综合应用解决方案提供商,致力于以物联网技术重塑电、水、气、热等能源的管理方式,以提供智慧能源管理完整解决方案为核心,并逐步向智慧消防、智慧路灯等领域拓展,是国内最早专业从事智慧公用事业的厂商之一">威胜信息</a></li></ul></div></div></td>`,
		`<td valign="top"><div class="cal_date"><span></span>10日</div><div class="cal_content"><div class="cal_item"><b>申 购</b><ul><li>无</li></ul></div><div class="cal_item"><b>上 市</b><ul><li>无</li></ul></div><div class="cal_item"><b>中签号</b><ul><li><a href="/xg/detail/688158.html" title="自主研发并提供计算,网络,存储等基础资源和构建在这些基础资源之上的基础IT架构产品,以及大数据,人工智能等产品,通过公有云,私有云,混合云三种模式为用户提供服务">优刻得</a></li></ul></div><div class="cal_item"><b>中签率</b><ul><li>无</li></ul></div><div class="cal_item"><b>缴款日</b><ul><li><a href="/xg/detail/688158.html" title="自主研发并提供计算,网络,存储等基础资源和构建在这些基础资源之上的基础IT架构产品,以及大数据,人工智能等产品,通过公有云,私有云,混合云三种模式为用户提供服务">优刻得</a></li></ul></div></div></td>`,
	}
)

func TestFindTodayCalendar(t *testing.T) {

	print := func(html string) {
		d, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Error(err)
		}
		r := FindTodayCalendar(d.Selection)
		t.Logf("len(%d), %v", len(r), r)
	}

	for _, h := range days {
		print(h)
	}
}
