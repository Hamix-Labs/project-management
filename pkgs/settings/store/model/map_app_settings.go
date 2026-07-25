package model

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/jsonmap"
)

// FromDomainAppSettings copies a domain row to its persistence model.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FromDomainAppSettings(d domain.AppSettings) AppSettings {
	return AppSettings{
		ID:                          d.ID,
		AgentPaused:                 d.AgentPaused,
		Runner:                      d.Runner,
		CursorBin:                   d.CursorBin,
		CursorModel:                 d.CursorModel,
		VerifyModel:                 d.VerifyModel,
		MaxRunDurationSeconds:       d.MaxRunDurationSeconds,
		StreamIdleStuckSeconds:      d.StreamIdleStuckSeconds,
		AgentPickupDelaySeconds:     d.AgentPickupDelaySeconds,
		DisplayTimezone:             d.DisplayTimezone,
		OptimisticMutationsEnabled:  d.OptimisticMutationsEnabled,
		SSEReplayEnabled:            d.SSEReplayEnabled,
		RunnerConfigs:               jsonmap.DatatypesFromRaw(d.RunnerConfigs),
		VerifyMaxRetries:            d.VerifyMaxRetries,
		VerifyCommandTimeoutSeconds: d.VerifyCommandTimeoutSeconds,
		CursorSessionResumeEnabled:  d.CursorSessionResumeEnabled,
		UpdatedAt:                   d.UpdatedAt,
	}
}

// ToDomainAppSettings copies a persistence row to domain.AppSettings.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ToDomainAppSettings(m AppSettings) domain.AppSettings {
	return domain.AppSettings{
		ID:                          m.ID,
		AgentPaused:                 m.AgentPaused,
		Runner:                      m.Runner,
		CursorBin:                   m.CursorBin,
		CursorModel:                 m.CursorModel,
		VerifyModel:                 m.VerifyModel,
		MaxRunDurationSeconds:       m.MaxRunDurationSeconds,
		StreamIdleStuckSeconds:      m.StreamIdleStuckSeconds,
		AgentPickupDelaySeconds:     m.AgentPickupDelaySeconds,
		DisplayTimezone:             m.DisplayTimezone,
		OptimisticMutationsEnabled:  m.OptimisticMutationsEnabled,
		SSEReplayEnabled:            m.SSEReplayEnabled,
		RunnerConfigs:               jsonmap.RawFromDatatypes(m.RunnerConfigs),
		VerifyMaxRetries:            m.VerifyMaxRetries,
		VerifyCommandTimeoutSeconds: m.VerifyCommandTimeoutSeconds,
		CursorSessionResumeEnabled:  m.CursorSessionResumeEnabled,
		UpdatedAt:                   m.UpdatedAt,
	}
}
