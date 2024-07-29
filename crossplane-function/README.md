# Crossplane function  builder

## Features
- Build a Crossplane function image
- Authenticate with the registry
- Push the image to a registry

## Usage
```bash
dagger -m <this-module-url> call with-registry-auth --address <registry-url> --username <registry-username> --secret <reg-pass-environment-name> build-crossplane-function --directory . --platform linux/amd64 --sock <docker-socket file>
```

