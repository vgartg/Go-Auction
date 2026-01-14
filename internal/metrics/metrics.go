package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    BidsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auction_bids_total",
        Help: "Total number of bids placed",
    }, []string{"lot_id"})

    ActiveLots = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "auction_active_lots",
        Help: "Current number of active lots",
    })

    LotClosuresTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "auction_lot_closures_total",
        Help: "Total number of closed lots",
    })
)