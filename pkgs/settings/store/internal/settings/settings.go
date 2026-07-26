package settings

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/settings/store/model"
	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Patch is the partial-update payload for app_settings. Pointer-typed
// fields distinguish "not provided" (nil) from "set to zero value"
// (e.g. *int = 0 means MaxRunDurationSeconds explicitly set to 0 == "no
// limit"; nil means "leave unchanged").
type Patch = contract.SettingsPatch

// Get returns the singleton app_settings row, creating it with
// domain.DefaultAppSettings on first read so callers always observe a
// populated value. The create-on-read is guarded by the unique CHECK
// constraint on id=1: a parallel Get from another goroutine that wins
// the insert race will simply re-read the row the loser created.
func Get(ctx context.Context, db *gorm.DB) (domain.AppSettings, error) {
	defer storekernel.DeferLatency(storekernel.OpGetAppSettings)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.settings.Get")
	if db == nil {
		return domain.AppSettings{}, errors.New("settings store: nil database")
	}
	var row model.AppSettings
	err := db.WithContext(ctx).First(&row, "id = ?", domain.AppSettingsRowID).Error
	if err == nil {
		if !row.OptimisticMutationsEnabled || !row.SSEReplayEnabled {
			t := true
			return Update(ctx, db, Patch{
				OptimisticMutationsEnabled: &t,
				SSEReplayEnabled:           &t,
			})
		}
		d := model.ToDomainAppSettings(row)
		if strings.TrimSpace(d.VerifyChatMode) == "" {
			mode := string(domain.DefaultVerifyChatMode)
			return Update(ctx, db, Patch{VerifyChatMode: &mode})
		}
		return d, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.AppSettings{}, fmt.Errorf("get app settings: %w", err)
	}
	seed := model.FromDomainAppSettings(domain.DefaultAppSettings())
	seed.UpdatedAt = time.Now().UTC()
	if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
		return domain.AppSettings{}, fmt.Errorf("seed app settings: %w", err)
	}
	if err := db.WithContext(ctx).First(&row, "id = ?", domain.AppSettingsRowID).Error; err != nil {
		return domain.AppSettings{}, fmt.Errorf("get app settings (post-seed): %w", err)
	}
	slog.Info("app settings seeded with defaults",
		"cmd", calltrace.LogCmd, "operation", "settings.store.settings.seeded",
		"agent_paused", row.AgentPaused,
		"runner", row.Runner,
		"cursor_bin", row.CursorBin,
		"max_run_duration_seconds", row.MaxRunDurationSeconds,
		"display_timezone", row.DisplayTimezone)
	return model.ToDomainAppSettings(row), nil
}

// Update applies a partial Patch to the singleton row inside a
// transaction. If the row doesn't exist yet it is created from
// domain.DefaultAppSettings before the patch is overlaid, so the first
// PATCH against a fresh DB is well-defined.
func Update(ctx context.Context, db *gorm.DB, patch Patch) (domain.AppSettings, error) {
	defer storekernel.DeferLatency(storekernel.OpUpdateAppSettings)()
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.settings.Update")
	if db == nil {
		return domain.AppSettings{}, errors.New("settings store: nil database")
	}
	if err := validatePatch(patch); err != nil {
		return domain.AppSettings{}, err
	}
	var out domain.AppSettings
	txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.AppSettings
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", domain.AppSettingsRowID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.FromDomainAppSettings(domain.DefaultAppSettings())
			row.UpdatedAt = time.Now().UTC()
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return fmt.Errorf("seed app settings during update: %w", err)
			}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", domain.AppSettingsRowID).Error; err != nil {
				return fmt.Errorf("re-read app settings after seed: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("lock app settings: %w", err)
		}
		drow := model.ToDomainAppSettings(row)
		applyPatch(&drow, patch)
		row = model.FromDomainAppSettings(drow)
		row.UpdatedAt = time.Now().UTC()
		if err := tx.Save(&row).Error; err != nil {
			return fmt.Errorf("save app settings: %w", err)
		}
		out = model.ToDomainAppSettings(row)
		return nil
	})
	if txErr != nil {
		return domain.AppSettings{}, txErr
	}
	return out, nil
}

func validatePatch(patch Patch) error {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.settings.validatePatch")
	if patch.Runner != nil {
		trimmed := strings.TrimSpace(*patch.Runner)
		if trimmed == "" {
			return fmt.Errorf("%w: runner must be non-empty", domain.ErrInvalidInput)
		}
	}
	if patch.MaxRunDurationSeconds != nil && *patch.MaxRunDurationSeconds < 0 {
		return fmt.Errorf("%w: max_run_duration_seconds must be >= 0", domain.ErrInvalidInput)
	}
	if patch.AgentPickupDelaySeconds != nil {
		v := *patch.AgentPickupDelaySeconds
		if v < 0 || v > 604800 {
			return fmt.Errorf("%w: agent_pickup_delay_seconds must be between 0 and 604800", domain.ErrInvalidInput)
		}
	}
	if patch.CursorModel != nil && len(strings.TrimSpace(*patch.CursorModel)) > 256 {
		return fmt.Errorf("%w: cursor_model too long (max 256)", domain.ErrInvalidInput)
	}
	if patch.VerifyModel != nil && len(strings.TrimSpace(*patch.VerifyModel)) > 256 {
		return fmt.Errorf("%w: verify_model too long (max 256)", domain.ErrInvalidInput)
	}
	if patch.VerifyChatMode != nil {
		normalized, ok := domain.NormalizeVerifyChatMode(*patch.VerifyChatMode)
		if !ok || normalized == "" {
			return fmt.Errorf("%w: verify_chat_mode must be same_chat or different_chat", domain.ErrInvalidInput)
		}
	}
	if patch.DisplayTimezone != nil {
		trimmed := strings.TrimSpace(*patch.DisplayTimezone)
		if trimmed != "" {
			if _, err := time.LoadLocation(trimmed); err != nil {
				return fmt.Errorf("%w: display_timezone %q is not a valid IANA timezone: %v", domain.ErrInvalidInput, trimmed, err)
			}
		}
	}
	if patch.VerifyMaxRetries != nil {
		v := *patch.VerifyMaxRetries
		if v < 0 {
			return fmt.Errorf("%w: verify_max_retries must be >= 0", domain.ErrInvalidInput)
		}
	}
	return nil
}

func applyPatch(row *domain.AppSettings, patch Patch) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.settings.applyPatch")
	if patch.AgentPaused != nil {
		row.AgentPaused = *patch.AgentPaused
	}
	if patch.Runner != nil {
		row.Runner = strings.TrimSpace(*patch.Runner)
	}
	if patch.CursorBin != nil {
		row.CursorBin = strings.TrimSpace(*patch.CursorBin)
	}
	if patch.CursorModel != nil {
		row.CursorModel = strings.TrimSpace(*patch.CursorModel)
	}
	if patch.VerifyModel != nil {
		row.VerifyModel = strings.TrimSpace(*patch.VerifyModel)
	}
	if patch.VerifyChatMode != nil {
		normalized, _ := domain.NormalizeVerifyChatMode(*patch.VerifyChatMode)
		row.VerifyChatMode = normalized
	}
	if patch.MaxRunDurationSeconds != nil {
		row.MaxRunDurationSeconds = *patch.MaxRunDurationSeconds
	}
	if patch.AgentPickupDelaySeconds != nil {
		row.AgentPickupDelaySeconds = *patch.AgentPickupDelaySeconds
	}
	if patch.DisplayTimezone != nil {
		row.DisplayTimezone = strings.TrimSpace(*patch.DisplayTimezone)
	}
	if patch.OptimisticMutationsEnabled != nil {
		row.OptimisticMutationsEnabled = *patch.OptimisticMutationsEnabled
	}
	if patch.SSEReplayEnabled != nil {
		row.SSEReplayEnabled = *patch.SSEReplayEnabled
	}
	if patch.RunnerConfigs != nil {
		row.RunnerConfigs = json.RawMessage(*patch.RunnerConfigs)
	}
	if patch.VerifyMaxRetries != nil {
		row.VerifyMaxRetries = *patch.VerifyMaxRetries
	}
	if patch.CursorSessionResumeEnabled != nil {
		row.CursorSessionResumeEnabled = *patch.CursorSessionResumeEnabled
	}

	dualWriteCursorToRunnerConfigs(row)
}

func dualWriteCursorToRunnerConfigs(row *domain.AppSettings) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "settings.store.settings.dualWriteCursorToRunnerConfigs")
	var configs map[string]json.RawMessage
	if len(row.RunnerConfigs) > 0 {
		_ = json.Unmarshal([]byte(row.RunnerConfigs), &configs)
	}
	if configs == nil {
		configs = make(map[string]json.RawMessage)
	}
	cursorCfg := map[string]string{
		"binary_path":   row.CursorBin,
		"default_model": row.CursorModel,
	}
	raw, err := json.Marshal(cursorCfg)
	if err != nil {
		slog.Warn("dual-write cursor config marshal failed",
			"cmd", calltrace.LogCmd, "operation", "settings.store.settings.dualWriteCursorToRunnerConfigs",
			"err", err)
		return
	}
	configs["cursor"] = raw
	merged, err := json.Marshal(configs)
	if err != nil {
		slog.Warn("dual-write runner configs marshal failed",
			"cmd", calltrace.LogCmd, "operation", "settings.store.settings.dualWriteCursorToRunnerConfigs",
			"err", err)
		return
	}
	row.RunnerConfigs = json.RawMessage(merged)
}
