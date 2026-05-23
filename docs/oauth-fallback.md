# OAuth Fallback Authentication

Plego implements a **three-tier OAuth fallback system** that provides flexible authentication while keeping the UX simple.

## UX Flow

```
START APP
    ↓
OAUTH (tries in order: token file → env var → interactive)
    ↓
DO APP (pipeline execution)
```

## Fallback Sequence

For each plugin that requires OAuth (e.g., Gmail), the system tries authentication in this order:

### 1. Token File (Default)
- Location: `~/.plego/gmail-token.json` (or custom `token` in config)
- Used when available from a previous run
- Automatically refreshed if expired

### 2. Environment Variable (Fallback)
- Primary: `{PLUGIN_NAME}_TOKEN`
  - For Gmail: `GMAIL_TOKEN`
  - For Slack: `SLACK_TOKEN`
  - For GitHub: `GITHUB_TOKEN`
- Accepts either:
  - **Full OAuth token** (JSON): `{"access_token": "...", "token_type": "Bearer", ...}`
  - **Raw access token** (string): `ya29.a0AfH6SMB...`

Set via:
```bash
export GMAIL_TOKEN='ya29.a0AfH6SMB...'
plego run -c config.yaml
```

Or in `.env`:
```
GMAIL_TOKEN=ya29.a0AfH6SMB...
```

### 3. Interactive OAuth Flow (Final)
- Opens browser for user authorization
- User pastes code back into terminal
- Token automatically saved to file for next run
- **No** interaction required on subsequent runs

## Configuration

Minimal config (uses all defaults):
```yaml
pipeline:
  outputs:
    - module: gmail
      to: you@example.com
      credentials: credentials.json
```

Custom token file location:
```yaml
pipeline:
  outputs:
    - module: gmail
      to: you@example.com
      credentials: credentials.json
      token: /custom/path/token.json
```

## Usage Examples

### First Run (Interactive OAuth)
```bash
$ plego run -c config.yaml
>> initializing authentication...
[gmail] no token found, starting OAuth flow...
[gmail] opening browser for authentication...
If browser doesn't open, visit:
  https://accounts.google.com/o/oauth2/auth?...

Paste authorization code: ya29.a0AfH6SMB...
[gmail] ✓ token saved
>> starting pipeline...
[filesystem] 5 items found
[gmail] published: Email 1
...
```

### Second Run (Token File)
```bash
$ plego run -c config.yaml
>> initializing authentication...
[gmail] using existing token from ~/.plego/gmail-token.json
>> starting pipeline...
[filesystem] 5 items found
[gmail] published: Email 1
...
```

### With Environment Variable (CI/CD)
```bash
export GMAIL_TOKEN='ya29.a0AfH6SMB...'
plego run -c config.yaml
# No browser interaction needed, uses env var token
```

## For CI/CD & Automation

1. **Generate a token** locally first run (interactive)
2. **Extract the token** from `~/.plego/gmail-token.json`
3. **Set as secret** in your CI/CD (GitHub Actions, GitLab CI, etc.)
4. **Run without interaction**:
   ```bash
   export GMAIL_TOKEN='<your-token>'
   plego run -c config.yaml
   ```

## Supported Plugins

Any plugin implementing the `Authorizer` interface gets automatic fallback:

- ✅ `gmail` - Gmail composer
- (More plugins can implement `Authorizer` in `core/auth.go`)

## Troubleshooting

**"no token found, starting OAuth flow" but browser doesn't open:**
- Copy the URL printed to terminal
- Manually paste into browser
- Paste code back

**"invalid token in GMAIL_TOKEN":**
- Verify env var is set: `echo $GMAIL_TOKEN`
- Check it's either valid JSON or a raw token string
- Not a file path—must be the actual token content

**"could not save token":**
- Check directory permissions: `~/.plego/` must be writable
- Or use `token:` config to specify custom location with writable path

## Security Notes

- Token files are saved with `0700` permissions (owner-only readable)
- Don't commit `gmail-token.json` to git
- In CI/CD, use secret management (GitHub Secrets, etc.)
- Env vars are slightly less secure than file permissions (visible in process listings)
- For production, consider service accounts instead of user tokens
