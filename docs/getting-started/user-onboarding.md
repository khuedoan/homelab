# User onboarding

=== "For user"

    ## Create user

    Ask an admin to create your account, provide the following information:

    - [ ] Full name (John Doe)
    - [ ] Select a username (`johndoe`)
    - [ ] Email address (`johndoe@example.com`)

    ## Install companion apps

    For all users:

    - [ ] [Password manager](#recommended-password-managers)
    - [ ] [Matrix chat client](https://matrix.org/clients) (optional, you can use the web version)

    For technical users:

    - [ ] [Docker](https://docs.docker.com/engine/install)
    - [ ] [Lens](https://k8slens.dev) (optional, you can use the included `kubectl` or `k9s` command in the tools container)

=== "For admin"

    Run the following script:

    ```sh
    ./script/onboard-user exampleuser "Example User" "user@example.com"
    ```

    Let the user scan the QR code or follow the link to set up passkeys or password + TOTP.

## Appendix

### Recommended password managers

- [Bitwarden](https://bitwarden.com/download) (easy to use, but requires an online account)
- [KeePassXC](https://keepassxc.org) (completely offline, but you'll need to sync manually)
