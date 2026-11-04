package connector

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"go.mau.fi/mautrix-meta/pkg/messagix/socket"
	"go.mau.fi/mautrix-meta/pkg/messagix/table"
)

const defaultActiveStatusTimeout = 5 * time.Minute

type activeStatus struct {
	lock       sync.Mutex
	foreground bool
	idleTimer  *time.Timer
	generation uint64
}

// markActive reports the app as foregrounded to Meta (if it isn't already) and
// restarts the idle timer that reports it as backgrounded again.
func (m *MetaClient) markActive(ctx context.Context) {
	if !m.Main.Config.SendActiveStatus || m.Client == nil {
		return
	}
	timeout := m.Main.Config.ActiveStatusTimeout
	if timeout <= 0 {
		timeout = defaultActiveStatusTimeout
	}
	m.activeStatus.lock.Lock()
	defer m.activeStatus.lock.Unlock()
	if m.activeStatus.idleTimer != nil {
		m.activeStatus.idleTimer.Stop()
	}
	m.activeStatus.generation++
	generation := m.activeStatus.generation
	m.activeStatus.idleTimer = time.AfterFunc(timeout, func() { m.markIdle(generation) })
	if !m.activeStatus.foreground {
		m.activeStatus.foreground = true
		go m.reportAppState(zerolog.Ctx(ctx).WithContext(m.Main.Bridge.BackgroundCtx), table.FOREGROUND)
	}
}

func (m *MetaClient) markIdle(generation uint64) {
	m.activeStatus.lock.Lock()
	defer m.activeStatus.lock.Unlock()
	if m.activeStatus.generation != generation {
		// There was activity (or a disconnect) after this timer was started
		return
	}
	m.activeStatus.idleTimer = nil
	if m.activeStatus.foreground {
		m.activeStatus.foreground = false
		go m.reportAppState(m.UserLogin.Log.WithContext(m.Main.Bridge.BackgroundCtx), table.BACKGROUND)
	}
}

// resetActiveStatus forgets the reported state without sending anything,
// so that the next activity after a reconnect reports foreground again.
func (m *MetaClient) resetActiveStatus() {
	m.activeStatus.lock.Lock()
	defer m.activeStatus.lock.Unlock()
	if m.activeStatus.idleTimer != nil {
		m.activeStatus.idleTimer.Stop()
		m.activeStatus.idleTimer = nil
	}
	m.activeStatus.generation++
	m.activeStatus.foreground = false
}

func (m *MetaClient) reportAppState(ctx context.Context, state table.AppState) {
	cli := m.Client
	if cli == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := cli.ExecuteTasks(ctx, &socket.ReportAppStateTask{AppState: state, RequestID: uuid.NewString()})
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Int64("app_state", int64(state)).Msg("Failed to report app state")
	} else {
		zerolog.Ctx(ctx).Debug().Int64("app_state", int64(state)).Msg("Reported app state")
	}
}
