# Commit Organizer

Analyze git changes and organize into logical commits with MR description.

## Usage

Invoke this skill when you have uncommitted changes that need organization.

**Trigger phrases:**
- "Organize my commits"
- "Review my changes"
- "Create an MR"
- "Help me commit these changes"

## Description

This skill analyzes your git working directory, groups related changes into logical commits, reviews code quality, and generates both commit commands and a merge request description.

**Key capabilities:**
- Analyzes git status and diffs
- Groups changes into logical, atomic commits
- Reviews code quality against project patterns
- Generates conventional commit messages
- Creates structured MR descriptions with test plans

## Workflow

When invoked, the skill:

1. **Gather**: Run git status and diff to understand all changes
2. **Read**: Review all changed and new files completely
3. **Group**: Organize changes into logical commits (dependencies, infrastructure, features, docs)
4. **Review**: Check code quality against project patterns
5. **Generate**: Output commit commands and MR description

**Output**:
- Ready-to-run commit commands with proper message format
- MR description with summary, categorized changes, and test plan

## Prerequisites

**Required:**
- Git repository initialized
- Uncommitted changes to organize
