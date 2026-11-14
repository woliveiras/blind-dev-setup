package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/target"
)

type wizardTarget struct {
	candidate target.Candidate
	details   *target.Details
}

func (w *wizard) discover() ([]wizardTarget, error) {
	if w.explicitTarget != "" {
		if w.deps.Inspect == nil {
			return nil, errors.New("inspeção de destino indisponível")
		}
		details, err := w.deps.Inspect(w.explicitTarget)
		if err != nil {
			return nil, err
		}
		return []wizardTarget{{details: &details}}, nil
	}
	if w.deps.ListTargets == nil {
		return nil, errors.New("procura de pendrives indisponível")
	}
	candidates, err := w.deps.ListTargets()
	if err != nil {
		return nil, err
	}
	choices := make([]wizardTarget, len(candidates))
	for i, candidate := range candidates {
		choices[i].candidate = candidate
	}
	return choices, nil
}

func (choice wizardTarget) path() string {
	if choice.details != nil {
		return choice.details.Path
	}
	return choice.candidate.RootPath()
}

func (choice wizardTarget) exists() bool {
	if choice.details != nil {
		return choice.details.DestinationExists
	}
	return choice.candidate.DestinationExists
}

func (w *wizard) issues(choice wizardTarget) []string {
	if choice.details != nil {
		details := *choice.details
		if choice.exists() {
			details.DestinationExists = false
			details.FreeBytes = uint64(w.deps.Manifest.MinimumFreeBytes)
		}
		if err := target.Validate(details, w.deps.Manifest.MinimumFreeBytes); err != nil {
			return []string{err.Error()}
		}
		return nil
	}
	c := choice.candidate
	minimum := w.deps.Manifest.MinimumFreeBytes
	if c.DestinationExists {
		minimum = 0
	}
	return c.PreparationIssues(minimum)
}

func (w *wizard) revalidate(chosen wizardTarget) error {
	fresh, err := w.discover()
	if err != nil {
		return err
	}
	for _, candidate := range fresh {
		if !strings.EqualFold(candidate.path(), chosen.path()) {
			continue
		}
		if chosen.details != nil {
			if candidate.details == nil || !strings.EqualFold(candidate.details.Volume, chosen.details.Volume) || !strings.EqualFold(candidate.details.SystemVolume, chosen.details.SystemVolume) || !strings.EqualFold(candidate.details.FileSystem, chosen.details.FileSystem) {
				return errors.New("a identificação do volume mudou")
			}
		} else {
			a, b := chosen.candidate, candidate.candidate
			if a.DiskNumber != b.DiskNumber || a.Model != b.Model || a.SizeBytes != b.SizeBytes || a.Label != b.Label || !strings.EqualFold(a.FileSystem, b.FileSystem) || a.IsSystem != b.IsSystem {
				return errors.New("o dispositivo ou volume associado à letra mudou")
			}
		}
		if candidate.exists() != chosen.exists() {
			return errors.New("a presença de uma instalação mudou")
		}
		if issues := w.issues(candidate); len(issues) > 0 {
			return errors.New(strings.Join(issues, "; "))
		}
		return nil
	}
	return errors.New("o destino selecionado não foi encontrado")
}

func (choice wizardTarget) label() string {
	if choice.details != nil {
		return fmt.Sprintf("%s · %s · %.1f GiB livres", choice.path(), choice.details.FileSystem, bytesInGiB(choice.details.FreeBytes))
	}
	c := choice.candidate
	return fmt.Sprintf("%s · %s · %.1f GiB · %s", valueOr(c.RootPath(), "sem letra"), valueOr(c.Model, "USB"), bytesInGiB(c.SizeBytes), valueOr(c.Label, "sem nome"))
}

func (choice wizardTarget) summary() string {
	if choice.details != nil {
		return choice.label()
	}
	c := choice.candidate
	return fmt.Sprintf("%s\n%s · %.1f GiB livres", choice.label(), valueOr(c.FileSystem, "sistema de arquivos não informado"), bytesInGiB(c.FreeBytes))
}
