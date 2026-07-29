// Package reports re-exports the shared side-channel contract from
// pkgs/agents/sidecar for harness-internal callers. New code should prefer
// importing sidecar directly; this package remains for import stability.
package reports

import "github.com/AlexsanderHamir/Hamix/pkgs/agents/sidecar"

var (
	ErrCriteriaReportMissing = sidecar.ErrCriteriaReportMissing
	ErrCriteriaReportInvalid = sidecar.ErrCriteriaReportInvalid
	ErrVerifyReportMissing   = sidecar.ErrVerifyReportMissing
	ErrVerifyReportInvalid   = sidecar.ErrVerifyReportInvalid
	ErrSubmitReceiptMissing  = sidecar.ErrSubmitReceiptMissing
	ErrSubmitReceiptInvalid  = sidecar.ErrSubmitReceiptInvalid
)

const CurrentSchemaVersion = sidecar.CurrentSchemaVersion

type CriteriaEntry = sidecar.CriteriaEntry
type VerifyEntry = sidecar.VerifyEntry
type CriteriaCommitClaim = sidecar.CriteriaCommitClaim
type SubmitReceipt = sidecar.SubmitReceipt

var (
	ReportCycleDir               = sidecar.ReportCycleDir
	CriteriaReportPath           = sidecar.CriteriaReportPath
	VerifyReportPath             = sidecar.VerifyReportPath
	CriteriaSubmitReceiptPath    = sidecar.CriteriaSubmitReceiptPath
	VerifySubmitReceiptPath      = sidecar.VerifySubmitReceiptPath
	EnsureReportCycleDir         = sidecar.EnsureReportCycleDir
	ScrubCycleArtifacts          = sidecar.ScrubCycleArtifacts
	CleanupReportDir             = sidecar.CleanupReportDir
	ParseCriteriaReportPartial   = sidecar.ParseCriteriaReportPartial
	ParseCriteriaReport          = sidecar.ParseCriteriaReport
	ParseCriteriaReportCommits   = sidecar.ParseCriteriaReportCommits
	ParseVerifyReport            = sidecar.ParseVerifyReport
	WriteCriteriaReport          = sidecar.WriteCriteriaReport
	WriteVerifyReport            = sidecar.WriteVerifyReport
	WriteSubmitReceipt           = sidecar.WriteSubmitReceipt
	RequireCriteriaSubmitReceipt = sidecar.RequireCriteriaSubmitReceipt
	RequireVerifySubmitReceipt   = sidecar.RequireVerifySubmitReceipt
)
