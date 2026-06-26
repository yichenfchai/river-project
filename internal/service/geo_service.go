package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type GeoService interface {
	GetCanalIntro(ctx context.Context, ip string) (*CanalIntroResponse, error)
	GetCanalIntroByCoords(ctx context.Context, lat, lng float64) (*CanalIntroResponse, error)
}

type CanalIntroResponse struct {
	IsCanalCity bool            `json:"is_canal_city"`
	City        string          `json:"city"`
	Province    string          `json:"province"`
	IP          string          `json:"ip"`
	Intro       *CanalCityIntro `json:"intro,omitempty"`
}

type CanalCityIntro struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Era     string `json:"era"`
	POI     string `json:"poi"`
}

type geoService struct {
	log    *zap.Logger
	cache  sync.Map
	client *http.Client
}

func NewGeoService(log *zap.Logger) GeoService {
	return &geoService{
		log:    log,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *geoService) GetCanalIntro(ctx context.Context, ip string) (*CanalIntroResponse, error) {
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		ip = ""
	}

	city, province := "", ""
	if ip != "" {
		if cached, ok := s.cache.Load(ip); ok {
			if cc, ok := cached.(*cachedLocation); ok && time.Since(cc.At) < 24*time.Hour {
				city = cc.City
				province = cc.Province
			}
		}
		if city == "" {
			city, province = s.lookupIP(ctx, ip)
			if city != "" {
				s.cache.Store(ip, &cachedLocation{City: city, Province: province, At: time.Now()})
			}
		}
	}

	resp := &CanalIntroResponse{
		IsCanalCity: false,
		City:        city,
		Province:    province,
		IP:          ip,
	}

	intro := s.matchCity(city, province)
	if intro != nil {
		resp.IsCanalCity = true
		resp.Intro = intro
	} else {
		resp.Intro = defaultIntro
	}

	return resp, nil
}

func (s *geoService) GetCanalIntroByCoords(ctx context.Context, lat, lng float64) (*CanalIntroResponse, error) {
	name, province := findNearestCity(lat, lng)
	resp := &CanalIntroResponse{
		IsCanalCity: name != "",
		City:        name,
		Province:    province,
	}

	if name != "" {
		if intro, ok := canalCities[name]; ok {
			resp.Intro = &intro
		}
	} else {
		resp.Intro = defaultIntro
	}
	return resp, nil
}

func (s *geoService) lookupIP(ctx context.Context, ip string) (string, string) {
	services := []string{
		fmt.Sprintf("http://whois.pconline.com.cn/ipJson.jsp?ip=%s&json=true", ip),
		fmt.Sprintf("http://ip-api.com/json/%s?fields=city,regionName,country&lang=zh-CN", ip),
	}

	var lastErr error
	for _, url := range services {
		city, province, err := s.tryLookup(ctx, url)
		if err == nil && city != "" {
			return city, province
		}
		lastErr = err
	}

	s.log.Warn("IP地理位置查询全部失败，使用默认欢迎页",
		zap.String("ip", ip),
		zap.Error(lastErr),
	)
	return "", ""
}

func (s *geoService) tryLookup(ctx context.Context, url string) (city, province string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		City       string `json:"city"`
		RegionName string `json:"regionName"`
		Pro        string `json:"pro"`
		Country    string `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if result.Pro != "" {
		return result.City, result.Pro, nil
	}
	if result.Country != "" && result.Country != "中国" && result.Country != "China" {
		return "", "", fmt.Errorf("非境内IP")
	}
	return result.City, result.RegionName, nil
}

func (s *geoService) matchCity(city, province string) *CanalCityIntro {
	clean := city
	clean = strings.TrimSuffix(clean, "市")
	clean = strings.TrimSuffix(clean, "区")
	clean = strings.TrimSuffix(clean, "县")
	if intro, ok := canalCities[clean]; ok {
		return &intro
	}
	return nil
}

var defaultIntro = &CanalCityIntro{
	Title: "千里运河，流动的史诗",
	Content: `京杭大运河始建于公元前486年，是世界上开凿最早、里程最长的人工运河。它南起余杭（今杭州），北至涿郡（今北京），全长约1797公里，贯穿海河、黄河、淮河、长江、钱塘江五大水系。2500年来，大运河见证了中华民族的兴衰荣辱，承载了南北经济文化交流的辉煌篇章。2014年，中国大运河成功入选世界遗产名录。无论你身在何处，运河的精神与文脉始终奔流不息。`,
	Era: "公元前486年至今",
	POI: "京杭大运河",
}

type cachedLocation struct {
	City     string
	Province string
	At       time.Time
}

type cityCoord struct {
	Name     string
	Province string
	Lat      float64
	Lng      float64
}

var cityCoords = []cityCoord{
	{"杭州", "浙江省", 30.29, 120.15},
	{"嘉兴", "浙江省", 30.77, 120.75},
	{"湖州", "浙江省", 30.87, 120.09},
	{"苏州", "江苏省", 31.30, 120.63},
	{"无锡", "江苏省", 31.57, 120.30},
	{"常州", "江苏省", 31.81, 119.97},
	{"镇江", "江苏省", 32.19, 119.43},
	{"南京", "江苏省", 32.06, 118.80},
	{"扬州", "江苏省", 32.39, 119.43},
	{"淮安", "江苏省", 33.51, 119.14},
	{"宿迁", "江苏省", 33.96, 118.28},
	{"徐州", "江苏省", 34.27, 117.19},
	{"枣庄", "山东省", 34.86, 117.55},
	{"济宁", "山东省", 35.38, 116.58},
	{"聊城", "山东省", 36.45, 115.99},
	{"德州", "山东省", 37.43, 116.36},
	{"沧州", "河北省", 38.30, 116.84},
	{"廊坊", "河北省", 39.52, 116.70},
	{"天津", "天津市", 39.14, 117.19},
	{"北京", "北京市", 39.90, 116.41},
	{"洛阳", "河南省", 34.68, 112.44},
	{"开封", "河南省", 34.79, 114.31},
}

func findNearestCity(lat, lng float64) (name, province string) {
	minDist := math.MaxFloat64
	for _, c := range cityCoords {
		d := haversine(lat, lng, c.Lat, c.Lng)
		if d < minDist && d < 200 {
			minDist = d
			name = c.Name
			province = c.Province
		}
	}
	return
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

var canalCities = map[string]CanalCityIntro{
	"杭州": c("东南形胜，三吴都会", `杭州，位于京杭大运河最南端，自隋朝开凿江南运河起便成为南北水运的终点。南宋定都临安后，运河两岸商贾云集、市井繁华，湖墅八景名扬天下，拱宸桥横跨运河四百年，见证了无数漕船往来与文人墨客的题咏。白居易笔下的「江南忆，最忆是杭州」，沉淀着千年水韵与人文。大运河与西湖双遗产交相辉映，使杭州成为名副其实的人间天堂。

今天的杭州段运河依旧舟楫如织，运河文化广场、中国京杭大运河博物馆等文化设施讲述着这条千年水道的辉煌历史。运河两岸的慢行绿道、历史文化街区与现代都市交相融合，续写着运河南端的新篇章。`, "隋朝至今", "拱宸桥、运河文化广场"),
	"嘉兴": c("嘉禾秀水，运河水乡", `嘉兴地处江南水乡腹地，京杭大运河由杭州北上一路经过嘉兴，河道纵横、桥梁如虹。自隋唐以来，嘉兴便是运河沿线重要的漕运中转站和粮食集散地，素有「嘉禾一穰，江淮为之康」的美誉。西塘古镇、乌镇等水乡明珠依运河而生，白墙黛瓦、小桥流水之间，千年的运河水韵从未断绝。

如今的嘉兴运河段保留了多处历史码头和古桥梁，运河文化带建设与水乡旅游深度融合，传统丝绸、蓝印花布等非遗项目沿运河传承，让这座水乡名城焕发出新的生机。`, "隋朝至今", "西塘古镇、乌镇"),
	"湖州": c("湖笔之乡，运河水韵", `湖州坐落在太湖南岸，京杭大运河穿境而过。唐代诗人张志和在湖州写下「西塞山前白鹭飞，桃花流水鳜鱼肥」的千古名句，描绘的正是运河水乡的田园风光。作为江南著名的蚕桑丝绸产地和湖笔之乡，大运河为湖州带来了繁荣的商业贸易和文化交流。南浔古镇的百年丝商巨宅，见证了运河经济的富庶与辉煌。

今天的湖州运河沿线保存完好的古镇群落和自然湿地，构成了独特的运河水乡景观带。南浔、双林、练市等古镇依运河而生，江南的水韵文脉在这里静静流淌。`, "唐朝至今", "南浔古镇"),
	"苏州": c("君到姑苏见，人家尽枕河", `苏州是京杭大运河江南段最重要的节点城市之一。自隋唐以来，苏州因运河而兴，成为江南经济文化的中心。唐代诗人杜荀鹤写道「君到姑苏见，人家尽枕河」，道尽了苏州与运河共生共荣的历史。山塘河、上塘河等运河支脉串联起苏州古城的水网肌理，阊门、盘门等水陆城门见证了运河带来的商贸繁荣。

苏州古城至今保留着水陆并行、河街相邻的双棋盘格局，大运河苏州段已被列为世界文化遗产。七里山塘的桨声灯影、虎丘的千年塔影，仍在述说着运河留给这座东方水城的无尽风华。`, "隋朝至今", "山塘河、阊门"),
	"无锡": c("太湖明珠，运河名城", `无锡位于太湖北岸，京杭大运河穿城而过。自隋唐大运河贯通以来，无锡便成为江南地区重要的漕运枢纽和粮食集散中心，明清时期更是全国四大米市之一。运河两岸的南长街、清名桥历史街区集中展示了无锡作为运河工商业重镇的辉煌历史。

古运河畔的惠山古镇、鼋头渚等名胜古迹与运河文脉紧密相连。如今的清名桥古运河景区成为无锡的城市文化名片，千年运河在这里与现代化城市和谐共生。`, "隋朝至今", "清名桥、南长街"),
	"常州": c("中吴要辅，运河咽喉", `常州地处江南运河中段，自古便是三吴襟带之邦，百越舟车之会。隋唐大运河贯通后，常州成为连接南北的重要节点，漕运繁忙、市井繁华。宋代大文豪苏轼曾为常州运河写下「多谢残灯不嫌客，孤舟一夜许相依」的诗句，描绘了运河客旅的独特情怀。

常州运河段至今保留着青果巷、篦箕巷等历史街区和多处古桥梁遗址。天宁寺的钟声、红梅公园的烟柳，与千年运河一同构成了这座江南名城独特的文化底色。`, "隋朝至今", "青果巷、东坡公园"),
	"南京": c("金陵帝王州，运河龙江关", `南京地处长江下游，自古便是南北漕运的咽喉要地。明代永乐迁都北京前，大运河的漕船直达南京龙江关码头，秦淮河畔千帆竞泊、商贾云集，江东门外的龙江宝船厂正是郑和下西洋的起点，运河与海洋在此交汇。石头城下的运河水道历经六朝兴替，见证了金陵的帝王气韵与市井繁华。

明代南京作为留都，大运河仍是连接南北的经济命脉，玄武湖畔的黄册库藏天下户籍，秦淮河上的画舫灯影浸润着运河带来的江南富庶。挹江门外的运河水道至今仍可寻觅当年漕运码头的遗迹，大运河与长江的千里回澜共同滋养了南京这座千年古都。`, "明朝至今", "龙江关、秦淮河"),
	"镇江": c("江河交汇，天下第一江山", `镇江地处长江与京杭大运河的交汇之处，自古便是南北水运的咽喉要冲。隋唐以降，江南运河的漕船从镇江入长江北上，运河与长江的十字交汇造就了镇江天下第一江山的军事与商贸地位。金山寺的法海传说、北固山的三国遗踪、西津渡的千年古渡，都印证了这座城市与大运河的深厚渊源。

镇江的谏壁船闸是江南运河的北端门户，至今仍在发挥着重要的航运功能。西津渡古街的青石板路历经千年打磨，静静诉说着大运河留给这座山水名城的传奇往事。`, "隋朝至今", "西津渡、谏壁船闸"),
	"扬州": c("淮左名都，漕运枢纽", `扬州地处江淮之间、运河与长江交汇之处，自吴王夫差开凿邗沟起便与大运河血脉相连。两汉时期扬州已是东南重镇，隋唐大运河的贯通更使扬州一跃成为全国最大的商业都会和对外贸易港口。唐代诗人李白写道「故人西辞黄鹤楼，烟花三月下扬州」，杜牧在此十年一觉扬州梦，大明寺的钟声、瘦西湖的烟柳、古运河畔的文昌阁，无一不浸染着千年的水汽与文脉。

明清时期，扬州因漕运和盐业再度繁荣，两淮盐商在此创造了独特的园林文化和美食文化。今天的扬州运河段已成为世界文化遗产的重要组成部分，古邗沟遗址、东关古渡、扬州中国大运河博物馆等串联起这座千年运河名城的文化记忆。`, "公元前486年至今", "古邗沟、扬州古运河渡口"),
	"淮安": c("运河之都，漕运中心", `淮安位于黄、淮、运三河交汇之处，明清时期是漕运总督署和河道总督署的驻地，被史学家誉为运河之都。清代淮安设漕运总督统管全国漕运事务，每年数万艘漕船和数十万漕运壮丁在此集结北上，清江浦的繁华盛极一时。周恩来总理出生于此，淮安的运河水韵涵养了一代伟人的童年记忆。

今天的大运河淮安段保留了清江浦楼、总督漕运公署遗址等多处历史遗迹。里运河文化长廊串联起河下古镇、古末口等景点，讲述着这座运河之都的千年辉煌。`, "明清至今", "清江浦、总督漕运公署遗址"),
	"宿迁": c("水城宿迁，运河新韵", `宿迁位于苏北平原，京杭大运河自徐州南下进入宿迁，纵贯全境。明清时期，宿迁是大运河重要的漕运中转地和水利枢纽，皂河古镇的乾隆行宫见证了康乾盛世六下江南均在此驻跸的荣光。骆马湖与运河相依相连，构成了水城宿迁独特的自然与人文景观。

大运河宿迁段至今仍发挥着重要的航运和水利功能。皂河古镇、乾隆行宫、项王故里等景点串联起宿迁的运河文化脉络，让这座苏北水城在千年运河的滋养下续写新篇。`, "明清至今", "皂河古镇、乾隆行宫"),
	"徐州": c("五省通衢，运河要冲", `徐州地处苏鲁豫皖四省交界，自古便是兵家必争之地和南北交通枢纽。隋唐大运河开通后，徐州成为通济渠上的重要节点城市，唐宋时期商贾云集、舟车辏辐，有五省通衢之誉。徐州是大运河沿线唯一一座拥有汉文化核心区、汉墓群和汉画像石的城市，运河文化与两汉文明在这里交汇融合。

大运河徐州段保留了窑湾古镇、吕梁洪等历史文化遗产。今天的徐州大运河文化带建设与彭祖文化、汉文化交相辉映，让这座千年古城在运河的滋养下焕发出新的活力。`, "隋唐至今", "窑湾古镇"),
	"枣庄": c("运河古城，台儿庄", `枣庄地处鲁南，京杭大运河从南四湖穿境而过。台儿庄古城是京杭大运河沿线唯一一座保存完好的明清运河古城，被誉为活的古运河和京杭运河仅存的遗产村庄。明清时期台儿庄是运河漕运的重要枢纽，商贾云集，拥有三大名城的气象。1938年台儿庄大战更让这座古城名垂青史。

台儿庄古城重建后完整复原了明清运河码头的繁华风貌，古运河穿城而过、拱桥石阶、水街相依，成为大运河文化带上最璀璨的明珠之一。`, "明清至今", "台儿庄古城"),
	"济宁": c("运河之都，孔孟之乡", `济宁地处京杭大运河中段，元代运河裁弯取直后，济宁一跃成为运河上的重要商业城市，被称为运河之都。南旺分水枢纽是明代水利工程的杰出代表，解决了大运河制高点水源难题，被联合国教科文组织誉为工业革命前世界上最伟大的水利工程。济宁不仅是运河重镇，更是儒家文化的发祥地，曲阜的孔庙、孔府、孔林与运河文脉交相辉映。

今天的大运河济宁段保留了南旺分水枢纽遗址、太白楼、竹竿巷等历史文化遗产。运河文化与儒家文明在这座千年古城交融共生，续写着运河之都的辉煌篇章。`, "元代至今", "南旺分水枢纽、太白楼"),
	"聊城": c("江北水城，运河明珠", `聊城位于山东西部，京杭大运河穿城而过，东昌湖环抱古城，被誉为江北水城。明清时期聊城是运河沿线九大商埠之一，山陕会馆见证了运河商贸的繁荣，光岳楼俯瞰着运河的千年沧桑。清代聊城商人走南闯北，天下聊商的美名远播。

聊城古城至今保留着城中有湖、湖中有城的独特格局。东昌湖、光岳楼、山陕会馆和大运河一起构成了这座江北水城的文化标识，运河文脉在这座千年古城中静静流淌。`, "明清至今", "山陕会馆、光岳楼"),
	"德州": c("九达天衢，运河门户", `德州位于山东省西北部，京杭大运河北上进入河北的必经要地。明清时期德州是运河上重要的漕运中转站和粮食储备中心，有九达天衢和神京门户之誉。苏禄王墓见证了明代中外文化交流的历史，而运河带来的皇家粮仓文化使德州至今享有扒鸡之乡的美名。

大运河德州段保留了四女寺船闸、德州码头等水利航运遗迹。运河文化带建设与现代化城市发展并行不悖，让这座九达天衢的运河名城在新时代续写着南北通达的华章。`, "明清至今", "苏禄王墓、四女寺船闸"),
	"沧州": c("运河狮城，武术之乡", `沧州地处河北省东南部，京杭大运河南运河段穿城而过。作为运河上的重要节点，沧州因漕运而兴，铁狮镇水的传说流传千年，古老的沧州武术沿运河传播大江南北。吴桥杂技从运河码头走出国门，被誉为天下杂技第一乡。

南运河沧州段至今保留着多处古码头和漕运遗址。沧州铁狮子、吴桥杂技大世界、南运河生态廊道串联起这座运河古城的历史记忆与文化新貌。`, "明清至今", "沧州铁狮子、吴桥杂技大世界"),
	"廊坊": c("京畿重镇，运河通衢", `廊坊位于京津之间，北运河穿境而过。作为京畿要地，廊坊自古以来便是连接北京与江南的漕运要道上的重要一站。清代北运河的漕船经廊坊进入通州，京畿之地的便利交通和运河文化为这片土地注入了独特的历史底蕴。

大运河廊坊段北运河是京杭大运河北端的重要组成部分。香河段的运河生态修复工程让北运河重现碧波荡漾的美景，运河文化在这座京畿新城中焕发新的生机。`, "清代至今", "北运河香河段"),
	"天津": c("九河下梢，运河明珠", `天津位于海河入海口，南运河、北运河在此汇入海河，三岔河口正是这座城市的发祥地。元代京杭大运河全线贯通后，天津从一个小渔村一跃成为华北最重要的漕运枢纽和商业都会。明清时期，天津作为京师门户和运河枢纽，商贾云集、会馆林立、妈祖文化与运河文化在此交融共生。

今天的大运河天津段保留了三岔河口、杨柳青古镇、大悲院等历史遗迹。运河文化、妈祖文化与哏都的独特城市性格共同塑造了天津开放包容的文化品格。`, "元代至今", "三岔河口、杨柳青古镇"),
	"北京": c("天府之国，运河京师", `北京作为京杭大运河北端的终点，自元代起大运河便源源不断地将江南的漕粮、丝绸、茶叶输送至京师。通州作为运河北端的漕运码头，自古就有一京、二卫、三通州之说。元明清三代，运河维系着帝国京师的命脉，通惠河、什刹海、积水潭曾是千帆竞泊的繁华码头。通州燃灯塔矗立千年，是漕船进入京城的最后航标。

今天的大运河北京段保留了通州燃灯塔、大运河森林公园、白浮泉遗址等众多文化遗产。运河文化带建设与大运河国家文化公园的推进，让这条千年水道在首都续写着新的历史篇章。`, "元代至今", "通州燃灯塔、什刹海"),
	"洛阳": c("神都洛阳，运河中枢", `洛阳地处天下之中，隋唐大运河以洛阳为中心，南达余杭、北通涿郡，奠定了隋唐帝国的经济命脉。含嘉仓是隋唐时期最大的国家粮仓，可储粮数百万石，大运河的漕船将江南的稻米源源不断输送至此。武则天迁都洛阳后，神都的繁华与运河的便利密不可分，龙门石窟的万千佛龛见证了大运河带来的文化交融与艺术繁荣。

含嘉仓遗址、洛河古道、龙门石窟等历史遗迹诉说着隋唐大运河与神都洛阳的辉煌往事。作为丝绸之路与大运河的交汇点，洛阳在中华文明发展史上的地位无可替代。`, "隋唐", "含嘉仓遗址、龙门石窟"),
	"开封": c("东京梦华，运河旧都", `开封，古称汴州、汴梁，隋唐大运河通济渠段自洛阳东来经过开封，使其成为运河上的重要节点城市。北宋定都开封后，运河迎来了最辉煌的时期，汴河、蔡河、五丈河、金水河四水贯都，《清明上河图》描绘的正是汴河两岸的繁华盛景。北宋东京城是当时世界上最大的城市，运河从中承载了帝国经济命脉与文化繁荣的重任。

今天的开封保留了汴河遗址、大相国寺等与运河相关的历史文化遗产。清明上河园全景再现了宋代运河码头的繁华市井，让这座千年古都的运河记忆在此重生。`, "隋唐至北宋", "汴河遗址、清明上河园"),
}

func c(title, content, era, poi string) CanalCityIntro {
	return CanalCityIntro{Title: title, Content: content, Era: era, POI: poi}
}
