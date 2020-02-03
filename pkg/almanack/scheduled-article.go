package almanack

import "time"

type ScheduledArticleService struct {
}

type ScheduledArticle struct {
	Article
	PubTime *time.Time
}
