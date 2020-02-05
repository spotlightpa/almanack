package almanack

import "time"

type ScheduledArticleService struct {
}

type ScheduledArticle struct {
	Article
	Body        string
	ScheduleFor *time.Time
}
