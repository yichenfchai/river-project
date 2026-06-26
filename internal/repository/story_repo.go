package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type StoryRepository interface {
	List(ctx context.Context, page, pageSize int) ([]model.Story, int64, error)
	FindByID(ctx context.Context, id string) (*model.Story, error)
	Seed(ctx context.Context) error
}

type storyRepo struct {
	db *gorm.DB
}

func NewStoryRepo(db *gorm.DB) StoryRepository {
	return &storyRepo{db: db}
}

func (r *storyRepo) List(ctx context.Context, page, pageSize int) ([]model.Story, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Story{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	offset := (page - 1) * pageSize
	var stories []model.Story
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&stories).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	return stories, total, nil
}

func (r *storyRepo) FindByID(ctx context.Context, id string) (*model.Story, error) {
	var story model.Story
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&story).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrStoryNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &story, nil
}

func (r *storyRepo) Seed(ctx context.Context) error {
	seeds := []model.Story{
		{
			ID: UUID5("grand-canal-story", "隋炀帝与通济渠"),
			Title: "隋炀帝与通济渠", Content: `公元605年，隋炀帝杨广下令开凿连通黄河与淮河的通济渠。这项宏大的水利工程动用了百余万民夫，西起洛阳，东至淮河，全长约千里。

通济渠的开凿并非凭空而来。隋炀帝利用了先秦时期的鸿沟、汴渠等旧有水系，加以拓宽疏浚，使南北水运得以贯通。史载"发河南诸郡男女百余万，开通济渠"，工程之浩大可见一斑。

通济渠的建成使洛阳成为全国水运枢纽。《资治通鉴》记载："自西苑引谷、洛水达于河，引河通于淮。"自此，江南的粮食、丝绸可以经水路直抵都城，北方军队和物资也可快速南下。

然而这项工程也付出了沉重代价。大量劳役导致民生凋敝，成为隋朝迅速覆亡的原因之一。唐人皮日休在《汴河怀古》中写道："尽道隋亡为此河，至今千里赖通波。若无水殿龙舟事，共禹论功不较多。"这首诗精辟地评价了运河功过——虽然工程劳民伤财加速了隋亡，但千年以来运河确实造福了无数后人。

今天，通济渠故道仍然存在于河南、安徽等地，部分河段还保留着航运功能，是研究古代水利工程的珍贵实证。`,
			Topic: "历史", Emoji: "🏰", AgeGroup: "青少年", Likes: 128,
		},
		{
			ID: UUID5("grand-canal-story", "运河边的白鹭家族"),
			Title: "运河边的白鹭家族", Content: `在扬州段的古运河畔，生活着一个庞大的白鹭家族。每年春天，数百只白鹭从南方飞来，在运河两岸的柳树林中筑巢繁衍。

清晨六点，天刚蒙蒙亮，第一缕阳光洒在运河水面。白鹭们从巢穴中醒来，开始一天的觅食活动。它们三五成群，或伫立浅滩等待鱼虾，或在水面低飞巡视。白鹭飞行时姿态优美，长腿后伸，脖颈弯曲成优雅的S形，宛如运河上的白色精灵。

扬州的鸟类学家李教授已经跟踪观察这个白鹭族群长达十五年。他的研究记录显示，随着运河水质改善，白鹭的数量从最初的一百余只增长到现在的四百多只。白鹭的存在成为运河生态环境改善的最好证明。

2015年，扬州市政府在运河沿岸设立了白鹭保护区，种植了大片芦苇和柳树，为白鹭提供理想栖息地。保护区管理员老张每天都要沿着运河走上十公里，记录白鹭的数量和活动情况。"这些白鹭是运河的孩子，"老张说，"它们在这里生活了几百年，我们有责任保护好它们的家园。"

运河不仅是人类文明的遗产，更是无数生灵的家园。白鹭与运河的故事，是大自然与人类文明和谐共处的生动写照。`,
			Topic: "生态", Emoji: "🦅", AgeGroup: "儿童", Likes: 95,
		},
		{
			ID: UUID5("grand-canal-story", "古代船闸的秘密"),
			Title: "古代船闸的秘密", Content: `京口闸、瓜洲闸——这些古老的名字背后，藏着中国古代水利工程最精妙的设计。

大运河横贯南北，沿途地势起伏。从杭州到北京，海拔相差数十米。古人如何让船只"翻山越岭"？答案就是——船闸。

船闸的原理看似简单：在两个水域之间修筑闸门，通过蓄水和放水来调节水位高度，使船只像"坐电梯"一样逐级攀升或下降。但其工程实现却极为复杂。宋代科学家沈括在《梦溪笔谈》中详细记载了淮安船闸的工作原理："以闸节水，叠石为堰，引水注之。水满则舟过，水落则舟止。"

山东南旺分水枢纽被誉为"运河之心"。这里地处大运河最高点，被称为"水脊"。元代水利专家郭守敬设计了"引汶济运"方案，将汶河水引入运河，配合多级船闸系统，解决了这个工程难题。工程中使用了精确的水平测量技术，误差不超过两厘米，这在那个没有GPS和激光测距的年代堪称奇迹。

明代永乐年间，工部尚书宋礼在南旺修建了"十里闸"。这座闸门采用条石砌筑，闸门用厚达两尺的松木板制成，可以承受巨大的水压。闸门两侧还设有"鱼道"，让鱼类能够迁徙通过，这种生态意识在古代工程中实属罕见。

2014年，大运河南旺枢纽遗址被列入世界文化遗产。站在古闸遗址前，仿佛还能听到数百年前船工的号子声和流水冲击闸门的轰鸣。`,
			Topic: "工程", Emoji: "🏗", AgeGroup: "成人", Likes: 210,
		},
		{
			ID: UUID5("grand-canal-story", "船工号子的回响"),
			Title: "船工号子的回响", Content: `"嘿——呦——嘿——呦——"这苍凉有力的号子声曾经响彻京杭大运河两岸，是运河文化中最动人心魄的音符。

船工号子源于千年的船运劳作。逆流而上时，船工们需要合力拉纤，号子就是他们统一节奏的力量源泉。号子的种类繁多：起锚号子、拉纤号子、摇橹号子、过闸号子……每种号子都有独特的节奏和唱词。

在山东台儿庄，80岁的李德厚老人是最后一批会唱完整运河号子的老船工。"我十二岁就上船干活，"老人回忆道，"那时候从台儿庄到扬州，一趟要走二十多天。逆水时全靠人拉，一天只能走二十里。"他唱了一段起锚号子，苍老的嗓音中透着穿透岁月的力量："太阳一出照九洲呀——嗨呦——拉起来呀扯起来呀——嗨呦——"

运河号子不仅是劳动号令，更是船工们的精神寄托。歌词里唱生活艰辛，也唱对家乡的思念。有学者研究发现，不同河段的号子风格迥异：山东段豪迈粗犷，江南段婉转悠扬，这反映了各地的民风差异。

2011年，"运河船工号子"被列入国家级非物质文化遗产名录。然而随着机械动力船只普及，传统船工号子正在快速消亡。如今，只有在台儿庄古城、扬州古运河等地的文化表演中，还能听到这千年的音符。

运河号子是非物质文化遗产中的活态史诗，它记录了无数船工的汗水和智慧，是一份不能丢失的文化记忆。`,
			Topic: "民俗", Emoji: "🎭", AgeGroup: "成人", Likes: 73,
		},
		{
			ID: UUID5("grand-canal-story", "南水北调与运河新生"),
			Title: "南水北调与运河新生", Content: `大运河在最辉煌的时代过去后，部分河段因淤塞和水源不足而断航。然而21世纪的南水北调工程，让这条千年运河焕发了新生。

南水北调东线工程于2002年正式开工，2013年建成通水。工程建设者巧妙利用了京杭大运河的既有河道，从长江下游扬州抽引江水，利用大运河及其平行河道逐级提水北送。这不仅节省了巨额开挖成本，更让断航多年的运河河段重获生机。

在江苏徐州，大运河南水北调段的照片令人震撼：宽阔的河面碧波荡漾，两岸绿树成荫，大型调水泵站巍峨矗立。这里已经从一条废弃的古河道变身为现代化的调水通道。

工程带来的不仅是水资源调配。随着水量增加，运河沿线生态环境显著改善。数据显示，东线工程通水后，沿线湖泊湿地面积增加了约12%，地下水水位回升了1至3米。运河重现百舸争流的景象，内河航运也随之复兴。

当然工程也面临挑战。如何在调水的同时保护运河文化遗产？如何在工程推进中维持生态平衡？这些问题考验着当代水利工程师们的智慧。采用地下管道穿越古城区域、为鱼类建设洄游通道、保留沿岸古闸遗址——这些措施体现了现代工程对文化遗产的尊重。

从隋唐的漕运命脉，到今天的调水干渠，大运河始终在适应时代需求而变迁。南水北调给了古老运河第二次生命，也让"运河文化"这个千年IP在新时代继续传承。`,
			Topic: "现代水利", Emoji: "🌊", AgeGroup: "青少年", Likes: 156,
		},
	}

	for _, story := range seeds {
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&story).Error; err != nil {
			return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
		}
	}
	return nil
}
