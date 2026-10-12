package almsvc

import (
	"context"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
)

func (svc Services) ConfigureGoogleCert(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	if svc.Gsvc.HasCert() {
		return nil
	}

	opt, err := svc.Queries.GetOption(ctx, "google-json")
	switch {
	case pgxutil.IsNotFound(err):
		l := almlog.FromContext(ctx)
		l.Warn("ConfigureGoogleCert: no certificate in database")
		return nil
	case err != nil:
		return err
	default:
		return svc.Gsvc.ConfigureCert(opt)
	}
}
