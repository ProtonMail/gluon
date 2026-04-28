package connectionlimiter

import (
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/sirupsen/logrus"
)

type ConnectionLimiter interface {
	//TryBind tries to bind a given sessionID to a client.
	TryBind(sessionID int, id imap.IMAPID, useFallback bool) (allowed bool, key Client, current int, max int)

	//Unbind unbinds a given sessionID from a client.
	Unbind(sessionID int)
}

type limiter struct {
	mu             sync.Mutex
	limits         Limits
	fallbackLimits Limits

	//sessionID mapping to normalized client key
	sessionClient map[int]Client

	//normalized key mapping to current open sessions
	clientCount map[Client]int

	log *logrus.Entry
}

func NewConnectionLimiter(limits, fallbackLimits Limits) ConnectionLimiter {
	return newLimiter(limits, fallbackLimits)
}

func newLimiter(limits, fallbackLimits Limits) *limiter {
	log := logrus.WithFields(logrus.Fields{
		"pkg":    "gluon/connectionlimiter",
		"limits": limits,
	})

	return &limiter{
		limits:         limits,
		fallbackLimits: fallbackLimits,
		sessionClient:  make(map[int]Client),
		clientCount:    make(map[Client]int),
		log:            log,
	}
}

func (l *limiter) TryBind(sessionID int, id imap.IMAPID, useFallback bool) (allowed bool, key Client, current int, max int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key = normalizeClientKey(id)
	if useFallback {
		l.log.WithField("fallbackLimits", l.fallbackLimits).Debug("Using fallback limits")
	}

	// already bound to this client, no-op allow
	if prev, ok := l.sessionClient[sessionID]; ok && prev == key {
		maxUsages := l.maxForKey(key, useFallback)

		l.log.WithFields(logrus.Fields{
			"sessionID": sessionID,
			"client":    key,
			"current":   l.clientCount[key],
			"max":       maxUsages,
		}).Info("Already bound to this client, no-op allow")

		return true, key, l.clientCount[key], maxUsages
	}

	// if rebind, release the old key first
	if prev, ok := l.sessionClient[sessionID]; ok {
		if c := l.clientCount[prev]; c > 0 {
			l.log.WithFields(logrus.Fields{
				"sessionID": sessionID,
				"client":    prev,
				"current":   c,
			}).Info("Releasing old client")

			l.clientCount[prev] = c - 1
		}

	}

	max = l.maxForKey(key, useFallback)
	cur := l.clientCount[key]

	if max > 0 && cur >= max {
		delete(l.sessionClient, sessionID)

		return false, key, cur, max
	}

	l.clientCount[key] = cur + 1
	l.sessionClient[sessionID] = key

	l.log.WithFields(logrus.Fields{
		"sessionID": sessionID,
		"client":    key,
		"current":   l.clientCount[key],
		"max":       max,
	}).Debug("Binding session to client")

	return true, key, l.clientCount[key], max
}

func (l *limiter) Unbind(sessionID int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key, ok := l.sessionClient[sessionID]
	if !ok {
		return
	}
	delete(l.sessionClient, sessionID)

	if c := l.clientCount[key]; c > 1 {
		l.clientCount[key] = c - 1

		l.log.WithFields(logrus.Fields{
			"sessionID": sessionID,
			"client":    key,
			"current":   l.clientCount[key],
		}).Debug("Unbinding session from client")

	} else {
		delete(l.clientCount, key)

		l.log.WithFields(logrus.Fields{
			"sessionID": sessionID,
			"client":    key,
		}).Debug("Unbinding session from client")
	}
}

// maxForKey returns the maximum allowed current connections for the client
// If the key is not found in the limits or the fallbackLimits, it will use the unknownClientLimit value,
// which is set separately.
func (l *limiter) maxForKey(key Client, useFallback bool) int {
	if useFallback {
		if limit, ok := l.fallbackLimits.PerClient[key]; ok {
			return limit
		}

		l.log.WithFields(logrus.Fields{
			"client": key,
			"limit":  l.fallbackLimits.UnknownLimit,
		}).Debug("Client key not found in fallbackLimits, using unknown client limit")

		return l.fallbackLimits.UnknownLimit
	} else {
		if limit, ok := l.limits.PerClient[key]; ok {
			return limit
		}
		l.log.WithFields(logrus.Fields{
			"client": key,
			"limit":  l.limits.UnknownLimit,
		}).Debug("Client key not found in limits, using unknown client limit")

		return l.limits.UnknownLimit
	}

}
