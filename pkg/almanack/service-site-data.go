package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/jackc/pgx/v4"
	"github.com/spotlightpa/almanack/internal/db"
	"golang.org/x/exp/slog"
)

func (svc Services) PopScheduledSiteChanges(ctx context.Context, loc string) error {
	l := slog.FromContext(ctx)
	err := svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (txerr error) {
		defer errutil.Trace(&txerr)

		configs, txerr := q.PopScheduledSiteChanges(ctx, loc)
		if txerr != nil {
			return txerr
		}

		var currentConfig *db.SiteDatum
		for _, config := range configs {
			if currentConfig == nil || config.ScheduleFor.After(currentConfig.ScheduleFor) {
				currentConfig = &config
			}
		}
		if currentConfig == nil {
			l.Info("Services.PopScheduledSiteChanges: no changes", "location", loc)
			return nil
		}
		l.Info("Services.PopScheduledSiteChanges: updating", "location", loc)

		return svc.PublishSiteConfig(ctx, currentConfig)
	})
	if err != nil {
		return err
	}
	return svc.Queries.CleanSiteData(ctx, loc)
}

func (svc Services) PublishSiteConfig(ctx context.Context, siteConfig *db.SiteDatum) (err error) {
	defer errutil.Trace(&err)

	data, err := json.MarshalIndent(siteConfig.Data, "", "  ")
	if err != nil {
		return err
	}
	msg := MessageForLoc(siteConfig.Key)
	if err = svc.ContentStore.UpdateFile(ctx, msg, siteConfig.Key, data); err != nil {

		return err
	}
	return nil
}

type ScheduledSiteConfig struct {
	ScheduleFor time.Time `json:"schedule_for"`
	Data        db.Map    `json:"data"`
}

func (svc Services) UpdateSiteConfig(ctx context.Context, loc string, configs []ScheduledSiteConfig) ([]db.SiteDatum, error) {
	var dbConfigs []db.SiteDatum
	err := svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (txerr error) {
		defer errutil.Trace(&txerr)

		// Clear existing future entries before upserting current/future entries
		if txerr = q.DeleteSiteData(ctx, loc); txerr != nil {
			return txerr
		}
		for _, config := range configs {
			if txerr = q.UpsertSiteData(ctx, db.UpsertSiteDataParams{
				Key:         loc,
				Data:        config.Data,
				ScheduleFor: config.ScheduleFor,
			}); txerr != nil {
				return txerr
			}
		}
		dbConfigs, txerr = q.GetSiteData(ctx, loc)
		if txerr != nil {
			return txerr
		}
		if len(dbConfigs) == 0 {
			return fmt.Errorf("no item is currently scheduled")
		}

		// GetSiteData must return presorted configs!
		currentConfig := &dbConfigs[0]
		if txerr = svc.PublishSiteConfig(ctx, currentConfig); txerr != nil {
			return txerr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dbConfigs, nil
}
