-- Grand Canal Guardian 数据库初始化
-- PostgreSQL 16

-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(32) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(128) UNIQUE NOT NULL,
    nickname VARCHAR(64) DEFAULT '',
    avatar_url VARCHAR(512) DEFAULT '',
    bio VARCHAR(500) DEFAULT '',
    role VARCHAR(16) DEFAULT 'user',
    points INTEGER DEFAULT 0,
    rank_title VARCHAR(32) DEFAULT '青铜守护者',
    status VARCHAR(16) DEFAULT 'active',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 帖子表
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    images TEXT DEFAULT '[]',
    tags TEXT DEFAULT '[]',
    topic VARCHAR(32) DEFAULT 'share',
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    status VARCHAR(16) DEFAULT 'published',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_topic ON posts(topic);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_deleted_at ON posts(deleted_at);


-- 点赞表
CREATE TABLE IF NOT EXISTS post_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

CREATE INDEX idx_post_likes_post_id ON post_likes(post_id);
CREATE INDEX idx_post_likes_user_id ON post_likes(user_id);

-- 评论表
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at);

-- 题目表
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    options TEXT NOT NULL,   -- JSON array
    answer INTEGER NOT NULL, -- 0-indexed correct option
    difficulty VARCHAR(16) DEFAULT 'easy',
    category VARCHAR(32) DEFAULT '运河文化',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_questions_difficulty ON questions(difficulty);

-- 答题记录表
CREATE TABLE IF NOT EXISTS quiz_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    question_id UUID NOT NULL,
    correct BOOLEAN NOT NULL,
    score INTEGER DEFAULT 0,
    streak INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_quiz_records_user_id ON quiz_records(user_id);
CREATE INDEX idx_quiz_records_created_at ON quiz_records(created_at);

-- 种子数据: 测试用户 (密码: test1234)
INSERT INTO users (id, username, password_hash, email, nickname, role) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@gcg.dev', '运河管理员', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'monitor', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'monitor@gcg.dev', '生态监测员小王', 'monitor'),
    ('00000000-0000-0000-0000-000000000003', 'testuser', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user@gcg.dev', '运河小舟', 'user')
ON CONFLICT (username) DO NOTHING;

-- 种子数据: 示例题目
INSERT INTO questions (content, options, answer, difficulty, category) VALUES
    ('京杭大运河全长约多少公里？', '["约1000公里","约1400公里","约1800公里","约2200公里"]', 2, 'easy', '运河历史'),
    ('京杭大运河开凿最早可以追溯到哪个时期？', '["秦朝","春秋战国","汉朝","唐朝"]', 1, 'easy', '运河历史'),
    ('京杭大运河南起杭州，北至哪个城市？', '["天津","北京","济南","沧州"]', 1, 'easy', '运河地理'),
    ('以下哪条河流不属于京杭大运河沟通的五大水系？', '["海河","黄河","淮河","珠江"]', 3, 'medium', '运河地理'),
    ('隋炀帝时期开凿的大运河以哪个城市为中心？', '["长安","洛阳","开封","扬州"]', 1, 'medium', '运河历史'),
    ('京杭大运河在哪个朝代达到鼎盛？', '["宋朝","元朝","明朝","清朝"]', 1, 'medium', '运河历史'),
    ('南水北调东线工程主要利用了大运河的哪一段？', '["江南运河","里运河","鲁运河","南运河"]', 1, 'hard', '现代水利'),
    ('大运河沿岸的世界文化遗产点不包括以下哪个？', '["苏州古典园林","曲阜三孔","杭州西湖","敦煌莫高窟"]', 3, 'hard', '文化遗产'),
    ('大运河生态保护中"以鱼养水"的主要原理是什么？', '["增加鱼类数量","滤食性鱼类净化水质","增加生物多样性","减少水草生长"]', 1, 'medium', '生态保护'),
    ('京杭大运河被列入《世界遗产名录》是在哪一年？', '["2012年","2014年","2016年","2018年"]', 1, 'hard', '文化遗产')
ON CONFLICT DO NOTHING;
