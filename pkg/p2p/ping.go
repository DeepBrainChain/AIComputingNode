package p2p

import (
	"time"

	"AIComputingNode/pkg/log"
)

func (hio *HostInfo) StartPingService() {
	if hio.PingService == nil {
		return
	}

	go func() {
		interval := 30 * time.Second
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-hio.PingCtx.Done():
				return
			case <-timer.C:
			}

			conns := hio.Host.Network().Conns()
			for _, conn := range conns {
				ch := hio.PingService.Ping(hio.PingCtx, conn.RemotePeer())
				for i := 0; i < 5; i++ {
					res := <-ch
					if res.Error != nil {
						log.Logger.Warnf("ping failed with %s : %v", conn.RemotePeer().String(), res.Error)
					} else {
						log.Logger.Debugf("ping %s in %v", conn.RemotePeer().String(), res.RTT)
					}
				}
			}
			timer.Reset(interval)
		}
	}()
}

func (hio *HostInfo) StopPingService() {
	hio.PingStopCancel()
}