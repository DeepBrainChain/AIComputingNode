package host

import (
	"context"
	"time"

	"AIComputingNode/pkg/log"
)

func (hio *HostInfo) StartPingService(ctx context.Context) {
	if hio.PingService == nil {
		return
	}

	go func() {
		interval := 60 * time.Second
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}

			conns := hio.Host.Network().Conns()
			for _, conn := range conns {
				pingCtx, pingStopCancel := context.WithCancel(ctx)
				ch := hio.PingService.Ping(pingCtx, conn.RemotePeer())
				for i := 0; i < 5; i++ {
					res := <-ch
					if res.Error != nil {
						log.Logger.Warnf("ping failed with %s : %v", conn.RemotePeer().String(), res.Error)
					} else {
						log.Logger.Debugf("ping %s in %v", conn.RemotePeer().String(), res.RTT)
					}
				}
				pingStopCancel()
			}
			timer.Reset(interval)
		}
	}()
}
