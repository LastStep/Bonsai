# CLI Conventions

> [!important]
> These rules apply to every CLI tool. Language-agnostic — enforce regardless of framework.

---

## Command Structure

- Pattern: `<program> <command> [subcommand] [flags] [arguments]`
- Verbs for commands, nouns for resources: `user create`, not `create-user`
- Max two nesting levels: `app db migrate` — never deeper
- Commands are kebab-case: `list-users`, not `listUsers` or `list_users`
- Sort subcommands alphabetically in help output

## Standard Flags

Every CLI must support these global flags:

- `-h`, `--help` — show help for any command or subcommand
- `-v`, `--verbose` — increase output detail (stackable: `-vvv`)
- `-V`, `--version` — print version and exit
- `-q`, `--quiet` — suppress non-essential output
- `-o`, `--output <format>` — output format (`text`, `json`, `table`)
- `-f`, `--force` — skip confirmation prompts
- `--no-color` — disable colored output

Flag conventions:
- Boolean flags: `--flag` enables, `--no-flag` disables
- Short flags: single character, one hyphen (`-v`). Reserve for the most common flags only.
- Long flags: descriptive, double hyphen (`--verbose`)
- Flags before arguments: `cmd --flag arg`, not `cmd arg --flag`

## Output

- **Stdout** for primary data output — this is what gets piped
- **Stderr** for progress, status messages, errors, prompts — never pollute stdout
- Default output is human-readable; `--output json` for machine consumption
- JSON output: always an object at the top level, consistent field naming, never bare arrays
- Tables: align columns, truncate long values with `...`, show column headers
- Respect `NO_COLOR` environment variable — disable ANSI codes when set
- Respect terminal width — wrap or truncate output for the current terminal

## Exit Codes

- `0` — success
- `1` — general error (application failure)
- `2` — usage error (invalid arguments, missing required flags)
- `126` — command found but not executable
- `127` — command not found
- `130` — interrupted (Ctrl+C / SIGINT)
- Never exit `0` on failure — even partial success should exit non-zero if the overall operation failed

## Errors

- Print errors to stderr, never stdout
- Include the command context: `error: user create: email "foo" is not valid`
- Suggest fixes when possible: `Did you mean 'deploy'?`
- Show the help hint on usage errors: `Run 'app user create --help' for usage`
- Never print stack traces unless `--verbose` is set
- Exit with the appropriate code (see above)

## Confirmation and Safety

- Destructive operations require interactive confirmation OR `--force`
- Show what will happen before asking: `This will delete 3 files. Continue? [y/N]`
- Default to the safe option: `[y/N]` means default is No
- Provide `--dry-run` for destructive commands — show what would happen without doing it
- Never silently overwrite files — prompt or require `--force`

## Configuration

- Location: `$XDG_CONFIG_HOME/<program>/config.yaml` (fallback `~/.config/<program>/config.yaml`)
- Also check `.<program>.yaml` in the current directory for project-local config
- Precedence: flags > environment variables > project config > user config > defaults
- Environment variables: `<PROGRAM>_<FLAG>` in SCREAMING_SNAKE_CASE
- Provide `config show` or `config path` command to display current config and source

## Interactive Features

- Detect TTY: if stdin is not a terminal, skip prompts and use defaults or fail
- Progress indicators for operations taking >2 seconds
- Spinners for indeterminate waits, progress bars for measurable progress
- Support piping: `app list | grep foo` must work (no ANSI codes, no prompts)
- Tab completion: provide shell completions for bash, zsh, and fish

## Help Text

- Every command has a one-line description shown in parent help
- `--help` shows: description, usage pattern, available subcommands, flags with defaults, examples
- Include at least one example per command
- Group related flags under headers in help output

## Versioning

- Follow SemVer: `MAJOR.MINOR.PATCH`
- `--version` prints: `<program> version <semver>` (optionally with commit hash and build date)
- Breaking changes in CLI interface = major version bump (changed flag names, removed commands, changed output format)
