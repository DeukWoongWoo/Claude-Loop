# LLM Council

Multi-LLM opinion aggregator that queries Codex (OpenAI), Gemini (Google), and Claude (Anthropic) for expert opinions, then synthesizes responses.

## Usage

Invoke this skill when you need diverse AI perspectives on a problem.

**Trigger phrases:**
- "Get opinions from multiple LLMs"
- "What do different AIs think about X?"
- "Compare AI perspectives on this"
- "I need a council decision"

## Description

This skill acts as a chairperson of a council, gathering and synthesizing opinions from multiple LLMs. It queries three different AI models in parallel and combines their insights into a unified recommendation.

**Key capabilities:**
- Queries Claude, Gemini, and Codex in parallel
- Validates reasoning and evidence from each response
- Builds consensus when opinions differ (up to 3 rounds)
- Synthesizes final recommendation with confidence levels

## Workflow

When invoked, the skill:

1. **Clarify**: Define the problem clearly, gather context, resolve ambiguities
2. **Query**: Call all three LLM agents in parallel (claude-query, gemini-query, codex-query)
3. **Validate**: Verify reasoning, check logical consistency, identify weak arguments
4. **Build Consensus**: If opinions conflict, re-query with counterarguments (max 3 rounds)
5. **Synthesize**: Deliver final answer with consensus points and recommendation

**Output**: A synthesis report containing:
- Consensus points with valid reasoning
- Resolved disagreements with explanations
- Unique insights backed by evidence
- Final recommendation with confidence level

## Prerequisites

**Required:**
- Access to Task tool for querying external LLM agents
- Gemini CLI installed
- Codex CLI installed

### CLI Installation

#### Gemini CLI

Install Google's official Gemini CLI:

**npm (requires Node.js 20+):**
```bash
npm install -g @google/gemini-cli
```

**Homebrew:**
```bash
brew install gemini-cli
```

For more details, see the [official GitHub repository](https://github.com/google-gemini/gemini-cli).

#### Codex CLI

Install OpenAI's official Codex CLI:

**npm:**
```bash
npm install -g @openai/codex
```

**Homebrew:**
```bash
brew install --cask codex
```

For more details, see the [official documentation](https://developers.openai.com/codex/cli).

#### Codex CLI Configuration

To enable web search in Codex, add the following to `~/.codex/config.toml`:

```toml
[features]
web_search_request = true

[shell_environment_policy]
network_access = "enabled"
approval_policy = "auto"
```

### Authentication

There are two ways to authenticate with each CLI:

#### Option 1: Subscription-based Authentication (Recommended)

Run each CLI command once to authenticate via your subscription. This stores credentials locally and does not require API keys.

```bash
# Gemini CLI - opens browser for Google account login
gemini

# Codex CLI - opens browser for OpenAI account login
codex
```

After authenticating once, the CLIs will use your subscription automatically.

#### Option 2: API Key Environment Variables

If you prefer API keys over subscription auth, set these environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `OPENAI_API_KEY` | OpenAI API key for Codex queries | No (if subscription auth used) |
| `GOOGLE_API_KEY` | Google API key for Gemini queries | No (if subscription auth used) |
