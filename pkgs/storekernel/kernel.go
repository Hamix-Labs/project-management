// Package storekernel holds domain-agnostic store helpers: Prometheus latency
// histogram (DeferLatency + Op* constants), ID allocation, SQL constraint
// classifiers, and sentinel-parameterized error/JSON mappers. Task-specific
// validators, audit append, and task-row loading live in their owning BCs
// (taskcore / taskcycles / taskevents).
package storekernel
