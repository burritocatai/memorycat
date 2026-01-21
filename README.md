# memorycat 🐱

A beautiful terminal UI application for saving and organizing your frequently used commands with AI-generated descriptions.

## Features

- 💾 Save commands with auto-generated descriptions using Claude AI
- 🎨 Beautiful terminal UI built with Bubble Tea
- ⌨️ Simple keyboard navigation
- 📝 Persistent storage in JSON format
- 🤖 AI-powered command descriptions
- 🔄 Fallback to manual description entry if AI generation fails
- 📋 Pipe commands directly from clipboard or stdin
- 📋 Copy saved commands back to clipboard with one keystroke
- 🔧 Template variables support with `{{variable_name}}` syntax

## Installation

```bash
go build -o memorycat
```

## Setup

Set your Anthropic API key as an environment variable:

```bash
export ANTHROPIC_API_KEY=your_api_key_here
```

## Usage

### Interactive Mode

Run the application:

```bash
./memorycat
```

### Pipe from Clipboard (macOS)

Save a command directly from your clipboard:

```bash
pbpaste | ./memorycat
```

Or save any command:

```bash
echo "kubectl get pods -A" | ./memorycat
```

### Keyboard Shortcuts

**List View:**
- `n` - Add a new command
- `enter` or `c` - Copy selected command to clipboard
- `d` - Delete selected command
- `↑/k` - Move up
- `↓/j` - Move down
- `q` - Quit

**Input Mode:**
- Type your command
- `Enter` - Save command (will generate description with AI)
- `Esc` - Cancel

## Storage

Commands are saved to: `~/.config/memorycat/commands.json`

## Examples

### Interactive Mode
1. Press `n` to add a new command
2. Type: `docker ps -a`
3. Press Enter
4. Claude AI generates: "List all Docker containers"
5. Command is saved and displayed in the list
6. Navigate to any saved command and press `c` to copy it to clipboard

**Note:** If AI generation fails (e.g., network issues, API errors), you'll be prompted to enter a description manually instead of losing your command.

### Piping from Clipboard
```bash
$ pbpaste | ./memorycat
Generating description for: docker ps -a
Saved: Lists all Docker containers including stopped ones
```

### Template Variables
Save commands with template variables using `{{variable_name}}` syntax:

1. Press `n` to add a new command
2. Type: `curl -O {{url}}`
3. Press Enter (command is saved with the template)
4. Later, press `c` to copy the command
5. The app will prompt you to enter a value for `{{url}}`
6. Enter the URL value and press Enter
7. The final command with substituted values is copied to clipboard

Example templates:
- `curl -O {{url}}` - Download a file from a URL
- `ssh {{user}}@{{host}}` - SSH to a server
- `docker exec -it {{container}} bash` - Enter a container
- `git clone {{repo}} {{destination}}` - Clone a repository

## Requirements

- Go 1.24+
- Anthropic API key
