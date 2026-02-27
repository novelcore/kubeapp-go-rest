# Kubecore DevOps Agent System

This directory contains the AI agent system for your Python REST API repository. The agents are designed to help developers throughout the complete development lifecycle.

## 🤖 Available Agents

### 1. **Welcome Agent** (`welcome-agent.mdc`)
- **Role**: Entry point and router
- **Triggers**: "welcome", "hello", "start", "help", "what can I do"
- **Purpose**: Greets developers and routes them to specialized agents

### 2. **New Endpoint Agent** (`new-endpoint-agent.mdc`)
- **Role**: API Builder
- **Triggers**: "new endpoint", "create endpoint", "add feature", "new API"
- **Purpose**: Creates FastAPI endpoints with PRD integration and GitFlow workflow

### 3. **Testing Agent** (`testing-agent.mdc`)
- **Role**: Quality Assurance Specialist
- **Triggers**: "test", "unit test", "testing", "validate", "coverage"
- **Purpose**: Generates unit tests and ensures code quality

### 4. **Release Management Agent** (`release-management-agent.mdc`) + **Release Manager** (Cursor & Claude)
- **Role**: Release Coordinator
- **Triggers**: "release", "deploy", "hotfix", "production", "staging"
- **Purpose**: Manages GitFlow releases and deployment monitoring
- **Agents**: Cursor rule in `.cursor/rules/`; specialized **release-manager** agent in `.cursor/agents/release-manager.md` (Cursor) and `.claude/agents/release-manager.md` (Claude Code) for GitOps promotion and overlay paths.

### 5. **PRD Agent** (`prd-agent.mdc`)
- **Role**: Product Requirements Specialist
- **Triggers**: "PRD", "requirements", "specification", "plan", "documentation"
- **Purpose**: Creates and manages Product Requirements Documents with TaskMaster MCP

## 🚀 How It Works

### Automatic Activation
Each agent has specific trigger words that will automatically activate it when mentioned in conversation:

```
Developer: "I want to create a new endpoint"
↓ Automatically routes to New Endpoint Agent

Developer: "I need unit tests"
↓ Automatically routes to Testing Agent

Developer: "I'm ready to release"
↓ Automatically routes to Release Management Agent
```

### Agent Configuration
All agents are configured with:
- **`alwaysApply: true`** - Ensures they're always available
- **Specific triggers** - Keywords that activate each agent
- **Global globs** - Apply to all files in the repository
- **Proper descriptions** - Clear role definitions

### MCP Integration
The agents integrate with:
- **TaskMaster MCP** - For PRD creation and task tracking
- **GitHub MCP** - For repository management and PR creation
- **Local Git** - For branch management and workflow

## 🎯 Workflow Example

```
1. Developer opens repository
   ↓ crusor-init.sh runs automatically
   ↓ Welcome message appears with options

2. Developer: "I want to create a user management endpoint"
   ↓ New Endpoint Agent activates
   ↓ Creates PRD, sets up branch, generates code

3. Developer: "Test this implementation"
   ↓ Testing Agent activates
   ↓ Generates comprehensive unit tests

4. Developer: "I'm ready to release"
   ↓ Release Management Agent activates
   ↓ Manages GitFlow release process
```

## 📋 Agent Capabilities

### New Endpoint Agent
- ✅ PRD creation with TaskMaster MCP
- ✅ GitFlow branch management
- ✅ FastAPI router and model generation
- ✅ Comprehensive testing setup
- ✅ GitHub PR creation

### Testing Agent
- ✅ Unit test generation
- ✅ Coverage analysis and reporting
- ✅ Quality metrics and validation
- ✅ Testing best practices guidance

### Release Management Agent
- ✅ GitFlow release workflow
- ✅ Release candidate creation
- ✅ Hotfix emergency procedures
- ✅ Deployment monitoring and validation
- ✅ Version management and tagging

### PRD Agent
- ✅ Comprehensive PRD creation
- ✅ TaskMaster MCP integration
- ✅ Implementation progress tracking
- ✅ Acceptance criteria validation
- ✅ Requirements lifecycle management

## 🔧 Technical Details

### File Structure
```
.cursor/rules/
├── welcome-agent.mdc          # Entry point and router
├── new-endpoint-agent.mdc     # API development specialist
├── testing-agent.mdc          # Quality assurance specialist
├── release-management-agent.mdc # Release workflow manager
├── prd-agent.mdc             # Requirements specialist
└── README.md                 # This documentation
```

### Initialization
The `.github/ai-automations/crusor-init.sh` script automatically:
1. Checks for the welcome agent file
2. Loads the welcome message
3. Opens Cursor with the agent system activated
4. Provides clear guidance on available capabilities

### Agent Communication
Agents can:
- Route between each other seamlessly
- Share context and state information
- Collaborate on complex workflows
- Provide consistent user experience

## 🎉 Benefits

### For Developers
- **Conversational interface** - Natural language interactions
- **Complete workflows** - From concept to deployment
- **Quality assurance** - Built-in testing and validation
- **GitFlow compliance** - Proper branching and release management

### For Teams
- **Consistent processes** - Standardized development workflows
- **Documentation** - Automatic PRD and requirement tracking
- **Quality gates** - Mandatory testing and validation
- **Traceability** - Complete feature lifecycle tracking

### For Projects
- **Reduced complexity** - Hide technical details from developers
- **Faster delivery** - Automated workflow management
- **Higher quality** - Comprehensive testing and validation
- **Better documentation** - Automatic PRD creation and tracking

## 🔄 Continuous Improvement

The agent system is designed to evolve with your needs:
- Add new agents for specific domains
- Extend existing agents with new capabilities
- Integrate with additional MCP servers
- Customize workflows for your team's needs

---

**Ready to start?** Just mention what you'd like to work on, and the appropriate agent will guide you through the process! 🚀 