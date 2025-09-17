package commands

import "context"

// Command represents a command in the command pattern
type Command interface {
	Execute(ctx context.Context) error
	Validate() error
}

// CommandHandler handles command execution with validation
type CommandHandler struct{}

// NewCommandHandler creates a new command handler
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

// Execute runs a command with validation
func (h *CommandHandler) Execute(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	return cmd.Execute(ctx)
}

// BatchExecute runs multiple commands in sequence
func (h *CommandHandler) BatchExecute(ctx context.Context, commands ...Command) error {
	for _, cmd := range commands {
		if err := h.Execute(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}