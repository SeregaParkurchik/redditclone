// metrics.go
package pg

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// db_запросы_всего — счетчик, который будет отслеживать общее количество запросов к базе данных.
// Метки "метод" и "успешно" позволяют фильтровать по типу запроса и его результату.
var dbQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "db_queries_total",
	Help: "Общее количество выполненных запросов к базе данных.",
}, []string{"method", "status"})

// db_время_выполнения_запроса — гистограмма, которая будет измерять время выполнения запросов.
// Позволяет получить распределение времени, а не только среднее.
var dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "db_query_duration_seconds",
	Help:    "Время выполнения запросов к базе данных.",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "status"})
