# Setting Up ledger-cli Credentials

## Prerequisites
- A Firebase project created (named `kaiser-ledger`)
- Firestore Database enabled in that project

## Step 1: Create Service Account

1. Open the Firebase Console:
   ```
   open https://console.firebase.google.com/project/kaiser-ledger/settings/serviceaccounts/adminsdk
   ```

2. Click **"Generate new private key"**

3. Save the JSON file to a safe location:
   ```
   ~/Downloads/kaiser-ledger-key.json
   ```

## Step 2: Set Environment Variable Permanently

Add to your shell profile (`~/.zshrc` is the default on macOS):

```bash
# Add this line to ~/.zshrc
echo 'export GOOGLE_APPLICATION_CREDENTIALS="$HOME/Downloads/kaiser-ledger-key.json"' >> ~/.zshrc

# Reload the shell config
source ~/.zshrc
```

## Step 3: Verify Setup

```bash
# Check the variable is set
echo $GOOGLE_APPLICATION_CREDENTIALS

# Test connection to Firestore
cd /Users/corinakaiser/Projects/ledger-cli
./ledger doctor
```

## Expected Output

If credentials are working:
```
Ledger Doctor Report
====================
Auth Status:     ok
Firestore:       true
Projects:        0
Stale Projects:  0

Everything looks healthy.
```

If credentials are missing:
- Exit code: 10
- Error: "failed to create Firestore client: google: could not find default credentials"

## Alternative: Per-Project Credentials

Instead of shell globals, create `.envrc` in your project root:

```bash
# Install direnv first (brew install direnv)
echo 'export GOOGLE_APPLICATION_CREDENTIALS=./service-account.json' > .envrc
direnv allow
```

This keeps credentials scoped to the project directory.