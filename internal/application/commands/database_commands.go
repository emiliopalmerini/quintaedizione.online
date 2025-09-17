package commands

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
)

// CreateClasseCommand handles creation of Classe entities
type CreateClasseCommand struct {
	classe   *domain.Classe
	registry repositories.RepositoryRegistry
}

// NewCreateClasseCommand creates a new command to create a classe
func NewCreateClasseCommand(classe *domain.Classe, registry repositories.RepositoryRegistry) *CreateClasseCommand {
	return &CreateClasseCommand{
		classe:   classe,
		registry: registry,
	}
}

// Validate ensures the command is valid
func (c *CreateClasseCommand) Validate() error {
	if c.classe == nil {
		return fmt.Errorf("classe cannot be nil")
	}

	if c.classe.Nome == "" {
		return fmt.Errorf("classe nome is required")
	}

	if c.classe.Slug == "" {
		return fmt.Errorf("classe slug is required")
	}

	return nil
}

// Execute runs the command
func (c *CreateClasseCommand) Execute(ctx context.Context) error {
	repo, err := c.registry.GetRepository("classe")
	if err != nil {
		return fmt.Errorf("failed to get classe repository: %w", err)
	}

	classeRepo, ok := repo.(domain.ClasseRepository)
	if !ok {
		return fmt.Errorf("repository is not a ClasseRepository")
	}

	return classeRepo.Save(ctx, c.classe)
}

// UpdateClasseCommand handles updating Classe entities
type UpdateClasseCommand struct {
	slug     string
	classe   *domain.Classe
	registry repositories.RepositoryRegistry
}

// NewUpdateClasseCommand creates a new command to update a classe
func NewUpdateClasseCommand(slug string, classe *domain.Classe, registry repositories.RepositoryRegistry) *UpdateClasseCommand {
	return &UpdateClasseCommand{
		slug:     slug,
		classe:   classe,
		registry: registry,
	}
}

// Validate ensures the command is valid
func (c *UpdateClasseCommand) Validate() error {
	if c.slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}

	if c.classe == nil {
		return fmt.Errorf("classe cannot be nil")
	}

	return nil
}

// Execute runs the command
func (c *UpdateClasseCommand) Execute(ctx context.Context) error {
	repo, err := c.registry.GetRepository("classe")
	if err != nil {
		return fmt.Errorf("failed to get classe repository: %w", err)
	}

	classeRepo, ok := repo.(domain.ClasseRepository)
	if !ok {
		return fmt.Errorf("repository is not a ClasseRepository")
	}

	return classeRepo.Update(ctx, c.slug, c.classe)
}

// DeleteClasseCommand handles deletion of Classe entities
type DeleteClasseCommand struct {
	slug     string
	registry repositories.RepositoryRegistry
}

// NewDeleteClasseCommand creates a new command to delete a classe
func NewDeleteClasseCommand(slug string, registry repositories.RepositoryRegistry) *DeleteClasseCommand {
	return &DeleteClasseCommand{
		slug:     slug,
		registry: registry,
	}
}

// Validate ensures the command is valid
func (c *DeleteClasseCommand) Validate() error {
	if c.slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	return nil
}

// Execute runs the command
func (c *DeleteClasseCommand) Execute(ctx context.Context) error {
	repo, err := c.registry.GetRepository("classe")
	if err != nil {
		return fmt.Errorf("failed to get classe repository: %w", err)
	}

	classeRepo, ok := repo.(domain.ClasseRepository)
	if !ok {
		return fmt.Errorf("repository is not a ClasseRepository")
	}

	return classeRepo.Delete(ctx, c.slug)
}