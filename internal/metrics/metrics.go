package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	BidsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auction_bids_total",
		Help: "Total number of accepted bids, labelled by lot_id",
	}, []string{"lot_id"})

	ActiveLots = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "auction_active_lots",
		Help: "Current number of active lots",
	})

	LotClosuresTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_lot_closures_total",
		Help: "Total number of closed lots",
	})

	OptimisticLockRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_optimistic_lock_retries_total",
		Help: "Total number of optimistic-lock collisions retried during bidding",
	})

	AntiSnipingExtensions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_anti_sniping_extensions_total",
		Help: "Total number of times a lot's closing time was extended due to a late bid",
	})

	BidLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "auction_bid_latency_seconds",
		Help:    "End-to-end latency of an accepted bid, in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration by route and status code",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method", "status"})

	RateLimitedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auction_rate_limited_total",
		Help: "Requests rejected by the rate limiter, labelled by scope",
	}, []string{"scope"})
)
