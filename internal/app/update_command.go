// Slash-command execution: builds CommandDeps from services + env state,
// runs commands through input.NewCommandController.
//
// Note: services are passed by reference so commands read live state at
// execution time. Snapshot values (UI dimensions, InputTokens, Cwd) are
// captured here because they don't change while a command runs.
package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/genai-io/gen-code/internal/app/input"
)

func (m *model) commandDeps() input.CommandDeps {
	return input.CommandDeps{
		// UI state
		Input:        &m.userInput,
		Conversation: &m.conv.ConversationModel,
		Tool:         &m.conv.Tool,
		Width:        m.env.Width,
		Height:       m.env.Height,
		Cwd:          m.env.CWD,
		InputTokens:  m.env.InputTokens,

		// Services
		Setting: m.services.Setting,
		LLM:     m.services.LLM,
		Session: m.services.Session,
		Command: m.services.Command,
		Skill:   m.services.Skill,
		Plugin:  m.services.Plugin,
		MCP:     m.services.MCP,
		Tracker: m.services.Tracker,
		Cron:    m.services.Cron,
		ToolSvc: m.services.Tool,

		// Env callbacks
		GetThinkingEffort: func() string { return m.env.EffectiveThinkingEffort() },
		SetThinkingEffort: func(effort string) { m.env.ThinkingEffort = effort },
		ResetTokens:       m.env.ResetTokens,

		// Model actions
		CommitMessages:          m.CommitMessages,
		SubmitToAgent:           m.SubmitToAgent,
		HandleSkillInvocation:   m.HandleSkillInvocation,
		StartExternalEditor:     m.StartExternalEditor,
		ReloadPluginBackedState: m.ReloadPluginBackedState,
		PersistSession:          m.PersistSession,
		InitTaskStorage:         m.InitTaskStorage,
		ReconfigureAgentTool:    m.ReconfigureAgentTool,
		StopAgentSession:        m.StopAgentSession,
		FireSessionEnd:          m.FireSessionEnd,
		BuildCompactRequest:     m.BuildCompactRequest,
		SpinnerTickCmd:          m.SpinnerTickCmd,
		ResetCronQueue:          m.ResetCronQueue,
		ForkSession:             m.forkSession,
	}
}

func (m *model) executeCommand(ctx context.Context, inputText string) (string, tea.Cmd, bool) {
	return input.NewCommandController(m.commandDeps()).Execute(ctx, inputText)
}
