package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carlmjohnson/errutil"
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

func (svc Service) UpdateSiteConfig(ctx context.Context, loc string, configs []ScheduledSiteConfig) (dbConfigs []db.SiteDatum, err error) {
	defer errutil.Trace(&err)

	// TODO: Add transactions
	// Clear existing future entries before upserting current/future entries
	if err = svc.Queries.DeleteSiteData(ctx, loc); err != nil {
		return nil, err
	}
	for _, config := range configs {
		if err = svc.Queries.UpsertSiteData(ctx, db.UpsertSiteDataParams{
			Key:         loc,
			Data:        config.Data,
			ScheduleFor: config.ScheduleFor,
		}); err != nil {
			return nil, err
		}
	}
	dbConfigs, err = svc.Queries.GetSiteData(ctx, loc)
	if err != nil {
		return nil, err
	}
	if len(dbConfigs) == 0 {
		return nil, fmt.Errorf("no item is currently scheduled")
	}
	// GetSiteData must return presorted configs!
	currentConfig := &dbConfigs[0]
	if err = svc.PublishSiteConfig(ctx, currentConfig); err != nil {
		return nil, err
	}
	return dbConfigs, nil
}
