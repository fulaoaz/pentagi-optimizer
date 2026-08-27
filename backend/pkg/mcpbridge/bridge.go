package mcpbridge

import (
	"context"
	"fmt"
	"net/http"

	"pentagi/pkg/controller"
	"pentagi/pkg/database"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// FlowBridge exposes PentAGI flow management tools via the MCP protocol.
type FlowBridge struct {
	srv     *server.MCPServer
	name    string
	version string
	fc      controller.FlowController
	db      database.Querier
}

func NewFlowBridge(name, version string, fc controller.FlowController, db database.Querier) *FlowBridge {
	fb := &FlowBridge{name: name, version: version, fc: fc, db: db}
	fb.srv = server.NewMCPServer(name, version,
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	fb.srv.AddTool(mcp.NewTool("get_flow_status",
		mcp.WithDescription("List all flows and their current status. Returns flow ID, title, status, and task count for each flow."),
		mcp.WithString("detail",
			mcp.Required(),
			mcp.Description("detail level: summary lists all flows; tasks shows tasks for a specific flow_id"),
			mcp.Enum("summary", "tasks"),
		),
		mcp.WithNumber("flow_id",
			mcp.Description("flow ID (required when detail=tasks)"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]any)
		detail, _ := args["detail"].(string)

		if detail == "summary" {
			flows := fb.fc.ListFlows(ctx)
			result := ""
			for _, f := range flows {
				status, _ := f.GetStatus(ctx)
				tasks := f.ListTasks(ctx)
				result += fmt.Sprintf("Flow %d: %s", f.GetFlowID(), f.GetTitle())
				result += fmt.Sprintf("\n  Status: %s", status)
				result += fmt.Sprintf("\n  Tasks: %d\n", len(tasks))
			}
			if result == "" {
				result = "No flows found."
			}
			return mcp.NewToolResultText(result), nil
		}

		flowID, _ := args["flow_id"].(float64)
		f, err := fb.fc.GetFlow(ctx, int64(flowID))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error: flow %d not found", int64(flowID))), nil
		}

		status, _ := f.GetStatus(ctx)
		tasks := f.ListTasks(ctx)
		result := fmt.Sprintf("Flow %d: %s", f.GetFlowID(), f.GetTitle())
		result += fmt.Sprintf("\nStatus: %s", status)
		result += fmt.Sprintf("\nTasks: %d\n", len(tasks))
		for _, t := range tasks {
			ts, _ := t.GetStatus(ctx)
			result += fmt.Sprintf("  Task %d: %s [%s]", t.GetTaskID(), t.GetTitle(), ts)
			if t.IsCompleted() {
				r, _ := t.GetResult(ctx)
				if len(r) > 200 {
					r = r[:200] + "..."
				}
				if r != "" {
					result += fmt.Sprintf("\n    Result: %s", r)
				}
			}
			result += "\n"
		}
		return mcp.NewToolResultText(result), nil
	})

	fb.srv.AddTool(mcp.NewTool("submit_flow_input",
		mcp.WithDescription("Submit text input to a flow as a new task goal"),
		mcp.WithNumber("flow_id", mcp.Required(), mcp.Description("ID of the flow")),
		mcp.WithString("input", mcp.Required(), mcp.Description("the task description or goal")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]any)
		flowID, _ := args["flow_id"].(float64)
		inputText, _ := args["input"].(string)

		f, err := fb.fc.GetFlow(ctx, int64(flowID))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error: flow %d not found", int64(flowID))), nil
		}

		err = f.PutInput(ctx, inputText, nil, nil)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error submitting input: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Input submitted to flow %d successfully", int64(flowID))), nil
	})

	fb.srv.AddTool(mcp.NewTool("stop_flow",
		mcp.WithDescription("Stop a running flow"),
		mcp.WithNumber("flow_id", mcp.Required(), mcp.Description("ID of the flow to stop")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]any)
		flowID, _ := args["flow_id"].(float64)

		err := fb.fc.StopFlow(ctx, int64(flowID))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error stopping flow: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Flow %d stopped successfully", int64(flowID))), nil
	})

	fb.srv.AddTool(mcp.NewTool("list_assistants",
		mcp.WithDescription("List all assistants in a flow"),
		mcp.WithNumber("flow_id", mcp.Required(), mcp.Description("ID of the flow")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]any)
		flowID, _ := args["flow_id"].(float64)

		f, err := fb.fc.GetFlow(ctx, int64(flowID))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error: flow %d not found", int64(flowID))), nil
		}

		assistants := f.ListAssistants(ctx)
		result := fmt.Sprintf("Assistants in flow %d:", int64(flowID))
		if len(assistants) == 0 {
			result += "\n  No assistants."
		}
		for _, a := range assistants {
			result += fmt.Sprintf("\n  Assistant %d", a.GetAssistantID())
		}
		return mcp.NewToolResultText(result), nil
	})

	return fb
}

func (fb *FlowBridge) GetHandler(serverURL string) http.Handler {
	return server.NewSSEServer(fb.srv, server.WithBaseURL(serverURL))
}

func (fb *FlowBridge) GetServer() *server.MCPServer { return fb.srv }

func logBridgeAction(log *logrus.Entry, action string, details map[string]any) {
	if log != nil {
		log.WithField("action", action).WithField("details", details).Debug("mcp bridge action")
	}
}
