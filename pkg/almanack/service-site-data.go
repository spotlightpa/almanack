package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/jackc/pgx/v4"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/common"
)

func (svc Service) PopScheduledSiteChanges(ctx context.Context, loc string) (err error) {
	defer errutil.Trace(&err)

	configs, err := svc.Queries.PopScheduledSiteChanges(ctx, loc)
	if err != nil {
		return err
	}

	var currentConfig *db.SiteDatum
	for _, config := range configs {
		if currentConfig == nil || config.ScheduleFor.After(currentConfig.ScheduleFor) {
			currentConfig = &config
		}
	}
	if currentConfig == nil {
		common.Logger.Printf("site data: no changes to %s", loc)
		return nil
	}
	common.Logger.Printf("site data: updating %s", loc)

	// TODO: rollback
	if err = svc.PublishSiteConfig(ctx, currentConfig); err != nil {
		return err
	}

	return svc.Queries.CleanSiteData(ctx, loc)
}

func (svc Service) PublishSiteConfig(ctx context.Context, siteConfig *db.SiteDatum) (err error) {
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

func (svc Service) UpdateSiteConfig(ctx context.Context, loc string, configs []ScheduledSiteConfig) ([]db.SiteDatum, error) {
	var dbConfigs []db.SiteDatum
	err := svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (err error) {
		defer errutil.Trace(&err)

		// Clear existing future entries before upserting current/future entries
		if err = q.DeleteSiteData(ctx, loc); err != nil {
			return err
		}
		for _, config := range configs {
			if err = q.UpsertSiteData(ctx, db.UpsertSiteDataParams{
				Key:         loc,
				Data:        config.Data,
				ScheduleFor: config.ScheduleFor,
			}); err != nil {
				return err
			}
		}
		dbConfigs, err = q.GetSiteData(ctx, loc)
		if err != nil {
			return err
		}
		if len(dbConfigs) == 0 {
			return fmt.Errorf("no item is currently scheduled")
		}

		// GetSiteData must return presorted configs!
		currentConfig := &dbConfigs[0]
		if err = svc.PublishSiteConfig(ctx, currentConfig); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		dbConfigs = nil
	}

	return
}
