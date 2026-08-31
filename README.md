# Database Blueprint

Database Blueprint exports database structure and data into readable `.sql`
files that can be committed to Git and recreated later.

Currently, SQL Server is supported. PostgreSQL, MySQL, and SQLite are planned
and are already defined in the database type configuration.

## Install and Update (Linux and macOS)

```bash
curl -fsSL \
  https://raw.githubusercontent.com/Hennie5229x/DatabaseBlueprint/refs/heads/main/install.sh |
  sh
```

Verify:

```bash
blue
```

The installer supports Linux `amd64` and macOS `arm64` (Apple Silicon). It
installs `blue` to `~/.local/bin`. If that directory is not already in your
`PATH`, add it to `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

On macOS, Gatekeeper may block the unsigned binary. If that happens, run the
`xattr` command printed by the installer.

## Windows and Intel macOS

Windows users can install the CLI from PowerShell. This installs the current
user's copy without requiring Administrator access:

```powershell
irm https://raw.githubusercontent.com/Hennie5229x/DatabaseBlueprint/main/install.ps1 | iex
```

Restart PowerShell or Command Prompt after installation, then verify:

```powershell
blue
```

The Windows installer supports `amd64` PCs. Intel macOS users must download or
build the executable manually.

Download or build the executable, then run it directly:

Windows:

```powershell
.\blue.exe
```

macOS/Linux:

```bash
./blue
```

## Getting Started

Create a connection:

```bash
blue add
```

View saved connections:

```bash
blue list
```

Test a connection:

```bash
blue test <connection-name>
```

Export the database:

```bash
blue script <connection-name>
```

This creates a folder named after the connection containing the database
scripts, data, and `RunOrder.json`. The folder can be committed to Git.

Recreate the database from the exported scripts:

```bash
blue create <connection-name>
```

The `create` command asks for confirmation before creating or overwriting the
target database.

## Useful Commands

Edit a connection:

```bash
blue edit <connection-name>
```

Delete a connection:

```bash
blue delete <connection-name>
```

Script only the database schema:

```bash
blue script <connection-name> --schema-only
```

Script only the database data:

```bash
blue script <connection-name> --data-only
```

## Uninstall (Linux and macOS)

```bash
curl -fsSL \
  https://raw.githubusercontent.com/Hennie5229x/DatabaseBlueprint/refs/heads/main/uninstall.sh |
  sh
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/Hennie5229x/DatabaseBlueprint/main/uninstall.ps1 | iex
```
