package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// PostRepository 帖子数据访问接口
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, opts PostListOptions) ([]model.Post, int64, error)
	Update(ctx context.Context, post *model.Post) error
	SoftDelete(ctx context.Context, id string) error

	ToggleLike(ctx context.Context, postID, userID string) (isLiked bool, likeCount int, err error)
	IsLikedByUser(ctx context.Context, postID, userID string) (bool, error)

	CreateComment(ctx context.Context, comment *model.Comment) error
	FindCommentByID(ctx context.Context, id string) (*model.Comment, error)
	ListComments(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error)
	SoftDeleteComment(ctx context.Context, id string) error

	UpdateStatus(ctx context.Context, id, status string) error

	// 批量查询作者信息（用于组装响应）
	FindUsersByIDs(ctx context.Context, ids []string) (map[string]model.UserJSON, error)

	// 统计
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	Seed(ctx context.Context) error
}

type PostListOptions struct {
	Page     int
	PageSize int
	Topic    string
	Tag      string
	Keyword  string
	Status   string
	Sort     string // "created_at" | "like_count"
}

type postRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(ctx context.Context, post *model.Post) error {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *postRepo) FindByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &post, nil
}

func (r *postRepo) List(ctx context.Context, opts PostListOptions) ([]model.Post, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Post{})

	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}
	if opts.Topic != "" {
		q = q.Where("topic = ?", opts.Topic)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", like, like)
	}
	if opts.Tag != "" {
		q = q.Where("tags LIKE ?", "%"+opts.Tag+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	order := "created_at DESC"
	switch opts.Sort {
	case "like_count":
		order = "like_count DESC"
	case "created_at_asc":
		order = "created_at ASC"
	}

	offset := (opts.Page - 1) * opts.PageSize
	var posts []model.Post
	if err := q.Order(order).Offset(offset).Limit(opts.PageSize).Find(&posts).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return posts, total, nil
}

func (r *postRepo) Update(ctx context.Context, post *model.Post) error {
	if err := r.db.WithContext(ctx).Save(post).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *postRepo) SoftDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Post{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	return nil
}

func (r *postRepo) ToggleLike(ctx context.Context, postID, userID string) (bool, int, error) {
	var like model.PostLike
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postID, userID).
		First(&like).Error

	if err == nil {
		if delErr := r.db.WithContext(ctx).Delete(&like).Error; delErr != nil {
			return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, delErr)
		}
		if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; updErr != nil {
			return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
		}
		var post model.Post
		_ = r.db.WithContext(ctx).Select("like_count").Where("id = ?", postID).First(&post).Error
		return false, post.LikeCount, nil
	}

	if err != gorm.ErrRecordNotFound {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	like = model.PostLike{PostID: postID, UserID: userID}
	if createErr := r.db.WithContext(ctx).Create(&like).Error; createErr != nil {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, createErr)
	}
	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; updErr != nil {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}
	var post model.Post
	_ = r.db.WithContext(ctx).Select("like_count").Where("id = ?", postID).First(&post).Error
	return true, post.LikeCount, nil
}

func (r *postRepo) IsLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count > 0, nil
}

func (r *postRepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", comment.PostID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; updErr != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}
	return nil
}

func (r *postRepo) FindCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrCommentNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &comment, nil
}

func (r *postRepo) ListComments(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	var comments []model.Comment
	if err := q.Order("created_at ASC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return comments, total, nil
}

func (r *postRepo) SoftDeleteComment(ctx context.Context, id string) error {
	var comment model.Comment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewDefault(apperrors.ErrCommentNotFound)
		}
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	result := r.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}

	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", comment.PostID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error; updErr != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}

	return nil
}

func (r *postRepo) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	return nil
}

func (r *postRepo) FindUsersByIDs(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
	if len(ids) == 0 {
		return map[string]model.UserJSON{}, nil
	}
	dedup := make([]string, 0, len(ids))
	seen := make(map[string]bool)
	for _, id := range ids {
		if !seen[id] {
			dedup = append(dedup, id)
			seen[id] = true
		}
	}

	var users []model.User
	if err := r.db.WithContext(ctx).
		Select("id, username, nickname, avatar_url, role").
		Where("id IN ?", dedup).
		Find(&users).Error; err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	result := make(map[string]model.UserJSON, len(users))
	for _, u := range users {
		result[u.ID] = model.UserJSON{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			Role:      u.Role,
		}
	}
	return result, nil
}

func (r *postRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func (r *postRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func (r *postRepo) seedLookupUsers(ctx context.Context) (userID, monitorID, adminID string, err error) {
	var users []model.User
	if err := r.db.WithContext(ctx).
		Where("username IN ?", []string{"111user", "111monitor", "111admin"}).
		Find(&users).Error; err != nil {
		return "", "", "", err
	}
	for _, u := range users {
		switch u.Username {
		case "111user":
			userID = u.ID
		case "111monitor":
			monitorID = u.ID
		case "111admin":
			adminID = u.ID
		}
	}
	return userID, monitorID, adminID, nil
}

func (r *postRepo) Seed(ctx context.Context) error {
	user111ID, monitor111ID, admin111ID, err := r.seedLookupUsers(ctx)
	if err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if user111ID == "" {
		return nil
	}

	type postSeed struct {
		id        string
		title     string
		content   string
		topic     string
		likeCount int
		comments  []string
	}

	postSeeds := []postSeed{
		{
			id:      UUID5("grand-canal-post", "1"),
			title:   "扬州古运河畔的生态调查",
			content: "上周沿着扬州段运河进行了生态观察。水质明显比前几年好了很多，河面上能看到白鹭在觅食。沿岸种植了大片芦苇，形成了完整的湿地生态系统。特别惊喜的是发现了几只震旦鸦雀，这种鸟对环境要求很高。古运河不仅是一条水道，更是一条生态走廊。希望大家一起来保护运河生态，让这条千年水道继续生机勃勃。",
			topic:   "ecology", likeCount: 42,
			comments: []string{
				"看到了白鹭！上个月我也去了，还拍到了翠鸟，运河生态确实在变好。",
				"震旦鸦雀很珍稀，能在这里看到说明运河水质真的改善了很多。",
				"我是扬州的，政府确实投入了很多资金治理运河，成效显著。",
				"建议标记一下具体的观鸟地点，方便大家去观察。",
			},
		},
		{
			id:      UUID5("grand-canal-post", "2"),
			title:   "运河边的非遗：台儿庄船工号子采风记录",
			content: "专程去台儿庄古城探访了船工号子的传承人李爷爷。老人家今年80岁了，是最后一批在运河上拉过纤的船工。他为我们唱了起锚号子和拉纤号子，苍凉的嗓音里是全运河船工的集体记忆。李爷爷说，现在会唱完整号子的人不超过五个了。台儿庄古城虽然有表演，但那和真实的船工号子是两回事。非遗保护不能只停留在表演层面，应该做好记录和传承。",
			topic:   "culture", likeCount: 67,
			comments: []string{
				"太珍贵了！老人家的声音就是活的历史，希望有关部门能够做好录音保存工作。",
				"我去台儿庄旅游时听过表演版，但和真实的船工号子确实不一样，文化不能只做表面功夫。",
				"非遗传承真的需要更多关注，很多老手艺和习俗都在慢慢消失。",
			},
		},
		{
			id:      UUID5("grand-canal-post", "3"),
			title:   "从杭州到苏州：骑行运河线路实测",
			content: "刚完成杭州到苏州的运河沿线骑行，全程约280公里，历时三天。路线大部分沿着运河绿道，路况不错，沿途经过塘栖古镇、乌镇、桐乡、嘉兴等地。推荐几个值得停留的点：塘栖古镇保存完好的古码头、乌镇段的运河夜景、嘉兴南湖。路上补给方便，但建议避开七八月高温时段。这条线路既锻炼了身体，又深度体验了运河文化，强烈推荐。",
			topic:   "share", likeCount: 89,
			comments: []string{
				"收藏了！一直想走这条线，你的攻略太有用了。请问住宿是住民宿还是自带帐篷？",
				"280公里三天有点赶，我之前走了四天，在乌镇多停了一天，感觉更舒服。",
				"塘栖古镇真的值得一逛，那些老码头的石板路很有年代感。",
			},
		},
		{
			id:      UUID5("grand-canal-post", "4"),
			title:   "大运河考古新发现：北宋古闸浮出水面",
			content: "最近在安徽宿州段大运河遗址考古有了重大发现——一座北宋时期的船闸遗址出土。考古队清理出了完整的石砌闸墙、木桩基础和陶质闸板碎片。最震撼的是闸墙上保留的绳缆磨损痕迹，证明了千百年来船只通行留下的岁月印记。这个发现填补了宋代运河工程技术的重要空白。考古领队表示，这座闸的规模和工艺水平远超预期，证明了北宋时期运河工程技术已经非常成熟。",
			topic:   "share", likeCount: 124,
			comments: []string{
				"太震撼了！能看到千年闸墙上的绳痕，感觉穿越了时空与古人对话。",
				"考古人的工作太不容易了，致敬。希望能尽快建立博物馆让更多人参观。",
			},
		},
		{
			id:      UUID5("grand-canal-post", "5"),
			title:   "运河水质监测志愿者的日常",
			content: "作为一名环境工程专业的学生，我加入了运河水质监测志愿者队伍。每周在固定河段采集水样，检测pH值、溶解氧、总磷等指标。最近三个月的数据显示，河段水质稳定在III类水标准以上。遇到的最大问题其实是沿岸的生活垃圾，虽然设置了垃圾桶但总有人随手丢弃。我们还在暑假组织了沿岸净滩活动，清理了200多公斤垃圾。保护运河需要每个人的参与。",
			topic:   "ecology", likeCount: 56,
			comments: []string{
				"为你们点赞！环保志愿活动很有意义，请问怎么加入？有联系方式吗？",
				"水质数据可以公开分享吗？我们学校的环保社团也想参考一下。",
				"随手丢垃圾的情况确实普遍，建议加强宣传和执法，在重点区域设立监控。",
			},
		},
	}

	now := time.Now()
	for _, ps := range postSeeds {
		post := model.Post{
			ID: ps.id, AuthorID: user111ID,
			Title: ps.title, Content: ps.content, Topic: ps.topic,
			LikeCount: ps.likeCount, CommentCount: len(ps.comments),
			Status: "approved", CreatedAt: now, UpdatedAt: now,
		}
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&post).Error; err != nil {
			return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
		}

		for idx, content := range ps.comments {
			commentAuthorID := user111ID
			if idx%3 == 1 {
				commentAuthorID = monitor111ID
			} else if idx%3 == 2 {
				commentAuthorID = admin111ID
			}

			comment := model.Comment{
				ID: UUID5("grand-canal-comment", fmt.Sprintf("%s-%d", ps.id, idx)),
				PostID: ps.id, AuthorID: commentAuthorID,
				Content: content, CreatedAt: now.Add(time.Duration(idx+1) * time.Minute),
			}
			if err := r.db.WithContext(ctx).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&comment).Error; err != nil {
				return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
			}
		}
	}

	return nil
}
