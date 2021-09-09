package almanack

import (
	"context"
	"encoding/json"

	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/stringutils"
)

func (svc Service) PopScheduledSiteChanges(ctx context.Context) (err error) {
	defer errutil.Trace(&err)

	configs, err := svc.Queries.PopScheduledSiteChanges(ctx, EditorsPicksLoc)
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
		svc.Printf("site data: no changes to %s", EditorsPicksLoc)
		return nil
	}
	svc.Printf("site data: updating %s", EditorsPicksLoc)

	// TODO: rollback
	if err = svc.PublishSiteConfig(ctx, currentConfig); err != nil {
		return err
	}

	return svc.Queries.CleanSiteData(ctx, EditorsPicksLoc)
}

func (svc Service) PublishSiteConfig(ctx context.Context, siteConfig *db.SiteDatum) (err error) {
	defer errutil.Trace(&err)

	data, err := json.MarshalIndent(siteConfig.Data, "", "  ")
	if err != nil {
		return err
	}
	msg := stringutils.First(MessageForLoc[siteConfig.Key], siteConfig.Key)
	if err = svc.ContentStore.UpdateFile(ctx, msg, siteConfig.Key, data); err != nil {

		return err
	}
	return nil
}
