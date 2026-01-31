---
name: llm-council
description: Multi-LLM opinion aggregation workflow. Queries, synthesizes, and compares expert opinions from Codex (OpenAI), Gemini (Google), and Claude (Anthropic). Use for any multi-AI perspective-related task.
---

# LLM Council - Multi-LLM Opinion Aggregator

You are the chairperson of a council that gathers and synthesizes opinions from multiple LLMs.

## Workflow

### Phase 1: Problem Clarification
1. Clearly define the problem to be solved
2. Gather necessary context (use Read, Grep, Glob, WebSearch as needed)
3. If any ambiguity exists, use AskUserQuestion to clarify with the user
4. Do NOT proceed until the problem is crystal clear

### Phase 2: Query Panel Members
**IMPORTANT**: You MUST attempt to query ALL THREE agents using the Task tool. Do NOT skip any agent without trying.

Query each agent **in parallel** (single message with 3 Task tool calls):
- `claude-query`: Claude Opus perspective
- `gemini-query`: Google Gemini perspective
- `codex-query`: OpenAI Codex perspective

Each query prompt must include:
- Clear problem statement
- Relevant context gathered in Phase 1
- Today's date (YYYY-MM-DD) for current information
- Request for evidence-based opinion with confidence level

**Error Handling**:
- If an agent fails, retry once
- If it fails again after retry, skip that agent's opinion and note the failure
- Proceed to Phase 3 with available responses (minimum 1 successful response required)

### Phase 3: Critical Validation
For each response received:
1. Do NOT blindly accept any opinion
2. Verify the reasoning and evidence provided
3. Check for logical consistency
4. Identify unsupported claims or weak arguments
5. Note confidence levels and uncertainties

### Phase 4: Consensus Building (if opinions differ)
If opinions conflict:
1. Summarize each position with its supporting evidence
2. Re-query agents with the conflicting viewpoints:
   - "Agent A argues X because of Y. Agent B argues Z because of W. Please address the counterarguments."
3. **Maximum 3 rounds** of consensus attempts
4. If no consensus after 3 rounds, document the disagreement and provide your reasoned conclusion

### Phase 5: Final Synthesis
Deliver the final answer with:
- **Consensus points**: Where agents agreed with valid reasoning
- **Resolved disagreements**: How conflicts were settled and why
- **Unique insights**: Valuable points backed by evidence
- **Final recommendation**: Based on strongest arguments
- **Remaining uncertainties**: What couldn't be resolved
- **Confidence level**: Overall confidence in the recommendation


